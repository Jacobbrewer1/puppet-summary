package entities

import (
	"fmt"
	"time"
)

type Duration time.Duration

// MarshalJSON marshals the duration to JSON.
// We want to marshal to a max of days (e.g. 1d2h3m4s)
func (d *Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Time().String() + `"`), nil
}

// UnmarshalJSON unmarshals the duration from JSON.
//
// Remove all quites from text if present with regex (e.g. "1h30m" -> 1h30m)
// reg := regexp.MustCompile(`"(.*)"`)
// text = reg.ReplaceAll(text, []byte("$1"))
// Parse the duration from a string
func (d *Duration) UnmarshalJSON(text []byte) error {
	t, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("%s is not in the duration format", text)
	}
	*d = Duration(t)
	return nil
}

// String returns the duration as a string.
func (d Duration) String() string {
	// Calulate the hours, minutes, and seconds.
	t := time.Duration(d)
	hours := t / time.Hour
	t -= hours * time.Hour
	minutes := t / time.Minute
	t -= minutes * time.Minute
	seconds := t / time.Second

	// Build the string.
	str := ""
	if hours > 0 {
		str += fmt.Sprintf("%dh", hours)
	}
	if minutes > 0 {
		str += fmt.Sprintf("%dm", minutes)
	}
	if seconds > 0 {
		str += fmt.Sprintf("%ds", seconds)
	}
	return str
}

// PrettyString returns the duration as a pretty string.
func (d Duration) PrettyString() string {
	// Calulate the days, hours, minutes, and seconds.
	days := d.Time() / (time.Hour * 24)
	hours := (d.Time() - (days * time.Hour * 24)) / time.Hour
	minutes := (d.Time() - (days * time.Hour * 24) - (hours * time.Hour)) / time.Minute
	seconds := (d.Time() - (days * time.Hour * 24) - (hours * time.Hour) - (minutes * time.Minute)) / time.Second

	// Build the string.
	str := ""
	if days > 0 {
		str += fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		str += fmt.Sprintf("%dh", hours)
	}
	if minutes > 0 {
		str += fmt.Sprintf("%dm", minutes)
	}
	if seconds > 0 {
		str += fmt.Sprintf("%ds", seconds)
	}
	return str
}

// Scan converts the string to a duration.
func (d *Duration) Scan(src any) error {
	str := fmt.Sprintf("%s", src)
	t, err := time.ParseDuration(str)
	if err != nil {
		return fmt.Errorf("%s is not in the duration format", str)
	}
	*d = Duration(t)
	return nil
}

func (d *Duration) Time() time.Duration {
	return time.Duration(*d)
}
