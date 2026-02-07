package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error represents an error from the Memegen API.
type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("memegen api: %s (HTTP %d)", e.Message, e.StatusCode)
}

// checkImageResponse validates responses from image-generating endpoints.
// These endpoints return errors as images, not JSON, so status code mapping is used.
func checkImageResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return &Error{
		StatusCode: resp.StatusCode,
		Message:    statusMessage(resp.StatusCode),
	}
}

// checkJSONResponse validates responses from JSON-returning endpoints.
// It tries to parse {"error":"..."} from the body, falling back to status code mapping.
func checkJSONResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	var apiErr struct {
		Error string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Error != "" {
		return &Error{
			StatusCode: resp.StatusCode,
			Message:    apiErr.Error,
		}
	}

	return &Error{
		StatusCode: resp.StatusCode,
		Message:    statusMessage(resp.StatusCode),
	}
}

// statusMessage maps HTTP status codes to human-readable error messages.
func statusMessage(code int) string {
	switch code {
	case http.StatusNotFound:
		return "template not found"
	case http.StatusRequestURITooLong:
		return "text too long (max 200 chars per line)"
	case http.StatusUnsupportedMediaType:
		return "could not download image URL"
	case http.StatusUnprocessableEntity:
		return "invalid style or missing image URL"
	case http.StatusTooManyRequests:
		return "rate limited, try again later"
	default:
		return fmt.Sprintf("unexpected error (HTTP %d)", code)
	}
}
