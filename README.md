# woovi-go-sdk

Go SDK for the Woovi (OpenPix) REST API.

Typed clients, `context.Context`, and cent-based money.
Zero external dependencies.

## Install

```bash
go get github.com/orce-me/woovi-go-sdk
```

Docs: https://orce-me.github.io/woovi-go-sdk/

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
		Value:         woovi.Cents(1990),
		Comment:       woovi.String("Pedido #4321"),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(result.Charge.BrCode)
	log.Println(result.Charge.PaymentLinkURL)
}
```

Send the AppID in `Authorization` with no `Bearer` prefix. Keep the AppID on the server.

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
go run ./examples/create_customer
go run ./examples/create_payment
go run ./examples/create_subscription
go run ./examples/create_pix_qrcode
go run ./examples/check_pix_key
go run ./examples/decode_emv
go run ./examples/validate_boleto
go run ./examples/webhook_server
```
