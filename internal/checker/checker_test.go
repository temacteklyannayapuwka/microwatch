package checker

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckStatus(t *testing.T) {
	for _, code := range []int{200, 204, 404, 503} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("unexpected method: %s", r.Method)
				}
				w.WriteHeader(code)
			}))
			defer server.Close()
			result := Check(context.Background(), server.Client(), server.URL)
			if result.Err != nil || result.StatusCode != code || result.Up != (code >= 200 && code < 300) {
				t.Fatalf("unexpected result: %+v", result)
			}
			if result.Duration <= 0 {
				t.Fatal("expected a positive duration")
			}
		})
	}
}

func TestCheckTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 50 * time.Millisecond
	result := Check(context.Background(), client, server.URL)
	var timeout net.Error
	if result.Up || !errors.As(result.Err, &timeout) || !timeout.Timeout() {
		t.Fatalf("expected timeout, got %+v", result)
	}
}

func TestCheckCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := Check(ctx, &http.Client{Timeout: time.Second}, "http://localhost")
	if result.Up || !errors.Is(result.Err, context.Canceled) {
		t.Fatalf("expected cancellation, got %+v", result)
	}
}

func TestCheckInvalidURL(t *testing.T) {
	result := Check(context.Background(), &http.Client{Timeout: time.Second}, "://invalid")
	if result.Up || result.Err == nil {
		t.Fatalf("expected URL error, got %+v", result)
	}
}

func TestCheckAllConcurrentAndOrdered(t *testing.T) {
	started := make(chan struct{}, 3)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started <- struct{}{}
		<-release
		switch r.URL.Path {
		case "/healthy":
			w.WriteHeader(http.StatusOK)
		case "/unavailable":
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()
	defer close(release)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan []Result, 1)
	go func() {
		done <- CheckAll(ctx, server.Client(), []string{
			server.URL + "/healthy",
			server.URL + "/unavailable",
			server.URL + "/empty",
		})
	}()
	for i := 0; i < 3; i++ {
		select {
		case <-started:
		case <-ctx.Done():
			t.Fatal("checks did not start concurrently")
		}
	}
	for i := 0; i < 3; i++ {
		release <- struct{}{}
	}
	select {
	case results := <-done:
		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}
		for i, code := range []int{http.StatusOK, http.StatusServiceUnavailable, http.StatusNoContent} {
			if results[i].Err != nil || results[i].StatusCode != code || results[i].Up != (code < 300) {
				t.Errorf("result %d: expected HTTP %d, got %+v", i, code, results[i])
			}
		}
	case <-ctx.Done():
		t.Fatal("checks did not finish")
	}
}

func TestCheckAllCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results := CheckAll(ctx, &http.Client{Timeout: time.Second}, []string{
		"http://localhost/first",
		"http://localhost/second",
	})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for i, result := range results {
		if result.Up || !errors.Is(result.Err, context.Canceled) {
			t.Errorf("result %d: expected cancellation, got %+v", i, result)
		}
	}
}

func TestCheckAllEmpty(t *testing.T) {
	if results := CheckAll(context.Background(), &http.Client{}, nil); len(results) != 0 {
		t.Fatalf("expected no results, got %+v", results)
	}
}
