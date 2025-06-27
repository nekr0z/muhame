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

func TestStringWithProto(t *testing.T) {
	a := &NetAddress{Host: "localhost", Port: 8080}
	assert.Equal(t, "http://localhost:8080", a.StringWithProto())
}

func TestSet_error(t *testing.T) {
	tests := []string{
		"no:such:thing",
		"bad:port",
	}
	a := &NetAddress{}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			err := a.Set(tt)
			assert.Error(t, err)
		})
	}
}

func TestUnmarshalText(t *testing.T) {
	a := &NetAddress{}
	err := a.UnmarshalText([]byte("localhost:8080"))
	assert.NoError(t, err)
	assert.Equal(t, "localhost", a.Host)
	assert.Equal(t, 8080, a.Port)
}
