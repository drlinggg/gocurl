package tests

import (
	"testing"

	cmdio "github.com/banakh/gocurl/cmd/gocurl/io"
	coretypes "github.com/banakh/gocurl/core/types"
)

func TestParseArgs_ExplicitMethod(t *testing.T) {
	req, err := cmdio.ParseArgs([]string{"GET", "api.example.com/users"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := req.Method.(coretypes.MethodGet); !ok {
		t.Errorf("expected MethodGet, got %T", req.Method)
	}
	if req.URL != "api.example.com/users" {
		t.Errorf("unexpected URL: %s", req.URL)
	}
}

func TestParseArgs_AutoMethodGet(t *testing.T) {
	req, _ := cmdio.ParseArgs([]string{"api.example.com"})
	if _, ok := req.Method.(coretypes.MethodGet); !ok {
		t.Errorf("expected MethodGet, got %T", req.Method)
	}
}

func TestParseArgs_AutoMethodPost(t *testing.T) {
	req, _ := cmdio.ParseArgs([]string{"api.example.com", "name=alex"})
	if _, ok := req.Method.(coretypes.MethodPost); !ok {
		t.Errorf("expected MethodPost, got %T", req.Method)
	}
}

func TestParseArgs_BodyFields(t *testing.T) {
	req, _ := cmdio.ParseArgs([]string{"api.example.com", "name=alex", "age:=30"})
	if len(req.Body) != 2 {
		t.Fatalf("expected 2 body fields, got %d", len(req.Body))
	}
	if req.Body[0].Key != "name" || req.Body[0].Value.(coretypes.StringValue).Val != "alex" {
		t.Errorf("unexpected body[0]: %+v", req.Body[0])
	}
	if req.Body[1].Key != "age" || req.Body[1].Value.(coretypes.StringValue).Val != "30" {
		t.Errorf("unexpected body[1]: %+v", req.Body[1])
	}
}

func TestParseArgs_Header(t *testing.T) {
	req, _ := cmdio.ParseArgs([]string{"api.example.com", "@Authorization=Bearer token"})
	if len(req.Headers) != 1 {
		t.Fatalf("expected 1 header, got %d", len(req.Headers))
	}
	if req.Headers[0].Key != "Authorization" {
		t.Errorf("unexpected header key: %s", req.Headers[0].Key)
	}
	if req.Headers[0].Value.(coretypes.StringValue).Val != "Bearer token" {
		t.Errorf("unexpected header value: %+v", req.Headers[0].Value)
	}
}

func TestParseArgs_Query(t *testing.T) {
	req, _ := cmdio.ParseArgs([]string{"api.example.com", "?page=2", "?limit=10"})
	if len(req.Query) != 2 {
		t.Fatalf("expected 2 query params, got %d", len(req.Query))
	}
	if req.Query[0].Key != "page" || req.Query[0].Value.(coretypes.StringValue).Val != "2" {
		t.Errorf("unexpected query[0]: %+v", req.Query[0])
	}
}

func TestParseArgs_Flags(t *testing.T) {
	req, _ := cmdio.ParseArgs([]string{"--http2", "--pretty", "-v", "-t", "15", "api.example.com"})
	if _, ok := req.HTTP.(coretypes.Http2); !ok {
		t.Errorf("expected Http2, got %T", req.HTTP)
	}
	if !req.Pretty {
		t.Error("expected Pretty=true")
	}
	if !req.Verbose {
		t.Error("expected Verbose=true")
	}
	if req.Timeout != 15 {
		t.Errorf("expected Timeout=15, got %d", req.Timeout)
	}
}

func TestParseArgs_Scheme(t *testing.T) {
	req, _ := cmdio.ParseArgs([]string{"--scheme", "http", "localhost:8080"})
	if _, ok := req.Scheme.(coretypes.SchemeHTTP); !ok {
		t.Errorf("expected SchemeHTTP, got %T", req.Scheme)
	}
}
