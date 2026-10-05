---
title: Money
description: Represent amounts in cents with the Money type.
---

All API amounts use **centavos** as integers.

```go
value := woovi.Cents(1990) // R$ 19,90
```

`Money` is a typed integer. Pass pointers when a field is optional:

```go
value := woovi.Cents(1000)
qr, err := client.PixQRCodes.Create(ctx, &woovi.PixQRCodeCreateParams{
	Name:  "Loja Física",
	Value: &value,
})
```

Do not send floating-point reais. Convert first:

```go
// wrong: 19.90
// right:
woovi.Cents(1990)
```
