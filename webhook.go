package woovi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// WebhookEvent is an OpenPix webhook event name.
type WebhookEvent string

const (
	WebhookEventChargeCreated                       WebhookEvent = "OPENPIX:CHARGE_CREATED"
	WebhookEventChargeCompleted                     WebhookEvent = "OPENPIX:CHARGE_COMPLETED"
	WebhookEventChargeCompletedNotSameCustomerPayer WebhookEvent = "OPENPIX:CHARGE_COMPLETED_NOT_SAME_CUSTOMER_PAYER"
	WebhookEventChargeExpired                       WebhookEvent = "OPENPIX:CHARGE_EXPIRED"
	WebhookEventTransactionReceived                 WebhookEvent = "OPENPIX:TRANSACTION_RECEIVED"
	WebhookEventTransactionRefundReceived           WebhookEvent = "OPENPIX:TRANSACTION_REFUND_RECEIVED"
	WebhookEventMovementConfirmed                   WebhookEvent = "OPENPIX:MOVEMENT_CONFIRMED"
	WebhookEventMovementFailed                      WebhookEvent = "OPENPIX:MOVEMENT_FAILED"
	WebhookEventMovementRemoved                     WebhookEvent = "OPENPIX:MOVEMENT_REMOVED"
	WebhookEventSubscriptionAuthorized              WebhookEvent = "OPENPIX:SUBSCRIPTION_AUTHORIZED"
	WebhookEventSubscriptionRejected                WebhookEvent = "OPENPIX:SUBSCRIPTION_REJECTED"
	WebhookEventSubscriptionCancelled               WebhookEvent = "OPENPIX:SUBSCRIPTION_CANCELLED"
	WebhookEventSubaccountCreated                   WebhookEvent = "OPENPIX:SUBACCOUNT_CREATED"
	WebhookEventRefundCreated                       WebhookEvent = "OPENPIX:REFUND_CREATED"
	WebhookEventDisputeCreated                      WebhookEvent = "OPENPIX:DISPUTE_CREATED"
	WebhookEventDisputeAccepted                     WebhookEvent = "OPENPIX:DISPUTE_ACCEPTED"
	WebhookEventDisputeRejected                     WebhookEvent = "OPENPIX:DISPUTE_REJECTED"
	WebhookEventDisputeCanceled                     WebhookEvent = "OPENPIX:DISPUTE_CANCELED"
	WebhookEventCashbackGiven                       WebhookEvent = "OPENPIX:CASHBACK_GIVEN"
	WebhookEventCashbackRedeemed                    WebhookEvent = "OPENPIX:CASHBACK_REDEEMED"
	WebhookEventTEDOutConfirmed                     WebhookEvent = "OPENPIX:TED_OUT_CONFIRMED"
	WebhookEventTEDOutRejected                      WebhookEvent = "OPENPIX:TED_OUT_REJECTED"
	WebhookEventTEDInConfirmed                      WebhookEvent = "OPENPIX:TED_IN_CONFIRMED"
	WebhookEventTEDRefundSentConfirmed              WebhookEvent = "OPENPIX:TED_REFUND_SENT_CONFIRMED"
	WebhookEventBoletoSettled                       WebhookEvent = "OPENPIX:BOLETO_SETTLED"
)

// WebhooksService is /api/v1/webhook.
type WebhooksService struct {
	client *Client
}

// Webhook is a registered OpenPix webhook endpoint.
type Webhook struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	URL           string       `json:"url"`
	Authorization string       `json:"authorization,omitempty"`
	IsActive      bool         `json:"isActive"`
	Event         WebhookEvent `json:"event"`
	CreatedAt     string       `json:"createdAt,omitempty"`
	UpdatedAt     string       `json:"updatedAt,omitempty"`
}

// WebhookCreateParams registers a webhook for one event name.
type WebhookCreateParams struct {
	Name          string       `json:"name"`
	Event         WebhookEvent `json:"event"`
	URL           string       `json:"url"`
	Authorization string       `json:"authorization,omitempty"`
	IsActive      bool         `json:"isActive"`
}

type webhookCreateBody struct {
	Webhook WebhookCreateParams `json:"webhook"`
}

type webhookCreateResult struct {
	Webhook Webhook `json:"webhook"`
}

type webhookListResult struct {
	Webhooks []Webhook `json:"webhooks"`
	PageInfo PageInfo  `json:"pageInfo"`
}

func (s *WebhooksService) Create(ctx context.Context, params *WebhookCreateParams, opts ...RequestOption) (*Webhook, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: WebhookCreateParams is required")
	}
	if params.Name == "" || params.URL == "" || params.Event == "" {
		return nil, fmt.Errorf("woovi: name, url, and event are required")
	}
	var out webhookCreateResult
	body := webhookCreateBody{Webhook: *params}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/webhook", nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Webhook, nil
}

func (s *WebhooksService) List(ctx context.Context, opts ...RequestOption) ([]Webhook, error) {
	var out webhookListResult
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/webhook", nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return out.Webhooks, nil
}

func (s *WebhooksService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	if id == "" {
		return fmt.Errorf("woovi: id is required")
	}
	path := "/api/v1/webhook/" + url.PathEscape(id)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil, nil, opts...)
}

func (s *WebhooksService) ListEvents(ctx context.Context, opts ...RequestOption) ([]WebhookEvent, error) {
	var out struct {
		Events []WebhookEvent `json:"events"`
	}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/webhook/events", nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return out.Events, nil
}

// ListIPs returns outbound IPs used by Woovi webhooks.
func (s *WebhooksService) ListIPs(ctx context.Context, opts ...RequestOption) ([]string, error) {
	var out struct {
		IPs []string `json:"ips"`
		IP  []string `json:"ip"`
	}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/webhook/ips", nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	if len(out.IPs) > 0 {
		return out.IPs, nil
	}
	return out.IP, nil
}

// ToggleActivation toggles webhook active state (GET).
func (s *WebhooksService) ToggleActivation(ctx context.Context, id string, opts ...RequestOption) (*Webhook, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out webhookCreateResult
	path := "/api/v1/webhook/" + url.PathEscape(id) + "/toggle-activation"
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Webhook, nil
}

// WebhookPublicKey is an RSA key for webhook signature verification.
type WebhookPublicKey struct {
	KeyIdentifier string `json:"key_identifier"`
	Key           string `json:"key"`
	IsCurrent     bool   `json:"is_current"`
}

// ListPublicKeys returns RSA public keys used to sign webhooks.
func (s *WebhooksService) ListPublicKeys(ctx context.Context, opts ...RequestOption) ([]WebhookPublicKey, error) {
	var out struct {
		PublicKeys []WebhookPublicKey `json:"public_keys"`
	}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/webhook/public-keys", nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return out.PublicKeys, nil
}
