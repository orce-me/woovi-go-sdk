package woovi_test

import (
	"testing"

	"github.com/orce-me/woovi-go-sdk"
)

func TestMoneyBRLString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   woovi.Money
		want string
	}{
		{woovi.Cents(0), "0.00"},
		{woovi.Cents(1990), "19.90"},
		{woovi.Cents(100), "1.00"},
		{woovi.Cents(-250), "-2.50"},
	}
	for _, tc := range cases {
		if got := tc.in.BRLString(); got != tc.want {
			t.Fatalf("BRLString(%d)=%q want %q", tc.in, got, tc.want)
		}
	}
}
