package types

import (
	"encoding/hex"
	"fmt"
	"time"
)

type Color [3]byte

func (c Color) ANSI() string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", c[0], c[1], c[2])
}

const Reset = "\033[0m"

func ParseColor(s string) (Color, error) {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 3 {
		return Color{}, &ErrInvalidColor{Value: s}
	}
	return Color{b[0], b[1], b[2]}, nil
}

type Colors struct {
	Status2xx Color
	Status4xx Color
	Status5xx Color
	Headers   Color
	Body      Color
	Elapsed   Color
}

type Value interface {
	fieldValue()
}

type StringValue struct{ Val string }
type FileValue struct{ Path string }

func (StringValue) fieldValue() {}
func (FileValue) fieldValue()   {}

type Field struct {
	Key   string
	Value Value
}

type Method interface{ method() }

type MethodGet     struct{}
type MethodPost    struct{}
type MethodPut     struct{}
type MethodPatch   struct{}
type MethodDelete  struct{}
type MethodOptions struct{}
type MethodHead    struct{}

func (MethodGet) method()     {}
func (MethodPost) method()    {}
func (MethodPut) method()     {}
func (MethodPatch) method()   {}
func (MethodDelete) method()  {}
func (MethodOptions) method() {}
func (MethodHead) method()    {}

func (MethodGet) String() string     { return "GET" }
func (MethodPost) String() string    { return "POST" }
func (MethodPut) String() string     { return "PUT" }
func (MethodPatch) String() string   { return "PATCH" }
func (MethodDelete) String() string  { return "DELETE" }
func (MethodOptions) String() string { return "OPTIONS" }
func (MethodHead) String() string    { return "HEAD" }

type Http interface{ httpVersion() }

type Http1 struct{}
type Http2 struct{}
type Http3 struct{}

func (Http1) httpVersion() {}
func (Http2) httpVersion() {}
func (Http3) httpVersion() {}

type Scheme interface{ scheme() }

type SchemeHTTP  struct{}
type SchemeHTTPS struct{}

func (SchemeHTTP) scheme()  {}
func (SchemeHTTPS) scheme() {}

type Status struct{ val int }

func NewStatus(code int) (Status, error) {
	if code < 0 || code > 999 {
		return Status{}, &ErrStatusRange{Code: code}
	}
	return Status{code}, nil
}

func (s Status) Code() int { return s.val }

type HTTPResponse struct {
	Status  Status
	Headers []Field
	Body    []byte
}

type Response struct {
	HTTPResponse
	Elapsed time.Duration
}

type HTTPRequest struct {
	Method  Method
	URL     string
	Headers []Field
	Query   []Field
	Body    []Field
	HTTP    Http
	Scheme  Scheme
	Timeout int
}

type Request struct {
	HTTPRequest
	Pretty  bool
	Verbose bool
}

func (r HTTPRequest) Validate() error {
	if r.URL == "" {
		return &ErrValidation{Field: "url"}
	}
	if r.Scheme == nil {
		return &ErrValidation{Field: "scheme"}
	}
	if r.HTTP == nil {
		return &ErrValidation{Field: "http version"}
	}
	return nil
}
