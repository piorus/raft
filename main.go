package main

import (
	"fmt"
	"github.com/piorus/raft/internal"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

func startServer(port string, errChan chan<- error) {
	fmt.Printf("starting server :%s\n", port)
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		errChan <- err
		return
	}
	fmt.Printf("listening on :%s\n", port)
	errChan <- http.Serve(l, nil)
}

func main() {
	cfg, err := internal.ParseConfiguration()

	if err != nil {
		log.Fatalln(err)
	}

	dbConn, err := internal.NewBoltDatabaseConn(fmt.Sprintf("log-%s.db", cfg.Port))
	if err != nil {
		log.Fatalln("failed to create BoltDatabaseConn. Error was:", err)
	}

	serverId, err := dbConn.GetServerId()
	if err != nil {
		log.Fatalln(err)
	}

	var logs []*internal.Log
	metadata := internal.Metadata{}

	raft := internal.NewRaft(logs, &metadata, cfg, dbConn)
	go raft.StartElectionTimer()
	go raft.WaitForElection()

	err = rpc.Register(raft)
	if err != nil {
		log.Fatalln(err)
	}
	rpc.HandleHTTP()

	errChan := make(chan error)
	go startServer(cfg.Port, errChan)
	time.Sleep(time.Millisecond * 1500)

	err = <-errChan
	if err != nil {
		log.Fatalln("failed to start rpc server, error was: ", err)
	}
}
