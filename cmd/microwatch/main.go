package main

import (
	"flag"
	"fmt"
	"os"

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
	fmt.Printf("Loaded %d services (timeout: %s)\n", len(cfg.Services), cfg.Timeout)
	for _, service := range cfg.Services {
		fmt.Printf("%s: %s\n", service.Name, service.URL)
	}
}
