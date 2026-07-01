package metrics

import (
	"testing"
	"time"
)

func TestCollector_RecordAndSnapshot(t *testing.T) {
	t.Parallel()

	c := New()

	c.RecordRequest(200, 10*time.Millisecond)
	c.RecordRequest(200, 30*time.Millisecond)
	c.RecordRequest(500, 20*time.Millisecond)

	snap := c.Snapshot()
	if snap.Requests != 3 {
		t.Fatalf("requests = %d, want 3", snap.Requests)
	}
	if snap.Errors != 1 {
		t.Fatalf("errors = %d, want 1", snap.Errors)
	}
	if snap.AverageLatency != 20 {
		t.Fatalf("average_latency = %d, want 20", snap.AverageLatency)
	}
}

func TestCollector_EmptySnapshot(t *testing.T) {
	t.Parallel()

	snap := New().Snapshot()
	if snap.Requests != 0 || snap.Errors != 0 || snap.AverageLatency != 0 {
		t.Fatalf("expected zero snapshot, got %+v", snap)
	}
}
