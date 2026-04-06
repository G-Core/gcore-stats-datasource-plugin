package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type HeaderSetter func(*http.Request)

var ErrResponseTooLarge = errors.New("http response body too large")

var DefaultMaxResponseBodyBytes int64 = 10 << 20 // 10 MiB

// ApplyJSONAuthHeaders applies the standard headers used by this plugin for JSON API calls.
// If apiKey is prefixed with "APIKey " or "Bearer ", it will be used as-is; otherwise it is treated as a raw key.
func ApplyJSONAuthHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Content-Type", "application/json")
	if apiKey == "" {
		return
	}
	if strings.HasPrefix(apiKey, "APIKey ") || strings.HasPrefix(apiKey, "Bearer ") {
		req.Header.Set("Authorization", apiKey)
		return
	}
	req.Header.Set("Authorization", "APIKey "+apiKey)
}

// DoRequest executes the request, reads the full body, and always closes the response body.
func DoRequest(client *http.Client, req *http.Request, setHeaders HeaderSetter) (int, []byte, error) {
	if setHeaders != nil {
		setHeaders(req)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, DefaultMaxResponseBodyBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	if int64(len(raw)) > DefaultMaxResponseBodyBytes {
		return resp.StatusCode, nil, ErrResponseTooLarge
	}

	return resp.StatusCode, raw, nil
}

func NewJSONRequest(
	ctx context.Context,
	method string,
	url string,
	body any,
) (*http.Request, error) {
	if body == nil {
		return http.NewRequestWithContext(ctx, method, url, nil)
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(b))
}

func statusAccepted(status int, accepted []int) bool {
	for _, a := range accepted {
		if status == a {
			return true
		}
	}
	return false
}

// DoJSONRequest executes an already-built request and unmarshals into out when status is accepted.
// If out is nil, it only validates status and returns raw body errors consistently.
func DoJSONRequest(
	client *http.Client,
	req *http.Request,
	out any,
	setHeaders HeaderSetter,
	handleStatusError func(status int, raw []byte) error,
	accepted ...int,
) ([]byte, error) {
	status, raw, err := DoRequest(client, req, setHeaders)
	if err != nil {
		return nil, err
	}

	if len(accepted) == 0 {
		accepted = []int{http.StatusOK}
	}

	if !statusAccepted(status, accepted) {
		if handleStatusError != nil {
			if err := handleStatusError(status, raw); err != nil {
				return raw, err
			}
		}
		return raw, nil
	}

	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return raw, err
		}
	}

	return raw, nil
}

// DoJSON executes a JSON request and unmarshals into out when status is accepted.
// If out is nil, it only validates status and returns raw body errors consistently.
func DoJSON(
	ctx context.Context,
	client *http.Client,
	method string,
	url string,
	body any,
	out any,
	setHeaders HeaderSetter,
	handleStatusError func(status int, raw []byte) error,
	accepted ...int,
) ([]byte, error) {
	req, err := NewJSONRequest(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	return DoJSONRequest(client, req, out, setHeaders, handleStatusError, accepted...)
}

