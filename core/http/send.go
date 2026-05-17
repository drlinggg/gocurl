package http

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	coretypes "github.com/banakh/gocurl/core/types"
)

type Sender struct{}

func (s *Sender) Send(req *coretypes.HTTPRequest) (*coretypes.Response, error) {
	rawURL, err := buildURL(req)
	if err != nil {
		return nil, err
	}

	body, err := buildBody(req.Body)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(fmt.Sprintf("%s", req.Method), rawURL, body)
	if err != nil {
		return nil, err
	}

	for _, h := range req.Headers {
		if sv, ok := h.Value.(coretypes.StringValue); ok {
			httpReq.Header.Set(h.Key, sv.Val)
		}
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{
		Timeout:   clientTimeout(req.Timeout),
		Transport: buildTransport(req),
	}

	start := time.Now()
	resp, err := client.Do(httpReq)
	elapsed := time.Since(start)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	status, err := coretypes.NewStatus(resp.StatusCode)
	if err != nil {
		return nil, err
	}

	headers := make([]coretypes.Field, 0, len(resp.Header))
	for k, vs := range resp.Header {
		for _, v := range vs {
			headers = append(headers, coretypes.Field{
				Key:   k,
				Value: coretypes.StringValue{Val: v},
			})
		}
	}

	return &coretypes.Response{
		HTTPResponse: coretypes.HTTPResponse{
			Status:  status,
			Headers: headers,
			Body:    respBody,
		},
		Elapsed: elapsed,
	}, nil
}

// buildURL склеивает scheme + host + query params.
// Если URL уже содержит схему (пользователь написал https://...) — не добавляем.
func buildURL(req *coretypes.HTTPRequest) (string, error) {
	scheme := "https"
	if _, ok := req.Scheme.(coretypes.SchemeHTTP); ok {
		scheme = "http"
	}

	raw := req.URL
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = scheme + "://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	if len(req.Query) > 0 {
		q := u.Query()
		for _, f := range req.Query {
			if sv, ok := f.Value.(coretypes.StringValue); ok {
				q.Set(f.Key, sv.Val)
			}
		}
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

// buildBody превращает []Field в JSON-тело запроса.
// Возвращает nil если полей нет — тогда Content-Type не ставим.
func buildBody(fields []coretypes.Field) (io.Reader, error) {
	if len(fields) == 0 {
		return nil, nil
	}
	m := make(map[string]any, len(fields))
	for _, f := range fields {
		switch v := f.Value.(type) {
		case coretypes.StringValue:
			m[f.Key] = v.Val
		case coretypes.FileValue:
			m[f.Key] = v.Path
		}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// buildTransport возвращает транспорт под нужную версию HTTP.
// HTTP/1.1: пустой TLSNextProto отключает автоапгрейд до HTTP/2 через ALPN.
// HTTP/2: дефолтный транспорт — Go сам договорится до HTTP/2 если сервер поддерживает.
// HTTP/3: пока не реализован.
func buildTransport(req *coretypes.HTTPRequest) http.RoundTripper {
	switch req.HTTP.(type) {
	case coretypes.Http1:
		return &http.Transport{
			TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
		}
	default:
		return http.DefaultTransport
	}
}

func clientTimeout(seconds int) time.Duration {
	if seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return 30 * time.Second
}
