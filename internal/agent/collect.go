package agent

import (
	"math/rand"
	"runtime"

	"github.com/nekr0z/muhame/internal/metrics"
)

func collectMetrics(q *queue, counter int64) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	mm := map[string]metrics.Metric{
		"Alloc":         metrics.Gauge(mem.Alloc),
		"BuckHashSys":   metrics.Gauge(mem.BuckHashSys),
		"GCCPUFraction": metrics.Gauge(mem.GCCPUFraction),
		"GCSys":         metrics.Gauge(mem.GCSys),
		"HeapAlloc":     metrics.Gauge(mem.HeapAlloc),
		"HeapIdle":      metrics.Gauge(mem.HeapIdle),
		"HeapInuse":     metrics.Gauge(mem.HeapInuse),
		"HeapObjects":   metrics.Gauge(mem.HeapObjects),
		"HeapReleased":  metrics.Gauge(mem.HeapReleased),
		"HeapSys":       metrics.Gauge(mem.HeapSys),
		"LastGC":        metrics.Gauge(mem.LastGC),
		"Lookups":       metrics.Gauge(mem.Lookups),
		"MCacheInuse":   metrics.Gauge(mem.MCacheInuse),
		"MCacheSys":     metrics.Gauge(mem.MCacheSys),
		"MSpanInuse":    metrics.Gauge(mem.MSpanInuse),
		"MSpanSys":      metrics.Gauge(mem.MSpanSys),
		"Mallocs":       metrics.Gauge(mem.Mallocs),
		"NextGC":        metrics.Gauge(mem.NextGC),
		"NumForcedGC":   metrics.Gauge(mem.NumForcedGC),
		"NumGC":         metrics.Gauge(mem.NumGC),
		"OtherSys":      metrics.Gauge(mem.OtherSys),
		"PauseTotalNs":  metrics.Gauge(mem.PauseTotalNs),
		"StackInuse":    metrics.Gauge(mem.StackInuse),
		"StackSys":      metrics.Gauge(mem.StackSys),
		"Sys":           metrics.Gauge(mem.Sys),
		"TotalAlloc":    metrics.Gauge(mem.TotalAlloc),

		"RandomValue": metrics.Gauge(rand.Float64()),
		"PollCount":   metrics.Counter(counter),
	}

	for k, v := range mm {
		q.push(queuedMetric{
			name: k,
			val:  v,
		})
	}
}
