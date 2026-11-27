package server

import (
	"sync"
	"testing"
	"time"
)

func TestMetricsAllowedOnlyFromLoopback(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "::1"} {
		if !metricsAllowed(ip) {
			t.Errorf("metricsAllowed(%q) = false, want true", ip)
		}
	}
	for _, ip := range []string{"10.0.0.5", "203.0.113.9", "", "not-an-ip"} {
		if metricsAllowed(ip) {
			t.Errorf("metricsAllowed(%q) = true, want false", ip)
		}
	}
}

func TestMetricsConcurrentRecording(t *testing.T) {
	var m metrics
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				m.record(j%5 == 0, time.Now())
			}
		}(i)
	}
	wg.Wait()

	requests, errors, last := m.snapshot()
	if requests != 800 {
		t.Fatalf("requests = %d, want 800", requests)
	}
	if errors != 160 {
		t.Fatalf("errors = %d, want 160", errors)
	}
	if last.IsZero() {
		t.Fatal("last request time was not recorded")
	}
}
