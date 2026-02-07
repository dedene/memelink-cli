package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAutomatic_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/buzz/when_the_code_works.png","generator":"Pattern","confidence":0.46}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.GenerateAutomatic(context.Background(), AutomaticRequest{Text: "when the code works"})
	require.NoError(t, err)
	assert.Equal(t, "https://api.memegen.link/images/buzz/when_the_code_works.png", resp.URL)
	assert.Equal(t, "Pattern", resp.Generator)
	assert.InDelta(t, 0.46, resp.Confidence, 0.001)
}

func TestGenerateAutomatic_SendsCorrectBody(t *testing.T) {
	var gotBody []byte
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://example.com/meme.png"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	_, err := c.GenerateAutomatic(context.Background(), AutomaticRequest{Text: "hello world"})
	require.NoError(t, err)

	assert.Equal(t, "application/json", gotCT)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	assert.Equal(t, "hello world", parsed["text"])
}

func TestGenerateAutomatic_SafeFlag(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://example.com/meme.png"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	_, err := c.GenerateAutomatic(context.Background(), AutomaticRequest{Text: "test", Safe: true})
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	assert.Equal(t, true, parsed["safe"])
}

func TestGenerateAutomatic_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.GenerateAutomatic(context.Background(), AutomaticRequest{Text: ""})
	require.Error(t, err)
	assert.Nil(t, resp)

	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	assert.Equal(t, "bad request", apiErr.Message)
}

func TestGenerateAutomatic_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	// Use a client without retries so we don't wait.
	c := &Client{
		http:      &http.Client{},
		baseURL:   srv.URL,
		userAgent: "memelink-cli/test",
	}

	resp, err := c.GenerateAutomatic(context.Background(), AutomaticRequest{Text: "test"})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "500")
}

// --- Generate (template-based) tests ---

func TestGenerate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/drake/top/bottom.jpg"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.Generate(context.Background(), GenerateRequest{
		TemplateID: "drake",
		Text:       []string{"top", "bottom"},
	})
	require.NoError(t, err)
	assert.Equal(t, "https://api.memegen.link/images/drake/top/bottom.jpg", resp.URL)
}

func TestGenerate_SendsCorrectBody(t *testing.T) {
	var gotBody []byte
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://example.com/meme.png"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	_, err := c.Generate(context.Background(), GenerateRequest{
		TemplateID: "drake",
		Text:       []string{"one", "two"},
		Extension:  "png",
		Font:       "impact",
		Layout:     "top",
		Style:      []string{"default"},
		Redirect:   false,
	})
	require.NoError(t, err)

	assert.Equal(t, "/images", gotPath)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	assert.Equal(t, "drake", parsed["template_id"])
	assert.Equal(t, []any{"one", "two"}, parsed["text"])
	assert.Equal(t, "png", parsed["extension"])
	assert.Equal(t, "impact", parsed["font"])
	assert.Equal(t, "top", parsed["layout"])
	assert.Equal(t, []any{"default"}, parsed["style"])
	assert.Equal(t, false, parsed["redirect"])
}

func TestGenerate_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"template 'xyz' not found"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.Generate(context.Background(), GenerateRequest{TemplateID: "xyz"})
	require.Error(t, err)
	assert.Nil(t, resp)

	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Equal(t, "template 'xyz' not found", apiErr.Message)
}

// --- GenerateCustom tests ---

func TestGenerateCustom_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://api.memegen.link/images/custom/hello.jpg"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.GenerateCustom(context.Background(), CustomRequest{
		Background: "https://example.com/bg.jpg",
		Text:       []string{"hello"},
	})
	require.NoError(t, err)
	assert.Equal(t, "https://api.memegen.link/images/custom/hello.jpg", resp.URL)
}

func TestGenerateCustom_SendsCorrectBody(t *testing.T) {
	var gotBody []byte
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"url":"https://example.com/meme.png"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	_, err := c.GenerateCustom(context.Background(), CustomRequest{
		Background: "https://example.com/bg.jpg",
		Text:       []string{"top", "bottom"},
		Extension:  "png",
		Font:       "impact",
		Layout:     "top",
		Style:      "default,animated",
		Redirect:   false,
	})
	require.NoError(t, err)

	assert.Equal(t, "/images/custom", gotPath)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &parsed))
	assert.Equal(t, "https://example.com/bg.jpg", parsed["background"])
	assert.Equal(t, []any{"top", "bottom"}, parsed["text"])
	assert.Equal(t, "png", parsed["extension"])
	assert.Equal(t, "impact", parsed["font"])
	// Style is a string (not array) for custom endpoint
	assert.Equal(t, "default,animated", parsed["style"])
	assert.Equal(t, false, parsed["redirect"])
}

func TestGenerateCustom_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnsupportedMediaType)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	resp, err := c.GenerateCustom(context.Background(), CustomRequest{
		Background: "https://invalid.example.com/bad.txt",
		Text:       []string{"hello"},
	})
	require.Error(t, err)
	assert.Nil(t, resp)

	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusUnsupportedMediaType, apiErr.StatusCode)
	assert.Equal(t, "could not download image URL", apiErr.Message)
}

// --- AppendQueryParams tests ---

func TestAppendQueryParams_Basic(t *testing.T) {
	params := url.Values{}
	params.Set("width", "400")

	result, err := AppendQueryParams("https://api.memegen.link/images/drake/top/bottom.jpg", params)
	require.NoError(t, err)

	u, _ := url.Parse(result)
	assert.Equal(t, "400", u.Query().Get("width"))
}

func TestAppendQueryParams_PreservesExisting(t *testing.T) {
	params := url.Values{}
	params.Set("width", "400")

	result, err := AppendQueryParams("https://api.memegen.link/images/drake/top/bottom.jpg?font=impact", params)
	require.NoError(t, err)

	u, _ := url.Parse(result)
	assert.Equal(t, "impact", u.Query().Get("font"))
	assert.Equal(t, "400", u.Query().Get("width"))
}

func TestAppendQueryParams_Empty(t *testing.T) {
	params := url.Values{}

	result, err := AppendQueryParams("https://api.memegen.link/images/drake/top/bottom.jpg", params)
	require.NoError(t, err)
	assert.Equal(t, "https://api.memegen.link/images/drake/top/bottom.jpg", result)
}

func TestAppendQueryParams_EncodesSpecialChars(t *testing.T) {
	params := url.Values{}
	params.Set("color", "#ff0000")

	result, err := AppendQueryParams("https://api.memegen.link/images/drake/top/bottom.jpg", params)
	require.NoError(t, err)

	u, _ := url.Parse(result)
	assert.Equal(t, "#ff0000", u.Query().Get("color"))
}

// --- ListTemplates tests ---

func TestListTemplates_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"drake","name":"Drake Hotline Bling","lines":2},{"id":"buzz","name":"Buzz Lightyear","lines":2}]`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	templates, err := c.ListTemplates(context.Background(), "")
	require.NoError(t, err)
	assert.Len(t, templates, 2)
	assert.Equal(t, "drake", templates[0].ID)
	assert.Equal(t, "Drake Hotline Bling", templates[0].Name)
}

func TestListTemplates_WithFilter(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"drake","name":"Drake Hotline Bling","lines":2}]`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	_, err := c.ListTemplates(context.Background(), "drake")
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "filter=drake")
}

func TestListTemplates_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := &Client{http: &http.Client{}, baseURL: srv.URL, userAgent: "test"}
	templates, err := c.ListTemplates(context.Background(), "")
	require.Error(t, err)
	assert.Nil(t, templates)
}

// --- GetTemplate tests ---

func TestGetTemplate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/templates/drake", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"drake","name":"Drake Hotline Bling","lines":2,"styles":["default","animated"],"blank":"https://api.memegen.link/images/drake.png"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	tmpl, err := c.GetTemplate(context.Background(), "drake")
	require.NoError(t, err)
	assert.Equal(t, "drake", tmpl.ID)
	assert.Equal(t, "Drake Hotline Bling", tmpl.Name)
	assert.Equal(t, 2, tmpl.Lines)
	assert.Contains(t, tmpl.Styles, "animated")
}

func TestGetTemplate_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"template 'xyz' not found"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	tmpl, err := c.GetTemplate(context.Background(), "xyz")
	require.Error(t, err)
	assert.Nil(t, tmpl)

	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 404, apiErr.StatusCode)
}

// --- ListFonts tests ---

func TestListFonts_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"impact","alias":"impact-alias","filename":"impact.ttf"},{"id":"arial","alias":null,"filename":"arial.ttf"}]`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	fonts, err := c.ListFonts(context.Background())
	require.NoError(t, err)
	assert.Len(t, fonts, 2)
	assert.Equal(t, "impact", fonts[0].ID)
	require.NotNil(t, fonts[0].Alias)
	assert.Equal(t, "impact-alias", *fonts[0].Alias)
	assert.Nil(t, fonts[1].Alias)
}

func TestListFonts_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := &Client{http: &http.Client{}, baseURL: srv.URL, userAgent: "test"}
	fonts, err := c.ListFonts(context.Background())
	require.Error(t, err)
	assert.Nil(t, fonts)
}

// --- GetFont tests ---

func TestGetFont_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/fonts/impact", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"impact","alias":"impact-alias","filename":"impact.ttf","_self":"https://api.memegen.link/fonts/impact"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	font, err := c.GetFont(context.Background(), "impact")
	require.NoError(t, err)
	assert.Equal(t, "impact", font.ID)
	require.NotNil(t, font.Alias)
	assert.Equal(t, "impact-alias", *font.Alias)
}

func TestGetFont_NilAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"arial","alias":null,"filename":"arial.ttf"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	font, err := c.GetFont(context.Background(), "arial")
	require.NoError(t, err)
	assert.Nil(t, font.Alias)
}

func TestGetFont_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"font 'xyz' not found"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "")
	font, err := c.GetFont(context.Background(), "xyz")
	require.Error(t, err)
	assert.Nil(t, font)
}
