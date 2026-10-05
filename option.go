package woovi

import (
	"log/slog"
	"net/http"
	"time"
)

// Option configures a Client.
type Option func(*Client)

// WithBaseURL sets the API base URL. Default is https://api.woovi.com.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets the HTTP client. Inject your own RoundTripper for tracing or proxies.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithLogger sets a slog logger for debug request logs. Pass nil to disable.
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) {
		c.logger = l
	}
}

// WithUserAgent sets a custom User-Agent suffix.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// WithRetry configures automatic retries for safe requests.
func WithRetry(cfg RetryConfig) Option {
	return func(c *Client) {
		c.retry = cfg
	}
}

// WithTimeout sets the default client timeout when the SDK creates the client.
// Ignored if WithHTTPClient is used.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.timeout = d
	}
}

// RetryConfig controls retry behavior.
type RetryConfig struct {
	// MaxRetries is retries after the first attempt. Zero disables. Default 2.
	MaxRetries int
	// InitialBackoff is the base delay before the first retry. Default 200ms.
	InitialBackoff time.Duration
	// MaxBackoff caps backoff delay. Default 5s.
	MaxBackoff time.Duration
}

func defaultRetry() RetryConfig {
	return RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 200 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
	}
}
