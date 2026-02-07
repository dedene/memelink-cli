package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// retryTransport wraps an http.RoundTripper with automatic retry on transient errors.
type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
	baseDelay  time.Duration
}

// RoundTrip implements http.RoundTripper with retry logic for 429 and 5xx responses.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := range t.maxRetries + 1 {
		// Clone body for retry (body is consumed on read).
		if req.Body != nil && req.GetBody != nil {
			body, bodyErr := req.GetBody()
			if bodyErr != nil {
				return nil, fmt.Errorf("cloning request body: %w", bodyErr)
			}

			req.Body = body
		}

		resp, err = t.base.RoundTrip(req)
		if err != nil {
			return nil, fmt.Errorf("round trip: %w", err)
		}

		if !shouldRetry(resp.StatusCode) {
			return resp, nil
		}

		if attempt < t.maxRetries {
			// Close response body before retry to prevent connection leak.
			_ = resp.Body.Close()

			delay := t.baseDelay * (1 << attempt) //nolint:gosec // attempt is bounded by maxRetries (small int)

			select {
			case <-req.Context().Done():
				return nil, fmt.Errorf("retry wait: %w", req.Context().Err())
			case <-time.After(delay):
			}
		}
	}

	return resp, nil
}

// shouldRetry returns true for status codes that warrant a retry.
func shouldRetry(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests ||
		(statusCode >= http.StatusInternalServerError && statusCode <= http.StatusGatewayTimeout)
}

// loggingTransport wraps an http.RoundTripper with slog debug logging.
type loggingTransport struct {
	base http.RoundTripper
}

// RoundTrip implements http.RoundTripper with request/response logging.
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	slog.Debug("http request", "method", req.Method, "url", req.URL.String())

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		slog.Debug("http error",
			"method", req.Method,
			"url", req.URL.String(),
			"error", err,
			"duration", time.Since(start),
		)

		return nil, fmt.Errorf("logging round trip: %w", err)
	}

	slog.Debug("http response",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.StatusCode,
		"duration", time.Since(start),
	)

	return resp, nil
}
