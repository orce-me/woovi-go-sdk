---
title: Webhooks
description: Register webhooks and verify Woovi signatures.
---

## Register

```go
hook, err := client.Webhooks.Create(ctx, &woovi.WebhookCreateParams{
	Name:     "charge paid",
	Event:    woovi.WebhookEventChargeCompleted,
	URL:      "https://example.com/webhooks/woovi",
	IsActive: true,
})
```

## Public keys and IPs

```go
keys, err := client.Webhooks.ListPublicKeys(ctx)
ips, err := client.Webhooks.ListIPs(ctx)
```

## Verify signatures

Use the `webhookverify` package:

```go
import "github.com/orce-me/woovi-go-sdk/webhookverify"

ok := webhookverify.VerifyRSA(payload, signatureBase64)
```

Prefer `x-webhook-signature` (RSA). You can also verify HMAC with your webhook secret.
