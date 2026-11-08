package gounits

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	Byte = 1
	KiB  = 1024 * Byte
	MiB  = 1024 * KiB
	GiB  = 1024 * MiB
	TiB  = 1024 * GiB
	PiB  = 1024 * TiB

	KB = 1000 * Byte
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB
)

// ParseBytes parses a human-readable byte size string (e.g. "10MB", "500KiB", "1.5GB", "64k").
func ParseBytes(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrInvalidByteString
	}

	// Split numeric part and unit part
	numEnd := 0
	for numEnd < len(s) {
		r := rune(s[numEnd])
		if unicode.IsDigit(r) || r == '.' || r == '-' {
			numEnd++
		} else {
			break
		}
	}

	if numEnd == 0 {
		return 0, ErrInvalidByteString
	}

	numStr := s[:numEnd]
	unitStr := strings.TrimSpace(s[numEnd:])

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, ErrInvalidByteString
	}

	multiplier := float64(1)
	switch strings.ToUpper(unitStr) {
	case "", "B", "BYTES":
		multiplier = 1
	case "K", "KIB":
		multiplier = KiB
	case "KB":
		multiplier = KB
	case "M", "MIB":
		multiplier = MiB
	case "MB":
		multiplier = MB
	case "G", "GIB":
		multiplier = GiB
	case "GB":
		multiplier = GB
	case "T", "TIB":
		multiplier = TiB
	case "TB":
		multiplier = TB
	case "P", "PIB":
		multiplier = PiB
	case "PB":
		multiplier = PB
	default:
		return 0, fmt.Errorf("%w: unrecognized unit %q", ErrInvalidByteString, unitStr)
	}

	return int64(val * multiplier), nil
}

// FormatBytes formats byte counts using binary IEC units (KiB, MiB, GiB, TiB).
func FormatBytes(b int64) string {
	const unit = 1024.0
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := unit, 0
	for n := b / int64(unit); n >= int64(unit); n /= int64(unit) {
		div *= unit
		exp++
	}
	suffixes := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}
	return fmt.Sprintf("%.2f %s", float64(b)/div, suffixes[exp])
}

// FormatBytesDecimal formats byte counts using decimal SI units (KB, MB, GB, TB).
func FormatBytesDecimal(b int64) string {
	const unit = 1000.0
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := unit, 0
	for n := b / int64(unit); n >= int64(unit); n /= int64(unit) {
		div *= unit
		exp++
	}
	suffixes := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.2f %s", float64(b)/div, suffixes[exp])
}
