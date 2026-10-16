package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/temacteklyannayapuwka/microwatch/internal/checker"
	"github.com/temacteklyannayapuwka/microwatch/internal/config"
)

type ServiceStatus struct {
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Up         bool      `json:"up"`
	StatusCode int       `json:"status_code"`
	DurationMS float64   `json:"duration_ms"`
	CheckedAt  time.Time `json:"checked_at"`
	Error      string    `json:"error,omitempty"`
}

type Handler struct {
	mu       sync.RWMutex
	services []ServiceStatus
}

func (h *Handler) Update(services []config.Service, results []checker.Result, at time.Time) {
	statuses := make([]ServiceStatus, len(services))
	for i, service := range services {
		result := results[i]
		statuses[i] = ServiceStatus{
			Name: service.Name, URL: service.URL, Up: result.Up,
			StatusCode: result.StatusCode, DurationMS: float64(result.Duration) / float64(time.Millisecond),
			CheckedAt: at.UTC(),
		}
		if result.Err != nil {
			statuses[i].Error = result.Err.Error()
		}
	}
	h.mu.Lock()
	h.services = statuses
	h.mu.Unlock()
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Status string `json:"status"`
		}{Status: "ok"})
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		h.mu.RLock()
		services := append([]ServiceStatus{}, h.services...)
		h.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Services []ServiceStatus `json:"services"`
		}{Services: services})
	})
	return mux
}
