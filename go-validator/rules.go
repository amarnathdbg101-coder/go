package validator

import (
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	// Precompiled regular expressions for maximum performance
	uuidRegex  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	phoneRegex = regexp.MustCompile(`^\+?[0-9]\d{7,14}$`) // Standard telephone format (e.g. +919876543210, 9876543210, or 01123456789)
	slugRegex  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

// IsEmail verifies if string matches valid RFC 5322 email syntax.
func IsEmail(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 254 {
		return false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s {
		return false
	}
	parts := strings.Split(s, "@")
	return len(parts) == 2 && strings.Contains(parts[1], ".")
}

// IsPhone checks if string is a valid national or international telephone number.
func IsPhone(s string) bool {
	clean := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "-", ""), "(", "")
	clean = strings.ReplaceAll(clean, ")", "")
	return phoneRegex.MatchString(clean)
}

// IsURL checks if string is a valid absolute HTTP or HTTPS URL.
func IsURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// IsUUID checks if string matches canonical UUID v4/v7 format.
func IsUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

// IsSlug checks if string is a safe lowercase URL slug (e.g. "gaming-mouse-rgb").
func IsSlug(s string) bool {
	return slugRegex.MatchString(s)
}

// MinLength checks rune character count (handles UTF-8 multi-byte characters safely).
func MinLength(s string, min int) bool {
	return utf8.RuneCountInString(strings.TrimSpace(s)) >= min
}

// MaxLength checks rune character count.
func MaxLength(s string, max int) bool {
	return utf8.RuneCountInString(strings.TrimSpace(s)) <= max
}

// IsAlphanumeric checks if all characters are letters or digits.
func IsAlphanumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// In checks if string equals any of the allowed options (case-insensitive).
func In(s string, allowed ...string) bool {
	for _, a := range allowed {
		if strings.EqualFold(strings.TrimSpace(s), strings.TrimSpace(a)) {
			return true
		}
	}
	return false
}

// IsStrongPassword checks:
// - Minimum 8 characters
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one numeric digit
// - At least one special symbol
func IsStrongPassword(s string) bool {
	if len(s) < 8 {
		return false
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range s {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

// NoHTML checks that input does not contain dangerous HTML or script tags (Anti-XSS).
func NoHTML(s string) bool {
	lower := strings.ToLower(s)
	return !strings.Contains(lower, "<script") &&
		!strings.Contains(lower, "</script") &&
		!strings.Contains(lower, "javascript:") &&
		!strings.Contains(lower, "<iframe")
}
