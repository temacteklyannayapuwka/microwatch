package checker

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func Watch(ctx context.Context, client *http.Client, endpoints []string, interval time.Duration, report func([]Result)) error {
	if interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	for {
		if ctx.Err() != nil {
			return nil
		}
		results := CheckAll(ctx, client, endpoints)
		if ctx.Err() != nil {
			return nil
		}
		report(results)
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}
