package main

import (
	"fmt"
	"github.com/piorus/raft/internal"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func startServer(port string) error {
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	fmt.Printf("listening on :%s\n", port)
	return http.Serve(l, nil)
}

func main() {
	cfg, err := internal.ParseConfiguration()

	if err != nil {
		log.Fatalln(err)
	}

	repo, err := internal.NewBoltRepository("data/log.db")
	if err != nil {
		log.Fatalln("failed to initialize bolt:", err)
	}

	var logs []*internal.Log
	metadata, err := repo.GetMetadata()

	if err != nil {
		log.Fatalln("failed to load metadata:", err)
	}

	// needed for the first run to ensure that kvs are properly stored in bolt
	fmt.Printf("saving metadata\n")
	err = repo.SaveMetadata(metadata)
	if err != nil {
		log.Fatalln("failed to save metadata:", err)
	}

	raft := internal.NewRaft(logs, metadata, cfg, repo)
	go raft.StartElectionTimer()
	go raft.WaitForElection()

	if err = rpc.Register(raft); err != nil {
		log.Fatalln(err)
	}

	if err = startServer(cfg.Port); err != nil {
		log.Fatalln("server crashed:", err)
	}
}
