package woovi

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// RequestOption configures a single API call.
type RequestOption func(*requestConfig)

type requestConfig struct {
	headers        http.Header
	query          url.Values
	idempotencyKey string
	timeout        time.Duration
}

func applyRequestOptions(opts []RequestOption) requestConfig {
	cfg := requestConfig{headers: make(http.Header)}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// WithHeader adds a header to a single request.
func WithHeader(key, value string) RequestOption {
	return func(c *requestConfig) {
		c.headers.Add(key, value)
	}
}

// WithQuery adds a query parameter to a single request.
func WithQuery(key, value string) RequestOption {
	return func(c *requestConfig) {
		if c.query == nil {
			c.query = make(url.Values)
		}
		c.query.Add(key, value)
	}
}

// WithCompanyBankAccount scopes a MASTER request to another company account.
func WithCompanyBankAccount(accountID string) RequestOption {
	return WithQuery("companyBankAccount", accountID)
}

// WithIdempotencyKey sets an idempotency key and enables retries for POST/PUT/PATCH.
func WithIdempotencyKey(key string) RequestOption {
	return func(c *requestConfig) {
		c.idempotencyKey = key
	}
}

// WithRequestTimeout sets a per-request timeout from the parent context.
func WithRequestTimeout(d time.Duration) RequestOption {
	return func(c *requestConfig) {
		c.timeout = d
	}
}

func (c *Client) withRequestContext(ctx context.Context, cfg requestConfig) (context.Context, context.CancelFunc) {
	if cfg.timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, cfg.timeout)
}
