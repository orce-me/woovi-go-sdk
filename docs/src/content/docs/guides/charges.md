---
title: Charges
description: Create, list, and manage Pix charges.
---

## Create

```go
result, err := client.Charges.Create(ctx, &woovi.ChargeCreateParams{
	CorrelationID: "pedido-4321",
	Value:         woovi.Cents(1990),
	Comment:       woovi.String("Pedido #4321"),
})
```

## Get

```go
charge, err := client.Charges.Get(ctx, "pedido-4321")
```

## List all pages

```go
for charge, err := range client.Charges.ListAll(ctx, &woovi.ChargeListParams{Limit: 50}) {
	if err != nil {
		log.Fatal(err)
	}
	log.Println(charge.CorrelationID, charge.Status)
}
```

## Delete

```go
err := client.Charges.Delete(ctx, "pedido-4321")
```
