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

func listPageHasMore(info PageInfo, itemCount, limit int) bool {
	if info.HasMore() {
		return true
	}
	if info.TotalCount == 0 && !info.HasNextPage && itemCount >= limit && limit > 0 {
		return true
	}
	return false
}
