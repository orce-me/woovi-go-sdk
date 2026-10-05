package woovi

import (
	"context"
	"encoding/json"
	"net/http"
)

// AuthService is GET /api/v1/validate-token (live; not in public OpenAPI yet).
type AuthService struct {
	client *Client
}

// TokenValidation is the TOKEN_VALIDATE_GET response. Extra keys stay in Raw.
type TokenValidation struct {
	IsValid     *bool           `json:"isValid,omitempty"`
	Valid       *bool           `json:"valid,omitempty"`
	Application json.RawMessage `json:"application,omitempty"`
	Company     json.RawMessage `json:"company,omitempty"`
	Scopes      []string        `json:"scopes,omitempty"`
	Raw         json.RawMessage `json:"-"`
}

// OK reports token validity from isValid or valid; else treats a filled 2xx body as accepted.
func (t TokenValidation) OK() bool {
	if t.IsValid != nil {
		return *t.IsValid
	}
	if t.Valid != nil {
		return *t.Valid
	}
	return len(t.Raw) > 0 || t.Application != nil || t.Company != nil || len(t.Scopes) > 0
}

// ValidateToken checks the client AppID. Requires TOKEN_VALIDATE_GET.
func (s *AuthService) ValidateToken(ctx context.Context, opts ...RequestOption) (*TokenValidation, error) {
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/validate-token", nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	out := &TokenValidation{Raw: append(json.RawMessage(nil), raw...)}
	_ = json.Unmarshal(raw, out)
	return out, nil
}
