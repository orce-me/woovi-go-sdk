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
	pixKey := os.Getenv("WOOVI_DESTINATION_PIX_KEY")
	if pixKey == "" {
		log.Fatal("set WOOVI_DESTINATION_PIX_KEY")
	}

	client, err := woovi.NewClient(appID)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	payment, err := client.Payments.Create(ctx, &woovi.PaymentCreateParams{
		Value:                woovi.Cents(100),
		DestinationAlias:     pixKey,
		DestinationAliasType: woovi.PixKeyTypeRandom,
		CorrelationID:        fmt.Sprintf("payment-%d", time.Now().Unix()),
		Comment:              "pagamento de exemplo",
		AutoApprove:          woovi.Bool(false),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("status:", payment.Status)
	fmt.Println("correlationID:", payment.CorrelationID)
}
