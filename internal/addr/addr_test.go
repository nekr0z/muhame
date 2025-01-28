package addr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_netAddress_Set(t *testing.T) {
	tests := []struct {
		name string
		str  string
		host string
		port int
	}{
		{
			name: "no host no port",
			str:  "",
			host: "localhost",
			port: 8080,
		},
		{
			name: "no host",
			str:  ":9090",
			host: "localhost",
			port: 9090,
		},
		{
			name: "no port",
			str:  "example.com",
			host: "example.com",
			port: 8080,
		},
		{
			name: "host and port",
			str:  "example.com:6677",
			host: "example.com",
			port: 6677,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NetAddress{}

			err := n.Set(tt.str)
			assert.NoError(t, err)
			assert.Equal(t, &NetAddress{Host: tt.host, Port: tt.port}, n)
		})
	}
}
