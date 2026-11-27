package builder

import (
	"strings"
	"testing"
	"time"
)

func TestWorkerStats(t *testing.T) {
	w := NewWorker(nil, nil, nil, nil)
	if err := w.Enqueue(Job{ProjectID: "p1"}); err != nil {
		t.Fatal(err)
	}
	if err := w.Enqueue(Job{ProjectID: "p2"}); err != nil {
		t.Fatal(err)
	}

	started := time.Now()
	w.beginJob()
	stats := w.Stats()
	if stats.Queued != 2 || stats.Active != 1 || stats.Total != 1 {
		t.Fatalf("in-flight stats = %+v", stats)
	}

	w.finishJob("build exploded", started)
	stats = w.Stats()
	if stats.Active != 0 || stats.Total != 1 || stats.Failed != 1 || stats.Succeeded != 0 {
		t.Fatalf("finished stats = %+v", stats)
	}
	if !strings.Contains(stats.LastFailure, "build exploded") || stats.LastMS < 0 {
		t.Fatalf("failure stats = %+v", stats)
	}

	w.beginJob()
	w.finishJob("", started)
	stats = w.Stats()
	if stats.Succeeded != 1 || stats.Failed != 1 || stats.LastFailure != "" {
		t.Fatalf("success after failure = %+v", stats)
	}
}
