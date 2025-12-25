package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsSupportedCurrency(t *testing.T) {
	testCases := []struct {
		currency string
		expected bool
	}{
		{USD, true},
		{EUR, true},
		{CAD, true},
		{"CNY", false},
		{"GBP", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.currency, func(t *testing.T) {
			result := IsSupportedCurrency(tc.currency)
			require.Equal(t, tc.expected, result)
		})
	}
}
