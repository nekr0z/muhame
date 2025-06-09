package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/muhame/internal/addr"
	"github.com/nekr0z/muhame/internal/storage"
)

var testConfigFilename = filepath.Join("testdata", "config.json")

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
		args []string
		env  []string
		want config
	}{
		{
			name: "default",
			want: config{
				address: addr.NetAddress{
					Host: "localhost",
					Port: 8080,
				},
				st: storage.Config{
					Interval: time.Second * 300,
					Filename: "metrics.sav",
					Restore:  true,
				},
			},
		},
		{
			name: "env and flag",
			args: []string{"-f", "flag.sav", "-d", "flag/db"},
			env:  []string{"FILE_STORAGE_PATH=env.sav"},
			want: config{
				address: addr.NetAddress{
					Host: "localhost",
					Port: 8080,
				},
				st: storage.Config{
					Interval:    time.Second * 300,
					Filename:    "env.sav",
					Restore:     true,
					DatabaseDSN: "flag/db",
				},
			},
		},
		{
			name: "flag and configfile",
			args: []string{"-f", "flag.sav", "-c", testConfigFilename, "-d", "flag/db"},
			want: config{
				address: addr.NetAddress{
					Host: "localhost",
					Port: 8080,
				},
				st: storage.Config{
					Interval:    time.Second * 300,
					Filename:    "flag.sav",
					Restore:     false,
					DatabaseDSN: "flag/db",
				},
			},
		},
		{
			name: "flag, env and configfile",
			args: []string{"-f", "flag.sav", "-d", "flag/db"},
			env:  []string{"FILE_STORAGE_PATH=env.sav", "CONFIG=" + testConfigFilename},
			want: config{
				address: addr.NetAddress{
					Host: "localhost",
					Port: 8080,
				},
				st: storage.Config{
					Interval:    time.Second * 300,
					Filename:    "env.sav",
					Restore:     false,
					DatabaseDSN: "flag/db",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origEnv := os.Environ()
			origArgs := os.Args

			os.Args = append([]string{"server"}, tt.args...)

			t.Cleanup(func() {
				setenv(origEnv)
				os.Args = origArgs
			})

			setenv(tt.env)

			got := newConfig()

			assert.Equal(t, tt.want, got)
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
