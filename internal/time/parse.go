package timex

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var durationRe = regexp.MustCompile(`^(\d+)(h|m|min|hr|hrs|mins|hours?|minutes?)$`)

// Parse attempts to parse a time string in ISO 8601 / RFC3339 formats
func Parse(input string, tz *time.Location) (time.Time, error) {
	if tz == nil {
		tz = time.Local
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, input, tz); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s (use ISO 8601 format, e.g. 2006-01-02 or 2006-01-02T15:04:05)", input)
}

// ParseDuration parses duration strings like "30m", "1h", "1h30m"
func ParseDuration(input string) (time.Duration, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// Try standard Go duration parsing first
	if d, err := time.ParseDuration(input); err == nil {
		return d, nil
	}

	// Try simple patterns like "30m", "1h", "2hr"
	matches := durationRe.FindStringSubmatch(input)
	if matches != nil {
		value, _ := strconv.Atoi(matches[1])
		unit := matches[2]

		switch {
		case strings.HasPrefix(unit, "h"):
			return time.Duration(value) * time.Hour, nil
		case strings.HasPrefix(unit, "m"):
			return time.Duration(value) * time.Minute, nil
		}
	}

	return 0, fmt.Errorf("unable to parse duration: %s", input)
}

// FormatTime formats a time for display
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// StartOfDay returns the start of the day for a given time
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day for a given time
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek returns the start of the week (Monday) for a given time
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday
	}
	return StartOfDay(t.AddDate(0, 0, -(weekday - 1)))
}

// EndOfWeek returns the end of the week (Sunday) for a given time
func EndOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return EndOfDay(t.AddDate(0, 0, 7-weekday))
}
