package validator

import (
	"encoding/base64"
	"encoding/hex"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	// Precompiled regular expressions for maximum performance
	uuidRegex         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	phoneRegex        = regexp.MustCompile(`^\+?[0-9]\d{7,14}$`)
	e164Regex         = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	indianMobileRegex = regexp.MustCompile(`^[6-9]\d{9}$`)
	slugRegex         = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	panRegex          = regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]$`)
	gstinRegex        = regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`)
	ifscRegex         = regexp.MustCompile(`^[A-Z]{4}0[A-Z0-9]{6}$`)
	macRegex          = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
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

// IsE164 checks if string conforms strictly to international ITU-T E.164 phone specification.
func IsE164(s string) bool {
	return e164Regex.MatchString(strings.TrimSpace(s))
}

// IsIndianMobile validates a 10-digit Indian mobile number (starts with 6, 7, 8, or 9).
func IsIndianMobile(s string) bool {
	clean := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s, "+91", ""), " ", ""))
	if strings.HasPrefix(clean, "0") && len(clean) == 11 {
		clean = clean[1:]
	}
	return indianMobileRegex.MatchString(clean)
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

// IsLuhn validates a credit/debit card number or IMEI using the Luhn Mod-10 algorithm.
func IsLuhn(number string) bool {
	clean := strings.ReplaceAll(strings.ReplaceAll(number, " ", ""), "-", "")
	if len(clean) < 8 || len(clean) > 19 {
		return false
	}

	sum := 0
	alternate := false

	for i := len(clean) - 1; i >= 0; i-- {
		r := rune(clean[i])
		if !unicode.IsDigit(r) {
			return false
		}
		digit := int(r - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// IsPAN validates an Indian Income Tax Permanent Account Number (PAN).
// Format: 5 uppercase letters, 4 digits, 1 uppercase letter.
func IsPAN(pan string) bool {
	pan = strings.TrimSpace(strings.ToUpper(pan))
	if len(pan) != 10 {
		return false
	}
	return panRegex.MatchString(pan)
}

// IsGSTIN validates a 15-character Indian Goods & Services Tax Identification Number.
func IsGSTIN(gstin string) bool {
	gstin = strings.TrimSpace(strings.ToUpper(gstin))
	if !gstinRegex.MatchString(gstin) {
		return false
	}
	// Verify embedded PAN (chars 2:12)
	return IsPAN(gstin[2:12])
}

// IsIFSC validates an Indian Financial System Code (IFSC) for banking transactions.
// Format: 4 letters + '0' + 6 alphanumeric characters.
func IsIFSC(ifsc string) bool {
	ifsc = strings.TrimSpace(strings.ToUpper(ifsc))
	if len(ifsc) != 11 {
		return false
	}
	return ifscRegex.MatchString(ifsc)
}

// IsMAC validates an IEEE 802 MAC address (colon or hyphen separated).
func IsMAC(mac string) bool {
	return macRegex.MatchString(strings.TrimSpace(mac))
}

// IsBase64 verifies if a string is valid standard or URL-safe base64.
func IsBase64(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}
	if _, err := base64.StdEncoding.DecodeString(s); err == nil {
		return true
	}
	if _, err := base64.URLEncoding.DecodeString(s); err == nil {
		return true
	}
	return false
}

// IsHex verifies if a string is valid hexadecimal.
func IsHex(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 || len(s)%2 != 0 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}
