package metrics

import (
	"sync/atomic"
	"time"
)

// Snapshot is the JSON shape exposed by GET /metrics.
type Snapshot struct {
	Requests       int64 `json:"requests"`
	Errors         int64 `json:"errors"`
	AverageLatency int64 `json:"average_latency"`
}

// Collector records per-request observations and exposes a point-in-time snapshot.
type Collector interface {
	RecordRequest(status int, latency time.Duration)
	Snapshot() Snapshot
}

type collector struct {
	requests       atomic.Int64
	errors         atomic.Int64
	totalLatencyNs atomic.Int64
}

// New returns a thread-safe in-memory metrics collector.
func New() Collector {
	return &collector{}
}

func (c *collector) RecordRequest(status int, latency time.Duration) {
	c.requests.Add(1)
	if status >= 400 {
		c.errors.Add(1)
	}
	c.totalLatencyNs.Add(latency.Nanoseconds())
}

func (c *collector) Snapshot() Snapshot {
	reqs := c.requests.Load()
	var avgMs int64
	if reqs > 0 {
		avgMs = c.totalLatencyNs.Load() / reqs / int64(time.Millisecond)
	}
	return Snapshot{
		Requests:       reqs,
		Errors:         c.errors.Load(),
		AverageLatency: avgMs,
	}
}
