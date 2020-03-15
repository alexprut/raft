package raft

import (
	"fmt"
	"math/rand"
	"time"
)

var candidateId int = 1 // FIXME this should be dynamic
var electionTimeout *time.Timer = nil
var leaderState string
var servers []int

type LogEntry struct {
	value int
	term  int
}

// Variables defined by the official Raft protocol
// Persistent state on all servers
var currentTerm int = 0 // latest term server has seen (initialized to 0 on first boot, increases monotonically)
var votedFor int = 0    // candidateId that received vote in current term (or null if none)
var log []LogEntry      // log entries; each entry contains command for state machine, and term when entry was received by leader (first index is 1)

// Volatile state on all servers
var commitIndex int = 0 // index of highest log entry known to be committed (initialized to 0, increases monotonically)
var lastApplied int = 0 // index of highest log entry applied to state machine (initialized to 0, increases monotonically)

// Volatile state on leaders
var nextIndex []int  // for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
var matchIndex []int // for each server, index of highest log entry known to be replicated on server (initialized to 0, increases monotonically)

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
	if term < currentTerm {
		return currentTerm, false
	}
	if leaderState == "leader" {
		leaderState = "follower"
		startNewElectionTimeout()
	}
	if votedFor == 0 || candidateId == 0 {
		return currentTerm, true
	}
	votedFor = candidateId
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
	if term < currentTerm {
		return currentTerm, false
	}
	if leaderState == "leader" {
		leaderState = "follower"
		startNewElectionTimeout()
	}
	return 0, true
}

// Returns a random duration between 150ms and 300ms
func getRandomDuration() time.Duration {
	return time.Duration(rand.Intn(150) + 150)
}

func startNewElectionTimeout() {
	if electionTimeout != nil {
		electionTimeout.Stop()
	}
	electionTimeout = time.AfterFunc(getRandomDuration(), func() {
		currentTerm++
		leaderState = "candidate"
		// TODO requestVote in parallel to each of the other servers
		votes := 1 // vote for himself
		votedFor = candidateId
		for range servers {
			_, voteGranted := requestVote(currentTerm, candidateId, len(log)-1, log[len(log)-1].term)
			if voteGranted {
				votes++
			}
		}
		if len(servers)/2 <= votes {
			leaderState = "leader"
			// Send heartbeat to all servers to establish leadership
			for range servers {
				appendEntries(currentTerm, candidateId, len(log)-1, log[len(log)-1].term, make([]int, 0), commitIndex)
			}
		}
	})
}

func init() {
	leaderState = "follower"
	fmt.Printf("Server started")
	fmt.Printf("State: " + leaderState)
	startNewElectionTimeout()
}
