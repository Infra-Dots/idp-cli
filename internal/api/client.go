package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the InfraDots API HTTP client.
type Client struct {
	Host   string
	Token  string
	http   *http.Client
}

// NewClient creates a Client with a 30s timeout.
func NewClient(host, token string) *Client {
	return &Client{
		Host:  strings.TrimRight(host, "/"),
		Token: token,
		http:  &http.Client{Timeout: 30 * time.Second},
	}
}

// APIError represents a non-2xx response from the API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}


func (c *Client) url(path string) string {
	return c.Host + path
}

func (c *Client) do(method, path string, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.url(path), bodyReader)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return c.parseError(resp.StatusCode, respBody)
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

func (c *Client) parseError(status int, body []byte) error {
	// Try to extract a "detail" or "error" field from the JSON response.
	var payload map[string]any
	msg := http.StatusText(status)
	if json.Unmarshal(body, &payload) == nil {
		for _, key := range []string{"detail", "error", "message"} {
			if v, ok := payload[key]; ok {
				msg = fmt.Sprintf("%v", v)
				break
			}
		}
	}

	switch status {
	case http.StatusUnauthorized:
		msg = "authentication failed — run `idp auth login`"
	case http.StatusForbidden:
		msg = "permission denied"
	case http.StatusNotFound:
		msg = "resource not found"
	}

	return &APIError{StatusCode: status, Message: msg}
}

func (c *Client) Get(path string, out any) error {
	return c.do(http.MethodGet, path, nil, out)
}

func (c *Client) Post(path string, body, out any) error {
	return c.do(http.MethodPost, path, body, out)
}

func (c *Client) Patch(path string, body, out any) error {
	return c.do(http.MethodPatch, path, body, out)
}

func (c *Client) Delete(path string) error {
	return c.do(http.MethodDelete, path, nil, nil)
}
