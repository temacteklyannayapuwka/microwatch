package checker

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	StatusCode int
	Duration   time.Duration
	Up         bool
	Err        error
}

func CheckAll(ctx context.Context, client *http.Client, endpoints []string) []Result {
	results := make([]Result, len(endpoints))
	var wg sync.WaitGroup
	for i, endpoint := range endpoints {
		wg.Add(1)
		go func(index int, url string) {
			defer wg.Done()
			results[index] = Check(ctx, client, url)
		}(i, endpoint)
	}
	wg.Wait()
	return results
}

func Check(ctx context.Context, client *http.Client, endpoint string) Result {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Result{Duration: time.Since(start), Err: err}
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{Duration: time.Since(start), Err: err}
	}
	resp.Body.Close()
	return Result{
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start),
		Up:         resp.StatusCode >= 200 && resp.StatusCode < 300,
	}
}
