package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	coretypes "github.com/banakh/gocurl/core/types"
)

func defaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gocurl"), nil
}

type record struct {
	Timestamp time.Time `json:"timestamp"`
	Method    string    `json:"method"`
	URL       string    `json:"url"`
	Status    int       `json:"status"`
	Elapsed   string    `json:"elapsed"`
}

type History struct {
	path string
}

func LoadHistory() (*History, error) {
	path := os.Getenv("GOCURL_DATA")
	if path == "" {
		dir, err := defaultDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(dir, "history.jsonl")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return &History{path: path}, nil
}

func (h *History) Save(req *coretypes.Request, resp *coretypes.Response) error {
	if resp == nil {
		return nil
	}
	f, err := os.OpenFile(h.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	r := record{
		Timestamp: time.Now().UTC(),
		Method:    fmt.Sprintf("%s", req.Method),
		URL:       req.URL,
		Status:    resp.Status.Code(),
		Elapsed:   resp.Elapsed.String(),
	}
	return json.NewEncoder(f).Encode(r)
}
