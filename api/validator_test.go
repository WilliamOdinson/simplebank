package api

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestValidCurrencies(t *testing.T) {
	v := validator.New()
	v.RegisterValidation("currency", validCurrencies)

	type testStruct struct {
		Currency string `validate:"currency"`
	}

	testCases := []struct {
		name     string
		currency string
		valid    bool
	}{
		{"USD", "USD", true},
		{"EUR", "EUR", true},
		{"CAD", "CAD", true},
		{"Invalid", "CNY", false},
		{"Empty", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(testStruct{Currency: tc.currency})
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestValidCurrencies_NonStringField(t *testing.T) {
	v := validator.New()
	v.RegisterValidation("currency", validCurrencies)

	type nonStringStruct struct {
		Currency int `validate:"currency"`
	}

	err := v.Struct(nonStringStruct{Currency: 123})
	require.Error(t, err)
}
