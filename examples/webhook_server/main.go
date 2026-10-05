package main

import (
	"log"
	"net/http"
	"os"

	"github.com/orce-me/woovi-go-sdk/webhookverify"
)

func main() {
	secret := os.Getenv("WOOVI_WEBHOOK_HMAC_SECRET")
	requireRSA := false // set true in production when validating x-webhook-signature

	mux := http.NewServeMux()
	mux.Handle("/webhooks/woovi", webhookverify.Handler(webhookverify.Options{
		HMACSecret: secret,
		RequireRSA: &requireRSA,
	}, func(w http.ResponseWriter, r *http.Request, event webhookverify.Event) {
		switch e := event.(type) {
		case *webhookverify.ChargeCompleted:
			if e.Charge != nil {
				log.Printf("paid correlationID=%s value=%s", e.Charge.CorrelationID, e.Charge.Value.BRLString())
			}
		default:
			log.Printf("event=%s", event.EventName())
		}
		w.WriteHeader(http.StatusOK)
	}))

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
