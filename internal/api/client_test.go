package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Client header tests ---

func TestClient_Headers(t *testing.T) {
	var gotUA, gotAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotAPIKey = r.Header.Get("X-API-KEY")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "secret-key")
	resp, err := c.Get(context.Background(), "/test")
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "memelink-cli/test", gotUA)
	assert.Equal(t, "secret-key", gotAPIKey)
}

func TestClient_NoAPIKey(t *testing.T) {
	var gotAPIKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("X-API-KEY")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.Get(context.Background(), "/test")
	require.NoError(t, err)
	resp.Body.Close()

	assert.Empty(t, gotAPIKey)
}

func TestClient_PostContentType(t *testing.T) {
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.Post(context.Background(), "/test", strings.NewReader(`{"key":"value"}`))
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "application/json", gotCT)
}

// --- Retry transport tests ---

func TestRetryTransport_Success(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.Get(context.Background(), "/ok")
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, int32(1), callCount.Load())
}

func TestRetryTransport_RetryOn5xx(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := callCount.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.Get(context.Background(), "/retry")
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), callCount.Load())
}

func TestRetryTransport_RetryOn429(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := callCount.Add(1)
		if n < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.Get(context.Background(), "/rate-limit")
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), callCount.Load())
}

func TestRetryTransport_NoRetryOn4xx(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.Get(context.Background(), "/bad")
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, int32(1), callCount.Load())
}

func TestRetryTransport_MaxRetries(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.Get(context.Background(), "/always-fail")
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	// initial + 3 retries = 4 calls
	assert.Equal(t, int32(4), callCount.Load())
}

func TestRetryTransport_ContextCancellation(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after first response so retry sleep gets interrupted.
	c := &Client{
		http: &http.Client{
			Transport: &retryTransport{
				base:       http.DefaultTransport,
				maxRetries: 3,
				baseDelay:  500 * time.Millisecond,
			},
		},
		baseURL:   srv.URL,
		userAgent: "memelink-cli/test",
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	resp, err := c.Get(ctx, "/cancel")
	if resp != nil {
		resp.Body.Close()
	}
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// --- Error detection tests ---

func TestCheckImageResponse_Success(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusOK}
	assert.NoError(t, checkImageResponse(resp))
}

func TestCheckImageResponse_StatusCodes(t *testing.T) {
	tests := []struct {
		code    int
		wantMsg string
	}{
		{404, "template not found"},
		{414, "text too long (max 200 chars per line)"},
		{415, "could not download image URL"},
		{422, "invalid style or missing image URL"},
		{429, "rate limited, try again later"},
		{500, "unexpected error (HTTP 500)"},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.code), func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.code}
			err := checkImageResponse(resp)
			require.Error(t, err)

			var apiErr *Error
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, tt.code, apiErr.StatusCode)
			assert.Equal(t, tt.wantMsg, apiErr.Message)
		})
	}
}

func TestCheckJSONResponse_Success(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"data":"ok"}`)),
	}
	assert.NoError(t, checkJSONResponse(resp))
}

func TestCheckJSONResponse_ErrorBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"error":"template 'xyz' not found"}`)),
	}

	err := checkJSONResponse(resp)
	require.Error(t, err)

	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 404, apiErr.StatusCode)
	assert.Equal(t, "template 'xyz' not found", apiErr.Message)
}

func TestCheckJSONResponse_FallbackToStatusMessage(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`not json`)),
	}

	err := checkJSONResponse(resp)
	require.Error(t, err)

	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "template not found", apiErr.Message)
}

func TestCheckJSONResponse_EmptyErrorField(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusUnprocessableEntity,
		Body:       io.NopCloser(strings.NewReader(`{"error":""}`)),
	}

	err := checkJSONResponse(resp)
	require.Error(t, err)

	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "invalid style or missing image URL", apiErr.Message)
}

// --- Context round-trip tests ---

func TestWithClient_RoundTrip(t *testing.T) {
	c := NewClient(ClientOptions{})
	ctx := context.Background()
	ctx = WithClient(ctx, c)

	got := ClientFromContext(ctx)
	assert.Same(t, c, got)
}

func TestClientFromContext_Nil(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ClientFromContext(ctx))
}

// --- ShouldRetry tests ---

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{200, false},
		{301, false},
		{400, false},
		{401, false},
		{404, false},
		{429, true},
		{500, true},
		{501, true},
		{502, true},
		{503, true},
		{504, true},
		{505, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, shouldRetry(tt.code), "status %d", tt.code)
	}
}

// --- Post body preservation on retry ---

func TestRetryTransport_PostBodyPreserved(t *testing.T) {
	var bodies []string
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		n := callCount.Add(1)
		if n < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	payload := `{"text":"hello"}`
	resp, err := c.Post(context.Background(), "/post", strings.NewReader(payload))
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, int32(2), callCount.Load())
	// Body should be identical on both attempts.
	require.Len(t, bodies, 2)
	assert.Equal(t, payload, bodies[0])
	assert.Equal(t, payload, bodies[1])
}

// --- Logging transport test ---

func TestLoggingTransport_WrapsBase(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Build client with verbose to exercise the logging path.
	c := &Client{
		http: &http.Client{
			Transport: &loggingTransport{
				base: &retryTransport{
					base:       http.DefaultTransport,
					maxRetries: 0,
					baseDelay:  1 * time.Millisecond,
				},
			},
		},
		baseURL:   srv.URL,
		userAgent: "memelink-cli/test",
	}

	resp, err := c.Get(context.Background(), "/log")
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// --- helpers ---

func newTestClient(baseURL, apiKey string) *Client {
	return &Client{
		http: &http.Client{
			Transport: &retryTransport{
				base:       http.DefaultTransport,
				maxRetries: 3,
				baseDelay:  1 * time.Millisecond,
			},
		},
		baseURL:   baseURL,
		apiKey:    apiKey,
		userAgent: "memelink-cli/test",
	}
}

func TestNewClient_DefaultBaseURL(t *testing.T) {
	c := NewClient(ClientOptions{})
	assert.Equal(t, DefaultBaseURL, c.baseURL)
}

func TestNewClient_CustomBaseURL(t *testing.T) {
	c := NewClient(ClientOptions{BaseURL: "https://custom.example.com"})
	assert.Equal(t, "https://custom.example.com", c.baseURL)
}

func TestNewClient_DefaultUserAgent(t *testing.T) {
	c := NewClient(ClientOptions{})
	assert.Equal(t, "memelink-cli/dev", c.userAgent)
}

func TestError_Error(t *testing.T) {
	e := &Error{StatusCode: 404, Message: "template not found"}
	assert.Equal(t, "memegen api: template not found (HTTP 404)", e.Error())
}

// --- Post with GetBody for retry ---

func TestPost_SetsGetBody(t *testing.T) {
	// Verify that Post requests using strings.NewReader work
	// because http.NewRequest auto-sets GetBody for *strings.Reader.
	var buf bytes.Buffer
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(&buf, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.Post(context.Background(), "/body", strings.NewReader("payload"))
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "payload", buf.String())
}
