package agent

import (
	"math/rand"
	"runtime"

	"github.com/nekr0z/muhame/internal/metrics"
)

func collectMetrics(q *queue, counter int64) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	q.push(queuedMetric{name: "Alloc", val: metrics.Gauge(mem.Alloc)})
	q.push(queuedMetric{name: "BuckHashSys", val: metrics.Gauge(mem.BuckHashSys)})
	q.push(queuedMetric{name: "GCCPUFraction", val: metrics.Gauge(mem.GCCPUFraction)})
	q.push(queuedMetric{name: "GCSys", val: metrics.Gauge(mem.GCSys)})
	q.push(queuedMetric{name: "HeapAlloc", val: metrics.Gauge(mem.HeapAlloc)})
	q.push(queuedMetric{name: "HeapIdle", val: metrics.Gauge(mem.HeapIdle)})
	q.push(queuedMetric{name: "HeapInuse", val: metrics.Gauge(mem.HeapInuse)})
	q.push(queuedMetric{name: "HeapObjects", val: metrics.Gauge(mem.HeapObjects)})
	q.push(queuedMetric{name: "HeapReleased", val: metrics.Gauge(mem.HeapReleased)})
	q.push(queuedMetric{name: "HeapSys", val: metrics.Gauge(mem.HeapSys)})
	q.push(queuedMetric{name: "LastGC", val: metrics.Gauge(mem.LastGC)})
	q.push(queuedMetric{name: "Lookups", val: metrics.Gauge(mem.Lookups)})
	q.push(queuedMetric{name: "MCacheInuse", val: metrics.Gauge(mem.MCacheInuse)})
	q.push(queuedMetric{name: "MCacheSys", val: metrics.Gauge(mem.MCacheSys)})
	q.push(queuedMetric{name: "MSpanInuse", val: metrics.Gauge(mem.MSpanInuse)})
	q.push(queuedMetric{name: "MSpanSys", val: metrics.Gauge(mem.MSpanSys)})
	q.push(queuedMetric{name: "Mallocs", val: metrics.Gauge(mem.Mallocs)})
	q.push(queuedMetric{name: "NextGC", val: metrics.Gauge(mem.NextGC)})
	q.push(queuedMetric{name: "NumForcedGC", val: metrics.Gauge(mem.NumForcedGC)})
	q.push(queuedMetric{name: "NumGC", val: metrics.Gauge(mem.NumGC)})
	q.push(queuedMetric{name: "OtherSys", val: metrics.Gauge(mem.OtherSys)})
	q.push(queuedMetric{name: "PauseTotalNs", val: metrics.Gauge(mem.PauseTotalNs)})
	q.push(queuedMetric{name: "StackInuse", val: metrics.Gauge(mem.StackInuse)})
	q.push(queuedMetric{name: "StackSys", val: metrics.Gauge(mem.StackSys)})
	q.push(queuedMetric{name: "Sys", val: metrics.Gauge(mem.Sys)})
	q.push(queuedMetric{name: "TotalAlloc", val: metrics.Gauge(mem.TotalAlloc)})

	q.push(queuedMetric{name: "RandomValue", val: metrics.Gauge(rand.Float64())})
	q.push(queuedMetric{name: "RandomValue", val: metrics.Counter(counter)})
}
