package woovi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// AnticipationStatus is the lifecycle of a receivable anticipation.
type AnticipationStatus string

const (
	AnticipationStatusPending    AnticipationStatus = "PENDING"
	AnticipationStatusProcessing AnticipationStatus = "PROCESSING"
	AnticipationStatusConfirmed  AnticipationStatus = "CONFIRMED"
	AnticipationStatusCanceled   AnticipationStatus = "CANCELED"
	AnticipationStatusPaid       AnticipationStatus = "PAID"
	AnticipationStatusOverdue    AnticipationStatus = "OVERDUE"
	AnticipationStatusFailed     AnticipationStatus = "FAILED"
)

// AnticipationsService is /api/v1/anticipation.
type AnticipationsService struct {
	client *Client
}

// Anticipation is a receivable advance request (amounts in cents).
type Anticipation struct {
	ID                   string             `json:"id"`
	Status               AnticipationStatus `json:"status,omitempty"`
	BeneficiaryTaxID     string             `json:"beneficiaryTaxID,omitempty"`
	RequestedAmount      Money              `json:"requestedAmount,omitempty"`
	FeeAmount            Money              `json:"feeAmount,omitempty"`
	NetAmount            Money              `json:"netAmount,omitempty"`
	FeeMode              string             `json:"feeMode,omitempty"`
	MonthlyFeePercentage float64            `json:"monthlyFeePercentage,omitempty"`
	DaysUntilDue         int                `json:"daysUntilDue,omitempty"`
	DueDate              time.Time          `json:"dueDate,omitempty"`
	ApprovedAt           *time.Time         `json:"approvedAt"`
	CancelledAt          *time.Time         `json:"cancelledAt"`
	CancelReason         *string            `json:"cancelReason"`
	EndToEndID           *string            `json:"endToEndId"`
	FailureCode          *string            `json:"failureCode"`
	FailureReason        *string            `json:"failureReason"`
	CreatedAt            time.Time          `json:"createdAt,omitempty"`
}

// AnticipationList is GET /anticipation.
type AnticipationList struct {
	Anticipations []Anticipation `json:"anticipations"`
	Count         int            `json:"count,omitempty"`
}

// AnticipationListParams filters anticipations by status.
type AnticipationListParams struct {
	Status AnticipationStatus
	Limit  int
}

// AnticipationBeneficiaryTaxID is CPF/CNPJ on an anticipation beneficiary.
type AnticipationBeneficiaryTaxID struct {
	TaxID string `json:"taxID"`
	Type  string `json:"type,omitempty"`
}

// AnticipationBeneficiary can receive anticipations (amounts in cents).
type AnticipationBeneficiary struct {
	Name                 string                       `json:"name,omitempty"`
	TaxID                AnticipationBeneficiaryTaxID `json:"taxID,omitempty"`
	IsActive             bool                         `json:"isActive,omitempty"`
	AvailableAmount      Money                        `json:"availableAmount,omitempty"`
	MaxAdvanceableAmount Money                        `json:"maxAdvanceableAmount,omitempty"`
	NotifyEmail          string                       `json:"notifyEmail,omitempty"`
	NotifyPhone          string                       `json:"notifyPhone,omitempty"`
	Verified             bool                         `json:"verified,omitempty"`
	CreatedAt            time.Time                    `json:"createdAt,omitempty"`
}

// AnticipationBeneficiaryCreateParams registers a beneficiary (taxID required).
type AnticipationBeneficiaryCreateParams struct {
	Name                 string `json:"name"`
	TaxID                string `json:"taxID"`
	NotifyEmail          string `json:"notifyEmail,omitempty"`
	NotifyPhone          string `json:"notifyPhone,omitempty"`
	AvailableAmount      *Money `json:"availableAmount,omitempty"`
	MaxAdvanceableAmount *Money `json:"maxAdvanceableAmount,omitempty"`
	CorrelationID        string `json:"correlationID,omitempty"`
}

// AnticipationBeneficiaryCreateResult is POST /anticipation/beneficiary.
type AnticipationBeneficiaryCreateResult struct {
	Beneficiary   AnticipationBeneficiary `json:"beneficiary"`
	CorrelationID string                  `json:"correlationID,omitempty"`
}

// AnticipationBalanceItem sets available/max advanceable amounts (cents) for one taxID.
type AnticipationBalanceItem struct {
	TaxID                string `json:"taxID"`
	AvailableAmount      Money  `json:"availableAmount"`
	MaxAdvanceableAmount Money  `json:"maxAdvanceableAmount"`
}

// AnticipationBalanceBatchParams syncs balances for up to 1000 taxIDs.
type AnticipationBalanceBatchParams struct {
	Items []AnticipationBalanceItem `json:"items"`
}

// AnticipationBalanceBatchResultItem is one taxID outcome from balance sync.
type AnticipationBalanceBatchResultItem struct {
	TaxID string `json:"taxID"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// AnticipationBalanceBatchResult is POST /anticipation/balance/batch.
type AnticipationBalanceBatchResult struct {
	Processed int                                  `json:"processed"`
	Succeeded int                                  `json:"succeeded"`
	Failed    int                                  `json:"failed"`
	Results   []AnticipationBalanceBatchResultItem `json:"results"`
}

type anticipationEnvelope struct {
	Anticipation Anticipation `json:"anticipation"`
}

type anticipationBeneficiaryEnvelope struct {
	Beneficiary AnticipationBeneficiary `json:"beneficiary"`
}

func (s *AnticipationsService) List(ctx context.Context, params *AnticipationListParams, opts ...RequestOption) (*AnticipationList, error) {
	q := url.Values{}
	if params != nil {
		if params.Status != "" {
			q.Set("status", string(params.Status))
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
	}
	var out AnticipationList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/anticipation", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// Approve approves a PENDING anticipation and triggers Pix Out.
func (s *AnticipationsService) Approve(ctx context.Context, id string, opts ...RequestOption) (*Anticipation, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out anticipationEnvelope
	path := "/api/v1/anticipation/" + url.PathEscape(id) + "/approve"
	if err := s.client.do(ctx, http.MethodPost, path, nil, map[string]any{}, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Anticipation, nil
}

// Reject rejects a PENDING anticipation.
func (s *AnticipationsService) Reject(ctx context.Context, id string, reason string, opts ...RequestOption) (*Anticipation, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	body := map[string]any{}
	if reason != "" {
		body["reason"] = reason
	}
	var out anticipationEnvelope
	path := "/api/v1/anticipation/" + url.PathEscape(id) + "/reject"
	if err := s.client.do(ctx, http.MethodPost, path, nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Anticipation, nil
}

func (s *AnticipationsService) CreateBeneficiary(ctx context.Context, params *AnticipationBeneficiaryCreateParams, opts ...RequestOption) (*AnticipationBeneficiaryCreateResult, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: AnticipationBeneficiaryCreateParams is required")
	}
	if params.Name == "" || params.TaxID == "" {
		return nil, fmt.Errorf("woovi: name and taxID are required")
	}
	var out AnticipationBeneficiaryCreateResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/anticipation/beneficiary", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AnticipationsService) ActivateBeneficiary(ctx context.Context, taxID string, opts ...RequestOption) (*AnticipationBeneficiary, error) {
	return s.toggleBeneficiary(ctx, taxID, "activate", opts...)
}

func (s *AnticipationsService) DeactivateBeneficiary(ctx context.Context, taxID string, opts ...RequestOption) (*AnticipationBeneficiary, error) {
	return s.toggleBeneficiary(ctx, taxID, "deactivate", opts...)
}

func (s *AnticipationsService) toggleBeneficiary(ctx context.Context, taxID, action string, opts ...RequestOption) (*AnticipationBeneficiary, error) {
	if taxID == "" {
		return nil, fmt.Errorf("woovi: taxID is required")
	}
	var out anticipationBeneficiaryEnvelope
	path := "/api/v1/anticipation/beneficiary/" + url.PathEscape(taxID) + "/" + action
	if err := s.client.do(ctx, http.MethodPost, path, nil, map[string]any{}, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Beneficiary, nil
}

func (s *AnticipationsService) SyncBalances(ctx context.Context, params *AnticipationBalanceBatchParams, opts ...RequestOption) (*AnticipationBalanceBatchResult, error) {
	if params == nil || len(params.Items) == 0 {
		return nil, fmt.Errorf("woovi: items are required")
	}
	if len(params.Items) > 1000 {
		return nil, fmt.Errorf("woovi: items must be at most 1000")
	}
	var out AnticipationBalanceBatchResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/anticipation/balance/batch", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}
