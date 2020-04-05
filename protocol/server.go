package protocol

import (
	"log"
	"math"
	"math/rand"
	"net"
	"net/rpc"
	"time"
)

type Server struct {
	url        string
	connection *rpc.Client
}

var id string // allocated on start and immutable after, as a convention the id of the server is the "host:port"
var electionTimeout *time.Timer = nil
var heartBeatTimeout *time.Timer = nil
var leaderState string
var servers map[string]Server = make(map[string]Server)

type LogEntry struct {
	Value int // contains command for state machine
	Term  int // term when entry was received by leader
}

// Variables defined by the official Raft protocol
// Persistent state on all servers
var currentTerm int = 0 // latest term server has seen (initialized to 0 on first boot, increases monotonically)
var votedFor string     // candidateId that received vote in current term (or null if none)
var logs []LogEntry     // log entries; each entry contains command for state machine, and term when entry was received by leader (first index is 1)

// Volatile state on all servers
var commitIndex int = -1 // index of highest log entry known to be committed (initialized to -1, increases monotonically)
var lastApplied int = -1 // index of highest log entry applied to state machine (initialized to -1, increases monotonically)

// Volatile state on leaders
var nextIndex map[string]int = make(map[string]int) // for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
var matchIndex []int                                // for each server, index of highest log entry known to be replicated on server (initialized to -1, increases monotonically)

var leader string

type RpcServer int

type RpcServerReply struct {
	Term    int
	Success bool
}

type RpcArgsRequestVote struct {
	Term         int
	CandidateId  string
	LastLogIndex int
	LastLogTerm  int
}

type RpcArgsAppendEntries struct {
	Term         int
	LeaderId     string
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

func (t *RpcServer) RequestVote(args RpcArgsRequestVote, reply *RpcServerReply) error {
	term, voteGranted := requestVote(args.Term, args.CandidateId, args.LastLogIndex, args.LastLogTerm)
	reply = &RpcServerReply{Term: term, Success: voteGranted}
	return nil
}

func (t *RpcServer) AppendEntries(args RpcArgsAppendEntries, reply *RpcServerReply) error {
	term, success := appendEntries(args.Term, args.LeaderId, args.PrevLogIndex, args.PrevLogTerm, args.Entries, args.LeaderCommit)
	reply = &RpcServerReply{term, success}
	return nil
}

// Invoked by candidates to gather votes
//
// Params:
//	term int: candidate’s term
//	candidateId string: candidate requesting vote
//	lastLogIndex int: index of candidate’s last log entry
//	lastLogTerm int: term of candidate’s last log entry
//
// Return:
//	term: currentTerm, for candidate to update itself
//	voteGranted: true means candidate received vote
func requestVote(term int, candidateId string, lastLogIndex int, lastLogTerm int) (int, bool) {
	startNewElectionTimeout()
	if term < currentTerm {
		return currentTerm, false
	}
	currentTerm = term // if one server's current term is smaller than the other's, then it updates to the larger value
	if leaderState != "follower" {
		toFollower()
	}

	// If votedFor is null or candidateId, and candidate’s log is at least as up-to-date as receiver’s log, grant vote
	if votedFor == "" ||
		(len(logs) > 0 && logs[len(logs)-1].Term < lastLogTerm) ||
		(len(logs) > 0 && logs[len(logs)-1].Term == lastLogTerm && len(logs)-1 <= lastLogIndex) {
		votedFor = candidateId
		log.Println("Voted for:", candidateId)
		return currentTerm, true
	}
	return currentTerm, false
}

// Invoked by leader to replicate log entries, also used as heartbeat
//
// Params:
//	term int: leader’s term
//	leaderId string: so follower can redirect clients
//	prevLogIndex int: index of log entry immediately preceding new ones
//	prevLogTerm int: term of prevLogIndex entry
//	entries []int: log entries to store (empty for heartbeat; may send more than one for efficiency)
//	leaderCommit int: leader’s commitIndex
//
// Return:
//	term: currentTerm, for leader to update itself
//	success: true if follower contained entry matching prevLogIndex and prevLogTerm
func appendEntries(term int, leaderId string, prevLogIndex int, prevLogTerm int, entries []LogEntry, leaderCommit int) (int, bool) {
	log.Println("Received append entries:", entries)
	startNewElectionTimeout()
	if term < currentTerm {
		return currentTerm, false
	}
	currentTerm = term // if one server's current term is smaller than the other's, then it updates to the larger value
	leader = leaderId
	if leaderState != "follower" {
		toFollower()
	}
	if len(logs) > 0 {
		log.Println("Elements present")
	}

	// If an existing entry conflicts with a new one (same index but different terms), delete the existing entry and all that follow it
	if len(logs) > 0 && (len(logs)-1 < prevLogIndex && logs[prevLogIndex].Term != prevLogTerm) {
		logs = logs[:prevLogIndex]
	}

	// Reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm
	if len(logs) > 0 && (len(logs)-1 < prevLogIndex || logs[prevLogIndex].Term != prevLogTerm) {
		return currentTerm, false
	}

	// Append any new entries not already in the log
	for i, e := range entries {
		if len(logs)-1 <= i+prevLogIndex {
			if logs[i+prevLogIndex].Term != e.Term {
				logs = logs[:i+prevLogIndex-1]

			}
		}
		logs = append(logs, e)
	}

	if leaderCommit > commitIndex {
		commitIndex = int(math.Min(float64(leaderCommit), float64(len(logs)-1)))
	}
	return currentTerm, true
}

// Returns a random duration between 150ms and 300ms
func getRandomDuration() time.Duration {
	return time.Millisecond * time.Duration(rand.Intn(5150)+150)
}

func stopElectionTimeout() {
	if electionTimeout != nil {
		electionTimeout.Stop()
	}
}

func startNewElectionTimeout() {
	stopElectionTimeout()
	electionTimeout = time.AfterFunc(getRandomDuration(), func() {
		startNewElectionTimeout() // reset election timer
		// on conversion to candidate start election
		toCandidate()
		currentTerm++
		log.Println("Term:", currentTerm)
		// TODO requestVote in parallel to each of the other servers, use coroutines
		votes := 1 // vote for himself
		votedFor = id
		// Send RequestVote RPCs to all other servers
		for _, s := range servers {
			if s.connection == nil {
				Connect(s.url)
			}
			if s.connection != nil {
				reply := RpcServerReply{}
				var args RpcArgsRequestVote
				if len(logs) == 0 {
					args = RpcArgsRequestVote{currentTerm, id, -1, -1}
				} else {
					args = RpcArgsRequestVote{currentTerm, id, len(logs) - 1, logs[len(logs)-1].Term}
				}
				err := s.connection.Call("RpcServer.RequestVote", args, &reply)
				if err != nil {
					log.Println("Error:", err)
				}
				if reply.Success {
					log.Printf("Received vote from: %s", s.url)
					votes++
				}
			}
		}
		log.Printf("Received votes: %d / %d", votes, len(servers)+1)

		// If votes received from majority of servers: become leader
		if (len(servers)+1)/2+1 <= votes {
			toLeader()
		}
	})
}

func startNewHeartBeatTimeout() {
	stopHeartBeatTimeout()
	heartBeatTimeout = time.AfterFunc(getRandomDuration()-time.Millisecond*500, func() {
		startNewHeartBeatTimeout()
		sendServerHeartbeat()
	})
}

func stopHeartBeatTimeout() {
	if heartBeatTimeout != nil {
		heartBeatTimeout.Stop()
	}
}

func sendServerHeartbeat() {
	log.Println("Sending heartbeat ...")
	for _, s := range servers {
		reply := RpcServerReply{}
		var args RpcArgsAppendEntries

		if len(logs) == 0 {
			args = RpcArgsAppendEntries{currentTerm, id, -1, -1, make([]LogEntry, 0), commitIndex}
		} else {
			args = RpcArgsAppendEntries{currentTerm, id, len(logs) - 1, logs[len(logs)-1].Term, make([]LogEntry, 0), commitIndex}
		}
		err := s.connection.Call("RpcServer.AppendEntries", args, &reply)
		if err != nil {
			log.Println("Error:", err)
		}
	}
}

func handleClientRequest(value int) (bool, string) {
	if leaderState == "leader" {
		log.Printf("Received: %d", value)
		logs = append(logs, LogEntry{value, currentTerm})
		return true, id
	}
	return false, leader
}

func Connect(url string) bool {
	log.Println("Connecting to server: " + url)
	connection, err := rpc.Dial("tcp", url)
	servers[url] = Server{url: url, connection: connection}
	if err != nil {
		log.Println("Unable Connecting to server: ", err)
		return false
	}
	idCurrentServer = url
	log.Println("Connected to server: " + url)
	return true
}

func toFollower() {
	stopHeartBeatTimeout()
	leaderState = "follower"
	log.Println("State:", leaderState)
}

func toCandidate() {
	stopHeartBeatTimeout()
	leaderState = "candidate"
	log.Println("State:", leaderState)
}

func toLeader() {
	// Stop election timeout
	electionTimeout.Stop()
	electionTimeout = nil
	leaderState = "leader"
	log.Println("State:", leaderState)
	// Volatile state on leader reinitialized after election
	for key, _ := range nextIndex {
		nextIndex[key] = len(logs)
	}
	for key, _ := range matchIndex {
		matchIndex[key] = 0
	}

	// Send heartbeat to all servers to establish leadership
	sendServerHeartbeat()
	startNewHeartBeatTimeout()
}

func Start(url string, urls []string) {
	id = url
	log.Println("Server started on: " + url)
	toFollower()
	log.Println("Term:", currentTerm)
	startNewElectionTimeout()
	for _, u := range urls {
		Connect(u)
	}
	addy, err := net.ResolveTCPAddr("tcp", url)
	if err != nil {
		log.Fatal(err)
	}

	inbound, err := net.ListenTCP("tcp", addy)
	if err != nil {
		log.Fatal(err)
	}

	rpc.Register(new(RpcClient))
	rpc.Register(new(RpcServer))
	rpc.Accept(inbound)
}
