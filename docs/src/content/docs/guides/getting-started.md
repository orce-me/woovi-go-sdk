---
title: Getting started
description: Install and create your first Woovi charge in Go.
---

## Requirements

- Go 1.23 or newer
- A Woovi AppID from the company API settings

## Install

```bash
go get github.com/orce-me/woovi-go-sdk
```

## Create a client

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

## Next steps

- [Authentication](./authentication/)
- [Money](./money/)
- [Charges](./charges/)
- [Webhooks](./webhooks/)
