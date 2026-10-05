package woovi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// DisputeStatus is a MED dispute lifecycle status.
type DisputeStatus string

const (
	DisputeStatusInReview DisputeStatus = "IN_REVIEW"
	DisputeStatusAccepted DisputeStatus = "ACCEPTED"
	DisputeStatusRejected DisputeStatus = "REJECTED"
	DisputeStatusCanceled DisputeStatus = "CANCELED"
)

// DisputeType is the dispute kind (MED).
type DisputeType string

const (
	DisputeTypeMED DisputeType = "MED"
)

// DisputesService is /api/v1/dispute.
type DisputesService struct {
	client *Client
}

// Dispute is a MED dispute (value in cents; id or endToEndId).
type Dispute struct {
	ID            string        `json:"id,omitempty"`
	Status        DisputeStatus `json:"status,omitempty"`
	Name          string        `json:"name,omitempty"`
	Email         string        `json:"email,omitempty"`
	PhoneNumber   string        `json:"phoneNumber,omitempty"`
	Value         Money         `json:"value,omitempty"`
	DisputeReason string        `json:"disputeReason,omitempty"`
	EndToEndID    string        `json:"endToEndId,omitempty"`
	Type          DisputeType   `json:"type,omitempty"`
	CreatedAt     time.Time     `json:"createdAt,omitempty"`
	UpdatedAt     time.Time     `json:"updatedAt,omitempty"`
}

// DisputeListParams filters disputes by RFC3339 start/end.
type DisputeListParams struct {
	Start string // RFC3339
	End   string // RFC3339
}

// DisputeList is GET /dispute.
type DisputeList struct {
	Disputes []Dispute `json:"disputes"`
	PageInfo PageInfo  `json:"pageInfo"`
}

// DisputeEvidenceDocument: set URL or FileID, never both.
type DisputeEvidenceDocument struct {
	URL           string `json:"url,omitempty"`
	FileID        string `json:"fileId,omitempty"`
	Description   string `json:"description,omitempty"`
	CorrelationID string `json:"correlationID,omitempty"`
	CreatedAt     string `json:"createdAt,omitempty"`
}

// DisputeEvidenceParams attaches evidence documents to a dispute.
type DisputeEvidenceParams struct {
	Documents []DisputeEvidenceDocument `json:"documents"`
}

// DisputeEvidenceResult is POST /dispute/{id}/evidence.
type DisputeEvidenceResult struct {
	Documents []DisputeEvidenceDocument `json:"documents"`
}

type disputeEnvelope struct {
	Dispute Dispute `json:"dispute"`
}

// Get retrieves a dispute by id or transaction endToEndId.
func (s *DisputesService) Get(ctx context.Context, id string, opts ...RequestOption) (*Dispute, error) {
	if id == "" {
		return nil, fmt.Errorf("woovi: id is required")
	}
	var out disputeEnvelope
	path := "/api/v1/dispute/" + url.PathEscape(id)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Dispute, nil
}

func (s *DisputesService) List(ctx context.Context, params *DisputeListParams, opts ...RequestOption) (*DisputeList, error) {
	q := url.Values{}
	if params != nil {
		if params.Start != "" {
			q.Set("start", params.Start)
		}
		if params.End != "" {
			q.Set("end", params.End)
		}
	}
	var out DisputeList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/dispute", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddEvidence attaches evidence. Prefer Files.Upload (DISPUTE_EVIDENCE), then FileID.
func (s *DisputesService) AddEvidence(ctx context.Context, disputeID string, params *DisputeEvidenceParams, opts ...RequestOption) (*DisputeEvidenceResult, error) {
	if disputeID == "" {
		return nil, fmt.Errorf("woovi: disputeID is required")
	}
	if params == nil || len(params.Documents) == 0 {
		return nil, fmt.Errorf("woovi: documents are required")
	}
	for i, doc := range params.Documents {
		hasURL := doc.URL != ""
		hasFile := doc.FileID != ""
		if hasURL == hasFile {
			return nil, fmt.Errorf("woovi: documents[%d] must set url or fileId, not both", i)
		}
	}
	var out DisputeEvidenceResult
	path := "/api/v1/dispute/" + url.PathEscape(disputeID) + "/evidence"
	if err := s.client.do(ctx, http.MethodPost, path, nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}
