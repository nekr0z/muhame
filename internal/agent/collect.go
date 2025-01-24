package agent

import (
	"math/rand"
	"runtime"

	"github.com/nekr0z/muhame/internal/metrics"
)

func collectMetrics(q *queue, counter int64) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	q.push(metric{name: "Alloc", m: metrics.Gauge(mem.Alloc)})
	q.push(metric{name: "BuckHashSys", m: metrics.Gauge(mem.BuckHashSys)})
	q.push(metric{name: "GCCPUFraction", m: metrics.Gauge(mem.GCCPUFraction)})
	q.push(metric{name: "GCSys", m: metrics.Gauge(mem.GCSys)})
	q.push(metric{name: "HeapAlloc", m: metrics.Gauge(mem.HeapAlloc)})
	q.push(metric{name: "HeapIdle", m: metrics.Gauge(mem.HeapIdle)})
	q.push(metric{name: "HeapInuse", m: metrics.Gauge(mem.HeapInuse)})
	q.push(metric{name: "HeapObjects", m: metrics.Gauge(mem.HeapObjects)})
	q.push(metric{name: "HeapReleased", m: metrics.Gauge(mem.HeapReleased)})
	q.push(metric{name: "HeapSys", m: metrics.Gauge(mem.HeapSys)})
	q.push(metric{name: "LastGC", m: metrics.Gauge(mem.LastGC)})
	q.push(metric{name: "Lookups", m: metrics.Gauge(mem.Lookups)})
	q.push(metric{name: "MCacheInuse", m: metrics.Gauge(mem.MCacheInuse)})
	q.push(metric{name: "MCacheSys", m: metrics.Gauge(mem.MCacheSys)})
	q.push(metric{name: "MSpanInuse", m: metrics.Gauge(mem.MSpanInuse)})
	q.push(metric{name: "MSpanSys", m: metrics.Gauge(mem.MSpanSys)})
	q.push(metric{name: "Mallocs", m: metrics.Gauge(mem.Mallocs)})
	q.push(metric{name: "NextGC", m: metrics.Gauge(mem.NextGC)})
	q.push(metric{name: "NumForcedGC", m: metrics.Gauge(mem.NumForcedGC)})
	q.push(metric{name: "NumGC", m: metrics.Gauge(mem.NumGC)})
	q.push(metric{name: "OtherSys", m: metrics.Gauge(mem.OtherSys)})
	q.push(metric{name: "PauseTotalNs", m: metrics.Gauge(mem.PauseTotalNs)})
	q.push(metric{name: "StackInuse", m: metrics.Gauge(mem.StackInuse)})
	q.push(metric{name: "StackSys", m: metrics.Gauge(mem.StackSys)})
	q.push(metric{name: "Sys", m: metrics.Gauge(mem.Sys)})
	q.push(metric{name: "TotalAlloc", m: metrics.Gauge(mem.TotalAlloc)})

	q.push(metric{name: "RandomValue", m: metrics.Gauge(rand.Float64())})
	q.push(metric{name: "RandomValue", m: metrics.Counter(counter)})
}
