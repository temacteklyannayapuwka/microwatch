package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/temacteklyannayapuwka/microwatch/internal/checker"
	"github.com/temacteklyannayapuwka/microwatch/internal/config"
)

func main() {
	path := flag.String("config", "configs/services.json", "path to service configuration")
	flag.Parse()
	cfg, err := config.Load(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	client := &http.Client{
		Timeout: cfg.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	failed := false
	for _, service := range cfg.Services {
		result := checker.Check(context.Background(), client, service.URL)
		status := "UP"
		if !result.Up {
			status = "DOWN"
			failed = true
		}
		if result.Err != nil {
			fmt.Printf("%s %s: %v\n", service.Name, status, result.Err)
			continue
		}
		fmt.Printf("%s %s: HTTP %d (%s)\n", service.Name, status, result.StatusCode, result.Duration)
	}
	if failed {
		os.Exit(1)
	}
}
