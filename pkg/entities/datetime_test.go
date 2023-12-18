package entities

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDatetime_MarshalJSON(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		d       Datetime
		want    string
		wantErr error
	}{
		{
			name:    "Datetime",
			d:       Datetime(now),
			want:    `"` + now.Format(time.RFC3339) + `"`,
			wantErr: nil,
		},
		{
			name:    "DatetimeZero",
			d:       Datetime(time.Time{}),
			want:    `"0001-01-01T00:00:00Z"`,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.MarshalJSON()
			require.Equal(t, tt.wantErr, err, "Datetime.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			require.Equal(t, tt.want, string(got), "Datetime.MarshalJSON() = %v, want %v", string(got), tt.want)
		})
	}
}

func TestDatetime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		d       Datetime
		data    []byte
		wantErr error
	}{
		{
			name:    "Date",
			d:       Datetime{},
			data:    []byte(`"2020-01-01T00:00:00Z"`),
			wantErr: nil,
		},
		{
			name:    "Invalid",
			d:       Datetime{},
			data:    []byte(`"2020-01-01"`),
			wantErr: errors.New("2020-01-01 is not in the RFC3339 format"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.d.UnmarshalJSON(tt.data)
			require.Equal(t, tt.wantErr, err, "Datetime.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestDatetime_Scan(t *testing.T) {
	now, err := time.Parse("2006-01-02", "2020-01-01")
	require.NoError(t, err, "Failed to parse time")

	tests := []struct {
		name    string
		d       Datetime
		data    any
		wantErr error
	}{
		{
			name:    "Date",
			d:       Datetime{},
			data:    now,
			wantErr: nil,
		},
		{
			name:    "Invalid",
			d:       Datetime{},
			data:    `2020-01-01T00:00:00Z`,
			wantErr: errors.New(`invalid scan, type string not supported for *entities.Datetime`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.d.Scan(tt.data)
			require.Equal(t, tt.wantErr, err, "Date.Scan() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestDatetime_String(t *testing.T) {
	now, err := time.Parse("2006-01-02", "2020-01-01")
	require.NoError(t, err, "Failed to parse time")

	tests := []struct {
		name string
		d    Datetime
		want string
	}{
		{
			name: "Date",
			d:    Datetime(now),
			want: "2020-01-01T00:00:00Z",
		},
		{
			name: "Invalid",
			d:    Datetime{},
			want: "0001-01-01T00:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.d.String(), "Datetime.String() = %v, want %v", tt.d.String(), tt.want)
		})
	}
}
