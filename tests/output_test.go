package tests

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	cmdio "github.com/drlinggg/gocurl/cmd/gocurl/io"
	coretypes "github.com/drlinggg/gocurl/core/types"
)

func makeResp(code int, body []byte, headers []coretypes.Field) *coretypes.Response {
	status, _ := coretypes.NewStatus(code)
	return &coretypes.Response{
		HTTPResponse: coretypes.HTTPResponse{
			Status:  status,
			Headers: headers,
			Body:    body,
		},
		Elapsed: 42 * time.Millisecond,
	}
}

func makeReq(pretty, verbose bool) *coretypes.Request {
	return &coretypes.Request{Pretty: pretty, Verbose: verbose}
}

func TestOutput_Status(t *testing.T) {
	for _, code := range []int{200, 404, 500} {
		var buf bytes.Buffer
		out := cmdio.NewOutputWriter(&buf, coretypes.Colors{})
		out.Write(makeReq(false, false), makeResp(code, nil, nil))
		got := buf.String()
		if !strings.Contains(got, "STATUS:") {
			t.Errorf("code %d: missing STATUS: label", code)
		}
		if !strings.Contains(got, fmt.Sprintf("%d", code)) {
			t.Errorf("code %d: missing status code in output", code)
		}
	}
}

func TestOutput_Elapsed(t *testing.T) {
	var buf bytes.Buffer
	out := cmdio.NewOutputWriter(&buf, coretypes.Colors{})
	out.Write(makeReq(false, false), makeResp(200, nil, nil))
	got := buf.String()
	if !strings.Contains(got, "TIME:") {
		t.Error("missing TIME: label")
	}
	if !strings.Contains(got, "42ms") {
		t.Errorf("missing elapsed, got %q", got)
	}
}

func TestOutput_BodyRaw(t *testing.T) {
	body := []byte(`{"name":"alex"}`)
	var buf bytes.Buffer
	out := cmdio.NewOutputWriter(&buf, coretypes.Colors{})
	out.Write(makeReq(false, false), makeResp(200, body, nil))
	got := buf.String()
	if !strings.Contains(got, "RESPONSE:") {
		t.Error("missing RESPONSE: label")
	}
	if !strings.Contains(got, `{"name":"alex"}`) {
		t.Errorf("missing raw body, got %q", got)
	}
}

func TestOutput_BodyPretty(t *testing.T) {
	body := []byte(`{"name":"alex"}`)
	var buf bytes.Buffer
	out := cmdio.NewOutputWriter(&buf, coretypes.Colors{})
	out.Write(makeReq(true, false), makeResp(200, body, nil))
	if !strings.Contains(buf.String(), "\t\"name\"") {
		t.Errorf("expected indented body, got %q", buf.String())
	}
}

func TestOutput_BodyEmpty(t *testing.T) {
	var buf bytes.Buffer
	out := cmdio.NewOutputWriter(&buf, coretypes.Colors{})
	out.Write(makeReq(false, false), makeResp(204, nil, nil))
	if strings.Contains(buf.String(), "RESPONSE:") {
		t.Error("expected no RESPONSE section for empty body")
	}
}

func TestOutput_HeadersVerbose(t *testing.T) {
	headers := []coretypes.Field{
		{Key: "Content-Type", Value: coretypes.StringValue{Val: "application/json"}},
	}
	var buf bytes.Buffer
	out := cmdio.NewOutputWriter(&buf, coretypes.Colors{})
	out.Write(makeReq(false, true), makeResp(200, nil, headers))
	got := buf.String()
	if !strings.Contains(got, "HEADERS:") {
		t.Error("missing HEADERS: label")
	}
	if !strings.Contains(got, "Content-Type: application/json") {
		t.Errorf("missing header, got %q", got)
	}
}

func TestOutput_BodyNoTrailingNewline(t *testing.T) {
	body := []byte("{\"name\":\"alex\"}\n")
	var buf bytes.Buffer
	out := cmdio.NewOutputWriter(&buf, coretypes.Colors{})
	out.Write(makeReq(false, false), makeResp(200, body, nil))
	got := buf.String()
	if strings.Contains(got, "\n\n\n") {
		t.Errorf("unexpected triple newline in output: %q", got)
	}
}

func TestOutput_HeadersHidden(t *testing.T) {
	headers := []coretypes.Field{
		{Key: "Content-Type", Value: coretypes.StringValue{Val: "application/json"}},
	}
	var buf bytes.Buffer
	out := cmdio.NewOutputWriter(&buf, coretypes.Colors{})
	out.Write(makeReq(false, false), makeResp(200, nil, headers))
	if strings.Contains(buf.String(), "Content-Type") {
		t.Errorf("expected headers hidden, got %q", buf.String())
	}
}
