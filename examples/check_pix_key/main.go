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
	pixKey := os.Getenv("WOOVI_PIX_KEY")
	if appID == "" || pixKey == "" {
		log.Fatal("set WOOVI_APP_ID and WOOVI_PIX_KEY")
	}

	client, err := woovi.NewClient(appID)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.PixKeys.Check(ctx, pixKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("pixKey:", result.PixKey)
	fmt.Println("type:", result.Type)
	fmt.Println("owner:", result.Owner.Name)
	fmt.Println("psp:", result.Owner.PSP)
	fmt.Println("pixKeyEndToEndId:", result.PixKeyEndToEndID)
}
