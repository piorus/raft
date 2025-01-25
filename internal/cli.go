package internal

import (
	"flag"
	"fmt"
	"strings"
)

type ServerIps []string

func (s *ServerIps) Set(value string) error {
	*s = append(*s, strings.Split(value, ",")...)
	return nil
}

func (s *ServerIps) String() string {
	return strings.Join(*s, ",")
}

type Configuration struct {
	Port      string
	ServerIps ServerIps
}

func ParseConfiguration() (*Configuration, error) {
	var serverIps ServerIps
	var port string

	flag.Var(&serverIps, "servers", "server ips")
	flag.StringVar(&port, "port", "", "Port")

	flag.Parse()

	if port == "" {
		return nil, fmt.Errorf("--Port is required")
	}

	return &Configuration{Port: port, ServerIps: serverIps}, nil
}
