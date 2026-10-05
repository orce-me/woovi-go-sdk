package webhookverify

import (
	"encoding/json"
	"fmt"

	"github.com/orce-me/woovi-go-sdk"
)

// Event is a parsed OpenPix webhook payload.
type Event interface {
	EventName() woovi.WebhookEvent
}

// Envelope is the common webhook JSON shape (event + optional charge).
type Envelope struct {
	Event   woovi.WebhookEvent `json:"event"`
	Charge  *woovi.Charge      `json:"charge,omitempty"`
	Pix     json.RawMessage    `json:"pix,omitempty"`
	Company json.RawMessage    `json:"company,omitempty"`
	Account json.RawMessage    `json:"account,omitempty"`
	Raw     json.RawMessage    `json:"-"`
}

func (e Envelope) EventName() woovi.WebhookEvent { return e.Event }

// ChargeCompleted is OPENPIX:CHARGE_COMPLETED.
type ChargeCompleted struct {
	Envelope
}

// ChargeCreated is OPENPIX:CHARGE_CREATED.
type ChargeCreated struct {
	Envelope
}

// ChargeExpired is OPENPIX:CHARGE_EXPIRED.
type ChargeExpired struct {
	Envelope
}

// UnknownEvent is any event without a dedicated type yet.
type UnknownEvent struct {
	Envelope
}

func ParseEvent(body []byte) (Event, error) {
	var env Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("webhookverify: decode event: %w", err)
	}
	env.Raw = append(json.RawMessage(nil), body...)
	switch env.Event {
	case woovi.WebhookEventChargeCompleted, woovi.WebhookEventChargeCompletedNotSameCustomerPayer:
		return &ChargeCompleted{Envelope: env}, nil
	case woovi.WebhookEventChargeCreated:
		return &ChargeCreated{Envelope: env}, nil
	case woovi.WebhookEventChargeExpired:
		return &ChargeExpired{Envelope: env}, nil
	default:
		return &UnknownEvent{Envelope: env}, nil
	}
}
