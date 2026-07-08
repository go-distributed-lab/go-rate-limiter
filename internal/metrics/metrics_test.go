package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetrics_Counters(t *testing.T) {
	m := NewMetrics("token_bucket")

	m.RecordAllowed()
	m.RecordAllowed()
	m.RecordDenied()

	if got := m.Allowed.Load(); got != 2 {
		t.Fatalf("expected allowed=2, got %d", got)
	}
	if got := m.Denied.Load(); got != 1 {
		t.Fatalf("expected denied=1, got %d", got)
	}
}

func TestMetrics_Handler(t *testing.T) {
	m := NewMetrics("sliding_window")
	m.RecordAllowed()
	m.RecordAllowed()
	m.RecordAllowed()
	m.RecordDenied()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	m.Handler()(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var snap Snapshot
	if err := json.NewDecoder(w.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if snap.Algorithm != "sliding_window" {
		t.Errorf("expected algorithm=sliding_window, got %s", snap.Algorithm)
	}
	if snap.Allowed != 3 {
		t.Errorf("expected allowed=3, got %d", snap.Allowed)
	}
	if snap.Denied != 1 {
		t.Errorf("expected denied=1, got %d", snap.Denied)
	}
	if snap.Total != 4 {
		t.Errorf("expected total=4, got %d", snap.Total)
	}
}

func TestMetrics_ConcurrentCounters(t *testing.T) {
	m := NewMetrics("fixed_window")
	done := make(chan struct{})

	for range 100 {
		go func() {
			m.RecordAllowed()
			done <- struct{}{}
		}()
	}
	for range 100 {
		<-done
	}

	if got := m.Allowed.Load(); got != 100 {
		t.Fatalf("expected allowed=100, got %d", got)
	}
}
