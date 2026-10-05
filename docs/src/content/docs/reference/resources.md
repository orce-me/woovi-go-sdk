---
title: Resources
description: Services exposed on the Woovi Go client.
---

Access resources as fields on `Client`:

```go
client.Charges
client.Customers
client.Payments
client.Webhooks
// ...
```

## Available services

- `Accounts`, `AccountRegisters`
- `Anticipations`, `Applications`, `Auth`
- `Boletos`, `CashbackFidelities`, `Charges`
- `Companies`, `Customers`, `Decode`
- `Disputes`, `Files`, `FundsRecoveries`, `Giftbacks`
- `Installments`, `Invoices`, `KYC`, `Limits`
- `Partners`, `Payments`, `PixAuths`, `PixKeys`, `PixQRCodes`
- `PSPs`, `Receipts`, `Refunds`, `Stablecoins`
- `Statements`, `Subaccounts`, `Subscriptions`
- `TEDs`, `Transactions`, `Transfers`, `Webhooks`

Paths follow the public OpenAPI when available. Giftback and `Auth.ValidateToken` use live paths that are not yet in the OpenAPI document.
