package internal

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/rpc"
	"time"
)

type Role string

const (
	Leader    Role = "leader"
	Follower  Role = "follower"
	Candidate Role = "candidate"
)

type Metadata struct {
	ServerId    string
	CurrentTerm int64
	VotedFor    string
	CommitIndex int64
	LastApplied int64
	NextIndex   []int64
	MatchIndex  []int64
}

type Configuration struct {
	Port            string
	ServerIps       ServerIps
	electionTimeout time.Duration
}

type Raft struct {
	logs                 []*Log
	role                 Role
	metadata             *Metadata
	cfg                  *Configuration
	db                   DatabaseConn
	resetElectionTimerCh chan bool
	startElectionCh      chan bool
	heartbeatCh          chan bool
}

type Log struct {
}

func NewRaft(logs []*Log, metadata *Metadata, cfg *Configuration, dbConn DatabaseConn) *Raft {
	return &Raft{
		logs:                 logs,
		role:                 Follower,
		metadata:             metadata,
		cfg:                  cfg,
		db:                   dbConn,
		resetElectionTimerCh: make(chan bool),
		startElectionCh:      make(chan bool),
		heartbeatCh:          make(chan bool),
	}
}

func withRandomOffset(duration time.Duration) time.Duration {
	return duration + (time.Duration(rand.Intn(150)) * time.Millisecond)
}

func (r *Raft) CurrentTerm() int64 {
	return r.metadata.CurrentTerm
}

func (r *Raft) SetCurrentTerm(currentTerm int64) {
	r.metadata.CurrentTerm = currentTerm
}

func (r *Raft) SetRole(role Role) {
	r.role = role
}

func (r *Raft) SetVotedFor(votedVor string) {
	r.metadata.VotedFor = votedVor
}

func (r *Raft) StartElectionTimer() {
	timer := time.NewTimer(withRandomOffset(r.cfg.electionTimeout))
	for {
		select {
		case <-timer.C:
			fmt.Printf("election timeout, starting voting\n")
			r.startElectionCh <- true
			return
		case <-r.resetElectionTimerCh:
			fmt.Printf("reset election timer\n")
			timer.Reset(withRandomOffset(r.cfg.electionTimeout))
		}
	}
}

func (r *Raft) WaitForElection() {
	for {
		<-r.startElectionCh
		r.StartElection()
	}
}

func DoRequestVote(serverIp string, args *RequestVoteArgs, replyCh chan<- *RequestVoteReply) {
	client, err := rpc.DialHTTP("tcp", serverIp)
	reply := RequestVoteReply{}

	if err != nil {
		log.Printf("error dialing %s. Error was: %s\n", serverIp, err)
		replyCh <- &reply
		return
	}
	fmt.Printf("successfully dialed %s\n", serverIp)
	err = client.Call("Raft.RequestVote", &args, &reply)
	if err != nil {
		log.Printf("RequestVote error: %s\n", err.Error())
	}
	replyCh <- &reply
}

func DoAppendEntries(serverIp string, args *AppendEntriesArgs, replyCh chan<- AppendEntriesReply) {
	client, err := rpc.DialHTTP("tcp", serverIp)
	reply := AppendEntriesReply{}

	if err != nil {
		log.Printf("error dialing %s. Error was: %s\n", serverIp, err)
		replyCh <- reply
		return
	}
	fmt.Printf("successfully dialed %s\n", serverIp)
	err = client.Call("Raft.AppendEntries", &args, &reply)
	if err != nil {
		log.Printf("AppendEntries error: %s\n", err.Error())
	}
	replyCh <- reply
}

func (r *Raft) DoHeartbeat(serverIp string) {
	DoAppendEntries(serverIp, &AppendEntriesArgs{Term: r.metadata.CurrentTerm, LeaderId: r.metadata.ServerId}, make(chan<- AppendEntriesReply))
}

func (r *Raft) StartElection() {
	fmt.Printf("starting new election\n")
	r.SetCurrentTerm(r.CurrentTerm() + 1)
	r.SetRole(Candidate)
	r.SetVotedFor(r.metadata.ServerId)

	args := RequestVoteArgs{Term: r.metadata.CurrentTerm}
	repliesCh := make(chan *RequestVoteReply, len(r.cfg.ServerIps))

	for _, ip := range r.cfg.ServerIps {
		go DoRequestVote(ip, &args, repliesCh)
	}

	numClients := len(r.cfg.ServerIps)
	quorum := int(math.Ceil(float64(numClients) / 2))
	numVotes := 1 // cuz voted for itself

	for reply := range repliesCh {
		if reply.VoteGranted {
			numVotes += 1
		}
		if numVotes >= quorum {
			r.SetRole(Leader)
			fmt.Printf("[%s] elected as a leader\n", r.cfg.Port)
		}
	}
	close(repliesCh)

	if !r.IsLeader() {
		go r.StartElectionTimer()
	} else {
		go r.StartHeartbeat()
	}
}

func (r *Raft) StartHeartbeat() {
	for {
		for _, serverIp := range r.cfg.ServerIps {
			go r.DoHeartbeat(serverIp)
		}
		time.Sleep(r.cfg.electionTimeout / 3)
	}
}

func (r *Raft) IsCandidate() bool {
	return r.role == Candidate
}

func (r *Raft) IsFollower() bool {
	return r.role == Follower
}

func (r *Raft) IsLeader() bool {
	return r.role == Leader
}
