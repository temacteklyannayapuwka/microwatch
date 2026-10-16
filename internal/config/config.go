package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
)

type Service struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Config struct {
	Timeout  time.Duration
	Services []Service
}

func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()
	return decode(file)
}

func decode(reader io.Reader) (Config, error) {
	var raw struct {
		Timeout  string    `json:"timeout"`
		Services []Service `json:"services"`
	}
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return Config{}, fmt.Errorf("config must contain a single JSON object")
	}
	timeout, err := time.ParseDuration(raw.Timeout)
	if err != nil || timeout <= 0 {
		return Config{}, fmt.Errorf("timeout must be a positive duration")
	}
	if len(raw.Services) == 0 {
		return Config{}, fmt.Errorf("at least one service is required")
	}
	names := make(map[string]bool, len(raw.Services))
	for i, service := range raw.Services {
		if strings.TrimSpace(service.Name) == "" || service.Name != strings.TrimSpace(service.Name) {
			return Config{}, fmt.Errorf("service %d must have a non-empty name without surrounding whitespace", i+1)
		}
		if names[service.Name] {
			return Config{}, fmt.Errorf("duplicate service name %q", service.Name)
		}
		names[service.Name] = true
		endpoint, err := url.Parse(service.URL)
		if err != nil || endpoint.Hostname() == "" || (endpoint.Scheme != "http" && endpoint.Scheme != "https") {
			return Config{}, fmt.Errorf("service %q must have an absolute HTTP or HTTPS URL", service.Name)
		}
		if endpoint.User != nil || endpoint.Fragment != "" {
			return Config{}, fmt.Errorf("service %q URL must not contain credentials or a fragment", service.Name)
		}
	}
	return Config{Timeout: timeout, Services: raw.Services}, nil
}
