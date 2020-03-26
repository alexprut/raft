package protocol

import (
	"log"
	"math"
	"math/rand"
	"net"
	"net/rpc"
	"time"
)

var id int = 1 // FIXME this should be allocated on start and immutable after, as a convention the id will be the host:port
var electionTimeout *time.Timer = nil
var leaderState string
var servers []int

type LogEntry struct {
	value int // contains command for state machine
	term  int // term when entry was received by leader
}

// Variables defined by the official Raft protocol
// Persistent state on all servers
var currentTerm int = 0 // latest term server has seen (initialized to 0 on first boot, increases monotonically)
var votedFor int = 0    // candidateId that received vote in current term (or null if none)
var logs []LogEntry     // log entries; each entry contains command for state machine, and term when entry was received by leader (first index is 1)

// Volatile state on all servers
var commitIndex int = 0 // index of highest log entry known to be committed (initialized to 0, increases monotonically)
var lastApplied int = 0 // index of highest log entry applied to state machine (initialized to 0, increases monotonically)

// Volatile state on leaders
var nextIndex []int  // for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
var matchIndex []int // for each server, index of highest log entry known to be replicated on server (initialized to 0, increases monotonically)

type RpcServer int

type RpcServerReply struct {
	term    int
	success bool
}

type RpcArgsRequestVote struct {
	term         int
	candidateId  int
	lastLogIndex int
	lastLogTerm  int
}

type RpcArgsAppendEntries struct {
	term         int
	leaderId     int
	prevLogIndex int
	prevLogTerm  int
	entries      []int
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
//	currentTerm int: candidate requesting vote
//	lastLogIndex int: index of candidate’s last log entry
//	lastLogTerm int: term of candidate’s last log entry
//
// Return:
//	term: currentTerm, for candidate to update itself
//	voteGranted: true means candidate received vote
func requestVote(term int, candidateId int, lastLogIndex int, lastLogTerm int) (int, bool) {
	startNewElectionTimeout()
	if term < currentTerm {
		leaderState = "follower"
		startNewElectionTimeout()
		return currentTerm, false
	}

	// If votedFor is null or candidateId, and candidate’s log is at least as up-to-date as receiver’s log, grant vote
	// TODO candidate’s log is at least as up-to-date as receiver’s log
	if votedFor == 0 || candidateId == 0 {
		votedFor = candidateId
		return currentTerm, true
	}
	return currentTerm, true
}

// Invoked by leader to replicate log entries, also used as heartbeat
//
// Params:
//	term int: leader’s term
//	leaderId int: so follower can redirect clients
//	prevLogIndex int: index of log entry immediately preceding new ones
//	prevLogTerm int: term of prevLogIndex entry
//	entries []int: log entries to store (empty for heartbeat; may send more than one for efficiency)
//	leaderCommit int: leader’s commitIndex
//
// Return:
//	term: currentTerm, for leader to update itself
//	success: true if follower contained entry matching prevLogIndex and prevLogTerm
func appendEntries(term int, leaderId int, prevLogIndex int, prevLogTerm int, entries []int, leaderCommit int) (int, bool) {
	startNewElectionTimeout()
	if term < currentTerm {
		return currentTerm, false
	}
	if leaderState == "leader" {
		leaderState = "follower"
		startNewElectionTimeout()
	}

	// TODO Reply false if log doesn’t contain an entry at prevLogIndex whose term matches prevLogTerm
	// TODO If an existing entry conflicts with a new one (same index  but different terms), delete the existing entry and all that follow it
	// TODO Append any new entries not already in the log

	if leaderCommit > commitIndex {
		commitIndex = int(math.Min(float64(leaderCommit), float64(len(logs)-1)))
	}
	return 0, true
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
		// TODO requestVote in parallel to each of the other servers
		votes := 1 // vote for himself
		votedFor = id
		for _, s := range connectionsServers {
			if s.connection == nil {
				Connect(s.url)
			}
			if s.connection != nil {
				reply := RpcServerReply{}
				s.connection.Call(
					"RpcServer.requestVote",
					RpcArgsRequestVote{currentTerm, id, len(logs) - 1, logs[len(logs)-1].term},
					reply)
				if reply.success {
					votes++
				}
			}
		}
		// if votes received from majority of servers: become leader
		if len(servers)/2 <= votes {
			leaderState = "leader"
			// Send heartbeat to all servers to establish leadership
			for range servers {
				appendEntries(currentTerm, id, len(logs)-1, logs[len(logs)-1].term, make([]int, 0), commitIndex)
			}
			// Volatile state on leader reinitialized after election
			for i, _ := range nextIndex {
				nextIndex[i] = len(logs)
			}
			for i, _ := range matchIndex {
				nextIndex[i] = 0
			}
		}
	})
}

type Server struct {
	url        string
	connection *rpc.Client
}

var connectionsServers map[string]Server = make(map[string]Server)

func handleClientRequest(value int) bool {
	log.Printf("Received: %d", value)
	logs = append(logs, LogEntry{value: value, term: currentTerm})
	return true
}

func Connect(url string) bool {
	log.Println("Connecting to server: " + url)
	connection, err := rpc.Dial("tcp", url)
	connectionsServers[url] = Server{url: url, connection: connection}
	if err != nil {
		log.Println("Unable Connecting to server: ", err)
		return false
	}
	idCurrentServer = url
	log.Println("Connected to server: " + url)
	return true
}

func Start(url string, urls []string) {
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
