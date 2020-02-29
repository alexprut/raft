package state

// Persistent state on all servers
var currentTerm int // latest term server has seen (initialized to 0 on first boot, increases monotonically)
var votedFor int    // candidateId that received vote in current term (or null if none)
var log []          // log entries; each entry contains command for state machine, and term when entry was received by leader (first index is 1)

// Volatile state on all servers
var commitIndex int // index of highest log entry known to be committed (initialized to 0, increases monotonically)
var lastApplied int // index of highest log entry applied to state machine (initialized to 0, increases monotonically)

// Volatile state on leaders
var nextIndex []  // for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
var matchIndex [] // for each server, index of highest log entry known to be replicated on server (initialized to 0, increases monotonically)
