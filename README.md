<p align="center">
  <img src="assets/woovi.png" alt="Woovi" width="96" />
</p>

<h1 align="center">woovi-go-sdk</h1>

<p align="center">
  <strong>Go SDK for the Woovi (OpenPix) REST API</strong>
</p>

<p align="center">
  Typed clients · <code>context.Context</code> · money in cents · zero dependencies
</p>

<p align="center">
  <a href="https://orce-me.github.io/woovi-go-sdk/"><img src="https://img.shields.io/badge/docs-GitHub%20Pages-03D69D?style=for-the-badge&logo=readthedocs&logoColor=white" alt="Docs" /></a>
  <a href="https://pkg.go.dev/github.com/orce-me/woovi-go-sdk"><img src="https://img.shields.io/badge/pkg.go.dev-reference-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go reference" /></a>
  <a href="https://github.com/orce-me/woovi-go-sdk/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/orce-me/woovi-go-sdk/ci.yml?branch=main&style=for-the-badge&label=CI" alt="CI" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-0A0F2C?style=for-the-badge" alt="MIT" /></a>
</p>

<p align="center">
  <a href="https://orce-me.github.io/woovi-go-sdk/"><b>Documentation →</b></a>
  ·
  <a href="https://developers.woovi.com/">Woovi API</a>
  ·
  <a href="https://pkg.go.dev/github.com/orce-me/woovi-go-sdk">GoDoc</a>
</p>

<p align="center">
  <img src="assets/sdk-usage.jpg" alt="Create a Pix charge with woovi-go-sdk" width="920" />
</p>

## Install

```bash
go get github.com/orce-me/woovi-go-sdk
```

Requires Go 1.23+.

## Quick start

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/orce-me/woovi-go-sdk"
)

func main() {
	client, err := woovi.NewClient(os.Getenv("WOOVI_APP_ID"))
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.Charges.Create(context.Background(), &woovi.ChargeCreateParams{
		CorrelationID: "pedido-4321",
		Value:         woovi.Cents(1990), // R$ 19,90
		Comment:       woovi.String("Pedido #4321"),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(result.Charge.BrCode)
	log.Println(result.Charge.PaymentLinkURL)
}
```

Send the AppID in `Authorization` with **no** `Bearer` prefix. Keep it on the server.

## Features

- Functional options (`WithBaseURL`, `WithHTTPClient`, `WithRetry`, …)
- Typed `APIError` with `errors.As` / `errors.Is`
- Pagination: `List`, `ListAll`, `ListPages` (`iter.Seq2`)
- Retry with full jitter for safe methods
- Hooks: `WithOnRequest`, `WithOnResponse`
- Webhook verify package (`RSA` + `HMAC`)

## Options

```go
client, err := woovi.NewClient(appID,
	woovi.WithBaseURL("https://api.woovi.com"),
	woovi.WithHTTPClient(http.DefaultClient),
	woovi.WithRetry(woovi.RetryConfig{MaxRetries: 2}),
	woovi.WithLogger(slog.Default()),
	woovi.WithOnResponse(func(req woovi.RequestInfo, resp woovi.ResponseInfo) {
		log.Println(req.Method, resp.StatusCode, resp.Duration)
	}),
)
```

## Examples

```bash
export WOOVI_APP_ID=your-app-id
go run ./examples/create_charge
go run ./examples/webhook_server
```

More guides: **[orce-me.github.io/woovi-go-sdk](https://orce-me.github.io/woovi-go-sdk/)**

## License

MIT
