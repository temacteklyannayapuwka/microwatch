package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/temacteklyannayapuwka/microwatch/internal/checker"
	"github.com/temacteklyannayapuwka/microwatch/internal/config"
	"github.com/temacteklyannayapuwka/microwatch/internal/monitor"
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
		var tracker monitor.Tracker
		if err := checker.Watch(ctx, client, endpoints, *interval, func(results []checker.Result) {
			printResults(cfg.Services, results)
			printTransitions(&tracker, cfg.Services, results, time.Now().UTC())
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

func printTransitions(tracker *monitor.Tracker, services []config.Service, results []checker.Result, at time.Time) {
	for i, service := range services {
		transition, changed := tracker.Observe(service.Name, results[i].Up, at)
		if !changed {
			continue
		}
		if transition.To.Up {
			fmt.Printf("%s %s RECOVERED (observed downtime %s)\n", at.Format(time.RFC3339Nano), service.Name, at.Sub(transition.From.Since))
		} else {
			fmt.Printf("%s %s BECAME DOWN\n", at.Format(time.RFC3339Nano), service.Name)
		}
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
