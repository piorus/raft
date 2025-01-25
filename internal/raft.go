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

type Raft struct {
	ServerId    string
	CurrentTerm int64
	VotedFor    string
	logs        []*Log
	//CommitIndex          int64
	//LastApplied          int64
	//NextIndex            []int64
	//MatchIndex           []int64
	Role                 Role
	resetElectionTimerCh chan bool
	startElectionCh      chan bool
	clients              []*rpc.Client
	dbConn               DatabaseConn
}

type Log struct {
}

func NewRaft(id string, dbConn DatabaseConn) *Raft {
	resetElectionTimerCh := make(chan bool)
	startElectionCh := make(chan bool)

	return &Raft{
		ServerId:             id,
		Role:                 Follower,
		resetElectionTimerCh: resetElectionTimerCh,
		startElectionCh:      startElectionCh,
		dbConn:               dbConn,
	}
}

func withRandomOffset(duration time.Duration) time.Duration {
	return duration + (time.Duration(rand.Intn(150)) * time.Millisecond)
}

func (r *Raft) StartElectionTimer(electionTimeout time.Duration) {
	timer := time.NewTimer(withRandomOffset(electionTimeout))
	for {
		select {
		case <-timer.C:
			fmt.Printf("[%s] election timeout\n", r.ServerId)
			r.startElectionCh <- true
			return
		case <-r.resetElectionTimerCh:
			timer.Reset(withRandomOffset(electionTimeout))
		}
	}
}

func (r *Raft) WaitForElection() {
	for {
		<-r.startElectionCh
		r.StartElection()
	}
}

func (r *Raft) StartElection() {
	fmt.Printf("starting new election\n")
	r.CurrentTerm += 1
	r.Role = Candidate
	r.VotedFor = r.ServerId

	args := RequestVoteArgs{Term: r.CurrentTerm}
	numClients := len(r.clients)
	majorityRequired := int(math.Floor(float64(numClients) / 2))
	// already voted for itself
	numVotes := 1

	for _, client := range r.clients {
		reply := RequestVoteReply{}
		// TODO: parallel
		err := client.Call("Raft.RequestVote", &args, &reply)
		if err != nil {
			log.Printf("RequestVote error: %s\n", err.Error())
			continue
		}
		if reply.VoteGranted {
			numVotes += 1
		}
		if numVotes >= majorityRequired {
			r.Role = Leader
			break
		}
	}

	go r.StartElectionTimer(1000 * time.Millisecond)
}

func (r *Raft) IsCandidate() bool {
	return r.Role == Candidate
}

func (r *Raft) IsFollower() bool {
	return r.Role == Follower
}

func (r *Raft) IsLeader() bool {
	return r.Role == Leader
}

//func (r * Raft) TriggerRequestVoteRequest(args *RequestVoteArgs) {
//
//}
