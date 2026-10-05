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

// LimitRequestStatus is IN_REVIEW, APPROVED, or REJECTED.
type LimitRequestStatus string

const (
	LimitRequestStatusInReview LimitRequestStatus = "IN_REVIEW"
	LimitRequestStatusApproved LimitRequestStatus = "APPROVED"
	LimitRequestStatusRejected LimitRequestStatus = "REJECTED"
)

// LimitsService is /api/v1/limits.
type LimitsService struct {
	client *Client
}

// AccountLimits are monetary ceilings in cents.
type AccountLimits struct {
	PixDayLimit                     Money  `json:"pixDayLimit,omitempty"`
	PixNightLimit                   Money  `json:"pixNightLimit,omitempty"`
	PixOutSameHolderDayLimit        Money  `json:"pixOutSameHolderDayLimit,omitempty"`
	PixOutDifferentHolderDayLimit   Money  `json:"pixOutDifferentHolderDayLimit,omitempty"`
	PixOutSameHolderNightLimit      Money  `json:"pixOutSameHolderNightLimit,omitempty"`
	PixOutDifferentHolderNightLimit Money  `json:"pixOutDifferentHolderNightLimit,omitempty"`
	PixInSameHolderDayLimit         Money  `json:"pixInSameHolderDayLimit,omitempty"`
	PixInDifferentHolderDayLimit    Money  `json:"pixInDifferentHolderDayLimit,omitempty"`
	PixInSameHolderNightLimit       Money  `json:"pixInSameHolderNightLimit,omitempty"`
	PixInDifferentHolderNightLimit  Money  `json:"pixInDifferentHolderNightLimit,omitempty"`
	DayStartAt                      string `json:"dayStartAt,omitempty"`
	NightStartAt                    string `json:"nightStartAt,omitempty"`
	BoletoEmissionLimit             int64  `json:"boletoEmissionLimit,omitempty"`
	BoletoMaximumValueLimit         Money  `json:"boletoMaximumValueLimit,omitempty"`
	StableInDayLimit                Money  `json:"stableInDayLimit,omitempty"`
	StableOutDayLimit               Money  `json:"stableOutDayLimit,omitempty"`
	TedInLimit                      Money  `json:"tedInLimit,omitempty"`
	TedOutLimit                     Money  `json:"tedOutLimit,omitempty"`
}

// LimitUsageCounter tracks used vs available amount for one limit window (cents).
type LimitUsageCounter struct {
	Period         string   `json:"period,omitempty"`
	TotalLimit     *Money   `json:"totalLimit"`
	UsedValue      Money    `json:"usedValue"`
	AvailableValue *Money   `json:"availableValue"`
	UsedPercentage *float64 `json:"usedPercentage"`
	ResetsAt       string   `json:"resetsAt,omitempty"`
}

// LimitUsage is current limit consumption returned with account limits.
type LimitUsage struct {
	Window         string                       `json:"window,omitempty"`
	AsOf           string                       `json:"asOf,omitempty"`
	Aggregate      map[string]LimitUsageCounter `json:"aggregate,omitempty"`
	PerTransaction map[string]*Money            `json:"perTransaction,omitempty"`
}

// AccountLimitsResult is GET /limits/{accountId}.
type AccountLimitsResult struct {
	Limits AccountLimits `json:"limits"`
	Usage  *LimitUsage   `json:"usage,omitempty"`
}

// LimitRequestDocumentInput references a previously uploaded file.
type LimitRequestDocumentInput struct {
	FileID string `json:"fileId"`
}

// LimitRequestCreateParams requests higher Pix day/night limits (cents) with documents.
type LimitRequestCreateParams struct {
	CompanyBankAccountID            string                      `json:"companyBankAccountId"`
	PixDayLimit                     Money                       `json:"pixDayLimit"`
	PixNightLimit                   Money                       `json:"pixNightLimit"`
	Documents                       []LimitRequestDocumentInput `json:"documents"`
	Description                     string                      `json:"description,omitempty"`
	LimitRequestReason              string                      `json:"limitRequestReason,omitempty"`
	PixOutSameHolderDayLimit        *Money                      `json:"pixOutSameHolderDayLimit,omitempty"`
	PixOutDifferentHolderDayLimit   *Money                      `json:"pixOutDifferentHolderDayLimit,omitempty"`
	PixOutSameHolderNightLimit      *Money                      `json:"pixOutSameHolderNightLimit,omitempty"`
	PixOutDifferentHolderNightLimit *Money                      `json:"pixOutDifferentHolderNightLimit,omitempty"`
	PixInSameHolderDayLimit         *Money                      `json:"pixInSameHolderDayLimit,omitempty"`
	PixInDifferentHolderDayLimit    *Money                      `json:"pixInDifferentHolderDayLimit,omitempty"`
	PixInSameHolderNightLimit       *Money                      `json:"pixInSameHolderNightLimit,omitempty"`
	PixInDifferentHolderNightLimit  *Money                      `json:"pixInDifferentHolderNightLimit,omitempty"`
}

// LimitRequestDocument is a file metadata snapshot on a limit request.
type LimitRequestDocument struct {
	FileName    string `json:"fileName,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

// LimitRequest is a limit-increase request under /api/v1/limits.
type LimitRequest struct {
	ID                   string                 `json:"id"`
	CompanyBankAccountID string                 `json:"companyBankAccountId,omitempty"`
	Status               LimitRequestStatus     `json:"status,omitempty"`
	RequestedLimits      AccountLimits          `json:"requestedLimits,omitempty"`
	ApprovedLimits       *AccountLimits         `json:"approvedLimits,omitempty"`
	Documents            []LimitRequestDocument `json:"documents,omitempty"`
	Description          string                 `json:"description,omitempty"`
	CreatedAt            time.Time              `json:"createdAt,omitempty"`
	UpdatedAt            time.Time              `json:"updatedAt,omitempty"`
}

// LimitRequestListParams paginates limit requests.
type LimitRequestListParams struct {
	Skip  int
	Limit int
}

// LimitRequestList is GET /limits/requests.
type LimitRequestList struct {
	LimitRequests []LimitRequest `json:"limitRequests"`
	PageInfo      PageInfo       `json:"pageInfo"`
	Skip          int
	Limit         int
}

type limitRequestEnvelope struct {
	LimitRequest LimitRequest `json:"limitRequest"`
}

func (s *LimitsService) Get(ctx context.Context, accountID string, opts ...RequestOption) (*AccountLimitsResult, error) {
	if accountID == "" {
		return nil, fmt.Errorf("woovi: accountID is required")
	}
	var out AccountLimitsResult
	path := "/api/v1/limits/" + url.PathEscape(accountID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateRequest opens a limit-increase request. Upload docs first via Files.Upload.
func (s *LimitsService) CreateRequest(ctx context.Context, params *LimitRequestCreateParams, opts ...RequestOption) (*LimitRequest, error) {
	if params == nil {
		return nil, fmt.Errorf("woovi: LimitRequestCreateParams is required")
	}
	if params.CompanyBankAccountID == "" {
		return nil, fmt.Errorf("woovi: companyBankAccountId is required")
	}
	if len(params.Documents) == 0 {
		return nil, fmt.Errorf("woovi: documents are required")
	}
	var out limitRequestEnvelope
	if err := s.client.do(ctx, http.MethodPost, "/api/v1/limits/request", nil, params, &out, opts...); err != nil {
		return nil, err
	}
	return &out.LimitRequest, nil
}

// GetRequest retrieves a limit-increase request (poll until decided).
func (s *LimitsService) GetRequest(ctx context.Context, limitRequestID string, opts ...RequestOption) (*LimitRequest, error) {
	if limitRequestID == "" {
		return nil, fmt.Errorf("woovi: limitRequestID is required")
	}
	var out limitRequestEnvelope
	path := "/api/v1/limits/request/" + url.PathEscape(limitRequestID)
	if err := s.client.do(ctx, http.MethodGet, path, nil, nil, &out, opts...); err != nil {
		return nil, err
	}
	return &out.LimitRequest, nil
}

func (s *LimitsService) ListRequests(ctx context.Context, params *LimitRequestListParams, opts ...RequestOption) (*LimitRequestList, error) {
	if params == nil {
		params = &LimitRequestListParams{}
	}
	skip := params.Skip
	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		return nil, fmt.Errorf("woovi: limit must be at most 100")
	}
	q := url.Values{}
	q.Set("skip", strconv.Itoa(skip))
	q.Set("limit", strconv.Itoa(limit))

	var out LimitRequestList
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/limits/request", q, nil, &out, opts...); err != nil {
		return nil, err
	}
	out.Skip = skip
	out.Limit = limit
	normalizePageInfo(&out.PageInfo, skip, limit)
	return &out, nil
}

func (s *LimitsService) ListAllRequests(ctx context.Context, params *LimitRequestListParams, opts ...RequestOption) iter.Seq2[*LimitRequest, error] {
	return func(yield func(*LimitRequest, error) bool) {
		p := LimitRequestListParams{}
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
			page, err := s.ListRequests(ctx, &p, opts...)
			if err != nil {
				yield(nil, err)
				return
			}
			for i := range page.LimitRequests {
				if !yield(&page.LimitRequests[i], nil) {
					return
				}
			}
			if !listPageHasMore(page.PageInfo, len(page.LimitRequests), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}
