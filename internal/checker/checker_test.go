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
