// Package addr implements network address flag.
package addr

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	defaultHost = "localhost"
	defaultPort = 8080
)

type NetAddress struct {
	Host string
	Port int
}

func (n *NetAddress) String() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

func (n *NetAddress) StringWithProto() string {
	return fmt.Sprintf("http://%s:%d", n.Host, n.Port)
}

func (n *NetAddress) Set(s string) error {
	n.Host, n.Port = defaultHost, defaultPort

	ss := strings.Split(s, ":")
	if len(ss) > 2 {
		return fmt.Errorf("failed to parse network address")
	}

	if len(ss) == 2 {
		port, err := strconv.Atoi(ss[1])
		if err != nil {
			return fmt.Errorf("failed to parse port: %w", err)
		}
		n.Port = port
	}

	if ss[0] != "" {
		n.Host = ss[0]
	}

	return nil
}
