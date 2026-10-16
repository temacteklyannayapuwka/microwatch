package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/temacteklyannayapuwka/microwatch/internal/checker"
	"github.com/temacteklyannayapuwka/microwatch/internal/config"
)

func main() {
	path := flag.String("config", "configs/services.json", "path to service configuration")
	interval := flag.Duration("interval", 0, "pause between check rounds; 0 checks once")
	flag.Parse()
	if *interval < 0 {
		fmt.Fprintln(os.Stderr, "interval must not be negative")
		os.Exit(1)
	}
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
	endpoints := make([]string, len(cfg.Services))
	for i, service := range cfg.Services {
		endpoints[i] = service.URL
	}
	if *interval > 0 {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := checker.Watch(ctx, client, endpoints, *interval, func(results []checker.Result) {
			printResults(cfg.Services, results)
		}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if printResults(cfg.Services, checker.CheckAll(context.Background(), client, endpoints)) {
		os.Exit(1)
	}
}

func printResults(services []config.Service, results []checker.Result) bool {
	failed := false
	for i, service := range services {
		result := results[i]
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
	return failed
}
