# woovi-go-sdk

Go SDK for the Woovi (OpenPix) REST API.

This SDK uses typed clients, `context.Context`, and cent-based money values.
Paths follow the public OpenAPI at `https://api.woovi.com/api/openapi.json`.
Types stay hand-written for idiomatic Go.

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

- Idiomatic `Client` with functional options
- Resources: `Accounts`, `AccountRegisters`, `Anticipations`, `Applications`, `Auth`, `Boletos`, `CashbackFidelities`, `Charges`, `Companies`, `Customers`, `Decode`, `Disputes`, `Files`, `FundsRecoveries`, `Giftbacks`, `Installments`, `Invoices`, `KYC`, `Limits`, `Partners`, `Payments`, `PixAuths`, `PixKeys`, `PixQRCodes`, `PSPs`, `Receipts`, `Refunds`, `Stablecoins`, `Statements`, `Subaccounts`, `Subscriptions`, `TEDs`, `Transactions`, `Transfers`, `Webhooks`
- MASTER helpers: `WithCompanyBankAccount`
- Observability hooks: `WithOnRequest`, `WithOnResponse`
- Method-specific params (`ChargeCreateParams`)
- Typed `APIError` with `errors.As` / `errors.Is`
- Pagination trio: `List`, `ListAll`, `ListPages` (`iter.Seq2`)
- Retry with full jitter for safe methods
- Injectable `http.Client` for tracing and proxies
- Webhook verify package (`RSA` + `HMAC`)
- Zero external dependencies

## Resources

### Charges

```go
result, err := client.Charges.Create(ctx, &woovi.ChargeCreateParams{
	CorrelationID: "pedido-4321",
	Value:         woovi.Cents(1990),
})

for charge, err := range client.Charges.ListAll(ctx, &woovi.ChargeListParams{Limit: 50}) {
	if err != nil {
		log.Fatal(err)
	}
	log.Println(charge.CorrelationID, charge.Status)
}
```

### Customers

```go
customer, err := client.Customers.Create(ctx, &woovi.CustomerCreateParams{
	Name:  "Ana",
	TaxID: "67200000051",
	Email: "ana@example.com",
})

updated, err := client.Customers.Update(ctx, customer.CorrelationID, &woovi.CustomerUpdateParams{
	Email: woovi.String("ana@new.example.com"),
})
```

### Payments (PixOut / BoletoOut)

```go
payment, err := client.Payments.Create(ctx, &woovi.PaymentCreateParams{
	Value:                woovi.Cents(100),
	DestinationAlias:     "c4249323-b4ca-43f2-8139-8232aab09b93",
	DestinationAliasType: woovi.PixKeyTypeRandom,
	CorrelationID:        "payment-1",
})

approved, err := client.Payments.Approve(ctx, payment.CorrelationID)
```

Decode a Pix copia-e-cola before paying:

```go
decoded, err := client.Decode.EMV(ctx, brCode)
```

Validate and pay a boleto:

```go
boleto, err := client.Boletos.Validate(ctx, "34195148200000003001095517077320772982609000")
payment, err := client.Payments.Create(ctx, &woovi.PaymentCreateParams{
	Type:          woovi.PaymentTypeBoleto,
	BoletoBarcode: boleto.Barcode,
	CorrelationID: "boleto-1",
})
```

### Subscriptions

```go
day := 5
sub, err := client.Subscriptions.Create(ctx, &woovi.SubscriptionCreateParams{
	Value: woovi.Cents(100),
	Customer: woovi.SubscriptionCustomerInput{
		Name:  "Dan",
		TaxID: "31324227036",
		Email: "email0@example.com",
		Phone: "5511999999999",
	},
	DayGenerateCharge: &day,
})
```

### Transactions

```go
tx, err := client.Transactions.Get(ctx, "E18236120202012032010s0133872GZA")

for item, err := range client.Transactions.ListAll(ctx, &woovi.TransactionListParams{
	Limit:  100,
	Charge: "charge-correlation-id",
}) {
	if err != nil {
		log.Fatal(err)
	}
	log.Println(item.EndToEndID, item.Value.BRLString())
}
```

### Refunds

```go
// By transaction endToEndId
refund, err := client.Refunds.Create(ctx, &woovi.RefundCreateParams{
	CorrelationID:         "refund-1",
	TransactionEndToEndID: "E18236120202012032010s0133872GZA",
	Value:                 woovi.Cents(1000),
})

// Or refund a charge
value := woovi.Cents(500)
chargeRefund, err := client.Charges.Refund(ctx, "pedido-4321", &woovi.ChargeRefundCreateParams{
	CorrelationID: "refund-charge-1",
	Value:         &value,
})
_ = refund
_ = chargeRefund
```

### Subaccounts

```go
sub, err := client.Subaccounts.Create(ctx, &woovi.SubaccountCreateParams{
	Name:   "seller-1",
	PixKey: "seller@example.com",
})

value := woovi.Cents(7000)
result, err := client.Subaccounts.Withdraw(ctx, sub.PixKey, &woovi.SubaccountWithdrawParams{
	Value: &value,
})
_ = result
```

Use `splits` on charge create to route value to subaccounts:

```go
result, err := client.Charges.Create(ctx, &woovi.ChargeCreateParams{
	CorrelationID: "pedido-split-1",
	Value:         woovi.Cents(10000),
	Splits: []woovi.ChargeSplit{{
		PixKey:    "seller@example.com",
		Value:     woovi.Cents(1500),
		SplitType: woovi.SplitTypeSubAccount,
	}},
})
```

### Accounts / Applications / Pix Keys (BAAS)

```go
account, err := client.Accounts.Create(ctx)

app, err := client.Applications.Create(ctx, &woovi.ApplicationCreateParams{
	AccountID: account.AccountID,
	Application: woovi.ApplicationInput{
		Name: "campaign-api",
		Type: woovi.ApplicationTypeAPI,
	},
})

// Use app.AppID with a new client for that account
campaign, err := woovi.NewClient(app.AppID)

key, err := campaign.PixKeys.Create(ctx, &woovi.PixKeyCreateParams{
	PixKey: "campanha@example.com",
	Type:   woovi.PixKeyTypeEmail,
})

checked, err := campaign.PixKeys.Check(ctx, "00000000191")
// store checked.PixKeyEndToEndID for Payments.Create
_ = key
```

### Pix QR Codes (static)

```go
value := woovi.Cents(1000)
qr, err := client.PixQRCodes.Create(ctx, &woovi.PixQRCodeCreateParams{
	Name:       "Loja Física",
	Identifier: "loja001abc",
	Value:      &value,
})
```

### Transfers / Statements / Account withdraw

```go
transfer, err := client.Transfers.Create(ctx, &woovi.TransferCreateParams{
	Value:         woovi.Cents(5000),
	FromPixKey:    "from@woovi.com",
	ToPixKey:      "to@woovi.com",
	CorrelationID: "repasse-1",
})

page, err := client.Statements.List(ctx, &woovi.StatementListParams{
	Start: "2026-08-01T00:00:00Z",
	End:   "2026-08-24T23:59:59Z",
	Limit: 100,
})

withdraw, err := client.Accounts.Withdraw(ctx, accountID, &woovi.AccountWithdrawParams{
	Value: woovi.Cents(7000),
})
_ = transfer
_ = page
_ = withdraw
```

### Subscriptions / Installments (Pix Automático)

```go
sub, err := client.Subscriptions.UpdateValue(ctx, globalID, woovi.Cents(250))
installments, err := client.Subscriptions.ListInstallments(ctx, globalID, 0, 100)
inst, err := client.Installments.Get(ctx, installmentGlobalID)
_, err = client.Installments.CreateCobr(ctx, installmentGlobalID, nil)
_, err = client.Subscriptions.Cancel(ctx, globalID)
_ = sub
_ = installments
_ = inst
```

### Company

```go
company, err := client.Companies.Get(ctx)
```

### Partners

```go
companies, err := client.Partners.ListCompanies(ctx)
company, err := client.Partners.GetCompany(ctx, "12345678000199")
affiliates, err := client.Partners.ListAffiliates(ctx)
_ = companies
_ = company
_ = affiliates
```

### Pix Auth / Account register

```go
auth, err := client.PixAuths.Create(ctx, &woovi.PixAuthCreateParams{
	CorrelationID: "signup-1",
	TaxID:         "12345678909",
})
got, err := client.PixAuths.Get(ctx, "signup-1")

reg, err := client.AccountRegisters.Get(ctx, "reg-1")
_, err = client.AccountRegisters.Delete(ctx, "reg-1")
_ = auth
_ = got
_ = reg
```

### Files / Disputes / Limits

```go
file, err := client.Files.Upload(ctx, &woovi.FileUploadParams{
	File:          bytes.NewReader(pdfBytes),
	FileName:      "evidence.pdf",
	ContentType:   "application/pdf",
	Purpose:       woovi.FilePurposeDisputeEvidence,
	CorrelationID: "evidence-1",
})

_, err = client.Disputes.AddEvidence(ctx, disputeID, &woovi.DisputeEvidenceParams{
	Documents: []woovi.DisputeEvidenceDocument{{
		FileID:      file.ID,
		Description: "Nota fiscal",
	}},
})

limits, err := client.Limits.Get(ctx, accountID)
req, err := client.Limits.CreateRequest(ctx, &woovi.LimitRequestCreateParams{
	CompanyBankAccountID: accountID,
	PixDayLimit:          woovi.Cents(5000000),
	PixNightLimit:        woovi.Cents(200000),
	Documents:            []woovi.LimitRequestDocumentInput{{FileID: file.ID}},
})
_ = limits
_ = req
```

### Cashback / Giftback / Auth / PSP / Subaccount ledger

```go
_, err = client.CashbackFidelities.Create(ctx, &woovi.CashbackFidelityCreateParams{
	TaxID: "31324227036",
	Value: woovi.Cents(1500),
})
bal, err := client.CashbackFidelities.Balance(ctx, "31324227036")

gift, err := client.Giftbacks.Balance(ctx, "31324227036")
token, err := client.Auth.ValidateToken(ctx)

psps, err := client.PSPs.List(ctx, &woovi.PSPListParams{Compe: "001"})

_, err = client.Subaccounts.Credit(ctx, "seller@example.com", &woovi.SubaccountMovementParams{
	Value: woovi.Cents(100),
})
_, err = client.Subaccounts.Debit(ctx, "seller@example.com", &woovi.SubaccountMovementParams{
	Value: woovi.Cents(50),
})
_ = bal
_ = gift
_ = token
_ = psps
```

### Pix key tokens / Charge image

```go
tokens, err := client.PixKeys.Tokens(ctx)
img, err := client.Charges.BRCodeImage(ctx, paymentLinkID, 768)
b64, err := client.Charges.QRCodeBase64(ctx, paymentLinkID, 768)
_ = tokens
_ = img
_ = b64
```

### TED / Receipt / Funds recovery / Invoice

```go
ted, err := client.TEDs.Create(ctx, &woovi.TEDCreateParams{
	CorrelationID: "payout-1",
	Value:         woovi.Cents(150050),
	AccountID:     accountID,
	Receiver: woovi.TEDParty{
		Name: "Joao", Document: "12345678901", ISPB: "87654321",
		Agency: 4321, Account: 98765, AccountType: woovi.TEDAccountTypeCACC,
	},
})

pdf, err := client.Receipts.Get(ctx, woovi.ReceiptTypePixIn, endToEndID)

fr, err := client.FundsRecoveries.Create(ctx, &woovi.FundsRecoveryCreateParams{
	TransactionEndToEndID: endToEndID,
	SituationType:         woovi.FundsRecoverySituationScam,
	Details:               "fake seller",
})

invoices, err := client.Invoices.List(ctx, nil)
invoice, err := client.Invoices.Create(ctx, &woovi.InvoiceCreateParams{
	Charge:        "charge-1",
	CorrelationID: "inv-1",
})
_, err = client.Invoices.UploadCertificate(ctx, &woovi.InvoiceCertificateParams{
	Pcks12:     certBase64,
	Passphrase: passphrase,
})
_, err = client.Invoices.PatchIntegration(ctx, map[string]any{"isActive": true})

btx, err := client.Boletos.ListTransactions(ctx, &woovi.BoletoTransactionListParams{
	Type: woovi.BoletoTransactionTypeIn,
})
_ = ted
_ = pdf
_ = fr
_ = invoices
_ = invoice
_ = btx
```

### Anticipation / KYC / Stablecoin

```go
list, err := client.Anticipations.List(ctx, &woovi.AnticipationListParams{
	Status: woovi.AnticipationStatusPending,
})
_, err = client.Anticipations.Approve(ctx, list.Anticipations[0].ID)

onboarding, err := client.KYC.CreateOnboarding(ctx, &woovi.KYCOnboardingParams{
	TaxID:         "12345678000199",
	CorrelationID: "merchant-1",
})

gross := woovi.Cents(10000)
deposit, err := client.Stablecoins.CreateDeposit(ctx, &woovi.StablecoinDepositCreateParams{
	GrossAmount:   &gross,
	Currency:      woovi.StablecoinUSDT,
	CorrelationID: "dep-1",
})
_, err = client.Stablecoins.ApproveDeposit(ctx, "dep-1")
_ = onboarding
_ = deposit
```

### Webhooks

```go
hook, err := client.Webhooks.Create(ctx, &woovi.WebhookCreateParams{
	Name:     "charge paid",
	Event:    woovi.WebhookEventChargeCompleted,
	URL:      "https://example.com/webhooks/woovi",
	IsActive: true,
})

events, err := client.Webhooks.ListEvents(ctx)
ips, err := client.Webhooks.ListIPs(ctx)
keys, err := client.Webhooks.ListPublicKeys(ctx)
_ = hook
_ = events
_ = ips
_ = keys
```

## Webhook verify

```go
import "github.com/orce-me/woovi-go-sdk/webhookverify"

requireRSA := true
mux.Handle("/webhooks/woovi", webhookverify.Handler(webhookverify.Options{
	HMACSecret: os.Getenv("WOOVI_WEBHOOK_HMAC_SECRET"),
	RequireRSA: &requireRSA,
}, func(w http.ResponseWriter, r *http.Request, event webhookverify.Event) {
	switch e := event.(type) {
	case *webhookverify.ChargeCompleted:
		_ = e.Charge.CorrelationID
	}
	w.WriteHeader(http.StatusOK)
}))
```

Use the raw request body for signature checks. Do not re-encode JSON.

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

## Status

v1.1 adds Giftback balance (`GET /api/v1/giftback/balance/{taxID}`) and
AppID token validation (`GET /api/v1/validate-token`).
These paths are live and match the `GIFTBACK_BALANCE_GET` and
`TOKEN_VALIDATE_GET` scopes, but they are still missing from the public OpenAPI.
Response shapes for `Auth.ValidateToken` are best-effort.
