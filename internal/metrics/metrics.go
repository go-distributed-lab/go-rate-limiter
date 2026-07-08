package metrics

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

type Metrics struct {
	Algorithm string
	Allowed   atomic.Int64
	Denied    atomic.Int64
	StartTime time.Time
}

func NewMetrics(algorithm string) *Metrics {
	return &Metrics{
		Algorithm: algorithm,
		StartTime: time.Now(),
	}
}

func (m *Metrics) RecordAllowed() {
	m.Allowed.Add(1)
}

func (m *Metrics) RecordDenied() {
	m.Denied.Add(1)
}

type Snapshot struct {
	Algorithm string  `json:"algorithm"`
	Allowed   int64   `json:"allowed"`
	Denied    int64   `json:"denied"`
	Total     int64   `json:"total"`
	UptimeSec float64 `json:"uptime_seconds"`
}

func (m *Metrics) Handler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed := m.Allowed.Load()
		denied := m.Denied.Load()
		total := allowed + denied
		uptime := time.Since(m.StartTime).Seconds()

		snapshot := Snapshot{
			Algorithm: m.Algorithm,
			Allowed:   allowed,
			Denied:    denied,
			Total:     total,
			UptimeSec: uptime,
		}

		w.Header().Set("Content-type", "application/json")
		json.NewEncoder(w).Encode(snapshot)
	})
}
