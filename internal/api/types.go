package api

// AutomaticRequest is the payload for POST /images/automatic.
type AutomaticRequest struct {
	Text string `json:"text"`
	Safe bool   `json:"safe"`
}

// AutomaticResponse is the response from POST /images/automatic.
type AutomaticResponse struct {
	URL        string  `json:"url"`
	Generator  string  `json:"generator,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// GenerateRequest is the payload for POST /images/{template_id}.
type GenerateRequest struct {
	TemplateID string   `json:"template_id"`
	Text       []string `json:"text"`
	Extension  string   `json:"extension,omitempty"`
	Font       string   `json:"font,omitempty"`
	Layout     string   `json:"layout,omitempty"`
	Style      []string `json:"style,omitempty"`
	Redirect   bool     `json:"redirect"` // always false
}

// CustomRequest is the payload for POST /images/custom.
type CustomRequest struct {
	Background string   `json:"background"`
	Text       []string `json:"text"`
	Extension  string   `json:"extension,omitempty"`
	Font       string   `json:"font,omitempty"`
	Layout     string   `json:"layout,omitempty"`
	Style      string   `json:"style,omitempty"`
	Redirect   bool     `json:"redirect"` // always false
}

// GenerateResponse is the response from template/custom generation endpoints.
type GenerateResponse struct {
	URL string `json:"url"`
}

// Template describes a meme template from the API.
type Template struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Lines    int      `json:"lines"`
	Overlays int      `json:"overlays"`
	Styles   []string `json:"styles"`
	Blank    string   `json:"blank"`
	Example  struct {
		Text []string `json:"text"`
		URL  string   `json:"url"`
	} `json:"example"`
	Source   string   `json:"source"`
	Keywords []string `json:"keywords"`
	Self     string   `json:"_self"`
}

// Font describes a font from the API.
type Font struct {
	ID       string  `json:"id"`
	Alias    *string `json:"alias"` // nullable in API
	Filename string  `json:"filename"`
	Self     string  `json:"_self"`
}
