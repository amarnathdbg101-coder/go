# Technical Specification: go-useragent

## 1. Domain & Purpose
`go-useragent` is a high-speed, zero-allocation HTTP `User-Agent` header parser in Go. It classifies inbound web requests by extracting the client's browser, browser version, operating system, device family (Desktop, Mobile, Tablet, Bot), and web crawler identities for authentication auditing and responsive routing without regex performance penalties.

## 2. Architecture & Design Principles
- **Fast Token & Substring Classification**: Bypasses slow regular expression overhead by utilizing prioritized string containment checks.
- **Strict Browser Hierarchy**: Enforces correct resolution precedence (Edge > Opera > Samsung Browser > Chrome > Firefox > Safari > IE) to eliminate false positives in composite user agent strings.
- **Granular Device Categorization**: Classifies requests into `DeviceDesktop`, `DeviceMobile`, `DeviceTablet`, `DeviceBot`, or `DeviceUnknown`.
- **Bot & Spider Identification**: Detects automated clients (Googlebot, Bingbot, curl, Postman, social media link scrapers).
- **HTTP Middleware Friendly**: Direct `FromRequest(r *http.Request)` integration.
- **Zero Heavy Dependencies**: Pure Go standard library (`strings`, `net/http`, `unicode`).

## 3. API Surface
- `type DeviceType string`
- `type UserAgent struct`
- `func Parse(raw string) *UserAgent`
- `func FromRequest(r *http.Request) *UserAgent`

## 4. Error Handling
- `ErrEmptyUserAgent`: Indicates blank or missing header (falls back safely to `DeviceUnknown`).

## 5. Verification & Testing
- Unit tests verify major desktop browsers (Chrome, Firefox, Edge), mobile platforms (iOS iPhone, Android), tablets (iPad, Android tablets), search engine crawlers, and HTTP request extractions.
