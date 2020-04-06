package protocol

import "testing"

func TestRequestVoteLowerTerm(t *testing.T) {
	currentTerm = 3
	_, isVoted := requestVote(2, "localhost:8000", 2, 2)
	if isVoted {
		t.Errorf("requestVote() is %t", isVoted)
	}
}

func TestRequestVoteSameTermNotVotedYet(t *testing.T) {
	currentTerm = 3
	votedFor = vote{}
	_, isVoted := requestVote(3, "localhost:8000", 2, 2)
	if !isVoted {
		t.Errorf("requestVote() is %t", isVoted)
	}
}

func TestRequestVoteSameTermAlreadyVoted(t *testing.T) {
	currentTerm = 3
	votedFor = vote{"localhost:8000", 3}
	_, isVoted := requestVote(3, "localhost:8000", 2, 2)
	if isVoted {
		t.Errorf("requestVote() is %t", isVoted)
	}
}

func TestAppendEntriesLowerTerm(t *testing.T) {
	currentTerm = 3
	_, isVoted := appendEntries(2, "localhost:8000", 2, 2, make([]LogEntry, 0), 2)
	if isVoted {
		t.Errorf("appendEntries() is %t", isVoted)
	}
}

// TODO AppendEntries heartbeat
// TODO nextIndex[] and matchIndex[] is reset after leader election