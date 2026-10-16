package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "services.json")
	data := `{"timeout":"3s","services":[{"name":"auth","url":"http://localhost:8081/health"},{"name":"orders","url":"https://example.com/health"}]}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Timeout != 3*time.Second || len(cfg.Services) != 2 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.Services[0].Name != "auth" || cfg.Services[1].URL != "https://example.com/health" {
		t.Fatalf("unexpected services: %+v", cfg.Services)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected an error for a missing file")
	}
}

func TestDecodeRejectsInvalidConfig(t *testing.T) {
	cases := map[string]string{
		"malformed JSON":        `{"timeout":`,
		"null":                  `null`,
		"unknown field":         `{"timeout":"3s","typo":true,"services":[{"name":"a","url":"http://localhost"}]}`,
		"unknown service field": `{"timeout":"3s","services":[{"name":"a","url":"http://localhost","typo":true}]}`,
		"trailing JSON":         `{"timeout":"3s","services":[]} {}`,
		"trailing garbage":      `{"timeout":"3s","services":[]} trailing`,
		"missing timeout":       `{"services":[{"name":"a","url":"http://localhost"}]}`,
		"invalid timeout":       `{"timeout":"soon","services":[]}`,
		"zero timeout":          `{"timeout":"0s","services":[]}`,
		"negative timeout":      `{"timeout":"-1s","services":[]}`,
		"empty services":        `{"timeout":"3s","services":[]}`,
		"empty name":            `{"timeout":"3s","services":[{"name":" ","url":"http://localhost"}]}`,
		"duplicate names":       `{"timeout":"3s","services":[{"name":"a","url":"http://localhost"},{"name":"a","url":"http://localhost:8081"}]}`,
		"relative URL":          `{"timeout":"3s","services":[{"name":"a","url":"/health"}]}`,
		"unsupported scheme":    `{"timeout":"3s","services":[{"name":"a","url":"ftp://localhost"}]}`,
		"missing host":          `{"timeout":"3s","services":[{"name":"a","url":"http:///health"}]}`,
		"credentials":           `{"timeout":"3s","services":[{"name":"a","url":"http://user:pass@localhost"}]}`,
		"fragment":              `{"timeout":"3s","services":[{"name":"a","url":"http://localhost/#health"}]}`,
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := decode(strings.NewReader(data)); err == nil {
				t.Fatal("expected invalid configuration to be rejected")
			}
		})
	}
}
