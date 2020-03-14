package raft

import (
	"fmt"
)

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

var leaderState string

type LogEntry struct {
	value int
	term  int
}

/**
Invoked by candidates to gather votes

Return:
	term: candidate's term
	voteGranted: true means candidate received vote
*/
func requestVote(term int, candidateId int, lastLogIndex int, lastLogTerm int) (int, bool) {
	if term < currentTerm {
		return currentTerm, false
	}
	if votedFor == 0 || candidateId == 0 {
		return currentTerm, true
	}
	return currentTerm, true
}

/**
Invoked by leader to replicate log entries, also used as heartbeat

Return:
	term: currentTerm, for leader to update itself
	success: true if follower contained entry matching prevLogIndex and prevLogTerm
*/
func appendEntries(term int, leaderId int, prevLogIndex int, prevLogTerm int, entries []int, leaderCommit int) (int, bool) {
	return 0, true
}

func init() {
	leaderState = "follower"
	fmt.Printf("Server started")
	fmt.Printf("State: " + leaderState)
}
