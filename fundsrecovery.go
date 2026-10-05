package woovi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// FundsRecoverySituationType is the fraud situation for a MED recovery.
type FundsRecoverySituationType string

const (
	FundsRecoverySituationScam             FundsRecoverySituationType = "SCAM"
	FundsRecoverySituationAccountTakeover  FundsRecoverySituationType = "ACCOUNT_TAKEOVER"
	FundsRecoverySituationCoercion         FundsRecoverySituationType = "COERCION"
	FundsRecoverySituationFraudulentAccess FundsRecoverySituationType = "FRAUDULENT_ACCESS"
	FundsRecoverySituationOther            FundsRecoverySituationType = "OTHER"
	FundsRecoverySituationUnknown          FundsRecoverySituationType = "UNKNOWN"
)

// FundsRecoveryStatus is a MED recovery status.
type FundsRecoveryStatus string

const (
	FundsRecoveryStatusCreated   FundsRecoveryStatus = "CREATED"
	FundsRecoveryStatusCompleted FundsRecoveryStatus = "COMPLETED"
	FundsRecoveryStatusCancelled FundsRecoveryStatus = "CANCELLED"
)

// FundsRecoveriesService is /api/v1/funds-recovery.
type FundsRecoveriesService struct {
	client *Client
}

// FundsRecoveryCreateParams opens a MED recovery for a transaction endToEndId.
type FundsRecoveryCreateParams struct {
	TransactionEndToEndID string                     `json:"transactionEndToEndId"`
	SituationType         FundsRecoverySituationType `json:"situationType"`
	Details               string                     `json:"details"`
}

// FundsRecovery is a MED funds recovery.
type FundsRecovery struct {
	RootTransactionID   string                     `json:"rootTransactionId,omitempty"`
	SituationType       FundsRecoverySituationType `json:"situationType,omitempty"`
	ReportDetails       string                     `json:"reportDetails,omitempty"`
	DictID              string                     `json:"dictId,omitempty"`
	Status              FundsRecoveryStatus        `json:"status,omitempty"`
	Direction           string                     `json:"direction,omitempty"`
	ReporterParticipant string                     `json:"reporterParticipant,omitempty"`
	CreationTime        time.Time                  `json:"creationTime,omitempty"`
	LastModified        time.Time                  `json:"lastModified,omitempty"`
	Events              []json.RawMessage          `json:"events,omitempty"`
	CreatedAt           time.Time                  `json:"createdAt,omitempty"`
	UpdatedAt           time.Time                  `json:"updatedAt,omitempty"`
}

// FundsRecoveryDispute is a dispute linked to a funds recovery leg.
type FundsRecoveryDispute struct {
	ID              string        `json:"id,omitempty"`
	Type            DisputeType   `json:"type,omitempty"`
	Status          DisputeStatus `json:"status,omitempty"`
	SentOrReceived  string        `json:"sentOrReceived,omitempty"`
	SituationType   string        `json:"situationType,omitempty"`
	Value           Money         `json:"value,omitempty"`
	EndToEndID      string        `json:"endToEndId,omitempty"`
	FundsRecoveryID string        `json:"fundsRecoveryId,omitempty"`
	DisputeReason   string        `json:"disputeReason,omitempty"`
	CreatedAt       time.Time     `json:"createdAt,omitempty"`
	UpdatedAt       time.Time     `json:"updatedAt,omitempty"`
}

// FundsRecoveryInfractionReport is a Bacen infraction report on a recovery leg.
type FundsRecoveryInfractionReport struct {
	BacenInfractionReportID string    `json:"bacenInfractionReportId,omitempty"`
	FundsRecoveryID         string    `json:"fundsRecoveryId,omitempty"`
	TransactionID           string    `json:"transactionId,omitempty"`
	Reason                  string    `json:"reason,omitempty"`
	SituationType           string    `json:"situationType,omitempty"`
	ReportDetails           string    `json:"reportDetails,omitempty"`
	Status                  string    `json:"status,omitempty"`
	ReporterParticipant     string    `json:"reporterParticipant,omitempty"`
	CounterpartyParticipant string    `json:"counterpartyParticipant,omitempty"`
	FraudMarkerID           string    `json:"fraudMarkerId,omitempty"`
	AnalysisResult          string    `json:"analysisResult,omitempty"`
	AnalysisDetails         string    `json:"analysisDetails,omitempty"`
	InfractionAmount        Money     `json:"infractionAmount,omitempty"`
	CreationTime            time.Time `json:"creationTime,omitempty"`
	LastModified            time.Time `json:"lastModified,omitempty"`
	CreatedAt               time.Time `json:"createdAt,omitempty"`
	UpdatedAt               time.Time `json:"updatedAt,omitempty"`
}

// FundsRecoveryRefundSolicitation is a Bacen refund request on a recovery leg.
type FundsRecoveryRefundSolicitation struct {
	BacenRefundID           string    `json:"bacenRefundId,omitempty"`
	FundsRecoveryID         string    `json:"fundsRecoveryId,omitempty"`
	EndToEndID              string    `json:"endToEndId,omitempty"`
	RefundTransactionID     string    `json:"refundTransactionId,omitempty"`
	RefundReason            string    `json:"refundReason,omitempty"`
	RefundAmount            Money     `json:"refundAmount,omitempty"`
	EffectiveRefundedAmount Money     `json:"EffectiveRefundedAmount,omitempty"`
	RefundedAt              time.Time `json:"refundedAt,omitempty"`
	Status                  string    `json:"status,omitempty"`
	AnalysisResult          string    `json:"analysisResult,omitempty"`
	RejectionReason         string    `json:"rejectionReason,omitempty"`
	ContestedParticipant    string    `json:"contestedParticipant,omitempty"`
	RequestingParticipant   string    `json:"requestingParticipant,omitempty"`
	CreationTime            time.Time `json:"creationTime,omitempty"`
	LastModified            time.Time `json:"lastModified,omitempty"`
	CreatedAt               time.Time `json:"createdAt,omitempty"`
	UpdatedAt               time.Time `json:"updatedAt,omitempty"`
}

func (s *FundsRecoveriesService) Create(ctx context.Context, params *FundsRecoveryCreateParams, opts ...RequestOption) (*FundsRecovery, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: FundsRecoveryCreateParams is required")
	}
	if params.TransactionEndToEndID == "" {
		return nil, fmt.Errorf("woovi: transactionEndToEndId is required")
	}
	if params.SituationType == "" {
		return nil, fmt.Errorf("woovi: situationType is required")
	}
	if params.Details == "" {
		return nil, fmt.Errorf("woovi: details is required")
	}
	var out FundsRecovery
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/funds-recovery", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a funds recovery by dictId (or rootTransactionId).
func (s *FundsRecoveriesService) Get(ctx context.Context, id string, opts ...RequestOption) (*FundsRecovery, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out FundsRecovery
	path := "/api/v1/funds-recovery/" + url.PathEscape(id)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *FundsRecoveriesService) Cancel(ctx context.Context, id string, opts ...RequestOption) (*FundsRecovery, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out FundsRecovery
	path := "/api/v1/funds-recovery/" + url.PathEscape(id) + "/cancel"
	if err := s.client.do(ctx, http.MethodPost, path, nil, map[string]any{}, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *FundsRecoveriesService) ListDisputes(ctx context.Context, id string, opts ...RequestOption) ([]FundsRecoveryDispute, error) {
	return decodeFundsRecoveryList[FundsRecoveryDispute](s, ctx, id, "disputes", opts...)
}

// ListInfractionReports lists Bacen infraction reports for a funds recovery.
func (s *FundsRecoveriesService) ListInfractionReports(ctx context.Context, id string, opts ...RequestOption) ([]FundsRecoveryInfractionReport, error) {
	return decodeFundsRecoveryList[FundsRecoveryInfractionReport](s, ctx, id, "infraction-reports", opts...)
}

// ListRefundSolicitations lists Bacen refund solicitations for a funds recovery.
func (s *FundsRecoveriesService) ListRefundSolicitations(ctx context.Context, id string, opts ...RequestOption) ([]FundsRecoveryRefundSolicitation, error) {
	return decodeFundsRecoveryList[FundsRecoveryRefundSolicitation](s, ctx, id, "refund-solicitations", opts...)
}

func decodeFundsRecoveryList[T any](s *FundsRecoveriesService, ctx context.Context, id, suffix string, opts ...RequestOption) ([]T, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/funds-recovery/" + url.PathEscape(id) + "/" + suffix
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	var arr []T
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var wrapped map[string]json.RawMessage
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, fmt.Errorf("woovi: decode funds recovery %s: %w", suffix, err)
	}
	for _, v := range wrapped {
		if err := json.Unmarshal(v, &arr); err == nil && len(arr) > 0 {
			return arr, nil
		}
	}
	var one T
	if err := json.Unmarshal(raw, &one); err == nil {
		return []T{one}, nil
	}
	return []T{}, nil
}
