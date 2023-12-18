package entities

import (
	"fmt"
	"regexp"
	"time"
)

// Date is a custom type for time.Time that marshals to and from RFC3339 date only.
type Date time.Time

// MarshalJSON marshals the date to JSON.
func (d *Date) MarshalJSON() ([]byte, error) {
	return d.MarshalText()
}

// MarshalText marshals the date to text.
func (d *Date) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

// UnmarshalJSON unmarshals the date from JSON.
func (d *Date) UnmarshalJSON(text []byte) error {
	// Remove " from text if present with regex (e.g. "2020-01-27" -> 2020-01-27)
	reg := regexp.MustCompile(`"(.*)"`)
	text = reg.ReplaceAll(text, []byte("$1"))
	return d.UnmarshalText(text)
}

// UnmarshalText unmarshals the date from text.
func (d *Date) UnmarshalText(text []byte) error {
	t, err := time.Parse(time.DateOnly, string(text))
	if err != nil {
		return err
	}
	*d = Date(t)
	return nil
}

// Scan scans the date from a database value.
func (d *Date) Scan(src any) error {
	t, ok := src.(time.Time)
	if !ok {
		return fmt.Errorf("invalid scan, type %T not supported for %T", src, d)
	}
	*d = Date(t)
	return nil
}

// String returns the date as a string.
func (d Date) String() string {
	return time.Time(d).Format(time.DateOnly)
}

// MySQL returns the date as a MySQL string.
func (d Date) Time() time.Time {
	return time.Time(d)
}
