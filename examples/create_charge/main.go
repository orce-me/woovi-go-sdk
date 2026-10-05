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

	result, err := client.Charges.Create(ctx, &woovi.ChargeCreateParams{
		CorrelationID: fmt.Sprintf("pedido-%d", time.Now().Unix()),
		Value:         woovi.Cents(1990),
		Comment:       woovi.String("Pedido de exemplo"),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("status:", result.Charge.Status)
	fmt.Println("value:", result.Charge.Value.BRLString())
	fmt.Println("brCode:", result.BrCode)
	fmt.Println("paymentLink:", result.Charge.PaymentLinkURL)
}
