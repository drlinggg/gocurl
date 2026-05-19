package io

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	coreio "github.com/drlinggg/gocurl/core/io"
	coretypes "github.com/drlinggg/gocurl/core/types"
	"golang.org/x/term"
)

var _ coreio.Output = (*cmdOutput)(nil)

type cmdOutput struct {
	w      io.Writer
	colors coretypes.Colors
	isTTY  bool
}

// NewOutputWriter is an alias for NewOutput, used in tests.
func NewOutputWriter(w io.Writer, colors coretypes.Colors) *cmdOutput {
	return NewOutput(w, colors)
}

func NewOutput(w io.Writer, colors coretypes.Colors) *cmdOutput {
	isTTY := false
	if f, ok := w.(*os.File); ok {
		isTTY = term.IsTerminal(int(f.Fd()))
	}
	return &cmdOutput{w: w, colors: colors, isTTY: isTTY}
}

func (o *cmdOutput) Write(req *coretypes.Request, resp *coretypes.Response) error {
	if err := o.writeBody(resp.Body, req.Pretty); err != nil {
		return err
	}
	o.writeMeta(resp.Status, resp.Elapsed)
	if req.Verbose && len(resp.Headers) > 0 {
		fmt.Fprintln(o.w)
		o.writeHeaders(resp.Headers)
	}
	return nil
}

func (o *cmdOutput) writeMeta(status coretypes.Status, elapsed time.Duration) {
	var parts []string
	if status.Code() > 0 {
		parts = append(parts, fmt.Sprintf("STATUS: %s", o.colorize(fmt.Sprintf("%d", status.Code()), o.pickStatusColor(status.Code()))))
	}
	if elapsed > 0 {
		parts = append(parts, fmt.Sprintf("TIME: %s", o.colorize(elapsed.String(), o.colors.Elapsed)))
	}
	if len(parts) > 0 {
		fmt.Fprintf(o.w, "%s\n", strings.Join(parts, "   "))
	}
}

func (o *cmdOutput) writeHeaders(headers []coretypes.Field) {
	fmt.Fprintln(o.w, "HEADERS:")
	for _, h := range headers {
		line := fmt.Sprintf("%s: %s", h.Key, h.Value.(coretypes.StringValue).Val)
		fmt.Fprintf(o.w, "  %s\n", o.colorize(line, o.colors.Headers))
	}
}

func (o *cmdOutput) writeBody(body []byte, pretty bool) error {
	if len(body) == 0 {
		return nil
	}
	out := body
	if pretty {
		var buf bytes.Buffer
		if err := json.Indent(&buf, body, "  ", "\t"); err != nil {
			return err
		}
		out = buf.Bytes()
	}
	out = bytes.TrimRight(out, "\n")
	fmt.Fprintf(o.w, "RESPONSE:\n  %s\n", o.colorize(string(out), o.colors.Body))
	return nil
}

func (o *cmdOutput) pickStatusColor(code int) coretypes.Color {
	switch {
	case code >= 200 && code < 300:
		return o.colors.Status2xx
	case code >= 400 && code < 500:
		return o.colors.Status4xx
	default:
		return o.colors.Status5xx
	}
}

func (o *cmdOutput) colorize(s string, c coretypes.Color) string {
	if !o.isTTY || c == (coretypes.Color{}) {
		return s
	}
	return c.ANSI() + s + coretypes.Reset
}
