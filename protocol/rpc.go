package protocol

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
	reply.Success = voteGranted
	reply.Term = term
	return nil
}

func (t *RpcServer) AppendEntries(args RpcArgsAppendEntries, reply *RpcServerReply) error {
	term, success := appendEntries(args.Term, args.LeaderId, args.PrevLogIndex, args.PrevLogTerm, args.Entries, args.LeaderCommit)
	reply.Term = term
	reply.Success = success
	return nil
}
