package woovi

// PageInfo is skip/limit pagination metadata from list endpoints.
type PageInfo struct {
	Skip            int  `json:"skip"`
	Limit           int  `json:"limit"`
	TotalCount      int  `json:"totalCount"`
	HasPreviousPage bool `json:"hasPreviousPage"`
	HasNextPage     bool `json:"hasNextPage"`
}

// HasMore reports whether another page is available.
// Prefers HasNextPage; otherwise uses totalCount.
func (p PageInfo) HasMore() bool {
	if p.HasNextPage {
		return true
	}
	if p.TotalCount > 0 && p.Limit > 0 {
		return p.Skip+p.Limit < p.TotalCount
	}
	return false
}

// listPageHasMore decides if ListAll/ListPages should fetch the next page.
// Uses HasNextPage or totalCount only. A full last page with hasNextPage=false
// must stop; guessing from itemCount caused infinite loops.
func listPageHasMore(info PageInfo, itemCount, limit int) bool {
	_ = itemCount
	if info.Limit <= 0 && limit > 0 {
		info.Limit = limit
	}
	return info.HasMore()
}
