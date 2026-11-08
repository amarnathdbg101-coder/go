package useragent

import (
	"net/http"
	"strings"
	"unicode"
)

// DeviceType defines the category of client device.
type DeviceType string

const (
	DeviceDesktop DeviceType = "Desktop"
	DeviceMobile  DeviceType = "Mobile"
	DeviceTablet  DeviceType = "Tablet"
	DeviceBot     DeviceType = "Bot"
	DeviceUnknown DeviceType = "Unknown"
)

// UserAgent represents parsed details from an HTTP User-Agent string.
type UserAgent struct {
	Raw            string
	Browser        string
	BrowserVersion string
	OS             string
	Device         DeviceType
	IsMobile       bool
	IsTablet       bool
	IsDesktop      bool
	IsBot          bool
}

// Parse extracts browser, OS, version, and device type from a raw User-Agent header string.
func Parse(raw string) *UserAgent {
	ua := &UserAgent{
		Raw:     raw,
		Browser: "Unknown",
		OS:      "Unknown",
		Device:  DeviceUnknown,
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ua
	}

	lower := strings.ToLower(trimmed)

	// 1. Detect Bots / Crawlers
	if detectBot(lower) {
		ua.IsBot = true
		ua.Device = DeviceBot
		ua.Browser = "Bot"
		ua.OS = "Bot"
		return ua
	}

	// 2. Detect OS
	ua.OS = detectOS(lower, trimmed)

	// 3. Detect Device Category (Tablet / Mobile / Desktop)
	if detectTablet(lower) {
		ua.Device = DeviceTablet
		ua.IsTablet = true
	} else if detectMobile(lower) {
		ua.Device = DeviceMobile
		ua.IsMobile = true
	} else {
		ua.Device = DeviceDesktop
		ua.IsDesktop = true
	}

	// 4. Detect Browser & Version
	ua.Browser, ua.BrowserVersion = detectBrowser(trimmed)

	return ua
}

// FromRequest extracts and parses the User-Agent header from an incoming HTTP request.
func FromRequest(r *http.Request) *UserAgent {
	if r == nil {
		return Parse("")
	}
	return Parse(r.UserAgent())
}

func detectBot(lower string) bool {
	botTokens := []string{
		"bot", "spider", "crawl", "slurp", "googlebot", "bingbot", "duckduckbot",
		"baiduspider", "yandex", "facebookexternalhit", "twitterbot", "whatsapp",
		"curl", "wget", "postmanruntime", "insomnia", "python-requests", "httpclient",
	}

	for _, token := range botTokens {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}

func detectOS(lower string, raw string) string {
	switch {
	case strings.Contains(lower, "windows nt 10.0") || strings.Contains(lower, "windows 10") || strings.Contains(lower, "windows 11"):
		return "Windows"
	case strings.Contains(lower, "windows"):
		return "Windows"
	case strings.Contains(lower, "iphone"):
		return "iOS"
	case strings.Contains(lower, "ipad"):
		return "iPadOS"
	case strings.Contains(lower, "macintosh") || strings.Contains(lower, "mac os x"):
		return "macOS"
	case strings.Contains(lower, "android"):
		return "Android"
	case strings.Contains(lower, "cros"):
		return "ChromeOS"
	case strings.Contains(lower, "linux"):
		return "Linux"
	default:
		return "Unknown"
	}
}

func detectTablet(lower string) bool {
	if strings.Contains(lower, "ipad") ||
		strings.Contains(lower, "tablet") ||
		strings.Contains(lower, "playbook") ||
		strings.Contains(lower, "silk") ||
		strings.Contains(lower, "kindle") {
		return true
	}
	// Android tablets typically do not contain "mobile"
	if strings.Contains(lower, "android") && !strings.Contains(lower, "mobile") {
		return true
	}
	return false
}

func detectMobile(lower string) bool {
	return strings.Contains(lower, "mobile") ||
		strings.Contains(lower, "iphone") ||
		strings.Contains(lower, "ipod") ||
		strings.Contains(lower, "android") ||
		strings.Contains(lower, "blackberry") ||
		strings.Contains(lower, "iemobile") ||
		strings.Contains(lower, "opera mini")
}

func detectBrowser(raw string) (string, string) {
	// Order is strictly prioritized because Chrome includes "Safari", Edge includes "Chrome", etc.
	switch {
	case strings.Contains(raw, "Edg/"):
		return "Edge", extractVersion(raw, "Edg/")
	case strings.Contains(raw, "Edge/"):
		return "Edge", extractVersion(raw, "Edge/")
	case strings.Contains(raw, "OPR/"):
		return "Opera", extractVersion(raw, "OPR/")
	case strings.Contains(raw, "SamsungBrowser/"):
		return "Samsung Internet", extractVersion(raw, "SamsungBrowser/")
	case strings.Contains(raw, "Chrome/"):
		return "Chrome", extractVersion(raw, "Chrome/")
	case strings.Contains(raw, "CriOS/"):
		return "Chrome", extractVersion(raw, "CriOS/")
	case strings.Contains(raw, "Firefox/"):
		return "Firefox", extractVersion(raw, "Firefox/")
	case strings.Contains(raw, "FxiOS/"):
		return "Firefox", extractVersion(raw, "FxiOS/")
	case strings.Contains(raw, "Safari/") && strings.Contains(raw, "Version/"):
		return "Safari", extractVersion(raw, "Version/")
	case strings.Contains(raw, "MSIE ") || strings.Contains(raw, "Trident/"):
		return "Internet Explorer", ""
	default:
		return "Other", ""
	}
}

func extractVersion(raw string, prefix string) string {
	idx := strings.Index(raw, prefix)
	if idx == -1 {
		return ""
	}

	start := idx + len(prefix)
	end := start
	for end < len(raw) && (unicode.IsDigit(rune(raw[end])) || raw[end] == '.') {
		end++
	}

	return raw[start:end]
}
