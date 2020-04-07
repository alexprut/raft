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

func TestAppendEntriesHeartbeat(t *testing.T) {
	currentTerm = 3
	if _, success := appendEntries(currentTerm, id, -1, -1, make([]LogEntry, 0), commitIndex); !success {
		t.Errorf("Heartbeat is: %v", success)
	}
	logs = append(logs, LogEntry{1,1})
	if _, success := appendEntries(currentTerm, id, len(logs) - 1, logs[len(logs)-1].Term, make([]LogEntry, 0), commitIndex); !success {
		t.Errorf("Heartbeat is: %v", success)
	}
	logs = make([]LogEntry, 0)
}

func TestFirstAppendEntry(t *testing.T) {
	currentTerm = 3
	entry := LogEntry{1,3}
	logs = append(logs, entry)
	if _, success := appendEntries(currentTerm, id, -1, -1, []LogEntry{entry}, commitIndex); !success {
		t.Errorf("Unable to add new entry")
	}
}

// TODO nextIndex[] and matchIndex[] is reset after leader election