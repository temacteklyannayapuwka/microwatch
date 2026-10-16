package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/temacteklyannayapuwka/microwatch/internal/checker"
	"github.com/temacteklyannayapuwka/microwatch/internal/config"
)

func TestRoutes(t *testing.T) {
	var handler Handler
	routes := handler.Routes()
	for _, tc := range []struct {
		method string
		path   string
		code   int
		body   string
	}{
		{http.MethodGet, "/health", http.StatusOK, "{\"status\":\"ok\"}\n"},
		{http.MethodGet, "/status", http.StatusOK, "{\"services\":[]}\n"},
		{http.MethodPost, "/status", http.StatusMethodNotAllowed, ""},
		{http.MethodPost, "/health", http.StatusMethodNotAllowed, ""},
		{http.MethodGet, "/missing", http.StatusNotFound, ""},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			routes.ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))
			if response.Code != tc.code {
				t.Fatalf("status %d, want %d", response.Code, tc.code)
			}
			if tc.body != "" && (response.Body.String() != tc.body || response.Header().Get("Content-Type") != "application/json") {
				t.Fatalf("unexpected JSON response: %s", response.Body.String())
			}
		})
	}
}

func TestStatusLatestRound(t *testing.T) {
	var handler Handler
	at := time.Date(2026, 10, 16, 12, 0, 0, 0, time.FixedZone("local", 3*60*60))
	services := []config.Service{{Name: "catalog", URL: "http://catalog"}, {Name: "orders", URL: "http://orders"}}
	handler.Update(services, []checker.Result{{Up: true, StatusCode: 204, Duration: 1500 * time.Microsecond}, {Err: errors.New("request timeout")}}, at)
	routes := handler.Routes()
	readStatus := func() []ServiceStatus {
		t.Helper()
		response := httptest.NewRecorder()
		routes.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/status", nil))
		var body struct {
			Services []ServiceStatus `json:"services"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return body.Services
	}
	first := readStatus()
	if len(first) != 2 || first[0].Name != "catalog" || !first[0].Up || first[0].StatusCode != 204 || first[0].DurationMS != 1.5 || first[0].Error != "" || !first[0].CheckedAt.Equal(at) || first[0].CheckedAt.Location() != time.UTC {
		t.Fatalf("unexpected status: %+v", first)
	}
	if first[1].Name != "orders" || first[1].Up || first[1].Error != "request timeout" || first[1].StatusCode != 0 {
		t.Fatalf("unexpected failed status: %+v", first[1])
	}
	handler.Update(services[:1], []checker.Result{{StatusCode: 503}}, at.Add(time.Second))
	latest := readStatus()
	if len(latest) != 1 || latest[0].Up || latest[0].StatusCode != 503 || !latest[0].CheckedAt.Equal(at.Add(time.Second)) {
		t.Fatalf("expected replacement round: %+v", latest)
	}
	response := httptest.NewRecorder()
	routes.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusOK {
		t.Fatal("monitor health must not depend on service availability")
	}
}

func TestConcurrentStatusReadsAndUpdates(t *testing.T) {
	var handler Handler
	routes := handler.Routes()
	services := []config.Service{{Name: "catalog", URL: "http://catalog"}}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				handler.Update(services, []checker.Result{{Up: true, StatusCode: 200}}, time.Now())
				response := httptest.NewRecorder()
				routes.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/status", nil))
				if !json.Valid(response.Body.Bytes()) {
					t.Error("invalid status JSON")
				}
			}
		}()
	}
	wg.Wait()
}
