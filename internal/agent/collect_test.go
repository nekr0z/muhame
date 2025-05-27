package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/muhame/internal/metrics"
)

func TestCollectBasicMetrics(t *testing.T) {
	q := &queue{}

	collectBasicMetrics(q, 15)

	var names []string

	for m := q.pop(); m != nil; m = q.pop() {
		names = append(names, m.name)

		if m.name == "PollCount" {
			assert.Equal(t, metrics.Counter(15), m.val)

		}
	}

	assert.Contains(t, names, "TotalAlloc")
}

func TestCollectAuxMetrics(t *testing.T) {
	q := &queue{}

	collectAuxMetrics(q)

	var names []string

	for m := q.pop(); m != nil; m = q.pop() {
		names = append(names, m.name)
	}

	assert.Contains(t, names, "FreeMemory")
}
