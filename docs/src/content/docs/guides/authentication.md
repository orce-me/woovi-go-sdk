---
title: Authentication
description: Authenticate the Go SDK with a Woovi AppID.
---

Send the AppID in the `Authorization` header with **no** `Bearer` prefix.

```go
client, err := woovi.NewClient(os.Getenv("WOOVI_APP_ID"))
```

Keep the AppID on the server. Do not ship it in browsers or mobile apps.

## Validate the AppID

```go
token, err := client.Auth.ValidateToken(ctx)
if err != nil {
	log.Fatal(err)
}
if !token.OK() {
	log.Fatal("invalid AppID")
}
```

Requires the `TOKEN_VALIDATE_GET` scope.

## Scope another company account

For MASTER AppIDs, pass `WithCompanyBankAccount` on a request:

```go
charge, err := client.Charges.Get(
	ctx,
	"pedido-4321",
	woovi.WithCompanyBankAccount(accountID),
)
```
