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
	Servers         []string
	electionTimeout time.Duration
}

type Raft struct {
	logs                 []*Log
	role                 Role
	metadata             *Metadata
	cfg                  *Configuration
	repo                 Repository
	resetElectionTimerCh chan bool
	startElectionCh      chan bool
	heartbeatCh          chan bool
}

type Log struct {
}

func NewRaft(logs []*Log, metadata *Metadata, cfg *Configuration, repo Repository) *Raft {
	return &Raft{
		logs:                 logs,
		role:                 Follower,
		metadata:             metadata,
		cfg:                  cfg,
		repo:                 repo,
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
	if err := r.repo.SaveMetadata(r.metadata); err != nil {
		panic(err)
	}
}

func (r *Raft) Role() Role {
	return r.role
}

func (r *Raft) SetRole(role Role) {
	r.role = role
}

func (r *Raft) VotedFor() string {
	return r.metadata.VotedFor
}

func (r *Raft) SetVotedFor(votedVor string) {
	r.metadata.VotedFor = votedVor
	if err := r.repo.SaveMetadata(r.metadata); err != nil {
		panic(err)
	}
}

func (r *Raft) ServerId() string {
	return r.metadata.ServerId
}

func (r *Raft) StartElectionTimer() {
	timer := time.NewTimer(withRandomOffset(r.cfg.electionTimeout))
	for {
		select {
		case <-timer.C:
			fmt.Printf("election timeout\n")
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
	r.SetVotedFor(r.ServerId())

	args := RequestVoteArgs{Term: r.metadata.CurrentTerm}
	repliesCh := make(chan *RequestVoteReply, len(r.cfg.Servers))

	for _, serverIp := range r.cfg.Servers {
		go DoRequestVote(serverIp, &args, repliesCh)
	}

	numClients := len(r.cfg.Servers)
	quorum := int(math.Ceil(float64(numClients) / 2))
	numVotes := 1 // voted for itself

	if numClients > 0 {
		timer := time.NewTimer(100 * time.Millisecond)
		numRemainingRetries := 3
		retryLimitReached := false

		for !retryLimitReached {
			select {
			case reply := <-repliesCh:
				if reply.VoteGranted {
					numVotes += 1
				}
				timer.Reset(100 * time.Millisecond)
			case <-timer.C:
				if numRemainingRetries == 0 {
					retryLimitReached = true
				}
				numRemainingRetries -= 1
			}
		}
		close(repliesCh)
	}

	fmt.Printf("numVotes = %d, quorum = %d\n", numVotes, quorum)
	if numVotes >= quorum {
		r.SetRole(Leader)
		fmt.Printf("[%s] elected as a leader\n", r.cfg.Port)
	}

	if r.IsLeader() {
		go r.StartHeartbeat()
	} else {
		go r.StartElectionTimer()
	}
}

func (r *Raft) StartHeartbeat() {
	fmt.Printf("starting heartbeat\n")
	for {
		fmt.Printf("heartbeat\n")
		for _, serverIp := range r.cfg.Servers {
			go r.DoHeartbeat(serverIp)
		}
		time.Sleep(r.cfg.electionTimeout / 2)
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
