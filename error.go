package woovi

import (
	"errors"
	"fmt"
	"net/http"
)

// Sentinel errors for common HTTP outcomes.
var (
	ErrUnauthorized = errors.New("woovi: unauthorized")
	ErrNotFound     = errors.New("woovi: not found")
	ErrRateLimited  = errors.New("woovi: rate limited")
)

// APIError is a non-2xx Woovi API response.
type APIError struct {
	StatusCode int
	Message    string
	Code       string
	Errors     []APIErrorItem
	RequestID  string
	Body       []byte
}

// APIErrorItem is one field-level error from the API body.
type APIErrorItem struct {
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" && len(e.Errors) > 0 {
		msg = e.Errors[0].Message
	}
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	if e.Code != "" && e.RequestID != "" {
		return fmt.Sprintf("woovi: %s (%d/%s) [request_id=%s]", msg, e.StatusCode, e.Code, e.RequestID)
	}
	if e.Code != "" {
		return fmt.Sprintf("woovi: %s (%d/%s)", msg, e.StatusCode, e.Code)
	}
	if e.RequestID != "" {
		return fmt.Sprintf("woovi: %s (%d) [request_id=%s]", msg, e.StatusCode, e.RequestID)
	}
	return fmt.Sprintf("woovi: %s (%d)", msg, e.StatusCode)
}

// Unwrap maps status codes to sentinel errors for errors.Is.
func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		return nil
	}
}

type errorEnvelope struct {
	Message   string         `json:"message"`
	Error     string         `json:"error"`
	ErrorCode string         `json:"errorCode"`
	Errors    []APIErrorItem `json:"errors"`
	Data      any            `json:"data"`
}
