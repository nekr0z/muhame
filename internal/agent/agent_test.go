package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/muhame/internal/addr"
)

var testConfigFilename = filepath.Join("testdata", "config.json")

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		args []string
		env  []string
		want Agent
	}{
		{
			name: "default",
			want: Agent{
				address: addr.NetAddress{
					Host: "localhost",
					Port: 8080,
				},
				reportInterval: 10,
				pollInterval:   2,
				workers:        1,
			},
		},
		{
			name: "env and flag",
			args: []string{"-r", "15", "-k", "flag-key"},
			env:  []string{"KEY=env-key"},
			want: Agent{
				address: addr.NetAddress{
					Host: "localhost",
					Port: 8080,
				},
				reportInterval: 15,
				pollInterval:   2,
				workers:        1,
				signKey:        "env-key",
			},
		},
		{
			name: "flag and configfile",
			args: []string{"-r", "15", "--config", testConfigFilename, "-k", "flag-key"},
			want: Agent{
				address: addr.NetAddress{
					Host: "localhost",
					Port: 8080,
				},
				reportInterval: 15,
				pollInterval:   4,
				workers:        1,
				signKey:        "flag-key",
			},
		},
		{
			name: "flag, env and configfile",
			args: []string{"-r", "15", "-k", "flag-key"},
			env:  []string{"REPORT_INTERVAL=17", "CONFIG=" + testConfigFilename},
			want: Agent{
				address: addr.NetAddress{
					Host: "localhost",
					Port: 8080,
				},
				reportInterval: 17,
				pollInterval:   4,
				workers:        1,
				signKey:        "flag-key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origEnv := os.Environ()
			origArgs := os.Args

			os.Args = append([]string{"agent"}, tt.args...)

			t.Cleanup(func() {
				setenv(origEnv)
				os.Args = origArgs
			})

			setenv(tt.env)

			got := New()

			assert.Equal(t, tt.want.address, got.address)
			assert.Equal(t, tt.want.reportInterval*time.Second, got.reportInterval)
			assert.Equal(t, tt.want.pollInterval*time.Second, got.pollInterval)
			assert.Equal(t, tt.want.workers, got.workers)
			assert.Equal(t, tt.want.pubKey, got.pubKey)
			assert.Equal(t, tt.want.signKey, got.signKey)
		})
	}
}

func setenv(env []string) {
	os.Clearenv()

	for _, e := range env {
		s := strings.SplitN(e, "=", 2)
		v := ""

		if len(s) == 2 {
			v = s[1]
		}

		os.Setenv(s[0], v)
	}
}
