package entities

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDate_MarshalJSON(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		d       Date
		want    string
		wantErr error
	}{
		{
			name:    "Date",
			d:       Date(now),
			want:    `"` + now.Format("2006-01-02") + `"`,
			wantErr: nil,
		},
		{
			name:    "DateZero",
			d:       Date(time.Time{}),
			want:    `"0001-01-01"`,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.MarshalJSON()
			require.Equal(t, tt.wantErr, err, "Date.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			require.Equal(t, tt.want, string(got), "Date.MarshalJSON() = %v, want %v", string(got), tt.want)
		})
	}
}

func TestDate_MarshalText(t *testing.T) {
	now, err := time.Parse("2006-01-02", "2020-01-01")
	require.NoError(t, err, "Failed to parse time")

	tests := []struct {
		name    string
		d       Date
		want    string
		wantErr error
	}{
		{
			name:    "Date",
			d:       Date(now),
			want:    `"2020-01-01"`,
			wantErr: nil,
		},
		{
			name:    "Invalid",
			d:       Date{},
			want:    `"0001-01-01"`,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.MarshalText()
			require.Equal(t, tt.wantErr, err, "Date.MarshalText() error = %v, wantErr %v", err, tt.wantErr)
			require.Equal(t, tt.want, string(got), "Date.MarshalText() = %v, want %v", string(got), tt.want)
		})
	}
}

func TestDate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		d       Date
		data    []byte
		wantErr error
	}{
		{
			name:    "Date",
			d:       Date{},
			data:    []byte(`"2020-01-01"`),
			wantErr: nil,
		},
		{
			name: "Invalid",
			d:    Date{},
			data: []byte(`"2020-01-01T00:00:00Z"`),
			wantErr: &time.ParseError{
				Layout:     "2006-01-02",
				Value:      "2020-01-01T00:00:00Z",
				LayoutElem: "",
				ValueElem:  "T00:00:00Z",
				Message:    ": extra text: \"T00:00:00Z\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.d.UnmarshalJSON(tt.data)
			require.Equal(t, tt.wantErr, err, "Date.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestDate_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		d       Date
		data    []byte
		wantErr error
	}{
		{
			name:    "Date",
			d:       Date{},
			data:    []byte(`2020-01-01`),
			wantErr: nil,
		},
		{
			name: "Invalid",
			d:    Date{},
			data: []byte(`"2020-01-01T00:00:00Z"`),
			wantErr: &time.ParseError{
				Layout:     `2006-01-02`,
				Value:      `"2020-01-01T00:00:00Z"`,
				LayoutElem: "2006",
				ValueElem:  `"2020-01-01T00:00:00Z"`,
				Message:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.d.UnmarshalText(tt.data)
			require.Equal(t, tt.wantErr, err, "Date.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestDate_Scan(t *testing.T) {
	now, err := time.Parse("2006-01-02", "2020-01-01")
	require.NoError(t, err, "Failed to parse time")

	tests := []struct {
		name    string
		d       Date
		data    any
		wantErr error
	}{
		{
			name:    "Date",
			d:       Date{},
			data:    now,
			wantErr: nil,
		},
		{
			name:    "Invalid",
			d:       Date{},
			data:    `2020-01-01T00:00:00Z`,
			wantErr: errors.New(`invalid scan, type string not supported for *entities.Date`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.d.Scan(tt.data)
			require.Equal(t, tt.wantErr, err, "Date.Scan() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func TestDate_String(t *testing.T) {
	now, err := time.Parse("2006-01-02", "2020-01-01")
	require.NoError(t, err, "Failed to parse time")

	tests := []struct {
		name string
		d    Date
		want string
	}{
		{
			name: "Date",
			d:    Date(now),
			want: "2020-01-01",
		},
		{
			name: "Invalid",
			d:    Date{},
			want: "0001-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.d.String(), "Date.String() = %v, want %v", tt.d.String(), tt.want)
		})
	}
}

func TestDate_MySQL(t *testing.T) {
	now, err := time.Parse("2006-01-02", "2020-01-01")
	require.NoError(t, err, "Failed to parse time")

	tests := []struct {
		name string
		d    Date
		want string
	}{
		{
			name: "Date",
			d:    Date(now),
			want: "2020-01-01",
		},
		{
			name: "Invalid",
			d:    Date{},
			want: "0001-01-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.d.Time().Format(time.DateOnly),
				"Date.MySQL() = %v, want %v", tt.d.Time().Format(time.DateOnly), tt.want)
		})
	}
}
