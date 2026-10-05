package woovi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// KYCService is /api/v1/kyc.
type KYCService struct {
	client *Client
}

// KYCTaxID is CPF/CNPJ on KYC payloads.
type KYCTaxID struct {
	TaxID string `json:"taxID"`
	Type  string `json:"type,omitempty"`
}

// KYCOnboardingParams starts hosted KYC for a merchant.
type KYCOnboardingParams struct {
	TaxID               string         `json:"taxID,omitempty"`
	CorrelationID       string         `json:"correlationID,omitempty"`
	RedirectURL         string         `json:"redirectUrl,omitempty"`
	Website             string         `json:"website,omitempty"`
	BusinessDescription string         `json:"businessDescription,omitempty"`
	Partner             *bool          `json:"partner,omitempty"`
	Company             map[string]any `json:"company,omitempty"`
	User                map[string]any `json:"user,omitempty"`
}

// KYCAccountRegister is the merchant register created by KYC onboarding.
type KYCAccountRegister struct {
	Status              string              `json:"status,omitempty"`
	OfficialName        string              `json:"officialName,omitempty"`
	TradeName           string              `json:"tradeName,omitempty"`
	TaxID               *KYCTaxID           `json:"taxID,omitempty"`
	CorrelationID       string              `json:"correlationID,omitempty"`
	Website             string              `json:"website,omitempty"`
	BusinessDescription string              `json:"businessDescription,omitempty"`
	Representatives     []KYCRepresentative `json:"representatives,omitempty"`
}

// KYCOnboardingResult is POST /kyc/onboarding.
type KYCOnboardingResult struct {
	LinkOnboarding  string              `json:"linkOnboarding,omitempty"`
	RedirectURL     string              `json:"redirectUrl,omitempty"`
	AccountRegister *KYCAccountRegister `json:"accountRegister,omitempty"`
}

// KYCSubmitParams submits an onboarding register for review.
type KYCSubmitParams struct {
	CorrelationID string `json:"correlationID"`
}

// KYCSubmitResult is POST /kyc/onboarding/submit.
type KYCSubmitResult struct {
	CorrelationID string    `json:"correlationID,omitempty"`
	Status        string    `json:"status,omitempty"`
	InReviewAt    time.Time `json:"inReviewAt,omitempty"`
}

// KYCDocumentInput references a previously uploaded file.
type KYCDocumentInput struct {
	Type   string `json:"type"`
	FileID string `json:"fileId"`
}

// KYCRejectedDocument is a document rejected during representative create.
type KYCRejectedDocument struct {
	Type   string `json:"type,omitempty"`
	FileID string `json:"fileId,omitempty"`
	Error  string `json:"error,omitempty"`
}

// KYCRepresentativeDocument is an identity document on a representative.
type KYCRepresentativeDocument struct {
	Type     string `json:"type,omitempty"`
	FileName string `json:"fileName,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	URL      string `json:"url,omitempty"`
}

// KYCRepresentative is a company legal representative in KYC.
type KYCRepresentative struct {
	ID                        string                      `json:"id,omitempty"`
	Name                      string                      `json:"name,omitempty"`
	TaxID                     *KYCTaxID                   `json:"taxID,omitempty"`
	Type                      string                      `json:"type,omitempty"`
	Active                    bool                        `json:"active,omitempty"`
	BirthDate                 string                      `json:"birthDate,omitempty"`
	Email                     string                      `json:"email,omitempty"`
	Phone                     string                      `json:"phone,omitempty"`
	Source                    string                      `json:"source,omitempty"`
	PixAuthenticationVerified bool                        `json:"pixAuthenticationVerified,omitempty"`
	Documents                 []KYCRepresentativeDocument `json:"documents,omitempty"`
	RequestDocuments          []string                    `json:"requestDocuments,omitempty"`
	Steps                     []string                    `json:"steps,omitempty"`
}

// KYCRepresentativeCreateParams adds a representative to an onboarding (correlationID).
type KYCRepresentativeCreateParams struct {
	CorrelationID string             `json:"correlationID"`
	Name          string             `json:"name"`
	TaxID         string             `json:"taxID"`
	Type          string             `json:"type,omitempty"`
	BirthDate     string             `json:"birthDate,omitempty"`
	Email         string             `json:"email,omitempty"`
	Phone         string             `json:"phone,omitempty"`
	Documents     []KYCDocumentInput `json:"documents,omitempty"`
}

// KYCRepresentativeCreateResult is POST /kyc/representatives.
type KYCRepresentativeCreateResult struct {
	Representative    KYCRepresentative     `json:"representative"`
	Reactivated       bool                  `json:"reactivated,omitempty"`
	RejectedDocuments []KYCRejectedDocument `json:"rejectedDocuments,omitempty"`
}

// KYCRepresentativeDocumentsParams attaches identity/selfie files.
type KYCRepresentativeDocumentsParams struct {
	CorrelationID    string             `json:"correlationID"`
	RepresentativeID string             `json:"representativeId"`
	Documents        []KYCDocumentInput `json:"documents"`
}

// KYCCompanyDocumentsParams attaches company documents.
type KYCCompanyDocumentsParams struct {
	CorrelationID string             `json:"correlationID"`
	Documents     []KYCDocumentInput `json:"documents"`
}

// KYCPixAuthenticationParams starts Pix auth for a representative.
type KYCPixAuthenticationParams struct {
	CorrelationID    string `json:"correlationID"`
	RepresentativeID string `json:"representativeId,omitempty"`
}

// KYCRFIAnswerParams answers FILE items on an open RFI.
type KYCRFIAnswerParams struct {
	CorrelationID string             `json:"correlationID"`
	Documents     []KYCDocumentInput `json:"documents"`
}

// KYCValidationCreateParams screens a CPF/CNPJ.
type KYCValidationCreateParams struct {
	TaxID         string `json:"taxID"`
	CorrelationID string `json:"correlationID,omitempty"`
}

// KYCValidation is a CPF/CNPJ screening result (extra keys stay in Raw).
type KYCValidation struct {
	CorrelationID string          `json:"correlationID,omitempty"`
	Status        string          `json:"status,omitempty"`
	TaxID         string          `json:"taxID,omitempty"`
	Raw           json.RawMessage `json:"-"`
}

func (s *KYCService) CreateOnboarding(ctx context.Context, params *KYCOnboardingParams, opts ...RequestOption) (*KYCOnboardingResult, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: KYCOnboardingParams is required")
	}
	var out KYCOnboardingResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/kyc/onboarding", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *KYCService) SubmitOnboarding(ctx context.Context, params *KYCSubmitParams, opts ...RequestOption) (*KYCSubmitResult, error) {
	if params == nil || params.CorrelationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	var out KYCSubmitResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/kyc/onboarding/submit", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *KYCService) ListRepresentatives(ctx context.Context, correlationID string, opts ...RequestOption) ([]KYCRepresentative, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	q := url.Values{}
	q.Set("correlationID", correlationID)
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/kyc/representatives", q, nil, &raw, opts...); err != nil {
		return nil, err
	}
	var wrapped struct {
		Representatives []KYCRepresentative `json:"representatives"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && len(wrapped.Representatives) > 0 {
		return wrapped.Representatives, nil
	}
	var arr []KYCRepresentative
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	return []KYCRepresentative{}, nil
}

func (s *KYCService) CreateRepresentative(ctx context.Context, params *KYCRepresentativeCreateParams, opts ...RequestOption) (*KYCRepresentativeCreateResult, error) {
	if params == nil || params.CorrelationID == "" || params.Name == "" || params.TaxID == "" {
		return nil, fmt.Errorf("woovi: correlationID, name and taxID are required")
	}
	var out KYCRepresentativeCreateResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/kyc/representatives", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *KYCService) AttachRepresentativeDocuments(ctx context.Context, params *KYCRepresentativeDocumentsParams, opts ...RequestOption) (*KYCRepresentativeCreateResult, error) {
	if params == nil || params.CorrelationID == "" || params.RepresentativeID == "" || len(params.Documents) == 0 {
		return nil, fmt.Errorf("woovi: correlationID, representativeId and documents are required")
	}
	var out KYCRepresentativeCreateResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/kyc/representatives/documents", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *KYCService) ListDocuments(ctx context.Context, correlationID string, opts ...RequestOption) (json.RawMessage, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	q := url.Values{}
	q.Set("correlationID", correlationID)
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/kyc/documents", q, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *KYCService) AttachDocuments(ctx context.Context, params *KYCCompanyDocumentsParams, opts ...RequestOption) (json.RawMessage, error) {
	if params == nil || params.CorrelationID == "" || len(params.Documents) == 0 {
		return nil, fmt.Errorf("woovi: correlationID and documents are required")
	}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/kyc/documents", nil, params, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *KYCService) GetRFI(ctx context.Context, correlationID string, opts ...RequestOption) (json.RawMessage, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	q := url.Values{}
	q.Set("correlationID", correlationID)
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/kyc/rfi", q, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

// AnswerRFI answers FILE items on an open RFI.
func (s *KYCService) AnswerRFI(ctx context.Context, params *KYCRFIAnswerParams, opts ...RequestOption) (json.RawMessage, error) {
	if params == nil || params.CorrelationID == "" || len(params.Documents) == 0 {
		return nil, fmt.Errorf("woovi: correlationID and documents are required")
	}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/kyc/rfi", nil, params, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *KYCService) CreatePixAuthentication(ctx context.Context, params *KYCPixAuthenticationParams, opts ...RequestOption) (json.RawMessage, error) {
	if params == nil || params.CorrelationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/kyc/pix-authentication", nil, params, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *KYCService) GetPixAuthentication(ctx context.Context, pixAuthenticationID string, opts ...RequestOption) (json.RawMessage, error) {
	if pixAuthenticationID == "" {
		return nil, fmt.Errorf("woovi: pixAuthenticationID is required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/kyc/pix-authentication/" + url.PathEscape(pixAuthenticationID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

// GetBCProtection reads the BC Protege+ gate for a register.
func (s *KYCService) GetBCProtection(ctx context.Context, correlationID string, opts ...RequestOption) (json.RawMessage, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	q := url.Values{}
	q.Set("correlationID", correlationID)
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/kyc/bc-protection", q, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

// ResendBCProtection re-queries BC Protege+ after the merchant disables it.
func (s *KYCService) ResendBCProtection(ctx context.Context, correlationID string, opts ...RequestOption) (json.RawMessage, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	body := map[string]string{"correlationID": correlationID}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/kyc/bc-protection/resend", nil, body, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *KYCService) CreateValidation(ctx context.Context, params *KYCValidationCreateParams, opts ...RequestOption) (*KYCValidation, error) {
	if params == nil || params.TaxID == "" {
		return nil, fmt.Errorf("woovi: taxID is required")
	}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/kyc-validation/taxid", nil, params, &raw, opts...); err != nil {
		return nil, err
	}
	out := &KYCValidation{Raw: append(json.RawMessage(nil), raw...)}
	_ = json.Unmarshal(raw, out)
	return out, nil
}

func (s *KYCService) GetValidation(ctx context.Context, correlationID string, opts ...RequestOption) (*KYCValidation, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationID is required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/kyc-validation/" + url.PathEscape(correlationID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	out := &KYCValidation{Raw: append(json.RawMessage(nil), raw...)}
	_ = json.Unmarshal(raw, out)
	return out, nil
}
