// Package validator provides input validation and sanitization for API requests.
package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// Max lengths for various fields.
const (
	MaxMessageLength  = 10000
	MaxNameLength     = 50
	MaxGroupLength    = 100
	MaxStatusLength   = 140
	MaxDocTitleLength = 200
	MaxBlockLength    = 50000
	MaxSearchLength   = 200
	MaxTagLength      = 50
	MaxFileSize       = 25 * 1024 * 1024 // 25 MB
)

// Allowed file MIME types.
var AllowedImageMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// AllowedFileMIME contains acceptable document MIME types.
var AllowedFileMIME = map[string]bool{
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"application/vnd.oasis.opendocument.text":                           true,
	"application/vnd.oasis.opendocument.spreadsheet":                    true,
}

// dangerousPatterns matches common XSS/injection patterns.
var dangerousPatterns = regexp.MustCompile(`(?i)<script|javascript:|on\w+\s*=|<iframe|<object|<embed|<link\s|<meta\s`)

// htmlTagPattern matches HTML tags.
var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// ValidatePhone checks if a string looks like an E.164 phone number.
func ValidatePhone(phone string) error {
	matched, _ := regexp.MatchString(`^\+[1-9]\d{6,14}$`, phone)
	if !matched {
		return fmt.Errorf("invalid phone format, expected E.164 (e.g., +628123456789)")
	}
	return nil
}

// SanitizeText strips HTML tags and dangerous patterns from text.
func SanitizeText(s string) string {
	// Remove HTML tags
	s = htmlTagPattern.ReplaceAllString(s, "")
	// Trim whitespace
	return strings.TrimSpace(s)
}

// ContainsDangerousContent checks for XSS/injection patterns.
func ContainsDangerousContent(s string) bool {
	return dangerousPatterns.MatchString(s)
}

// ValidateLength checks that a string does not exceed the given max length.
func ValidateLength(field, value string, maxLen int) error {
	if len(value) > maxLen {
		return fmt.Errorf("%s exceeds maximum length of %d characters", field, maxLen)
	}
	return nil
}

// ValidateRequired checks that a string is not empty after trimming.
func ValidateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}

// ValidateFileMIME checks that a MIME type is allowed for upload.
func ValidateFileMIME(mimeType string) error {
	if AllowedImageMIME[mimeType] || AllowedFileMIME[mimeType] {
		return nil
	}
	return fmt.Errorf("file type %s is not allowed", mimeType)
}

// ValidateFileSize checks that a file does not exceed the maximum size.
func ValidateFileSize(size int64) error {
	if size > MaxFileSize {
		return fmt.Errorf("file size %d exceeds maximum of %d bytes", size, MaxFileSize)
	}
	return nil
}

// SanitizeMessage sanitizes a chat message: strip HTML, enforce length.
func SanitizeMessage(content string) (string, error) {
	content = SanitizeText(content)
	if err := ValidateLength("message", content, MaxMessageLength); err != nil {
		return "", err
	}
	return content, nil
}

// SanitizeName sanitizes a user/group name.
func SanitizeName(name string, maxLen int) (string, error) {
	name = SanitizeText(name)
	if err := ValidateRequired("name", name); err != nil {
		return "", err
	}
	if err := ValidateLength("name", name, maxLen); err != nil {
		return "", err
	}
	return name, nil
}
