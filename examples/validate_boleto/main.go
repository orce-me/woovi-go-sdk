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
	barcode := os.Getenv("WOOVI_BOLETO_BARCODE")
	if appID == "" || barcode == "" {
		log.Fatal("set WOOVI_APP_ID and WOOVI_BOLETO_BARCODE")
	}

	client, err := woovi.NewClient(appID)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	boleto, err := client.Boletos.Validate(ctx, barcode)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("value:", boleto.TotalValue.BRLString())
	fmt.Println("expires:", boleto.ExpiresDate)
	fmt.Println("beneficiary:", boleto.FinalBeneficiary.Name, boleto.FinalBeneficiary.TaxID)
}
