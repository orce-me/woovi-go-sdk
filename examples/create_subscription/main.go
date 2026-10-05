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

	day := 5
	sub, err := client.Subscriptions.Create(ctx, &woovi.SubscriptionCreateParams{
		Value: woovi.Cents(100),
		Customer: woovi.SubscriptionCustomerInput{
			Name:  "Dan",
			TaxID: os.Getenv("WOOVI_CUSTOMER_TAX_ID"),
			Email: "email0@example.com",
			Phone: "5511999999999",
		},
		DayGenerateCharge: &day,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("globalID:", sub.GlobalID)
	fmt.Println("value:", sub.Value.BRLString())
}
