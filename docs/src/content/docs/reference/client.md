---
title: Client options
description: Configure the Woovi Go client with functional options.
---

```go
client, err := woovi.NewClient(
	os.Getenv("WOOVI_APP_ID"),
	woovi.WithBaseURL("https://api.woovi.com"),
	woovi.WithHTTPClient(http.DefaultClient),
	woovi.WithRetry(woovi.RetryConfig{MaxRetries: 2}),
	woovi.WithLogger(slog.Default()),
	woovi.WithOnResponse(func(req woovi.RequestInfo, resp woovi.ResponseInfo) {
		log.Println(req.Method, resp.StatusCode, resp.Duration)
	}),
)
```

| Option | Purpose |
| --- | --- |
| `WithBaseURL` | Override API host (sandbox, proxy) |
| `WithHTTPClient` | Inject tracing, proxies, or custom transport |
| `WithRetry` | Retry safe methods with full jitter |
| `WithTimeout` | Request timeout |
| `WithLogger` | `slog` logger |
| `WithOnRequest` / `WithOnResponse` | Observability hooks |
| `WithUserAgent` | Custom User-Agent suffix |

The default base URL is `https://api.woovi.com`.
