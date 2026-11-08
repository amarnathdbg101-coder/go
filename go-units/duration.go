package gounits

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ParseHumanDuration parses duration strings including days ('d') and weeks ('w') (e.g. "2w 3d 4h", "48h", "30m").
func ParseHumanDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ErrInvalidDuration
	}

	var total time.Duration
	var numBuf strings.Builder

	runes := []rune(s)
	n := len(runes)

	for i := 0; i < n; i++ {
		r := runes[i]
		if unicode.IsSpace(r) {
			continue
		}

		if unicode.IsDigit(r) || r == '.' || r == '-' {
			numBuf.WriteRune(r)
			continue
		}

		// Found unit letter
		if numBuf.Len() == 0 {
			return 0, ErrInvalidDuration
		}

		val, err := strconv.ParseFloat(numBuf.String(), 64)
		if err != nil {
			return 0, ErrInvalidDuration
		}
		numBuf.Reset()

		// Read unit token (could be single char 'd' or multi-char 'ms', 'us', 'ns')
		var unitBuf strings.Builder
		for i < n && unicode.IsLetter(runes[i]) {
			unitBuf.WriteRune(runes[i])
			i++
		}
		i-- // step back for outer loop increment

		unit := strings.ToLower(unitBuf.String())
		var unitDur time.Duration

		switch unit {
		case "w", "week", "weeks":
			unitDur = 7 * 24 * time.Hour
		case "d", "day", "days":
			unitDur = 24 * time.Hour
		case "h", "hr", "hour", "hours":
			unitDur = time.Hour
		case "m", "min", "minute", "minutes":
			unitDur = time.Minute
		case "s", "sec", "second", "seconds":
			unitDur = time.Second
		case "ms", "millisec", "millisecond":
			unitDur = time.Millisecond
		case "us", "µs", "microsecond":
			unitDur = time.Microsecond
		case "ns", "nanosecond":
			unitDur = time.Nanosecond
		default:
			return 0, fmt.Errorf("%w: unknown duration unit %q", ErrInvalidDuration, unit)
		}

		total += time.Duration(val * float64(unitDur))
	}

	if numBuf.Len() > 0 {
		return 0, ErrInvalidDuration
	}

	return total, nil
}

// FormatHumanDuration formats time.Duration into concise human-readable components (e.g. "2d 4h 30m").
func FormatHumanDuration(d time.Duration) string {
	if d < 0 {
		return "-" + FormatHumanDuration(-d)
	}
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1000.0)
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}

	days := int(d / (24 * time.Hour))
	d -= time.Duration(days) * 24 * time.Hour

	hours := int(d / time.Hour)
	d -= time.Duration(hours) * time.Hour

	minutes := int(d / time.Minute)
	d -= time.Duration(minutes) * time.Minute

	seconds := int(d / time.Second)

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 && days == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	if len(parts) == 0 {
		return "0s"
	}

	return strings.Join(parts, " ")
}
