package phone_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/pkg/phone"
)

func TestNormalize_IndonesianNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"leading zero", "081234567890", "+6281234567890"},
		{"with country code", "+6281234567890", "+6281234567890"},
		{"with 62 prefix", "6281234567890", "+6281234567890"},
		{"with spaces", "0812 3456 7890", "+6281234567890"},
		{"with dashes", "0812-3456-7890", "+6281234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := phone.Normalize(tt.input, "ID")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalize_InternationalNumbers(t *testing.T) {
	result, err := phone.Normalize("+14155552671", "")
	require.NoError(t, err)
	assert.Equal(t, "+14155552671", result)
}

func TestNormalize_InvalidNumbers(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"too short", "0812"},
		{"empty", ""},
		{"letters only", "abcdef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := phone.Normalize(tt.input, "ID")
			assert.Error(t, err)
		})
	}
}

func TestValidate(t *testing.T) {
	assert.True(t, phone.Validate("+6281234567890"))
	assert.True(t, phone.Validate("+14155552671"))
	assert.False(t, phone.Validate("081234567890"))
	assert.False(t, phone.Validate("invalid"))
	assert.False(t, phone.Validate(""))
}

func TestHash_Consistency(t *testing.T) {
	h1 := phone.Hash("+6281234567890")
	h2 := phone.Hash("+6281234567890")
	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 64) // SHA-256 hex = 64 chars

	// Different phones should produce different hashes
	h3 := phone.Hash("+6289876543210")
	assert.NotEqual(t, h1, h3)
}
