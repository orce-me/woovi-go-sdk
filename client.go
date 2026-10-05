package woovi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.woovi.com"

// Client is the Woovi API client. Create with NewClient and reuse it.
type Client struct {
	appID      string
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
	userAgent  string
	retry      RetryConfig
	timeout    time.Duration
	onRequest  OnRequestFunc
	onResponse OnResponseFunc

	Accounts           *AccountsService
	AccountRegisters   *AccountRegistersService
	Anticipations      *AnticipationsService
	Applications       *ApplicationsService
	Auth               *AuthService
	Boletos            *BoletosService
	CashbackFidelities *CashbackFidelitiesService
	Charges            *ChargesService
	Companies          *CompaniesService
	Customers          *CustomersService
	Decode             *DecodeService
	Disputes           *DisputesService
	Files              *FilesService
	FundsRecoveries    *FundsRecoveriesService
	Giftbacks          *GiftbacksService
	Installments       *InstallmentsService
	Invoices           *InvoicesService
	KYC                *KYCService
	Limits             *LimitsService
	Partners           *PartnersService
	Payments           *PaymentsService
	PixAuths           *PixAuthsService
	PixKeys            *PixKeysService
	PixQRCodes         *PixQRCodesService
	PSPs               *PSPsService
	Receipts           *ReceiptsService
	Refunds            *RefundsService
	Stablecoins        *StablecoinsService
	Statements         *StatementsService
	Subaccounts        *SubaccountsService
	Subscriptions      *SubscriptionsService
	TEDs               *TEDsService
	Transactions       *TransactionsService
	Transfers          *TransfersService
	Webhooks           *WebhooksService
}

// NewClient returns a Client authenticated with appID.
func NewClient(appID string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(appID) == "" {
		return nil, fmt.Errorf("woovi: appID is required")
	}

	c := &Client{
		appID:     appID,
		baseURL:   defaultBaseURL,
		logger:    nil,
		userAgent: "woovi-go-sdk/" + Version,
		retry:     defaultRetry(),
		timeout:   30 * time.Second,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}

	c.baseURL = strings.TrimRight(c.baseURL, "/")

	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: c.timeout}
	}

	transport := c.httpClient.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	wrapped := wrapTransport(transport, c.appID, c.userAgent, c.retry, c.logger, c.onRequest, c.onResponse)
	c.httpClient = &http.Client{
		Transport:     wrapped,
		CheckRedirect: c.httpClient.CheckRedirect,
		Jar:           c.httpClient.Jar,
		Timeout:       c.httpClient.Timeout,
	}

	c.Accounts = &AccountsService{client: c}
	c.AccountRegisters = &AccountRegistersService{client: c}
	c.Anticipations = &AnticipationsService{client: c}
	c.Applications = &ApplicationsService{client: c}
	c.Auth = &AuthService{client: c}
	c.Boletos = &BoletosService{client: c}
	c.CashbackFidelities = &CashbackFidelitiesService{client: c}
	c.Charges = &ChargesService{client: c}
	c.Companies = &CompaniesService{client: c}
	c.Customers = &CustomersService{client: c}
	c.Decode = &DecodeService{client: c}
	c.Disputes = &DisputesService{client: c}
	c.Files = &FilesService{client: c}
	c.FundsRecoveries = &FundsRecoveriesService{client: c}
	c.Giftbacks = &GiftbacksService{client: c}
	c.Installments = &InstallmentsService{client: c}
	c.Invoices = &InvoicesService{client: c}
	c.KYC = &KYCService{client: c}
	c.Limits = &LimitsService{client: c}
	c.Partners = &PartnersService{client: c}
	c.Payments = &PaymentsService{client: c}
	c.PixAuths = &PixAuthsService{client: c}
	c.PixKeys = &PixKeysService{client: c}
	c.PixQRCodes = &PixQRCodesService{client: c}
	c.PSPs = &PSPsService{client: c}
	c.Receipts = &ReceiptsService{client: c}
	c.Refunds = &RefundsService{client: c}
	c.Stablecoins = &StablecoinsService{client: c}
	c.Statements = &StatementsService{client: c}
	c.Subaccounts = &SubaccountsService{client: c}
	c.Subscriptions = &SubscriptionsService{client: c}
	c.TEDs = &TEDsService{client: c}
	c.Transactions = &TransactionsService{client: c}
	c.Transfers = &TransfersService{client: c}
	c.Webhooks = &WebhooksService{client: c}
	return c, nil
}

// Do sends a raw JSON request to path (relative to base URL).
// Encodes in and decodes out when non-nil.
func (c *Client) Do(ctx context.Context, method, path string, in, out any, opts ...RequestOption) error {
	return c.do(ctx, method, path, nil, in, out, opts...)
}

func (c *Client) do(
	ctx context.Context,
	method, path string,
	query url.Values,
	in, out any,
	opts ...RequestOption,
) error {
	var body io.Reader
	contentType := ""
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("woovi: encode request: %w", err)
		}
		body = bytes.NewReader(b)
		contentType = "application/json"
	}
	_, err := c.doHTTP(ctx, method, path, query, contentType, body, out, opts...)
	return err
}

func (c *Client) doHTTP(
	ctx context.Context,
	method, path string,
	query url.Values,
	contentType string,
	body io.Reader,
	out any,
	opts ...RequestOption,
) ([]byte, error) {
	if ctx == nil {
		return nil, fmt.Errorf("woovi: context is required")
	}

	cfg := applyRequestOptions(opts)
	ctx, cancel := c.withRequestContext(ctx, cfg)
	defer cancel()

	if len(cfg.query) > 0 {
		if query == nil {
			query = make(url.Values)
		}
		for k, vals := range cfg.query {
			for _, v := range vals {
				query.Add(k, v)
			}
		}
	}

	u, err := c.resolveURL(path, query)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, fmt.Errorf("woovi: create request: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, vals := range cfg.headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	if cfg.idempotencyKey != "" {
		req.Header.Set(headerIdempotencyKey, cfg.idempotencyKey)
	}

	if c.logger != nil {
		c.logger.Debug("woovi: request", "method", method, "url", u)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("woovi: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("woovi: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, c.apiError(resp, respBody)
	}

	if out == nil || len(respBody) == 0 || resp.StatusCode == http.StatusNoContent {
		return respBody, nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return nil, fmt.Errorf("woovi: decode response: %w", err)
	}
	return respBody, nil
}

func (c *Client) doBytes(
	ctx context.Context,
	method, path string,
	query url.Values,
	opts ...RequestOption,
) ([]byte, string, error) {
	if ctx == nil {
		return nil, "", fmt.Errorf("woovi: context is required")
	}

	cfg := applyRequestOptions(opts)
	ctx, cancel := c.withRequestContext(ctx, cfg)
	defer cancel()

	if len(cfg.query) > 0 {
		if query == nil {
			query = make(url.Values)
		}
		for k, vals := range cfg.query {
			for _, v := range vals {
				query.Add(k, v)
			}
		}
	}

	u, err := c.resolveURL(path, query)
	if err != nil {
		return nil, "", err
	}

	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return nil, "", fmt.Errorf("woovi: create request: %w", err)
	}
	for k, vals := range cfg.headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("woovi: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("woovi: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", c.apiError(resp, respBody)
	}
	return respBody, resp.Header.Get("Content-Type"), nil
}

func (c *Client) resolveURL(path string, query url.Values) (string, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	raw := c.baseURL + path
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("woovi: invalid URL %q: %w", raw, err)
	}
	if query != nil {
		q := u.Query()
		for k, vals := range query {
			for _, v := range vals {
				q.Add(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

func (c *Client) apiError(resp *http.Response, body []byte) error {
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Body:       body,
		RequestID:  resp.Header.Get(headerRequestID),
	}
	var env errorEnvelope
	if err := json.Unmarshal(body, &env); err == nil {
		apiErr.Message = env.Message
		apiErr.Code = env.ErrorCode
		apiErr.Errors = env.Errors
		if apiErr.Message == "" {
			apiErr.Message = env.Error
		}
		if apiErr.Message == "" && len(env.Errors) > 0 {
			apiErr.Message = env.Errors[0].Message
		}
	}
	if apiErr.Message == "" {
		apiErr.Message = strings.TrimSpace(string(body))
	}
	return apiErr
}
