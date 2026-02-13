package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		phone string
		valid bool
	}{
		{"+6281234567890", true},
		{"+1234567890", true},
		{"081234567890", false},
		{"+0123456", false},
		{"abc", false},
		{"", false},
	}
	for _, tt := range tests {
		err := ValidatePhone(tt.phone)
		if tt.valid {
			assert.NoError(t, err, "phone=%s", tt.phone)
		} else {
			assert.Error(t, err, "phone=%s", tt.phone)
		}
	}
}

func TestSanitizeText(t *testing.T) {
	assert.Equal(t, "hello", SanitizeText("  hello  "))
	assert.Equal(t, "hello world", SanitizeText("<b>hello</b> world"))
}

func TestContainsDangerousContent(t *testing.T) {
	assert.True(t, ContainsDangerousContent("<script>alert(1)</script>"))
	assert.True(t, ContainsDangerousContent("javascript:void(0)"))
	assert.False(t, ContainsDangerousContent("hello world"))
}

func TestValidateLength(t *testing.T) {
	assert.NoError(t, ValidateLength("name", "hi", 10))
	assert.Error(t, ValidateLength("name", "a very long string", 5))
}

func TestValidateRequired(t *testing.T) {
	assert.NoError(t, ValidateRequired("name", "value"))
	assert.Error(t, ValidateRequired("name", ""))
	assert.Error(t, ValidateRequired("name", "   "))
}

func TestValidateFileMIME(t *testing.T) {
	assert.NoError(t, ValidateFileMIME("image/jpeg"))
	assert.NoError(t, ValidateFileMIME("application/pdf"))
	assert.Error(t, ValidateFileMIME("text/plain"))
}

func TestValidateFileSize(t *testing.T) {
	assert.NoError(t, ValidateFileSize(1024))
	assert.Error(t, ValidateFileSize(MaxFileSize+1))
}

func TestSanitizeMessage(t *testing.T) {
	msg, err := SanitizeMessage("hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello", msg)

	longMsg := strings.Repeat("a", MaxMessageLength+1)
	_, err = SanitizeMessage(longMsg)
	assert.Error(t, err)
}

func TestSanitizeName(t *testing.T) {
	name, err := SanitizeName("John", MaxNameLength)
	assert.NoError(t, err)
	assert.Equal(t, "John", name)

	_, err = SanitizeName("", MaxNameLength)
	assert.Error(t, err)

	longName := strings.Repeat("a", MaxNameLength+1)
	_, err = SanitizeName(longName, MaxNameLength)
	assert.Error(t, err)
}
