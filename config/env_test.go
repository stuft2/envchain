package config

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoadParsesTaggedConfig(t *testing.T) {
	type testConfig struct {
		Port      int               `env:"PORT,default=8080"`
		Debug     bool              `env:"DEBUG,default=false"`
		Timeout   time.Duration     `env:"TIMEOUT,default=5s"`
		BaseURL   *url.URL          `env:"BASE_URL,required"`
		Aliases   []string          `env:"ALIASES,default=api,worker"`
		Labels    map[string]string `env:"LABELS,default=team=platform,service=envchain"`
		StartedAt time.Time         `env:"STARTED_AT,layout=2006-01-02"`
		Mode      string            `env:"MODE,default=dev,oneof=dev|prod"`
		MaxBytes  int64             `env:"MAX_BYTES,default=2MiB,format=bytes"`
	}

	t.Setenv("PORT", "9090")
	t.Setenv("DEBUG", "true")
	t.Setenv("TIMEOUT", "30s")
	t.Setenv("BASE_URL", "https://example.com")
	t.Setenv("ALIASES", "web, jobs")
	t.Setenv("LABELS", "team=core,service=api")
	t.Setenv("STARTED_AT", "2026-03-10")
	t.Setenv("MODE", "prod")
	t.Setenv("MAX_BYTES", "4MiB")

	var cfg testConfig
	if err := Load(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 9090 {
		t.Fatalf("expected port 9090, got %d", cfg.Port)
	}
	if !cfg.Debug {
		t.Fatal("expected debug=true")
	}
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("expected timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.BaseURL == nil || cfg.BaseURL.String() != "https://example.com" {
		t.Fatalf("expected parsed url, got %#v", cfg.BaseURL)
	}
	if want := []string{"web", "jobs"}; !reflect.DeepEqual(cfg.Aliases, want) {
		t.Fatalf("expected aliases %v, got %v", want, cfg.Aliases)
	}
	if want := map[string]string{"team": "core", "service": "api"}; !reflect.DeepEqual(cfg.Labels, want) {
		t.Fatalf("expected labels %v, got %v", want, cfg.Labels)
	}
	if want := time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC); !cfg.StartedAt.Equal(want) {
		t.Fatalf("expected started_at %v, got %v", want, cfg.StartedAt)
	}
	if cfg.Mode != "prod" {
		t.Fatalf("expected mode prod, got %q", cfg.Mode)
	}
	if cfg.MaxBytes != 4*1024*1024 {
		t.Fatalf("expected max bytes 4194304, got %d", cfg.MaxBytes)
	}
}

func TestLoadUsesDefaultsForUnsetFields(t *testing.T) {
	type testConfig struct {
		Port    int           `env:"PORT,default=8080"`
		Timeout time.Duration `env:"TIMEOUT,default=5s"`
		Names   []string      `env:"NAMES,default=api,worker"`
	}

	var cfg testConfig
	if err := Load(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", cfg.Port)
	}
	if cfg.Timeout != 5*time.Second {
		t.Fatalf("expected default timeout 5s, got %v", cfg.Timeout)
	}
	if want := []string{"api", "worker"}; !reflect.DeepEqual(cfg.Names, want) {
		t.Fatalf("expected default names %v, got %v", want, cfg.Names)
	}
}

func TestLoadRejectsMissingRequiredField(t *testing.T) {
	type testConfig struct {
		BaseURL string `env:"BASE_URL,required"`
	}

	var cfg testConfig
	err := Load(&cfg)
	if err == nil {
		t.Fatal("expected error for missing required env")
	}
	if got := err.Error(); !strings.Contains(got, `BASE_URL`) || !strings.Contains(got, "required") {
		t.Fatalf("expected required env error, got %q", got)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	type testConfig struct {
		Port int `env:"PORT"`
	}

	t.Setenv("PORT", "not-a-number")

	var cfg testConfig
	err := Load(&cfg)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if got := err.Error(); !strings.Contains(got, `PORT`) || !strings.Contains(got, `"not-a-number"`) {
		t.Fatalf("expected key and value in parse error, got %q", got)
	}
}

func TestLoadRejectsInvalidLoaderTargets(t *testing.T) {
	type testConfig struct {
		Port int `env:"PORT"`
	}

	var cfg testConfig

	if err := Load(cfg); err == nil {
		t.Fatal("expected error for non-pointer target")
	}

	var ptr *testConfig
	if err := Load(ptr); err == nil {
		t.Fatal("expected error for nil pointer target")
	}
}

func TestLoadHonorsIgnoredFieldsAndCustomSeparators(t *testing.T) {
	type testConfig struct {
		Raw   string            `env:"-"`
		Names []string          `env:"NAMES,sep=|"`
		Meta  map[string]string `env:"META,entrysep=;,kvsep=:"`
	}

	t.Setenv("NAMES", "api|worker")
	t.Setenv("META", "team:platform;service:envchain")

	cfg := testConfig{Raw: "keep-me"}
	if err := Load(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Raw != "keep-me" {
		t.Fatalf("expected ignored field to remain unchanged, got %q", cfg.Raw)
	}
	if want := []string{"api", "worker"}; !reflect.DeepEqual(cfg.Names, want) {
		t.Fatalf("expected custom separated names %v, got %v", want, cfg.Names)
	}
	if want := map[string]string{"team": "platform", "service": "envchain"}; !reflect.DeepEqual(cfg.Meta, want) {
		t.Fatalf("expected custom separated meta %v, got %v", want, cfg.Meta)
	}
}

func TestLoadParsesNestedStructs(t *testing.T) {
	type serverConfig struct {
		Host string `env:"HOST,default=127.0.0.1"`
		Port int    `env:"PORT,default=8080"`
	}
	type appConfig struct {
		Name   string `env:"APP_NAME,default=envchain"`
		Server serverConfig
		Admin  *serverConfig
	}

	t.Setenv("APP_NAME", "api")
	t.Setenv("PORT", "9090")
	t.Setenv("HOST", "0.0.0.0")

	var cfg appConfig
	if err := Load(&cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "api" {
		t.Fatalf("expected name api, got %q", cfg.Name)
	}
	if cfg.Server.Host != "0.0.0.0" || cfg.Server.Port != 9090 {
		t.Fatalf("expected nested server config to load, got %+v", cfg.Server)
	}
	if cfg.Admin == nil {
		t.Fatal("expected nested pointer struct to be allocated")
	}
	if cfg.Admin.Host != "0.0.0.0" || cfg.Admin.Port != 9090 {
		t.Fatalf("expected nested admin config to load, got %+v", cfg.Admin)
	}
}

func TestLoadAggregatesFieldErrors(t *testing.T) {
	type credentialsConfig struct {
		Token string `env:"TOKEN,required"`
	}
	type appConfig struct {
		Port        int      `env:"PORT"`
		BaseURL     *url.URL `env:"BASE_URL,required"`
		Credentials credentialsConfig
		Labels      map[string]string `env:"LABELS"`
	}

	t.Setenv("PORT", "not-a-number")
	t.Setenv("LABELS", "broken")

	var cfg appConfig
	err := Load(&cfg)
	if err == nil {
		t.Fatal("expected aggregated error")
	}

	got := err.Error()
	for _, want := range []string{
		`field Port: env "PORT" value "not-a-number"`,
		`field BaseURL: environment variable "BASE_URL" is required`,
		`field Credentials.Token: environment variable "TOKEN" is required`,
		`field Labels: env "LABELS" value "broken"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected aggregated error to contain %q, got %q", want, got)
		}
	}
}
