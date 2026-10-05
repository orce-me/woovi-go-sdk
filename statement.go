package woovi

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// StatementEntryType is CREDIT or DEBIT on the account statement.
type StatementEntryType string

const (
	StatementEntryCredit StatementEntryType = "CREDIT"
	StatementEntryDebit  StatementEntryType = "DEBIT"
)

// StatementsService is GET /api/v1/statement.
type StatementsService struct {
	client *Client
}

// StatementEntry is one ledger line (value and balance in cents).
type StatementEntry struct {
	ID            string             `json:"id"`
	Time          time.Time          `json:"time"`
	Description   string             `json:"description,omitempty"`
	Balance       Money              `json:"balance"`
	Value         Money              `json:"value"`
	Type          StatementEntryType `json:"type"`
	OperationType string             `json:"operationType,omitempty"`
}

// StatementListParams paginates statements (RFC3339 start/end; optional companyBankAccount).
type StatementListParams struct {
	Skip               int
	Limit              int
	Start              string // RFC3339
	End                string // RFC3339
	CompanyBankAccount string
}

// StatementList is one page of GET /statement.
type StatementList struct {
	Statements []StatementEntry `json:"statements"`
	PageInfo   PageInfo         `json:"pageInfo"`
	Skip       int
	Limit      int
}

func (s *StatementsService) List(ctx context.Context, params *StatementListParams, opts ...RequestOption) (*StatementList, error) {
	if params == nil {
		params = &StatementListParams{}
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
	if params.Start != "" {
		q.Set("start", params.Start)
	}
	if params.End != "" {
		q.Set("end", params.End)
	}
	if params.CompanyBankAccount != "" {
		q.Set("companyBankAccount", params.CompanyBankAccount)
	}

	raw := json.RawMessage{}
	if err := s.client.do(ctx, http.MethodGet, "/api/v1/statement", q, nil, &raw, opts...); err != nil {
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
		return nil, fmt.Errorf("woovi: decode statement: %w", err)
	}
	out.Statements = flat
	normalizePageInfo(&out.PageInfo, skip, limit)
	return out, nil
}

func (s *StatementsService) ListAll(ctx context.Context, params *StatementListParams, opts ...RequestOption) iter.Seq2[*StatementEntry, error] {
	return func(yield func(*StatementEntry, error) bool) {
		for page, err := range s.ListPages(ctx, params, opts...) {
			if err != nil {
				yield(nil, err)
				return
			}
			for i := range page.Statements {
				if !yield(&page.Statements[i], nil) {
					return
				}
			}
		}
	}
}

func (s *StatementsService) ListPages(ctx context.Context, params *StatementListParams, opts ...RequestOption) iter.Seq2[*StatementList, error] {
	return func(yield func(*StatementList, error) bool) {
		p := StatementListParams{}
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
			if !listPageHasMore(page.PageInfo, len(page.Statements), page.Limit) {
				return
			}
			p.Skip += p.Limit
		}
	}
}
