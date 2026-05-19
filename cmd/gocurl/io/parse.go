package io

import (
	"strconv"
	"strings"

	coretypes "github.com/drlinggg/gocurl/core/types"
)

func ParseArgs(args []string) (*coretypes.Request, error) {
	var httpReq coretypes.HTTPRequest
	var pretty, verbose bool

	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "GET" || args[i] == "POST" || args[i] == "PUT" ||
			args[i] == "PATCH" || args[i] == "DELETE" || args[i] == "OPTIONS" || args[i] == "HEAD":
			methods := map[string]coretypes.Method{
				"GET":     coretypes.MethodGet{},
				"POST":    coretypes.MethodPost{},
				"PUT":     coretypes.MethodPut{},
				"PATCH":   coretypes.MethodPatch{},
				"DELETE":  coretypes.MethodDelete{},
				"OPTIONS": coretypes.MethodOptions{},
				"HEAD":    coretypes.MethodHead{},
			}
			httpReq.Method = methods[args[i]]

		case args[i] == "--http1":
			httpReq.HTTP = coretypes.Http1{}
		case args[i] == "--http2":
			httpReq.HTTP = coretypes.Http2{}
		case args[i] == "--http3":
			httpReq.HTTP = coretypes.Http3{}
		case args[i] == "--pretty":
			pretty = true
		case args[i] == "-v":
			verbose = true
		case args[i] == "--scheme":
			i++
			if args[i] == "http" {
				httpReq.Scheme = coretypes.SchemeHTTP{}
			} else {
				httpReq.Scheme = coretypes.SchemeHTTPS{}
			}
		case args[i] == "-t":
			i++
			httpReq.Timeout, _ = strconv.Atoi(args[i])
		case args[i] == "-o":
			i++

		case strings.HasPrefix(args[i], "@"):
			key, val, _ := strings.Cut(args[i][1:], "=")
			httpReq.Headers = append(httpReq.Headers, coretypes.Field{
				Key:   key,
				Value: coretypes.StringValue{Val: val},
			})

		case strings.HasPrefix(args[i], "?"):
			key, val, _ := strings.Cut(args[i][1:], "=")
			httpReq.Query = append(httpReq.Query, coretypes.Field{
				Key:   key,
				Value: coretypes.StringValue{Val: val},
			})

		case strings.Contains(args[i], ":="):
			key, val, _ := strings.Cut(args[i], ":=")
			httpReq.Body = append(httpReq.Body, coretypes.Field{
				Key:   key,
				Value: coretypes.StringValue{Val: val},
			})

		case strings.Contains(args[i], "="):
			key, val, _ := strings.Cut(args[i], "=")
			httpReq.Body = append(httpReq.Body, coretypes.Field{
				Key:   key,
				Value: coretypes.StringValue{Val: val},
			})

		default:
			if httpReq.URL == "" {
				httpReq.URL = args[i]
			} else {
				httpReq.URL += args[i]
			}
		}
	}

	if httpReq.Method == nil {
		if len(httpReq.Body) > 0 {
			httpReq.Method = coretypes.MethodPost{}
		} else {
			httpReq.Method = coretypes.MethodGet{}
		}
	}

	return &coretypes.Request{
		HTTPRequest: httpReq,
		Pretty:      pretty,
		Verbose:     verbose,
	}, nil
}
