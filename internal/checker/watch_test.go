package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatchRepeatsAfterFailure(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var rounds []bool
	err := Watch(ctx, server.Client(), []string{server.URL}, time.Millisecond, func(results []Result) {
		if len(results) != 1 || results[0].Err != nil {
			t.Fatalf("unexpected results: %+v", results)
		}
		rounds = append(rounds, results[0].Up)
		if len(rounds) == 3 {
			cancel()
		}
	})
	if err != nil || len(rounds) != 3 || rounds[0] || !rounds[1] || !rounds[2] {
		t.Fatalf("expected DOWN, UP, UP; got %v, error %v", rounds, err)
	}
}

func TestWatchCancellationDuringPause(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	rounds := 0
	done := make(chan error, 1)
	go func() {
		done <- Watch(ctx, &http.Client{}, nil, time.Hour, func([]Result) {
			rounds++
			cancel()
		})
	}()
	select {
	case err := <-done:
		if err != nil || rounds != 1 {
			t.Fatalf("expected one round, got %d, error %v", rounds, err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("monitor did not stop during pause")
	}
}

func TestWatchCancellationDuringRequest(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- Watch(ctx, server.Client(), []string{server.URL}, time.Hour, func([]Result) {
			t.Error("canceled round must not be reported")
		})
	}()
	select {
	case <-started:
		cancel()
	case <-ctx.Done():
		t.Fatal("request did not start immediately")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("active request was not canceled")
	}
}

func TestWatchRejectsInvalidInterval(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Second} {
		err := Watch(context.Background(), &http.Client{}, nil, interval, func([]Result) {
			t.Error("invalid interval must not start a round")
		})
		if err == nil {
			t.Fatalf("expected error for interval %s", interval)
		}
	}
}

func TestWatchAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Watch(ctx, &http.Client{}, nil, time.Second, func([]Result) {
		t.Error("canceled monitor must not report results")
	}); err != nil {
		t.Fatal(err)
	}
}
