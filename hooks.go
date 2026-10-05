package woovi

import (
	"net/http"
	"time"
)

// RequestInfo describes an outbound API call for hooks.
type RequestInfo struct {
	Method  string
	URL     string
	Attempt int
}

// ResponseInfo describes an HTTP response for hooks.
type ResponseInfo struct {
	StatusCode int
	Duration   time.Duration
	RequestID  string
	Err        error
}

// OnRequestFunc runs before each HTTP attempt.
type OnRequestFunc func(info RequestInfo)

// OnResponseFunc runs after each HTTP attempt.
type OnResponseFunc func(req RequestInfo, resp ResponseInfo)

func WithOnRequest(fn OnRequestFunc) Option {
	return func(c *Client) {
		c.onRequest = fn
	}
}

func WithOnResponse(fn OnResponseFunc) Option {
	return func(c *Client) {
		c.onResponse = fn
	}
}

type hookTransport struct {
	next       http.RoundTripper
	onRequest  OnRequestFunc
	onResponse OnResponseFunc
}

func (t hookTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	info := RequestInfo{
		Method:  req.Method,
		URL:     req.URL.String(),
		Attempt: attemptFromContext(req.Context()),
	}
	if t.onRequest != nil {
		t.onRequest(info)
	}

	start := time.Now()
	resp, err := t.next.RoundTrip(req)
	ri := ResponseInfo{
		Duration: time.Since(start),
		Err:      err,
	}
	if resp != nil {
		ri.StatusCode = resp.StatusCode
		ri.RequestID = resp.Header.Get(headerRequestID)
	}
	if t.onResponse != nil {
		t.onResponse(info, ri)
	}
	return resp, err
}
