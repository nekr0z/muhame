package metrics_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/muhame/internal/metrics"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		t       string
		v       string
		want    metrics.Metric
		wantErr bool
	}{
		{
			name:    "gauge",
			t:       "gauge",
			v:       "1.2",
			want:    metrics.Gauge(1.2),
			wantErr: false,
		},
		{
			name:    "counter",
			t:       "counter",
			v:       "12",
			want:    metrics.Counter(12),
			wantErr: false,
		},
		{
			name:    "parse error",
			t:       "counter",
			v:       "3.14",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "wrong type",
			t:       "elephant",
			v:       "0",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := metrics.Parse(tc.t, tc.v)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.want, got)
		})
	}
}
