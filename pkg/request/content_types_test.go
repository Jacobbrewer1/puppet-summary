package request

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestContentType_String tests the String method of the ContentType type.
func Test_getContentType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ContentType
	}{
		{
			name:  "empty",
			input: "",
			want:  ContentTypeJSON,
		},
		{
			name:  "json",
			input: "application/json",
			want:  ContentTypeJSON,
		},
		{
			name:  "xml",
			input: "application/xml",
			want:  ContentTypeXML,
		},
		{
			name:  "html",
			input: "text/html",
			want:  ContentTypeHTML,
		},
		{
			name:  "text",
			input: "text/plain",
			want:  ContentTypeText,
		},
		{
			name:  "png",
			input: "image/png",
			want:  ContentTypePng,
		},
		{
			name:  "invalid",
			input: "invalid",
			want:  ContentTypeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, getContentType(tt.input))
		})
	}
}
