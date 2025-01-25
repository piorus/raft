package internal

import "fmt"

type RequestVoteArgs struct {
	Term         int64
	CandidateId  string
	LastLogIndex int64
	LastLogTerm  int64
}

type RequestVoteReply struct {
	Term        int64
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int64
	LeaderId     string
	PrevLogIndex int64
	PrevLogTerm  int64
	Entries      []*Log
}

type AppendEntriesReply struct {
	Term    int64
	Success bool
}

type InstallSnapshotArgs struct {
	Term              int64
	LeaderId          string
	LastIncludedIndex int64
	LastIncludedTerm  int64
	Offset            int64
	Data              []byte
	Done              bool
}

type InstallSnapshotReply struct {
	term int64
}

func (r *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	fmt.Printf("[RPC] RequestVote Args: %+v Raft: %+v\n", args, r)

	if r.CurrentTerm > args.Term {
		reply.VoteGranted = false
		reply.Term = r.CurrentTerm
		return nil
	} else if r.CurrentTerm < args.Term && r.IsCandidate() {
		r.CurrentTerm = args.Term
		r.Role = Follower
	}

	if (r.VotedFor == "" || r.VotedFor == args.CandidateId) && args.LastLogIndex >= r. {

	}

	return nil
}

func (r *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	fmt.Printf("[RPC] AppendEntries Args: %+v Raft: %+v\n", args, r)

	if args.Term < r.CurrentTerm {
		reply.Success = false
		reply.Term = r.CurrentTerm
		return nil
	}

	r.resetElectionTimerCh <- true

	return nil
}

func (r *Raft) InstallSnapshot(args *InstallSnapshotArgs, reply *InstallSnapshotReply) error {
	fmt.Printf("[RPC] InstallSnapshot Args: %+v Raft: %+v\n", args, r)

	return nil
}
