package woovi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// StablecoinCurrency is USDT, USDC, or BRLA.
type StablecoinCurrency string

const (
	StablecoinUSDT StablecoinCurrency = "USDT"
	StablecoinUSDC StablecoinCurrency = "USDC"
	StablecoinBRLA StablecoinCurrency = "BRLA"
)

// StablecoinNetwork is the chain for deposit/payout (POLYGON, ETHEREUM, …).
type StablecoinNetwork string

const (
	StablecoinNetworkPolygon  StablecoinNetwork = "POLYGON"
	StablecoinNetworkEthereum StablecoinNetwork = "ETHEREUM"
	StablecoinNetworkBase     StablecoinNetwork = "BASE"
	StablecoinNetworkCelo     StablecoinNetwork = "CELO"
	StablecoinNetworkTron     StablecoinNetwork = "TRON"
	StablecoinNetworkBNB      StablecoinNetwork = "BNB"
)

// StablecoinsService is /api/v1/stablecoin.
type StablecoinsService struct {
	client *Client
}

// StablecoinQuote is FX/fees for a stablecoin operation (wooviFee in cents when set).
type StablecoinQuote struct {
	InputAmount    float64 `json:"inputAmount,omitempty"`
	InputCurrency  string  `json:"inputCurrency,omitempty"`
	OutputAmount   float64 `json:"outputAmount,omitempty"`
	OutputCurrency string  `json:"outputCurrency,omitempty"`
	Rate           float64 `json:"rate,omitempty"`
	Fee            float64 `json:"fee,omitempty"`
	WooviFee       *Money  `json:"wooviFee,omitempty"`
	ProviderFee    *Money  `json:"providerFee,omitempty"`
}

// StablecoinDepositCreateParams: send GrossAmount (preferred) or Value (legacy), not both.
type StablecoinDepositCreateParams struct {
	GrossAmount              *Money             `json:"grossAmount,omitempty"`
	Value                    *Money             `json:"value,omitempty"`
	Currency                 StablecoinCurrency `json:"currency"`
	Network                  StablecoinNetwork  `json:"network,omitempty"`
	SubAccountID             string             `json:"subAccountId,omitempty"`
	CorrelationID            string             `json:"correlationId,omitempty"`
	DestinationWalletAddress string             `json:"destinationWalletAddress,omitempty"`
}

// StablecoinDeposit is a BRL→crypto deposit intent.
type StablecoinDeposit struct {
	Status        string           `json:"status,omitempty"`
	DepositID     string           `json:"depositId,omitempty"`
	CorrelationID string           `json:"correlationId,omitempty"`
	Expiration    time.Time        `json:"expiration,omitempty"`
	Quote         *StablecoinQuote `json:"quote,omitempty"`
}

// StablecoinPayoutCreateParams converts crypto to Pix (value in cents, pixKey).
type StablecoinPayoutCreateParams struct {
	Value         Money              `json:"value"`
	Currency      StablecoinCurrency `json:"currency"`
	PixKey        string             `json:"pixKey"`
	CorrelationID string             `json:"correlationId,omitempty"`
	PixMessage    string             `json:"pixMessage,omitempty"`
}

// StablecoinPayout is a crypto→Pix payout (extra keys stay in Raw).
type StablecoinPayout struct {
	Status        string           `json:"status,omitempty"`
	PayoutID      string           `json:"payoutId,omitempty"`
	CorrelationID string           `json:"correlationId,omitempty"`
	PixKey        string           `json:"pixKey,omitempty"`
	Quote         *StablecoinQuote `json:"quote,omitempty"`
	Raw           json.RawMessage  `json:"-"`
}

// StablecoinSwapCreateParams creates a swap on INTERNAL float.
type StablecoinSwapCreateParams map[string]any

// StablecoinSubaccount is a KYB-scoped stablecoin wallet (extra keys stay in Raw).
type StablecoinSubaccount struct {
	ID            string          `json:"id,omitempty"`
	SubAccountID  string          `json:"subAccountId,omitempty"`
	Account       string          `json:"account,omitempty"`
	Status        string          `json:"status,omitempty"`
	ReasonCode    string          `json:"reasonCode,omitempty"`
	Reason        string          `json:"reason,omitempty"`
	CorrelationID string          `json:"correlationId,omitempty"`
	CreatedAt     time.Time       `json:"createdAt,omitempty"`
	Raw           json.RawMessage `json:"-"`
}

// StablecoinSubaccountCreateParams requests a KYB subaccount.
type StablecoinSubaccountCreateParams struct {
	AccountRegisterID    string `json:"accountRegisterId,omitempty"`
	CompanyBankAccountID string `json:"companyBankAccountId,omitempty"`
}

// StablecoinLimitDocumentType is a limit-increase comprovante kind.
type StablecoinLimitDocumentType string

const (
	StablecoinLimitProofFinancialCapacity StablecoinLimitDocumentType = "PROOF_OF_FINANCIAL_CAPACITY"
	StablecoinLimitProofAddressCompany    StablecoinLimitDocumentType = "PROOF_OF_ADDRESS_COMPANY"
	StablecoinLimitProofAddressUBO        StablecoinLimitDocumentType = "PROOF_OF_ADDRESS_UBO"
)

// StablecoinLimitDocumentUploadParams gets a pre-signed upload URL.
type StablecoinLimitDocumentUploadParams struct {
	Type     StablecoinLimitDocumentType `json:"type"`
	FileName string                      `json:"fileName"`
	MimeType string                      `json:"mimeType"`
}

// CreateDeposit creates a Pix→stable deposit (PENDING until ApproveDeposit).
func (s *StablecoinsService) CreateDeposit(ctx context.Context, params *StablecoinDepositCreateParams, opts ...RequestOption) (*StablecoinDeposit, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: StablecoinDepositCreateParams is required")
	}
	if params.Currency == "" {
		return nil, fmt.Errorf("woovi: currency is required")
	}
	if params.GrossAmount == nil && params.Value == nil {
		return nil, fmt.Errorf("woovi: grossAmount or value is required")
	}
	var out StablecoinDeposit
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/stablecoin/deposit", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *StablecoinsService) ApproveDeposit(ctx context.Context, correlationID string, opts ...RequestOption) (*StablecoinDeposit, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationId is required")
	}
	body := map[string]string{"correlationId": correlationID}
	var out StablecoinDeposit
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/stablecoin/deposit/approve", nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *StablecoinsService) QuoteDeposit(ctx context.Context, query url.Values, opts ...RequestOption) (json.RawMessage, error) {
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/stablecoin/quote", query, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

// CreatePayout creates a stable→Pix payout (PENDING until ApprovePayout).
func (s *StablecoinsService) CreatePayout(ctx context.Context, params *StablecoinPayoutCreateParams, opts ...RequestOption) (*StablecoinPayout, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: StablecoinPayoutCreateParams is required")
	}
	if params.Value <= 0 || params.Currency == "" || params.PixKey == "" {
		return nil, fmt.Errorf("woovi: value, currency and pixKey are required")
	}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/stablecoin/payout", nil, params, &raw, opts...); err != nil {
		return nil, err
	}
	out := &StablecoinPayout{Raw: append(json.RawMessage(nil), raw...)}
	_ = json.Unmarshal(raw, out)
	return out, nil
}

// ApprovePayout opens the provider ticket for a pending payout.
func (s *StablecoinsService) ApprovePayout(ctx context.Context, correlationID string, opts ...RequestOption) (*StablecoinPayout, error) {
	if correlationID == "" {
		return nil, fmt.Errorf("woovi: correlationId is required")
	}
	body := map[string]string{"correlationId": correlationID}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/stablecoin/payout/approve", nil, body, &raw, opts...); err != nil {
		return nil, err
	}
	out := &StablecoinPayout{Raw: append(json.RawMessage(nil), raw...)}
	_ = json.Unmarshal(raw, out)
	return out, nil
}

func (s *StablecoinsService) GetPayout(ctx context.Context, payoutID string, opts ...RequestOption) (*StablecoinPayout, error) {
	if payoutID == "" {
		return nil, fmt.Errorf("woovi: payoutId is required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/stablecoin/payout/" + url.PathEscape(payoutID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	out := &StablecoinPayout{Raw: append(json.RawMessage(nil), raw...)}
	_ = json.Unmarshal(raw, out)
	return out, nil
}

func (s *StablecoinsService) ListPayouts(ctx context.Context, query url.Values, opts ...RequestOption) (json.RawMessage, error) {
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/stablecoin/payout", query, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) QuotePayout(ctx context.Context, value Money, currency StablecoinCurrency, opts ...RequestOption) (json.RawMessage, error) {
	if value <= 0 || currency == "" {
		return nil, fmt.Errorf("woovi: value and currency are required")
	}
	q := url.Values{}
	q.Set("value", strconv.FormatInt(value.Int64(), 10))
	q.Set("currency", string(currency))
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/stablecoin/payout/quote", q, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) CreateSwap(ctx context.Context, params StablecoinSwapCreateParams, opts ...RequestOption) (json.RawMessage, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: swap params are required")
	}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/stablecoin/swap", nil, params, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) GetSwap(ctx context.Context, swapID string, opts ...RequestOption) (json.RawMessage, error) {
	if swapID == "" {
		return nil, fmt.Errorf("woovi: swapId is required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/stablecoin/swap/" + url.PathEscape(swapID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) ListSwaps(ctx context.Context, query url.Values, opts ...RequestOption) (json.RawMessage, error) {
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/stablecoin/swap", query, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) QuoteSwap(ctx context.Context, query url.Values, opts ...RequestOption) (json.RawMessage, error) {
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/stablecoin/swap/quote", query, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) ListWallets(ctx context.Context, opts ...RequestOption) (json.RawMessage, error) {
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/stablecoin/wallets", nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) ListSubaccounts(ctx context.Context, opts ...RequestOption) ([]StablecoinSubaccount, error) {
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/stablecoin/subaccount", nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	var wrapped struct {
		SubAccounts []StablecoinSubaccount `json:"subAccounts"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && len(wrapped.SubAccounts) > 0 {
		return wrapped.SubAccounts, nil
	}
	var arr []StablecoinSubaccount
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	return []StablecoinSubaccount{}, nil
}

// CreateSubaccount requests a KYB stablecoin subaccount.
func (s *StablecoinsService) CreateSubaccount(ctx context.Context, params *StablecoinSubaccountCreateParams, opts ...RequestOption) (*StablecoinSubaccount, error) {
	if params == nil {
		params = &StablecoinSubaccountCreateParams{}
	}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/stablecoin/subaccount", nil, params, &raw, opts...); err != nil {
		return nil, err
	}
	out := &StablecoinSubaccount{Raw: append(json.RawMessage(nil), raw...)}
	_ = json.Unmarshal(raw, out)
	return out, nil
}

func (s *StablecoinsService) GetSubaccount(ctx context.Context, subAccountID string, opts ...RequestOption) (*StablecoinSubaccount, error) {
	if subAccountID == "" {
		return nil, fmt.Errorf("woovi: subAccountId is required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/stablecoin/subaccount/" + url.PathEscape(subAccountID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	out := &StablecoinSubaccount{Raw: append(json.RawMessage(nil), raw...)}
	_ = json.Unmarshal(raw, out)
	return out, nil
}

func (s *StablecoinsService) GetSubaccountBalances(ctx context.Context, subAccountID string, opts ...RequestOption) (json.RawMessage, error) {
	if subAccountID == "" {
		return nil, fmt.Errorf("woovi: subAccountId is required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/stablecoin/subaccount/" + url.PathEscape(subAccountID) + "/balances"
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) ListSubaccountWallets(ctx context.Context, subAccountID string, opts ...RequestOption) (json.RawMessage, error) {
	if subAccountID == "" {
		return nil, fmt.Errorf("woovi: subAccountId is required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/stablecoin/subaccount/" + url.PathEscape(subAccountID) + "/wallets"
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

// CreateUSDKYB starts USD KYB for a subaccount.
func (s *StablecoinsService) CreateUSDKYB(ctx context.Context, body any, opts ...RequestOption) (json.RawMessage, error) {
	if body == nil {
		body = map[string]any{}
	}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/stablecoin/subaccount/kyb/usd", nil, body, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

// GetUSDKYB reads USD KYB status.
func (s *StablecoinsService) GetUSDKYB(ctx context.Context, query url.Values, opts ...RequestOption) (json.RawMessage, error) {
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/stablecoin/subaccount/kyb/usd", query, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

// CreateUSDKYBDocumentUpload gets a pre-signed URL for a USD KYB document.
func (s *StablecoinsService) CreateUSDKYBDocumentUpload(ctx context.Context, body any, opts ...RequestOption) (json.RawMessage, error) {
	if body == nil {
		return nil, fmt.Errorf("woovi: body is required")
	}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/stablecoin/subaccount/kyb/usd/document", nil, body, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) GetLimits(ctx context.Context, query url.Values, opts ...RequestOption) (json.RawMessage, error) {
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/stablecoin/limit", query, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

// CreateLimitDocumentUpload gets a pre-signed URL for a limit-increase document.
func (s *StablecoinsService) CreateLimitDocumentUpload(ctx context.Context, params *StablecoinLimitDocumentUploadParams, opts ...RequestOption) (json.RawMessage, error) {
	if params == nil || params.Type == "" || params.FileName == "" || params.MimeType == "" {
		return nil, fmt.Errorf("woovi: type, fileName and mimeType are required")
	}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/stablecoin/limit/document", nil, params, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) CreateLimitRequest(ctx context.Context, body any, opts ...RequestOption) (json.RawMessage, error) {
	if body == nil {
		return nil, fmt.Errorf("woovi: body is required")
	}
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/stablecoin/limit/request", nil, body, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) GetLimitRequest(ctx context.Context, limitRequestID string, opts ...RequestOption) (json.RawMessage, error) {
	if limitRequestID == "" {
		return nil, fmt.Errorf("woovi: limitRequestId is required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/stablecoin/limit/request/" + url.PathEscape(limitRequestID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) ListLimitRequests(ctx context.Context, query url.Values, opts ...RequestOption) (json.RawMessage, error) {
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/stablecoin/limit/request", query, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *StablecoinsService) AttachLimitRequestDocument(ctx context.Context, limitRequestID string, body any, opts ...RequestOption) (json.RawMessage, error) {
	if limitRequestID == "" || body == nil {
		return nil, fmt.Errorf("woovi: limitRequestId and body are required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/stablecoin/limit/request/" + url.PathEscape(limitRequestID) + "/document"
	if err := s.client.do(ctx, http.MethodPost, path, nil, body, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}
