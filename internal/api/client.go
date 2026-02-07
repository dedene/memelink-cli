// Package api provides an HTTP client for the Memegen.link API.
package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultBaseURL is the Memegen.link API base URL.
const DefaultBaseURL = "https://api.memegen.link"

// ClientOptions configures a new Client.
type ClientOptions struct {
	BaseURL   string
	APIKey    string
	Verbose   bool
	UserAgent string
}

// Client wraps an HTTP client for Memegen API calls.
type Client struct {
	http      *http.Client
	baseURL   string
	apiKey    string
	userAgent string
}

// NewClient builds a Client with retry transport and optional verbose logging.
func NewClient(opts ClientOptions) *Client {
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	ua := opts.UserAgent
	if ua == "" {
		ua = "memelink-cli/dev"
	}

	var transport http.RoundTripper = &retryTransport{
		base:       http.DefaultTransport,
		maxRetries: 3,
		baseDelay:  1 * time.Second,
	}

	if opts.Verbose {
		transport = &loggingTransport{base: transport}
	}

	return &Client{
		http: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		baseURL:   baseURL,
		apiKey:    opts.APIKey,
		userAgent: ua,
	}
}

// do executes an HTTP request with standard headers.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	if c.apiKey != "" {
		req.Header.Set("X-API-KEY", c.apiKey)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return resp, nil
}

// Get performs a GET request against the API.
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request against the API with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

type clientCtxKey struct{}

// WithClient stores a Client in the context.
func WithClient(ctx context.Context, cl *Client) context.Context {
	return context.WithValue(ctx, clientCtxKey{}, cl)
}

// ClientFromContext retrieves the Client from the context.
func ClientFromContext(ctx context.Context) *Client {
	if v := ctx.Value(clientCtxKey{}); v != nil {
		if cl, ok := v.(*Client); ok {
			return cl
		}
	}

	return nil
}
