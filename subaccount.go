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

// SubaccountsService is /api/v1/subaccount.
type SubaccountsService struct {
	client *Client
}

// SubaccountCreateParams creates a subaccount (name + pixKey).
type SubaccountCreateParams struct {
	Name   string `json:"name"`
	PixKey string `json:"pixKey"`
}

// SubaccountWithdrawParams withdraws from a subaccount. Omit Value for the full balance.
type SubaccountWithdrawParams struct {
	Value *Money `json:"value,omitempty"`
}

// SubaccountTransferParams moves value (cents) between subaccount Pix keys.
type SubaccountTransferParams struct {
	Value          Money      `json:"value"`
	FromPixKey     string     `json:"fromPixKey"`
	FromPixKeyType PixKeyType `json:"fromPixKeyType,omitempty"`
	ToPixKey       string     `json:"toPixKey"`
	ToPixKeyType   PixKeyType `json:"toPixKeyType,omitempty"`
}

// SubaccountMovementParams credits or debits a subaccount (value in cents).
type SubaccountMovementParams struct {
	Value       Money  `json:"value"`
	Description string `json:"description,omitempty"`
}

// SubaccountLedgerResult is POST /subaccount/{pixKey}/credit|debit.
type SubaccountLedgerResult struct {
	PixKey      string `json:"pixKey,omitempty"`
	Value       Money  `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
	Success     string `json:"success,omitempty"`
}

// Subaccount is a split ledger keyed by Pix key (balance in cents).
type Subaccount struct {
	ID              string `json:"id,omitempty"`
	Name            string `json:"name"`
	PixKey          string `json:"pixKey"`
	Balance         Money  `json:"balance,omitempty"`
	WithdrawBlocked bool   `json:"withdrawBlocked,omitempty"`
}

// SubaccountWithdrawResult is POST /subaccount/{pixKey}/withdraw.
type SubaccountWithdrawResult struct {
	Transaction SubaccountMovement `json:"transaction"`
}

// SubaccountMovement is a withdraw or transfer movement on a subaccount.
type SubaccountMovement struct {
	Status           string `json:"status,omitempty"`
	Value            Money  `json:"value,omitempty"`
	CorrelationID    string `json:"correlationID,omitempty"`
	DestinationAlias string `json:"destinationAlias,omitempty"`
	Comment          string `json:"comment,omitempty"`
}

// SubaccountListParams paginates GET /subaccount.
type SubaccountListParams struct {
	Skip  int
	Limit int
}

// SubaccountList is one page of GET /subaccount.
type SubaccountList struct {
	SubAccounts []Subaccount `json:"subAccounts"`
	PageInfo    PageInfo     `json:"pageInfo"`
	Skip        int
	Limit       int
}

func (s *SubaccountsService) Create(ctx context.Context, params *SubaccountCreateParams, opts ...RequestOption) (*Subaccount, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: SubaccountCreateParams is required")
	}
	if params.Name == "" || params.PixKey == "" {
		return nil, fmt.Errorf("woovi: name and pixKey are required")
	}

	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/subaccount", nil, params, &raw, opts...); err != nil {
		return nil, err
	}
	return decodeSubaccount(raw)
}

func (s *SubaccountsService) Get(ctx context.Context, pixKey string, opts ...RequestOption) (*Subaccount, error) {
	if pixKey == "" {
		return nil, fmt.Errorf("woovi: pixKey is required")
	}
	raw := json.RawMessage{}
	path := "/api/v1/subaccount/" + url.PathEscape(pixKey)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	return decodeSubaccount(raw)
}

func (s *SubaccountsService) Delete(ctx context.Context, pixKey string, opts ...RequestOption) error {
	if pixKey == "" {
		return fmt.Errorf("woovi: pixKey is required")
	}
	path := "/api/v1/subaccount/" + url.PathEscape(pixKey)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil, nil, opts...)
}

// Withdraw sends balance to the subaccount Pix key.
func (s *SubaccountsService) Withdraw(ctx context.Context, pixKey string, params *SubaccountWithdrawParams, opts ...RequestOption) (*SubaccountWithdrawResult, error) {
	if pixKey == "" {
		return nil, fmt.Errorf("woovi: pixKey is required")
	}
	var body any
	if params != nil {
		body = params
	}
	var out SubaccountWithdrawResult
	path := "/api/v1/subaccount/" + url.PathEscape(pixKey) + "/withdraw"
	if err := s.client.do(ctx, http.MethodPost, path, nil, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *SubaccountsService) Transfer(ctx context.Context, params *SubaccountTransferParams, opts ...RequestOption) error {
	if params == nil {
		return fmt.Errorf("woovi: SubaccountTransferParams is required")
	}
	if params.Value <= 0 {
		return fmt.Errorf("woovi: value must be greater than zero")
	}
	if params.FromPixKey == "" || params.ToPixKey == "" {
		return fmt.Errorf("woovi: fromPixKey and toPixKey are required")
	}
	return s.client.do(ctx, http.MethodPost, "/api/v1/subaccount/transfer", nil, params, nil, opts...)
}

func (s *SubaccountsService) Credit(ctx context.Context, pixKey string, params *SubaccountMovementParams, opts ...RequestOption) (*SubaccountLedgerResult, error) {
	return s.ledger(ctx, pixKey, "credit", params, opts...)
}

func (s *SubaccountsService) Debit(ctx context.Context, pixKey string, params *SubaccountMovementParams, opts ...RequestOption) (*SubaccountLedgerResult, error) {
	return s.ledger(ctx, pixKey, "debit", params, opts...)
}

func (s *SubaccountsService) ledger(ctx context.Context, pixKey, action string, params *SubaccountMovementParams, opts ...RequestOption) (*SubaccountLedgerResult, error) {
	if pixKey == "" {
		return nil, fmt.Errorf("woovi: pixKey is required")
	}
	if params == nil || params.Value <= 0 {
		return nil, fmt.Errorf("woovi: value must be greater than zero")
	}
	var out SubaccountLedgerResult
	path := "/api/v1/subaccount/" + url.PathEscape(pixKey) + "/" + action
	if err := s.client.do(ctx, http.MethodPost, path, nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubaccountStatementParams paginates a subaccount statement (RFC3339 start/end).
type SubaccountStatementParams struct {
	Skip  int
	Limit int
	Start string // RFC3339
	End   string // RFC3339
}

func (s *SubaccountsService) Statement(ctx context.Context, pixKey string, params *SubaccountStatementParams, opts ...RequestOption) (*StatementList, error) {
	if pixKey == "" {
		return nil, fmt.Errorf("woovi: pixKey is required")
	}
	if params == nil {
		params = &SubaccountStatementParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))
	if params.Start != "" {
		q.Set("start", params.Start)
	}
	if params.End != "" {
		q.Set("end", params.End)
	}

	raw := json.RawMessage{}
	path := "/api/v1/subaccount/" + url.PathEscape(pixKey) + "/statement"
	if err := s.client.do(ctx, http.MethodGet, path, q, nil, &raw, opts...); err != nil {
		return nil, err
	}

	out := &StatementList{Skip: skip, Limit: limit}
	var wrapped struct {
		Statements []StatementEntry `json:"statements"`
		PageInfo   PageInfo         `json:"pageInfo"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && (len(wrapped.Statements) > 0 || wrapped.PageInfo.Limit > 0 || wrapped.PageInfo.TotalCount > 0) {
		out.Statements = wrapped.Statements
		out.PageInfo = wrapped.PageInfo
		normalizePageInfo(&out.PageInfo, skip, limit)
		return out, nil
	}
	var flat []StatementEntry
	if err := json.Unmarshal(raw, &flat); err != nil {
		return nil, fmt.Errorf("woovi: decode subaccount statement: %w", err)
	}
	out.Statements = flat
	normalizePageInfo(&out.PageInfo, skip, limit)
	return out, nil
}

func (s *SubaccountsService) List(ctx context.Context, params *SubaccountListParams, opts ...RequestOption) (*SubaccountList, error) {
	if params == nil {
		params = &SubaccountListParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))

	var out SubaccountList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/subaccount", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func (s *SubaccountsService) ListAll(ctx context.Context, params *SubaccountListParams, opts ...RequestOption) iter.Seq2[*Subaccount, error] {
	return func(yield func(*Subaccount, error) bool) {
		for page, err := range s.ListPages(ctx, params, opts...) {
			if err != nil {
				yield(nil, err)
				return
			}
			for i := range page.SubAccounts {
				if !yield(&page.SubAccounts[i], nil) {
					return
				}
			}
		}
	}
}

func (s *SubaccountsService) ListPages(ctx context.Context, params *SubaccountListParams, opts ...RequestOption) iter.Seq2[*SubaccountList, error] {
	return func(yield func(*SubaccountList, error) bool) {
		p := SubaccountListParams{}
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
			if !listPageHasMore(page.PageInfo, len(page.SubAccounts), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}

func decodeSubaccount(raw json.RawMessage) (*Subaccount, error) {
	var wrapped struct {
		SubAccount  *Subaccount `json:"SubAccount"`
		Subaccount  *Subaccount `json:"subAccount"`
		SubAccounts *Subaccount `json:"subaccount"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, fmt.Errorf("woovi: decode subaccount: %w", err)
	}
	switch {
	case wrapped.SubAccount != nil:
		return wrapped.SubAccount, nil
	case wrapped.Subaccount != nil:
		return wrapped.Subaccount, nil
	case wrapped.SubAccounts != nil:
		return wrapped.SubAccounts, nil
	}

	var flat Subaccount
	if err := json.Unmarshal(raw, &flat); err != nil {
		return nil, fmt.Errorf("woovi: decode subaccount: %w", err)
	}
	if flat.Name == "" && flat.PixKey == "" && flat.ID == "" {
		return nil, fmt.Errorf("woovi: empty subaccount response")
	}
	return &flat, nil
}
