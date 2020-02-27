package rpc

func requestVote(term, candidateId, lastLogIndex, lastLogTerm) (term, voteGranted) {
}

func appendEntries(term, leaderId, prevLogIndex, prevLogTerm, entries[], leaderCommit) (term, success) {
}

func installSnapshot(term, leaderId, lastIncludedIndex, lastIncludedTerm, offset, data[], done) (term) {
}
