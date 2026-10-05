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
	emv := os.Getenv("WOOVI_EMV")
	if appID == "" || emv == "" {
		log.Fatal("set WOOVI_APP_ID and WOOVI_EMV")
	}

	client, err := woovi.NewClient(appID)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.Decode.EMV(ctx, emv)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("merchant:", result.EMV.MerchantName)
	fmt.Println("amount:", result.EMV.TransactionAmount)
	if result.EMV.MerchantAccountInformationPix != nil {
		fmt.Println("pixKey:", result.EMV.MerchantAccountInformationPix.PixKey)
		fmt.Println("url:", result.EMV.MerchantAccountInformationPix.URL)
	}
	if result.CobLocation != nil {
		fmt.Println("cob valid:", result.CobLocation.IsValid)
	}
	if result.RecLocation != nil {
		fmt.Println("rec valid:", result.RecLocation.IsValid)
	}
}
