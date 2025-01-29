package internal

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type ServerIps []string

func (s *ServerIps) Set(value string) error {
	*s = append(*s, strings.Split(value, ",")...)
	return nil
}

func (s *ServerIps) String() string {
	return strings.Join(*s, ",")
}

func ParseConfiguration() (*Configuration, error) {
	var port string
	flag.StringVar(&port, "port", "", "Port")
	flag.Parse()

	if port == "" {
		return nil, fmt.Errorf("--port is required")
	}

	electionTimeoutRaw := os.Getenv("ELECTION_TIMEOUT")

	if electionTimeoutRaw == "" {
		electionTimeoutRaw = "1s"
	}

	electionTimeout, err := time.ParseDuration(electionTimeoutRaw)

	if err != nil {
		return nil, err
	}

	serversRaw := os.Getenv("SERVERS")
	var servers []string

	if serversRaw != "" {
		servers = strings.Split(serversRaw, ",")
	}

	return &Configuration{Port: port, Servers: servers, electionTimeout: electionTimeout}, nil
}
