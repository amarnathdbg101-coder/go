package valrules

import (
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	regexPAN          = regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]$`)
	regexGSTIN        = regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`)
	regexIFSC         = regexp.MustCompile(`^[A-Z]{4}0[A-Z0-9]{6}$`)
	regexE164         = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	regexIndianMobile = regexp.MustCompile(`^[6-9]\d{9}$`)
	regexMAC          = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	regexUUID         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-8][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
)

// StateCodes contains standard Indian GST state code prefixes.
var StateCodes = map[string]string{
	"01": "Jammu & Kashmir",
	"02": "Himachal Pradesh",
	"03": "Punjab",
	"04": "Chandigarh",
	"05": "Uttarakhand",
	"06": "Haryana",
	"07": "Delhi",
	"08": "Rajasthan",
	"09": "Uttar Pradesh",
	"10": "Bihar",
	"11": "Sikkim",
	"12": "Arunachal Pradesh",
	"13": "Nagaland",
	"14": "Manipur",
	"15": "Mizoram",
	"16": "Tripura",
	"17": "Meghalaya",
	"18": "Assam",
	"19": "West Bengal",
	"20": "Jharkhand",
	"21": "Odisha",
	"22": "Chhattisgarh",
	"23": "Madhya Pradesh",
	"24": "Gujarat",
	"26": "Dadra & Nagar Haveli and Daman & Diu",
	"27": "Maharashtra",
	"29": "Karnataka",
	"30": "Goa",
	"31": "Lakshadweep",
	"32": "Kerala",
	"33": "Tamil Nadu",
	"34": "Puducherry",
	"35": "Andaman & Nicobar Islands",
	"36": "Telangana",
	"37": "Andhra Pradesh",
	"38": "Ladakh",
}

// GSTINDetails contains parsed and validated GSTIN metadata.
type GSTINDetails struct {
	GSTIN     string
	StateCode string
	StateName string
	PAN       string
	EntityNum string
	Checksum  string
}

// ValidateLuhn verifies a string of digits against the Mod-10 Luhn algorithm.
// Commonly used for credit/debit card numbers (Visa, Mastercard, RuPay, Amex) and IMEI numbers.
func ValidateLuhn(number string) bool {
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

// ValidatePAN validates an Indian Income Tax Permanent Account Number (PAN).
// Format: 5 letters, 4 digits, 1 letter (e.g., ABCDE1234F).
// 4th character indicates entity type:
// P: Individual, C: Company, H: HUF, A: AOP, T: Trust, B: BOI, L: Local, J: Artificial Juridical, G: Govt.
func ValidatePAN(pan string) bool {
	pan = strings.TrimSpace(strings.ToUpper(pan))
	if len(pan) != 10 {
		return false
	}
	return regexPAN.MatchString(pan)
}

// ValidateGSTIN validates a 15-character Indian Goods & Services Tax Identification Number.
// Format: 2-digit state code + 10-char PAN + 1 entity digit + 'Z' + 1 checksum character.
func ValidateGSTIN(gstin string) (GSTINDetails, error) {
	gstin = strings.TrimSpace(strings.ToUpper(gstin))
	if !regexGSTIN.MatchString(gstin) {
		return GSTINDetails{}, ErrInvalidGSTIN
	}

	stateCode := gstin[:2]
	stateName, ok := StateCodes[stateCode]
	if !ok {
		return GSTINDetails{}, ErrInvalidGSTIN
	}

	pan := gstin[2:12]
	if !ValidatePAN(pan) {
		return GSTINDetails{}, ErrInvalidGSTIN
	}

	return GSTINDetails{
		GSTIN:     gstin,
		StateCode: stateCode,
		StateName: stateName,
		PAN:       pan,
		EntityNum: string(gstin[12]),
		Checksum:  string(gstin[14]),
	}, nil
}

// ValidateIFSC validates an Indian Financial System Code (IFSC) assigned by RBI.
// Format: 4 letters (bank code) + '0' (reserved) + 6 alphanumeric characters (branch code).
func ValidateIFSC(ifsc string) bool {
	ifsc = strings.TrimSpace(strings.ToUpper(ifsc))
	if len(ifsc) != 11 {
		return false
	}
	return regexIFSC.MatchString(ifsc)
}

// ValidateE164 validates an international phone number conforming to ITU-T E.164.
// Format: '+' followed by country code and subscriber number (max 15 digits).
func ValidateE164(phone string) bool {
	phone = strings.TrimSpace(phone)
	return regexE164.MatchString(phone)
}

// ValidateIndianMobile validates a 10-digit Indian mobile number (must start with 6, 7, 8, or 9).
func ValidateIndianMobile(mobile string) bool {
	clean := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(mobile, "+91", ""), " ", ""))
	if strings.HasPrefix(clean, "0") && len(clean) == 11 {
		clean = clean[1:]
	}
	return regexIndianMobile.MatchString(clean)
}

// ValidateMAC validates an IEEE 802 MAC-48 address (colon or hyphen separated).
func ValidateMAC(mac string) bool {
	return regexMAC.MatchString(strings.TrimSpace(mac))
}

// ValidateUUID validates a standard RFC 4122 / RFC 9562 UUID string.
func ValidateUUID(uuidStr string) bool {
	return regexUUID.MatchString(strings.TrimSpace(uuidStr))
}

// ValidateBase64 verifies if a string is valid standard or URL-safe base64 encoded.
func ValidateBase64(s string) bool {
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

// ValidateHex verifies if a string consists entirely of hexadecimal digits.
func ValidateHex(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 || len(s)%2 != 0 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// ValidateURL validates if a URL string contains a valid scheme (http/https/ftp/ws/wss) and valid host.
func ValidateURL(rawURL string) bool {
	rawURL = strings.TrimSpace(rawURL)
	if len(rawURL) == 0 {
		return false
	}
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Host == "" || u.Scheme == "" {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" && scheme != "ftp" && scheme != "ws" && scheme != "wss" {
		return false
	}
	// Check port if present
	if u.Port() != "" {
		port, err := strconv.Atoi(u.Port())
		if err != nil || port < 1 || port > 65535 {
			return false
		}
	}
	return true
}
