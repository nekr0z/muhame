package addr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_netAddress_Set(t *testing.T) {
	tests := []struct {
		name      string
		flagValue string
		host      string
		port      int
	}{
		{
			name:      "no host no port",
			flagValue: "",
			host:      "localhost",
			port:      8080,
		},
		{
			name:      "no host",
			flagValue: ":9090",
			host:      "localhost",
			port:      9090,
		},
		{
			name:      "no port",
			flagValue: "example.com",
			host:      "example.com",
			port:      8080,
		},
		{
			name:      "host and port",
			flagValue: "example.com:6677",
			host:      "example.com",
			port:      6677,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NetAddress{}

			err := n.Set(tt.flagValue)
			assert.NoError(t, err)
			assert.Equal(t, &NetAddress{Host: tt.host, Port: tt.port}, n)
		})
	}
}
