package woovi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ApplicationType selects the application kind (API).
type ApplicationType string

const (
	ApplicationTypeAPI ApplicationType = "API"
)

// ApplicationsService is /api/v1/application.
type ApplicationsService struct {
	client *Client
}

// ApplicationCreateParams creates an AppID under an accountId.
type ApplicationCreateParams struct {
	AccountID   string           `json:"accountId"`
	Application ApplicationInput `json:"application"`
}

// ApplicationInput is name, type, and optional scopes for create.
type ApplicationInput struct {
	Name   string          `json:"name"`
	Type   ApplicationType `json:"type"`
	Scopes []string        `json:"scopes,omitempty"`
}

// Application is an API credential (AppID / clientId / clientSecret).
type Application struct {
	Name               string          `json:"name"`
	IsActive           bool            `json:"isActive"`
	Type               ApplicationType `json:"type"`
	ClientID           string          `json:"clientId,omitempty"`
	ClientSecret       string          `json:"clientSecret,omitempty"`
	AppID              string          `json:"-"`
	CompanyBankAccount string          `json:"companyBankAccount,omitempty"`
	Scopes             []string        `json:"scopes,omitempty"`
}

func (a *Application) UnmarshalJSON(data []byte) error {
	type alias Application
	aux := struct {
		alias
		AppIDCamel string `json:"appID"`
		AppIDLower string `json:"appId"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*a = Application(aux.alias)
	switch {
	case aux.AppIDCamel != "":
		a.AppID = aux.AppIDCamel
	case aux.AppIDLower != "":
		a.AppID = aux.AppIDLower
	}
	return nil
}

// ApplicationRotateParams rotates clientSecret for a clientId (master AppID).
type ApplicationRotateParams struct {
	ClientID string `json:"clientId"`
}

// ApplicationScopesResult is GET /application/scopes.
type ApplicationScopesResult struct {
	Scopes []string          `json:"scopes"`
	Groups []json.RawMessage `json:"groups,omitempty"`
	Raw    json.RawMessage   `json:"-"`
}

type applicationEnvelope struct {
	Application Application `json:"application"`
}

// Create creates an API application. Use a master or account AppID with create rights.
func (s *ApplicationsService) Create(ctx context.Context, params *ApplicationCreateParams, opts ...RequestOption) (*Application, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: ApplicationCreateParams is required")
	}
	if params.AccountID == "" {
		return nil, fmt.Errorf("woovi: accountId is required")
	}
	if params.Application.Name == "" {
		return nil, fmt.Errorf("woovi: application name is required")
	}
	if params.Application.Type == "" {
		params.Application.Type = ApplicationTypeAPI
	}

	var out applicationEnvelope
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/application", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Application, nil
}

// Delete deactivates the current AppID application. Master apps cannot be deleted via API.
func (s *ApplicationsService) Delete(ctx context.Context, opts ...RequestOption) error {
	var out struct {
		Success bool `json:"success"`
	}
	return s.client.do(ctx, http.MethodDelete, "/api/v1/application", nil, nil, &out, opts...)
}

// RotateSecret rotates clientSecret. Requires MASTER AppID and APPLICATION_ROTATE_POST.
func (s *ApplicationsService) RotateSecret(ctx context.Context, params *ApplicationRotateParams, opts ...RequestOption) (*Application, error) {
	if params == nil || params.ClientID == "" {
		return nil, fmt.Errorf("woovi: clientId is required")
	}
	var out applicationEnvelope
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/application/rotate-secret", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.Application, nil
}

// ListScopes lists requestable scopes. No scope required.
func (s *ApplicationsService) ListScopes(ctx context.Context, opts ...RequestOption) (*ApplicationScopesResult, error) {
	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/application/scopes", nil, nil, &raw, opts...); err != nil {
		return nil, err
	}
	out := &ApplicationScopesResult{Raw: append(json.RawMessage(nil), raw...)}
	_ = json.Unmarshal(raw, out)
	if len(out.Scopes) == 0 {
		var alt struct {
			ScopeGroups []struct {
				Scopes []string `json:"scopes"`
			} `json:"scopeGroups"`
			Groups []struct {
				Scopes []string `json:"scopes"`
			} `json:"groups"`
		}
		if err := json.Unmarshal(raw, &alt); err == nil {
			for _, g := range alt.ScopeGroups {
				out.Scopes = append(out.Scopes, g.Scopes...)
			}
			for _, g := range alt.Groups {
				out.Scopes = append(out.Scopes, g.Scopes...)
			}
		}
	}
	return out, nil
}
