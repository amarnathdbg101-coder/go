package gocron

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Schedule defines the contract for computing the next activation time.
type Schedule interface {
	Next(t time.Time) time.Time
}

// IntervalSchedule fires at fixed duration intervals.
type IntervalSchedule struct {
	duration time.Duration
}

// Every creates a schedule that triggers every duration d.
func Every(d time.Duration) Schedule {
	return &IntervalSchedule{duration: d}
}

func (s *IntervalSchedule) Next(t time.Time) time.Time {
	return t.Add(s.duration)
}

// CronSchedule implements standard 5-field Unix cron parsing.
type CronSchedule struct {
	expr     string
	minutes  map[int]bool
	hours    map[int]bool
	days     map[int]bool
	months   map[int]bool
	weekdays map[int]bool
}

// ParseCron parses standard 5-field cron strings: "minute hour dom month dow".
// Example: "*/15 * * * *" (every 15 minutes), "0 0 * * *" (midnight).
func ParseCron(expr string) (*CronSchedule, error) {
	fields := strings.Fields(strings.TrimSpace(expr))
	if len(fields) != 5 {
		return nil, fmt.Errorf("%w: got %d fields in %q", ErrInvalidCronExpr, len(fields), expr)
	}

	cs := &CronSchedule{expr: expr}
	var err error

	if cs.minutes, err = parseField(fields[0], 0, 59); err != nil {
		return nil, fmt.Errorf("minute field: %w", err)
	}
	if cs.hours, err = parseField(fields[1], 0, 23); err != nil {
		return nil, fmt.Errorf("hour field: %w", err)
	}
	if cs.days, err = parseField(fields[2], 1, 31); err != nil {
		return nil, fmt.Errorf("dom field: %w", err)
	}
	if cs.months, err = parseField(fields[3], 1, 12); err != nil {
		return nil, fmt.Errorf("month field: %w", err)
	}
	if cs.weekdays, err = parseField(fields[4], 0, 6); err != nil {
		return nil, fmt.Errorf("dow field: %w", err)
	}

	return cs, nil
}

func parseField(field string, min, max int) (map[int]bool, error) {
	res := make(map[int]bool)

	// Comma separated items
	parts := strings.Split(field, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		step := 1

		// Step values (e.g. */5 or 10-30/5)
		if strings.Contains(part, "/") {
			subParts := strings.Split(part, "/")
			if len(subParts) != 2 {
				return nil, ErrInvalidCronExpr
			}
			part = subParts[0]
			s, err := strconv.Atoi(subParts[1])
			if err != nil || s <= 0 {
				return nil, ErrInvalidField
			}
			step = s
		}

		if part == "*" {
			for i := min; i <= max; i += step {
				res[i] = true
			}
			continue
		}

		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, ErrInvalidCronExpr
			}
			start, err1 := strconv.Atoi(rangeParts[0])
			end, err2 := strconv.Atoi(rangeParts[1])
			if err1 != nil || err2 != nil || start < min || end > max || start > end {
				return nil, ErrInvalidField
			}
			for i := start; i <= end; i += step {
				res[i] = true
			}
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil || val < min || val > max {
			return nil, ErrInvalidField
		}
		res[val] = true
	}

	return res, nil
}

// Next finds the next closest activation minute.
func (cs *CronSchedule) Next(t time.Time) time.Time {
	// Round up to next minute (seconds & nanoseconds zeroed)
	t = t.Truncate(time.Minute).Add(time.Minute)

	// Search up to 5 years in future to avoid infinite loop
	limit := t.AddDate(5, 0, 0)

	for t.Before(limit) {
		if !cs.months[int(t.Month())] {
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
			continue
		}

		if !cs.days[t.Day()] || !cs.weekdays[int(t.Weekday())] {
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
			continue
		}

		if !cs.hours[t.Hour()] {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())
			continue
		}

		if !cs.minutes[t.Minute()] {
			t = t.Add(time.Minute)
			continue
		}

		return t
	}

	return time.Time{}
}
