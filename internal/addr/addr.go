// Package addr implements network address flag.
package addr

import (
	"fmt"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port int
}

func (n *NetAddress) String() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

func (n *NetAddress) StringWithProto() string {
	return fmt.Sprintf("https://%s:%d", n.Host, n.Port)
}

func (n *NetAddress) Set(flagValue string) error {
	s := strings.Split(flagValue, ":")
	if len(s) > 2 {
		return fmt.Errorf("failed to parse network address")
	}

	if len(s) < 2 {
		n.Port = 8080
	} else {
		var err error

		n.Port, err = strconv.Atoi(s[1])
		if err != nil {
			return fmt.Errorf("failed to parse port: %w", err)
		}
	}

	n.Host = s[0]
	if n.Host == "" {
		n.Host = "localhost"
	}

	return nil
}
