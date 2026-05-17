package tests

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	coretypes "github.com/banakh/gocurl/core/types"
	"github.com/banakh/gocurl/storage"
)

const testTOML = `
[default]
timeout      = 10
http_version = 1
scheme       = "https"

[default.colors]
status_2xx = "00c853"
status_4xx = "ffab00"
status_5xx = "ff1744"
headers    = "888888"
body       = "eeeeee"
elapsed    = "555555"

[default.headers]
X-Default = "yes"

[github]
base         = "https://api.github.com"
http_version = 2

[github.headers]
Accept = "application/vnd.github+json"
`

func writeTempTOML(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "presets-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(testTOML)
	f.Close()
	return f.Name()
}

func loadTestConfig(t *testing.T) *storage.Config {
	t.Helper()
	t.Setenv("GOCURL_CONFIG", writeTempTOML(t))
	cfg, err := storage.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestConfig_DefaultTimeout(t *testing.T) {
	cfg := loadTestConfig(t)
	req := &coretypes.Request{HTTPRequest: coretypes.HTTPRequest{URL: "example.com"}}
	out, err := cfg.FillRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if out.Timeout != 10 {
		t.Errorf("expected Timeout=10, got %d", out.Timeout)
	}
}

func TestConfig_DefaultScheme(t *testing.T) {
	cfg := loadTestConfig(t)
	req := &coretypes.Request{HTTPRequest: coretypes.HTTPRequest{URL: "example.com"}}
	out, _ := cfg.FillRequest(req)
	if _, ok := out.Scheme.(coretypes.SchemeHTTPS); !ok {
		t.Errorf("expected SchemeHTTPS, got %T", out.Scheme)
	}
}

func TestConfig_DefaultHTTP(t *testing.T) {
	cfg := loadTestConfig(t)
	req := &coretypes.Request{HTTPRequest: coretypes.HTTPRequest{URL: "example.com"}}
	out, _ := cfg.FillRequest(req)
	if _, ok := out.HTTP.(coretypes.Http1); !ok {
		t.Errorf("expected Http1, got %T", out.HTTP)
	}
}

func TestConfig_PresetMatch(t *testing.T) {
	cfg := loadTestConfig(t)
	req := &coretypes.Request{HTTPRequest: coretypes.HTTPRequest{URL: "https://api.github.com/repos"}}
	out, _ := cfg.FillRequest(req)
	if _, ok := out.HTTP.(coretypes.Http2); !ok {
		t.Errorf("expected Http2 from github preset, got %T", out.HTTP)
	}
}

func TestConfig_CLIOverridesPreset(t *testing.T) {
	cfg := loadTestConfig(t)
	req := &coretypes.Request{HTTPRequest: coretypes.HTTPRequest{
		URL:     "example.com",
		Timeout: 99,
		Scheme:  coretypes.SchemeHTTP{},
	}}
	out, _ := cfg.FillRequest(req)
	if out.Timeout != 99 {
		t.Errorf("expected CLI timeout=99, got %d", out.Timeout)
	}
	if _, ok := out.Scheme.(coretypes.SchemeHTTP); !ok {
		t.Errorf("expected CLI SchemeHTTP, got %T", out.Scheme)
	}
}

func TestConfig_Colors(t *testing.T) {
	cfg := loadTestConfig(t)
	colors := cfg.Colors()
	if colors.Status2xx == (coretypes.Color{}) {
		t.Error("expected Status2xx color to be set")
	}
}

func TestConfig_MissingFile(t *testing.T) {
	t.Setenv("GOCURL_CONFIG", filepath.Join(t.TempDir(), "nonexistent.toml"))
	cfg, err := storage.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if cfg == nil {
		t.Error("expected non-nil config")
	}
}

func TestHistory_SaveCreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	t.Setenv("GOCURL_DATA", path)
	h, err := storage.LoadHistory()
	if err != nil {
		t.Fatal(err)
	}
	req := &coretypes.Request{HTTPRequest: coretypes.HTTPRequest{Method: coretypes.MethodGet{}, URL: "example.com"}}
	status, _ := coretypes.NewStatus(200)
	resp := &coretypes.Response{HTTPResponse: coretypes.HTTPResponse{Status: status}, Elapsed: time.Millisecond}
	h.Save(req, resp)
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected history file to exist: %v", err)
	}
}

func TestHistory_SaveWritesRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	t.Setenv("GOCURL_DATA", path)
	h, _ := storage.LoadHistory()

	req := &coretypes.Request{HTTPRequest: coretypes.HTTPRequest{Method: coretypes.MethodPost{}, URL: "api.example.com"}}
	status, _ := coretypes.NewStatus(201)
	resp := &coretypes.Response{HTTPResponse: coretypes.HTTPResponse{Status: status}, Elapsed: 42 * time.Millisecond}
	h.Save(req, resp)

	f, _ := os.Open(path)
	defer f.Close()
	var r map[string]any
	json.NewDecoder(f).Decode(&r)

	if r["method"] != "POST" {
		t.Errorf("expected method=POST, got %v", r["method"])
	}
	if r["url"] != "api.example.com" {
		t.Errorf("expected url=api.example.com, got %v", r["url"])
	}
	if r["status"] != float64(201) {
		t.Errorf("expected status=201, got %v", r["status"])
	}
}

func TestHistory_SaveAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	t.Setenv("GOCURL_DATA", path)
	h, _ := storage.LoadHistory()

	status, _ := coretypes.NewStatus(200)
	resp := &coretypes.Response{HTTPResponse: coretypes.HTTPResponse{Status: status}, Elapsed: time.Millisecond}

	h.Save(&coretypes.Request{HTTPRequest: coretypes.HTTPRequest{Method: coretypes.MethodGet{}, URL: "a.com"}}, resp)
	h.Save(&coretypes.Request{HTTPRequest: coretypes.HTTPRequest{Method: coretypes.MethodGet{}, URL: "b.com"}}, resp)

	f, _ := os.Open(path)
	defer f.Close()
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Text() != "" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 records, got %d", count)
	}
}
