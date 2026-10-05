package woovi

import "fmt"

// Money is an amount in Brazilian real cents. Never use float64 for Pix values.
type Money int64

// Cents returns Money from an integer cent amount.
func Cents(v int64) Money { return Money(v) }

// Int64 returns the amount in cents.
func (m Money) Int64() int64 { return int64(m) }

// BRLString formats the amount as a BRL decimal string (e.g. "19.90").
func (m Money) BRLString() string {
	neg := m < 0
	if neg {
		m = -m
	}
	reais := m / 100
	cents := m % 100
	s := fmt.Sprintf("%d.%02d", reais, cents)
	if neg {
		return "-" + s
	}
	return s
}
