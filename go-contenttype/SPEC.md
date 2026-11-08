# Technical Specification: go-contenttype

## 1. Domain & Purpose
`go-contenttype` is a high-speed HTTP content negotiation and media type parsing library for Go. It parses `Content-Type`, `Accept`, `Accept-Language`, and `Accept-Encoding` headers, evaluating quality factors ($q$-values) and type specificity to negotiate the optimal representation for REST APIs and API gateways.

## 2. Architecture & Design Principles
- **RFC 9110 / RFC 7231 Media Type Parser**: Parses MIME types, sub-types, character sets, and parameters (e.g. `charset=utf-8`, `boundary=...`).
- **Quality Factor ($q$-value) & Specificity Sorting**: Accurately orders accepted media types by priority weight ($0.0 \le q \le 1.0$) with tie-breaking preference for specific types over wildcards (`application/json` > `application/*` > `*/*`).
- **Flexible Negotiation Engines**:
  - `Negotiate()`: Matches client `Accept` against server-supported formats.
  - `NegotiateLanguage()`: Matches client `Accept-Language` against supported locales (with sub-tag prefix matching, e.g. `hi` $\to$ `hi-IN`).
  - `NegotiateEncoding()`: Selects supported compression algorithms (`br`, `gzip`, `deflate`).
- **Zero Heavy Dependencies**: Pure standard library (`strings`, `strconv`, `sort`).

## 3. API Surface
- `type MediaType struct`
- `func ParseContentType(header string) (*MediaType, error)`
- `func ParseAccept(header string) []MediaType`
- `func Negotiate(acceptHeader string, supported []string) (string, bool)`
- `func NegotiateLanguage(acceptLangHeader string, supported []string) (string, bool)`
- `func NegotiateEncoding(acceptEncHeader string, supported []string) (string, bool)`

## 4. Error Handling
- `ErrEmptyHeader`: Header string was empty or whitespace.
- `ErrInvalidMediaType`: Media type format lacks required `type/subtype` delimiter.

## 5. Verification & Testing
- Unit tests verify header parsing with parameters, multi-type quality sorting, exact and wildcard negotiation, language prefix matching, and encoding selection.
