package woovi

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

type attemptKey struct{}

func withAttempt(ctx context.Context, n int) context.Context {
	return context.WithValue(ctx, attemptKey{}, n)
}

func attemptFromContext(ctx context.Context) int {
	if v, ok := ctx.Value(attemptKey{}).(int); ok {
		return v
	}
	return 0
}

const (
	headerAuthorization  = "Authorization"
	headerUserAgent      = "User-Agent"
	headerIdempotencyKey = "Idempotency-Key"
	headerRetryAfter     = "Retry-After"
	headerRequestID      = "X-Request-Id"
)

func wrapTransport(
	base http.RoundTripper,
	appID, userAgent string,
	retry RetryConfig,
	logger *slog.Logger,
	onRequest OnRequestFunc,
	onResponse OnResponseFunc,
) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	rt := base
	rt = authTransport{next: rt, appID: appID, userAgent: userAgent}
	if onRequest != nil || onResponse != nil {
		rt = hookTransport{next: rt, onRequest: onRequest, onResponse: onResponse}
	}
	rt = retryTransport{next: rt, cfg: retry, logger: logger}
	return rt
}

type authTransport struct {
	next      http.RoundTripper
	appID     string
	userAgent string
}

func (t authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	if r.Header.Get(headerAuthorization) == "" {
		r.Header.Set(headerAuthorization, t.appID)
	}
	if r.Header.Get(headerUserAgent) == "" {
		r.Header.Set(headerUserAgent, t.userAgent)
	}
	if r.Header.Get("Accept") == "" {
		r.Header.Set("Accept", "application/json")
	}
	return t.next.RoundTrip(r)
}

type retryTransport struct {
	next   http.RoundTripper
	cfg    RetryConfig
	logger *slog.Logger
}

func (t retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	maxRetries := t.cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		_ = req.Body.Close()
	}

	attempt := 0
	for {
		r := req.Clone(withAttempt(req.Context(), attempt))
		if body != nil {
			r.Body = io.NopCloser(bytes.NewReader(body))
			r.ContentLength = int64(len(body))
			r.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(body)), nil
			}
		}

		resp, err := t.next.RoundTrip(r)
		if err != nil {
			if attempt >= maxRetries || !canRetry(req, r) {
				return nil, err
			}
			if !t.sleep(req, attempt, 0) {
				return nil, err
			}
			attempt++
			continue
		}

		if !shouldRetryStatus(resp.StatusCode) || attempt >= maxRetries || !canRetry(req, r) {
			return resp, nil
		}

		retryAfter := parseRetryAfter(resp.Header.Get(headerRetryAfter))
		_ = resp.Body.Close()
		if t.logger != nil {
			t.logger.Debug("woovi: retrying request",
				"method", req.Method,
				"path", req.URL.Path,
				"status", resp.StatusCode,
				"attempt", attempt+1,
			)
		}
		if !t.sleep(req, attempt, retryAfter) {
			return nil, req.Context().Err()
		}
		attempt++
	}
}

func canRetry(original, cloned *http.Request) bool {
	switch original.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete:
		return true
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return cloned.Header.Get(headerIdempotencyKey) != ""
	default:
		return false
	}
}

func shouldRetryStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func (t retryTransport) sleep(req *http.Request, attempt int, retryAfter time.Duration) bool {
	delay := retryAfter
	if delay <= 0 {
		base := t.cfg.InitialBackoff
		if base <= 0 {
			base = 200 * time.Millisecond
		}
		maxBackoff := t.cfg.MaxBackoff
		if maxBackoff <= 0 {
			maxBackoff = 5 * time.Second
		}
		delay = base << attempt
		if delay > maxBackoff {
			delay = maxBackoff
		}
		if delay > 0 {
			delay = time.Duration(rand.Int64N(int64(delay) + 1))
		}
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-req.Context().Done():
		return false
	case <-timer.C:
		return true
	}
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d < 0 {
			return 0
		}
		return d
	}
	return 0
}
