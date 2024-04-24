package entities

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// Datetime represents a datetime.
type Datetime time.Time

// MarshalJSON implements the json.Marshaler interface.
func (d *Datetime) MarshalJSON() ([]byte, error) {
	// Marshal the time.
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Datetime) UnmarshalJSON(text []byte) error {
	// Remove " from text if present with regex (e.g. "2020-01-01T00:00:00Z" -> 2020-01-01T00:00:00Z)
	reg := regexp.MustCompile(`"(.*)"`)
	text = reg.ReplaceAll(text, []byte("$1"))

	// Parse the time.
	t, err := time.Parse(time.RFC3339, string(text))
	if err != nil {
		return fmt.Errorf("%s is not in the RFC3339 format", text)
	}
	*d = Datetime(t)
	return nil
}

func (d *Datetime) MarshalBSON() ([]byte, error) {
	if d == nil {
		return bson.Marshal(nil)
	}
	return bson.Marshal(d.String())
}

func (d *Datetime) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if d == nil {
		return bson.TypeNull, nil, nil
	}
	return bson.MarshalValue(d.String())
}

func (d *Datetime) UnmarshalBSON(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}
	str := strings.Trim(string(bytes), `"`)
	// Remove all escape characters.
	str = strings.ReplaceAll(str, `\`, ``)

	// Transform \u0015\u0000\u0000\u00002017-07-29T23:17:01Z\u0000 to 2017-07-29T23:17:01Z
	str = regexp.MustCompile(`[^a-zA-Z0-9-:]`).ReplaceAllString(str, ``)

	// Parse the time.
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return fmt.Errorf("%s is not in the RFC3339 format", str)
	}
	*d = Datetime(t)
	return nil
}

// Scan implements the sql.Scanner interface.
func (d *Datetime) Scan(src any) error {
	// Parse the time.
	str := fmt.Sprintf("%s", src)
	t, err := time.Parse(time.DateTime, str)
	if err != nil {
		return fmt.Errorf("%s is not in the RFC3339 format", str)
	}
	*d = Datetime(t)
	return nil
}

// String implements the fmt.Stringer interface.
func (d Datetime) String() string {
	return time.Time(d).Format(time.RFC3339)
}

func (d Datetime) Time() time.Time {
	return time.Time(d)
}
