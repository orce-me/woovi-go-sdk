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

	customer, err := client.Customers.Create(ctx, &woovi.CustomerCreateParams{
		Name:  "Ana",
		TaxID: os.Getenv("WOOVI_CUSTOMER_TAX_ID"),
		Email: "ana@example.com",
		Phone: "5511999999999",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("correlationID:", customer.CorrelationID)
	if customer.TaxID != nil {
		fmt.Println("taxID:", customer.TaxID.TaxID, customer.TaxID.Type)
	}
}
