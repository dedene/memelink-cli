package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GenerateAutomatic posts text to /images/automatic and returns the
// auto-selected meme URL along with generator metadata.
func (c *Client) GenerateAutomatic(ctx context.Context, req AutomaticRequest) (*AutomaticResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling automatic request: %w", err)
	}

	resp, err := c.Post(ctx, "/images/automatic", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("posting automatic image: %w", err)
	}
	defer resp.Body.Close()

	if err := checkJSONResponse(resp); err != nil {
		return nil, err
	}

	var out AutomaticResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding automatic response: %w", err)
	}

	return &out, nil
}

// Generate posts a template-based meme request to POST /images and returns
// the generated meme URL.
func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling generate request: %w", err)
	}

	resp, err := c.Post(ctx, "/images", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("posting template image: %w", err)
	}
	defer resp.Body.Close()

	if err := checkJSONResponse(resp); err != nil {
		return nil, err
	}

	var out GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding generate response: %w", err)
	}

	return &out, nil
}

// GenerateCustom posts a custom-background meme request to POST /images/custom
// and returns the generated meme URL.
func (c *Client) GenerateCustom(ctx context.Context, req CustomRequest) (*GenerateResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling custom request: %w", err)
	}

	resp, err := c.Post(ctx, "/images/custom", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("posting custom image: %w", err)
	}
	defer resp.Body.Close()

	if err := checkJSONResponse(resp); err != nil {
		return nil, err
	}

	var out GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding custom response: %w", err)
	}

	return &out, nil
}

// AppendQueryParams merges the given url.Values onto baseURL's existing query
// string. It is used to add presentation params (color, width, etc.) to the
// meme URL returned by the API.
func AppendQueryParams(baseURL string, params url.Values) (string, error) {
	if len(params) == 0 {
		return baseURL, nil
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parsing URL: %w", err)
	}

	q := u.Query()
	for k, vs := range params {
		for _, v := range vs {
			q.Set(k, v)
		}
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}

// ListTemplates fetches all meme templates from GET /templates.
// The optional filter query-param narrows results server-side.
func (c *Client) ListTemplates(ctx context.Context, filter string) ([]Template, error) {
	path := "/templates"
	if filter != "" {
		path += "?filter=" + url.QueryEscape(filter)
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("listing templates: %w", err)
	}
	defer resp.Body.Close()

	if err := checkJSONResponse(resp); err != nil {
		return nil, err
	}

	var out []Template
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding templates: %w", err)
	}

	return out, nil
}

// GetTemplate fetches a single template by ID from GET /templates/{id}.
func (c *Client) GetTemplate(ctx context.Context, id string) (*Template, error) {
	resp, err := c.Get(ctx, "/templates/"+url.PathEscape(id))
	if err != nil {
		return nil, fmt.Errorf("getting template %q: %w", id, err)
	}
	defer resp.Body.Close()

	if err := checkJSONResponse(resp); err != nil {
		return nil, err
	}

	var out Template
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding template %q: %w", id, err)
	}

	return &out, nil
}

// ListFonts fetches all fonts from GET /fonts.
func (c *Client) ListFonts(ctx context.Context) ([]Font, error) {
	resp, err := c.Get(ctx, "/fonts")
	if err != nil {
		return nil, fmt.Errorf("listing fonts: %w", err)
	}
	defer resp.Body.Close()

	if err := checkJSONResponse(resp); err != nil {
		return nil, err
	}

	var out []Font
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding fonts: %w", err)
	}

	return out, nil
}

// GetFont fetches a single font by ID from GET /fonts/{id}.
func (c *Client) GetFont(ctx context.Context, id string) (*Font, error) {
	resp, err := c.Get(ctx, "/fonts/"+url.PathEscape(id))
	if err != nil {
		return nil, fmt.Errorf("getting font %q: %w", id, err)
	}
	defer resp.Body.Close()

	if err := checkJSONResponse(resp); err != nil {
		return nil, err
	}

	var out Font
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decoding font %q: %w", id, err)
	}

	return &out, nil
}
