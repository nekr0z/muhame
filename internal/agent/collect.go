package agent

import (
	"math/rand"
	"runtime"

	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/nekr0z/muhame/internal/metrics"
)

func collectBasicMetrics(q *queue, counter int64) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	mm := map[string]metrics.Metric{
		"Alloc":         metrics.Gauge(mem.Alloc),
		"BuckHashSys":   metrics.Gauge(mem.BuckHashSys),
		"Frees":         metrics.Gauge(mem.Frees),
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

	pushAll(mm, q)
}

func collectAuxMetrics(q *queue) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}

	cpu, err := load.Avg()
	if err != nil {
		panic(err)
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	mm := map[string]metrics.Metric{
		"FreeMemory":      metrics.Gauge(vm.Available),
		"TotalMemory":     metrics.Gauge(vm.Total),
		"CPUUtilization1": metrics.Gauge(cpu.Load1),
	}

	pushAll(mm, q)
}

func pushAll(mm map[string]metrics.Metric, q *queue) {
	for k, v := range mm {
		q.push(queuedMetric{
			name: k,
			val:  v,
		})
	}
}
