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

// NetAddress represents network address.
type NetAddress struct {
	Host string
	Port int
}

// String satisfies fmt.Stringer.
func (n *NetAddress) String() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

// StringWithProto returns network address with protocol (currently only HTTP).
func (n *NetAddress) StringWithProto() string {
	return fmt.Sprintf("http://%s:%d", n.Host, n.Port)
}

// Set implements flag.Value.
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

// UnmarshalText satisfies encoding.TextUnmarshaler.
func (n *NetAddress) UnmarshalText(text []byte) error {
	return n.Set(string(text))
}
