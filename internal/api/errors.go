package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

// APIError represents an RFC 7807 Problem Details error from the HyperFleet API.
type APIError struct {
	Type      string                    `json:"type"`
	Title     string                    `json:"title"`
	Status    int                       `json:"status"`
	Detail    string                    `json:"detail"`
	Instance  string                    `json:"instance,omitempty"`
	Code      string                    `json:"code,omitempty"`
	Timestamp string                    `json:"timestamp,omitempty"`
	TraceID   string                    `json:"trace_id,omitempty"`
	Errors    []resource.ValidationError `json:"errors,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%d] %s: %s", e.Status, e.Title, e.Detail)
}

// parseError reads a non-2xx HTTP response and returns an *APIError.
func parseError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			Status: resp.StatusCode,
			Title:  resp.Status,
			Detail: fmt.Sprintf("failed to read response body: %v", err),
		}
	}

	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "application/problem+json") || strings.Contains(ct, "application/json") {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Title != "" {
			if apiErr.Status == 0 {
				apiErr.Status = resp.StatusCode
			}
			return &apiErr
		}
	}

	// Non-JSON body
	detail := string(body)
	if len(detail) > 500 {
		detail = detail[:500] + "... [truncated]"
	}

	// HTML detection
	trimmed := strings.TrimSpace(detail)
	if strings.HasPrefix(trimmed, "<!") || strings.HasPrefix(strings.ToLower(trimmed), "<html") {
		detail = "Received HTML response (possibly not the HyperFleet API). Verify the API URL with 'hf env show'.\n" + detail
	}

	return &APIError{
		Status: resp.StatusCode,
		Title:  resp.Status,
		Detail: detail,
	}
}
