# Technical Specification: go-units

## 1. Domain & Purpose
`go-units` is a human-readable parsing and formatting engine for data sizes (bytes, KB, MB, GB, KiB, MiB, GiB) and human durations (supporting days `d` and weeks `w` which standard Go `time.ParseDuration` lacks). It is essential for configuration parsing, storage quota enforcement, and console analytics reporting.

## 2. Architecture & Design Principles
- **Dual Data Size Standards**: Distinguishes between binary IEC units (`KiB`, `MiB`, `GiB` = $1024^n$) and decimal SI units (`KB`, `MB`, `GB` = $1000^n$).
- **Extended Duration Parsing**: Supports compound expressions including weeks (`w`), days (`d`), hours (`h`), minutes (`m`), seconds (`s`), and sub-second units (`ms`, `us`, `ns`).
- **Concise Formatting**: Formats durations into human-friendly strings (`"2d 4h 30m"` instead of `"52h30m0s"`).
- **Zero Heavy Dependencies**: Pure Go standard library (`strconv`, `strings`, `fmt`, `time`, `unicode`).

## 3. API Surface
- `func ParseBytes(s string) (int64, error)`
- `func FormatBytes(b int64) string` (IEC binary: KiB, MiB, GiB)
- `func FormatBytesDecimal(b int64) string` (SI decimal: KB, MB, GB)
- `func ParseHumanDuration(s string) (time.Duration, error)`
- `func FormatHumanDuration(d time.Duration) string`

## 4. Error Handling
- `ErrInvalidByteString`: Unrecognized or unparseable byte string.
- `ErrInvalidDuration`: Unrecognized duration unit.

## 5. Verification & Testing
- Unit tests verify byte parsing across SI and IEC units, human duration parsing with days/weeks, format scaling, and invalid input rejections.
