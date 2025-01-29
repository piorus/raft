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

	fmt.Printf("starting server :%s\n", port)
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

	repo, err := internal.NewBoltRepository(fmt.Sprintf("log-%s.db", cfg.Port))
	if err != nil {
		log.Fatalln("failed to create BoltRepository. Error was:", err)
	}
	defer repo.Cleanup()

	var logs []*internal.Log
	metadata, err := repo.GetMetadata()
	if err != nil {
		log.Fatalln("failed to get metadata. err: ", err)
	}

	raft := internal.NewRaft(logs, metadata, cfg, repo)
	go raft.StartElectionTimer()
	go raft.WaitForElection()

	if err = rpc.Register(raft); err != nil {
		log.Fatalln(err)
	}

	if err = startServer(cfg.Port); err != nil {
		log.Fatalln("failed to start rpc server, error was: ", err)
	}
}
