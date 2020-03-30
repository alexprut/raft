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
	value int // contains command for state machine
	term  int // term when entry was received by leader
}

// Variables defined by the official Raft protocol
// Persistent state on all servers
var currentTerm int = 0 // latest term server has seen (initialized to 0 on first boot, increases monotonically)
var votedFor string     // candidateId that received vote in current term (or null if none)
var logs []LogEntry     // log entries; each entry contains command for state machine, and term when entry was received by leader (first index is 1)

// Volatile state on all servers
var commitIndex int = 0 // index of highest log entry known to be committed (initialized to 0, increases monotonically)
var lastApplied int = 0 // index of highest log entry applied to state machine (initialized to 0, increases monotonically)

// Volatile state on leaders
var nextIndex map[string]int = make(map[string]int) // for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
var matchIndex []int                                // for each server, index of highest log entry known to be replicated on server (initialized to 0, increases monotonically)

var leader string

type RpcServer int

type RpcServerReply struct {
	term    int
	success bool
}

type RpcArgsRequestVote struct {
	term         int
	candidateId  string
	lastLogIndex int
	lastLogTerm  int
}

type RpcArgsAppendEntries struct {
	term         int
	leaderId     string
	prevLogIndex int
	prevLogTerm  int
	entries      []LogEntry
	leaderCommit int
}

func (t *RpcServer) RequestVote(args RpcArgsRequestVote, reply *RpcServerReply) error {
	term, voteGranted := requestVote(args.term, args.candidateId, args.lastLogIndex, args.lastLogTerm)
	reply = &RpcServerReply{term: term, success: voteGranted}
	return nil
}

func (t *RpcServer) AppendEntries(args RpcArgsAppendEntries, reply *RpcServerReply) error {
	term, success := appendEntries(args.term, args.leaderId, args.prevLogIndex, args.prevLogTerm, args.entries, args.leaderCommit)
	reply = &RpcServerReply{term: term, success: success}
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
		leaderState = "follower"
		return currentTerm, false
	}

	// If votedFor is null or candidateId, and candidate’s log is at least as up-to-date as receiver’s log, grant vote
	if votedFor == "" ||
		(len(logs) > 0 && logs[len(logs)-1].term < lastLogTerm) ||
		(len(logs) > 0 && logs[len(logs)-1].term == lastLogTerm && len(logs)-1 <= lastLogIndex) {
		votedFor = candidateId
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
	startNewElectionTimeout()
	if term < currentTerm {
		return currentTerm, false
	}
	leader = leaderId
	if leaderState == "leader" {
		leaderState = "follower"
	}

	// If an existing entry conflicts with a new one (same index but different terms), delete the existing entry and all that follow it
	if len(logs)-1 < prevLogIndex && logs[prevLogIndex].term != prevLogTerm {
		logs = logs[:prevLogIndex]
	}

	// Reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm
	if len(logs)-1 < prevLogIndex || logs[prevLogIndex].term != prevLogTerm {
		return currentTerm, false
	}

	// Append any new entries not already in the log
	for i, e := range entries {
		if len(logs)-1 <= i+prevLogIndex {
			if logs[i+prevLogIndex].term != e.term {
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
	return time.Millisecond * time.Duration(rand.Intn(150)+150)
}

func startNewElectionTimeout() {
	if electionTimeout != nil {
		electionTimeout.Stop()
	}
	electionTimeout = time.AfterFunc(getRandomDuration(), func() {
		// on conversion to candidate start election
		leaderState = "candidate"
		currentTerm++
		startNewElectionTimeout() // reset election timer
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
				s.connection.Call(
					"RpcServer.RequestVote",
					RpcArgsRequestVote{currentTerm, id, len(logs) - 1, logs[len(logs)-1].term},
					reply)
				if reply.success {
					votes++
				}
			}
		}
		// If votes received from majority of servers: become leader
		if len(servers)/2 <= votes {
			electionTimeout = nil
			leaderState = "leader"
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
	})
}

func startNewHeartBeatTimeout() {
	if heartBeatTimeout != nil {
		heartBeatTimeout.Stop()
	}
	heartBeatTimeout = time.AfterFunc(getRandomDuration() - time.Millisecond * 50, func() {
		sendServerHeartbeat()
	})
}

func sendServerHeartbeat() {
	for _, s := range servers {
		reply := RpcServerReply{}
		s.connection.Call(
			"RpcServer.AppendEntries",
			RpcArgsAppendEntries{currentTerm, id, len(logs) - 1, logs[len(logs)-1].term, make([]LogEntry, 0), commitIndex},
			reply)
	}
}

func handleClientRequest(value int) bool {
	if leaderState == "leader" {
		log.Printf("Received: %d", value)
		logs = append(logs, LogEntry{value: value, term: currentTerm})
		return true
	}
	return false
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

func Start(url string, urls []string) {
	id = url
	leaderState = "follower"
	log.Println("Server started on: " + url)
	log.Println("State: " + leaderState)
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
