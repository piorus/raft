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
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
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

	dbConn, err := internal.NewBoltDatabaseConn("log.db")
	if err != nil {
		log.Fatalln(err)
	}

	serverId, err := dbConn.GetServerId()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("serverId: %s\n", serverId)

	raft := internal.NewRaft(serverId, dbConn)
	go raft.StartElectionTimer(1000 * time.Millisecond)
	go raft.WaitForElection()

	err = rpc.Register(raft)
	if err != nil {
		log.Fatalln(err)
	}
	rpc.HandleHTTP()

	errChan := make(chan error)
	go startServer(cfg.Port, errChan)
	time.Sleep(time.Millisecond * 1500)

	var clients []*rpc.Client

	for _, ip := range cfg.ServerIps {
		client, err := rpc.DialHTTP("tcp", ip)
		fmt.Printf("successfully dialed client: %s\n", ip)

		if err != nil {
			log.Fatalln("error dialing ", ip, " error was: ", err)
		}
		clients = append(clients, client)
	}

	//args := internal.RequestVoteArgs{Term: raft.CurrentTerm}
	//var reply internal.RequestVoteReply
	//
	//err = clients[0].Call("Raft.RequestVote", &args, &reply)
	//if err != nil {
	//	log.Fatalln(err)
	//}

	err = <-errChan
	if err != nil {
		log.Fatalln("failed to start rpc server, error was: ", err)
	}
}
