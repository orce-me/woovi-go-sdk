package woovi

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PixKeysService is /api/v1/pix-keys and DICT check.
type PixKeysService struct {
	client *Client
}

const (
	// PixKeyTypeEVP is EVP/random as returned by some Pix key endpoints.
	PixKeyTypeEVP PixKeyType = "EVP"
)

// PixKeyCreateParams registers a Pix key on the account.
type PixKeyCreateParams struct {
	PixKey string     `json:"pixKey"`
	Type   PixKeyType `json:"type"`
}

// PixKey is a DICT key owned by the account.
type PixKey struct {
	Key  string     `json:"key"`
	Type PixKeyType `json:"type"`
}

// PixKeyListResult is GET /pix-keys.
type PixKeyListResult struct {
	PixKeys []PixKey `json:"pixKeys"`
	Account *Account `json:"account,omitempty"`
}

// PixKeyOwner is DICT owner data from check.
type PixKeyOwner struct {
	Name    string `json:"name"`
	TaxID   string `json:"taxID"`
	PSP     string `json:"psp"`
	Branch  string `json:"branch"`
	Account string `json:"account"`
}

// PixKeyCheckResult is the DICT verification response.
type PixKeyCheckResult struct {
	PixKeyEndToEndID string      `json:"pixKeyEndToEndId"`
	PixKey           string      `json:"pixKey"`
	Type             PixKeyType  `json:"type"`
	Owner            PixKeyOwner `json:"owner"`
}

// PixKeyWithdrawParams sets the account withdraw Pix key.
type PixKeyWithdrawParams struct {
	PixKey string `json:"pixKey"`
}

// PixKeyTokens is the DICT lookup token-bucket balance.
type PixKeyTokens struct {
	Tokens             int       `json:"tokens"`
	MaxTokens          int       `json:"maxTokens"`
	RefreshRate        int       `json:"refreshRate"`
	NextRefresh        time.Time `json:"nextRefresh"`
	TokensAfterRefresh int       `json:"tokensAfterRefresh"`
}

// PixKeyTokenLogOperation is REMOVE, ADD, or REFILL on the DICT token bucket.
type PixKeyTokenLogOperation string

const (
	PixKeyTokenLogRemove PixKeyTokenLogOperation = "REMOVE"
	PixKeyTokenLogAdd    PixKeyTokenLogOperation = "ADD"
	PixKeyTokenLogRefill PixKeyTokenLogOperation = "REFILL"
)

// PixKeyTokenLog is one DICT token-bucket ledger entry.
type PixKeyTokenLog struct {
	Operation    PixKeyTokenLogOperation `json:"operation"`
	Reason       string                  `json:"reason,omitempty"`
	Tokens       int                     `json:"tokens,omitempty"`
	TokensBefore int                     `json:"tokensBefore,omitempty"`
	TokensAfter  int                     `json:"tokensAfter,omitempty"`
	EndToEndID   string                  `json:"endToEndId,omitempty"`
	PixKey       string                  `json:"pixKey,omitempty"`
	CreatedAt    time.Time               `json:"createdAt,omitempty"`
}

// PixKeyTokenLogListParams paginates DICT token logs (optional companyBankAccount).
type PixKeyTokenLogListParams struct {
	Skip               int
	Limit              int
	CompanyBankAccount string
}

// PixKeyTokenLogList is GET /pix-keys/tokens/logs.
type PixKeyTokenLogList struct {
	Logs     []PixKeyTokenLog `json:"logs"`
	PageInfo PageInfo         `json:"pageInfo"`
	Skip     int
	Limit    int
}

func (s *PixKeysService) Create(ctx context.Context, params *PixKeyCreateParams, opts ...RequestOption) (*PixKey, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: PixKeyCreateParams is required")
	}
	if params.PixKey == "" || params.Type == "" {
		return nil, fmt.Errorf("woovi: pixKey and type are required")
	}

	raw := struct {
		PixKey *PixKey    `json:"pixKey"`
		Key    string     `json:"key"`
		Type   PixKeyType `json:"type"`
	}{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/pix-keys", nil, params, &raw, opts...); err != nil {
		return nil, err
	}
	if raw.PixKey != nil {
		return raw.PixKey, nil
	}
	if raw.Key == "" {
		return nil, fmt.Errorf("woovi: empty pix key response")
	}
	return &PixKey{Key: raw.Key, Type: raw.Type}, nil
}

// List returns Pix keys and account balance for the current AppID.
func (s *PixKeysService) List(ctx context.Context, opts ...RequestOption) (*PixKeyListResult, error) {
	var out PixKeyListResult
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/pix-keys", nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PixKeysService) Delete(ctx context.Context, pixKey string, opts ...RequestOption) error {
	if pixKey == "" {
		return fmt.Errorf("woovi: pixKey is required")
	}
	path := "/api/v1/pix-keys/" + url.PathEscape(pixKey)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil, nil, opts...)
}

// Check verifies a Pix key via POST body. Prefer for emails and special characters.
func (s *PixKeysService) Check(ctx context.Context, pixKey string, opts ...RequestOption) (*PixKeyCheckResult, error) {
	if pixKey == "" {
		return nil, fmt.Errorf("woovi: pixKey is required")
	}
	body := map[string]string{"pixKey": pixKey}
	var out PixKeyCheckResult
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/pix-keys/check", nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// CheckPath verifies a Pix key via GET path.
func (s *PixKeysService) CheckPath(ctx context.Context, pixKey string, opts ...RequestOption) (*PixKeyCheckResult, error) {
	if pixKey == "" {
		return nil, fmt.Errorf("woovi: pixKey is required")
	}
	var out PixKeyCheckResult
	path := "/api/v1/pix-keys/" + url.PathEscape(pixKey) + "/check"
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PixKeysService) SetDefault(ctx context.Context, pixKey string, opts ...RequestOption) (*PixKey, error) {
	if pixKey == "" {
		return nil, fmt.Errorf("woovi: pixKey is required")
	}
	raw := struct {
		PixKey *PixKey    `json:"pixKey"`
		Key    string     `json:"key"`
		Type   PixKeyType `json:"type"`
	}{}
	path := "/api/v1/pix-keys/" + url.PathEscape(pixKey) + "/default"
	if err := s.client.do(ctx, http.MethodPut, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	if raw.PixKey != nil {
		return raw.PixKey, nil
	}
	if raw.Key == "" {
		return &PixKey{Key: pixKey}, nil
	}
	return &PixKey{Key: raw.Key, Type: raw.Type}, nil
}

// SetWithdrawKey registers a new withdraw Pix key (POST).
func (s *PixKeysService) SetWithdrawKey(ctx context.Context, params *PixKeyWithdrawParams, opts ...RequestOption) (*Account, error) {
	return s.withdrawKey(ctx, http.MethodPost, params, opts...)
}

// UpdateWithdrawKey selects an existing withdraw Pix key (PUT).
func (s *PixKeysService) UpdateWithdrawKey(ctx context.Context, params *PixKeyWithdrawParams, opts ...RequestOption) (*Account, error) {
	return s.withdrawKey(ctx, http.MethodPut, params, opts...)
}

func (s *PixKeysService) withdrawKey(ctx context.Context, method string, params *PixKeyWithdrawParams, opts ...RequestOption) (*Account, error) {
	if params == nil || params.PixKey == "" {
		return nil, fmt.Errorf("woovi: pixKey is required")
	}
	var out accountEnvelope
	if err := s.client.do(ctx, method, "/api/v1/pix-keys/withdraw", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Account, nil
}

// Tokens returns the DICT lookup token-bucket balance.
func (s *PixKeysService) Tokens(ctx context.Context, opts ...RequestOption) (*PixKeyTokens, error) {
	var out PixKeyTokens
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/pix-keys/tokens", nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *PixKeysService) ListTokenLogs(ctx context.Context, params *PixKeyTokenLogListParams, opts ...RequestOption) (*PixKeyTokenLogList, error) {
	if params == nil {
		params = &PixKeyTokenLogListParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}
	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))
	if params.CompanyBankAccount != "" {
		q.Set("companyBankAccount", params.CompanyBankAccount)
	}
	var out PixKeyTokenLogList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/pix-keys/tokens/logs", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func (s *PixKeysService) ListAllTokenLogs(ctx context.Context, params *PixKeyTokenLogListParams, opts ...RequestOption) iter.Seq2[*PixKeyTokenLog, error] {
	return func(yield func(*PixKeyTokenLog, error) bool) {
		p := PixKeyTokenLogListParams{}
		if params != nil {
			p = *params
		}
		if p.Limit <= 0 {
			p.Limit = 50
		}
		for {
			if err := ctx.Err(); err != nil {
				yield(nil, err)
				return
			}
			page, err := s.ListTokenLogs(ctx, &p, opts...)
			if err != nil {
				yield(nil, err)
				return
			}
			for i := range page.Logs {
				if !yield(&page.Logs[i], nil) {
					return
				}
			}
			if !listPageHasMore(page.PageInfo, len(page.Logs), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}
