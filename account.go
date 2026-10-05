package woovi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AccountsService is POST/GET/DELETE /api/v1/account.
type AccountsService struct {
	client *Client
}

// AccountBalance holds ledger totals in cents (total, blocked, available).
type AccountBalance struct {
	Total     Money `json:"total"`
	Blocked   Money `json:"blocked"`
	Available Money `json:"available"`
}

// AccountPixKeyWithdraw is the Pix key used for account withdraw.
type AccountPixKeyWithdraw struct {
	PixKey string     `json:"pixKey"`
	Type   PixKeyType `json:"type"`
}

// Account is a Woovi bank account under /api/v1/account.
type Account struct {
	AccountID      string                 `json:"accountId"`
	TaxID          string                 `json:"taxId,omitempty"`
	IsDefault      bool                   `json:"isDefault,omitempty"`
	Balance        *AccountBalance        `json:"balance,omitempty"`
	PixKeyWithdraw *AccountPixKeyWithdraw `json:"pixKeyWithdraw"`
}

// AccountList is GET /account when the API returns an accounts array.
type AccountList struct {
	Accounts []Account `json:"accounts"`
}

type accountEnvelope struct {
	Account Account `json:"account"`
}

// AccountWithdrawParams sends value (cents) to the account withdraw Pix key.
type AccountWithdrawParams struct {
	Value Money `json:"value"`
}

// AccountWithdrawResult is POST /account/{id}/withdraw.
type AccountWithdrawResult struct {
	Withdraw AccountWithdraw `json:"withdraw"`
}

// AccountWithdraw is the withdraw payload with updated account and Pix movement.
type AccountWithdraw struct {
	Account     *Account                    `json:"account,omitempty"`
	Transaction *AccountWithdrawTransaction `json:"transaction,omitempty"`
}

// AccountWithdrawTransaction is the Pix out created by account withdraw (endToEndId).
type AccountWithdrawTransaction struct {
	EndToEndID string `json:"endToEndId,omitempty"`
	Value      Money  `json:"value,omitempty"`
}

// Create creates a bank account. Requires a master AppID.
func (s *AccountsService) Create(ctx context.Context, opts ...RequestOption) (*Account, error) {
	var out accountEnvelope
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/account", nil, map[string]any{}, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Account, nil
}

func (s *AccountsService) Get(ctx context.Context, accountID string, opts ...RequestOption) (*Account, error) {
	if accountID == "" {
		return nil, fmt.Errorf("woovi: accountID is required")
	}
	var out accountEnvelope
	path := "/api/v1/account/" + url.PathEscape(accountID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Account, nil
}

func (s *AccountsService) List(ctx context.Context, opts ...RequestOption) ([]Account, error) {
	raw := struct {
		Accounts []Account `json:"accounts"`
		Account  *Account  `json:"account"`
	}{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/account", nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	if len(raw.Accounts) > 0 {
		return raw.Accounts, nil
	}
	if raw.Account != nil {
		return []Account{*raw.Account}, nil
	}
	return []Account{}, nil
}

func (s *AccountsService) Delete(ctx context.Context, accountID string, opts ...RequestOption) error {
	if accountID == "" {
		return fmt.Errorf("woovi: accountID is required")
	}
	path := "/api/v1/account/" + url.PathEscape(accountID)
	return s.client.do(ctx, http.MethodDelete, path, nil, nil, nil, opts...)
}

// Withdraw sends value to the account withdraw Pix key.
func (s *AccountsService) Withdraw(ctx context.Context, accountID string, params *AccountWithdrawParams, opts ...RequestOption) (*AccountWithdrawResult, error) {
	if accountID == "" {
		return nil, fmt.Errorf("woovi: accountID is required")
	}
	if params == nil || params.Value <= 0 {
		return nil, fmt.Errorf("woovi: value must be greater than zero")
	}
	var out AccountWithdrawResult
	path := "/api/v1/account/" + url.PathEscape(accountID) + "/withdraw"
	if err := s.client.do(ctx, http.MethodPost, path, nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}
