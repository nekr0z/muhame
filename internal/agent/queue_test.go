package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndpoint(t *testing.T) {
	type args struct {
		addr       string
		metricType string
		name       string
		value      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			args: args{"http://localhost:8080", "counter", "test", "1"},
			want: "http://localhost:8080/update/counter/test/1",
		},
		{
			name: "trailing slash",
			args: args{"http://localhost:8080/", "gauge", "test", "1.1"},
			want: "http://localhost:8080/update/gauge/test/1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := endpoint(tt.args.addr, tt.args.metricType, tt.args.name, tt.args.value); got != tt.want {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}
