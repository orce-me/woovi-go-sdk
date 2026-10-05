package woovi

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
)

// SubscriptionStatus is ACTIVE or INACTIVE.
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "ACTIVE"
	SubscriptionStatusInactive SubscriptionStatus = "INACTIVE"
)

// SubscriptionType selects subscription kind (RECURRENT).
type SubscriptionType string

const (
	SubscriptionTypeRecurrent SubscriptionType = "RECURRENT"
)

// SubscriptionChargeType is PIX or BOLETO for each cycle.
type SubscriptionChargeType string

const (
	SubscriptionChargeTypePix    SubscriptionChargeType = "PIX"
	SubscriptionChargeTypeBoleto SubscriptionChargeType = "BOLETO"
)

// SubscriptionsService is /api/v1/subscription.
type SubscriptionsService struct {
	client *Client
}

// SubscriptionCustomerInput is required customer data on subscription create.
type SubscriptionCustomerInput struct {
	Name    string   `json:"name"`
	TaxID   string   `json:"taxID"`
	Email   string   `json:"email"`
	Phone   string   `json:"phone"`
	Address *Address `json:"address,omitempty"`
}

// SubscriptionCreateParams creates a subscription (value in cents).
type SubscriptionCreateParams struct {
	Value             Money                     `json:"value"`
	Customer          SubscriptionCustomerInput `json:"customer"`
	DayGenerateCharge *int                      `json:"dayGenerateCharge,omitempty"`
	CorrelationID     string                    `json:"correlationID,omitempty"`
	Type              SubscriptionType          `json:"type,omitempty"`
	ChargeType        SubscriptionChargeType    `json:"chargeType,omitempty"`
}

// PixRecurring is Pix Automático metadata on a subscription.
type PixRecurring struct {
	RecurrencyID string `json:"recurrencyId,omitempty"`
	EMV          string `json:"emv,omitempty"`
	Journey      string `json:"journey,omitempty"`
	Status       string `json:"status,omitempty"`
}

// Subscription is a recurring charge plan (value in cents; globalID).
type Subscription struct {
	GlobalID          string                 `json:"globalID"`
	Value             Money                  `json:"value"`
	Customer          *Customer              `json:"customer,omitempty"`
	DayGenerateCharge int                    `json:"dayGenerateCharge,omitempty"`
	Status            SubscriptionStatus     `json:"status,omitempty"`
	CorrelationID     string                 `json:"correlationID,omitempty"`
	Type              SubscriptionType       `json:"type,omitempty"`
	ChargeType        SubscriptionChargeType `json:"chargeType,omitempty"`
	PixRecurring      *PixRecurring          `json:"pixRecurring,omitempty"`
}

// SubscriptionPaymentBook is the payment-book URL for boleto subscriptions.
type SubscriptionPaymentBook struct {
	URL string          `json:"url,omitempty"`
	Raw json.RawMessage `json:"-"`
}

// SubscriptionListParams paginates GET /subscription.
type SubscriptionListParams struct {
	Skip  int
	Limit int
}

// SubscriptionList is one page of GET /subscription.
type SubscriptionList struct {
	Subscriptions []Subscription `json:"subscriptions"`
	PageInfo      PageInfo       `json:"pageInfo"`
	Skip          int
	Limit         int
}

type subscriptionEnvelope struct {
	Subscription Subscription `json:"subscription"`
}

func (s *SubscriptionsService) Create(ctx context.Context, params *SubscriptionCreateParams, opts ...RequestOption) (*Subscription, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: SubscriptionCreateParams is required")
	}
	if params.Value <= 0 {
		return nil, fmt.Errorf("woovi: value must be greater than zero")
	}
	if params.Customer.Name == "" || params.Customer.TaxID == "" {
		return nil, fmt.Errorf("woovi: customer name and taxID are required")
	}
	if params.DayGenerateCharge != nil {
		d := *params.DayGenerateCharge
		if d < 0 || d > 27 {
			return nil, fmt.Errorf("woovi: dayGenerateCharge must be between 0 and 27")
		}
	}

	var out subscriptionEnvelope
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/subscriptions", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Subscription, nil
}

func (s *SubscriptionsService) Get(ctx context.Context, globalID string, opts ...RequestOption) (*Subscription, error) {
	if globalID == "" {
		return nil, fmt.Errorf("woovi: globalID is required")
	}
	var out subscriptionEnvelope
	path := "/api/v1/subscriptions/" + url.PathEscape(globalID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Subscription, nil
}

func (s *SubscriptionsService) Cancel(ctx context.Context, id string, opts ...RequestOption) (*Subscription, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out subscriptionEnvelope
	path := "/api/v1/subscriptions/" + url.PathEscape(id) + "/cancel"
	if err := s.client.do(ctx, http.MethodPut, path, nil, map[string]any{}, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Subscription, nil
}

func (s *SubscriptionsService) UpdateValue(ctx context.Context, id string, value Money, opts ...RequestOption) (*Subscription, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	if value <= 0 {
		return nil, fmt.Errorf("woovi: value must be greater than zero")
	}
	var out subscriptionEnvelope
	path := "/api/v1/subscriptions/" + url.PathEscape(id) + "/value"
	body := map[string]Money{"value": value}
	if err := s.client.do(ctx, http.MethodPut, path, nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Subscription, nil
}

func (s *SubscriptionsService) PaymentBook(ctx context.Context, id string, opts ...RequestOption) (*SubscriptionPaymentBook, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/subscriptions/" + url.PathEscape(id) + "/payment-book"
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	out := &SubscriptionPaymentBook{Raw: append(json.RawMessage(nil), raw...)}
	_ = json.Unmarshal(raw, out)
	return out, nil
}

func (s *SubscriptionsService) List(ctx context.Context, params *SubscriptionListParams, opts ...RequestOption) (*SubscriptionList, error) {
	if params == nil {
		params = &SubscriptionListParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))

	var out SubscriptionList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/subscriptions", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func (s *SubscriptionsService) ListAll(ctx context.Context, params *SubscriptionListParams, opts ...RequestOption) iter.Seq2[*Subscription, error] {
	return func(yield func(*Subscription, error) bool) {
		for page, err := range s.ListPages(ctx, params, opts...) {
			if err != nil {
				yield(nil, err)
				return
			}
			for i := range page.Subscriptions {
				if !yield(&page.Subscriptions[i], nil) {
					return
				}
			}
		}
	}
}

func (s *SubscriptionsService) ListPages(ctx context.Context, params *SubscriptionListParams, opts ...RequestOption) iter.Seq2[*SubscriptionList, error] {
	return func(yield func(*SubscriptionList, error) bool) {
		p := SubscriptionListParams{}
		if params != nil {
			p = *params
		}
		if p.Limit <= 0 {
			p.Limit = 100
		}
		for {
			if err := ctx.Err(); err != nil {
				yield(nil, err)
				return
			}
			page, err := s.List(ctx, &p, opts...)
			if err != nil {
				yield(nil, err)
				return
			}
			if !yield(page, nil) {
				return
			}
			if !listPageHasMore(page.PageInfo, len(page.Subscriptions), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}
