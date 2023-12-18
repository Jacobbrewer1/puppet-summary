package request

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    *Message
	}{
		{
			name:    "NewMessage",
			message: "test",
			want:    &Message{Message: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMessage(tt.message)
			require.Equal(t, tt.want, got, "NewMessage() = %v, want %v", got, tt.want)
		})
	}
}
