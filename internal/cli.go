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
	var serverIps ServerIps
	var port string

	flag.Var(&serverIps, "servers", "server ips")
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

	return &Configuration{Port: port, ServerIps: serverIps, electionTimeout: electionTimeout}, nil
}
