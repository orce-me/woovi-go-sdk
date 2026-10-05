package woovi

// String returns a pointer to s for optional JSON fields.
func String(s string) *string { return &s }

// Int returns a pointer to v for optional JSON fields.
func Int(v int) *int { return &v }

// Int64 returns a pointer to v for optional JSON fields.
func Int64(v int64) *int64 { return &v }

// Bool returns a pointer to v for optional JSON fields.
func Bool(v bool) *bool { return &v }
