package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/orce-me/woovi-go-sdk"
)

func main() {
	appID := os.Getenv("WOOVI_APP_ID")
	if appID == "" {
		log.Fatal("set WOOVI_APP_ID")
	}

	client, err := woovi.NewClient(appID)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	value := woovi.Cents(1000)
	qr, err := client.PixQRCodes.Create(ctx, &woovi.PixQRCodeCreateParams{
		Name:       "QRCode Exemplo",
		Identifier: fmt.Sprintf("qr%d", time.Now().Unix()%1_000_000_000),
		Value:      &value,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("correlationID:", qr.CorrelationID)
	fmt.Println("brCode:", qr.BrCode)
	fmt.Println("paymentLink:", qr.PaymentLinkURL)
}
