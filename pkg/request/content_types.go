package request

// ContentType represents the content type of the request.
type ContentType string

const (
	// ContentTypeJSON is the JSON content type. This is the default.
	ContentTypeJSON ContentType = "application/json"

	// ContentTypeXML is the XML content type.
	ContentTypeXML ContentType = "application/xml"

	// ContentTypeHTML is the HTML content type.
	ContentTypeHTML ContentType = "text/html"

	// ContentTypeText is the text content type.
	ContentTypeText ContentType = "text/plain"

	// ContentTypePng is the png content type.
	ContentTypePng ContentType = "image/png"
)

// String returns the string representation of the ContentType.
func (c ContentType) String() string {
	return string(c)
}

// IsIn returns true if the ContentType is in the list of content types.
func (c ContentType) IsIn(contentTypes ...ContentType) bool {
	if contentTypes == nil {
		return false
	}
	for _, ct := range contentTypes {
		if ct == c {
			return true
		}
	}
	return false
}

// getContentType returns the content type from the string.
func getContentType(contentType string) ContentType {
	switch contentType {
	case "application/json":
		return ContentTypeJSON
	case "application/xml":
		return ContentTypeXML
	case "text/html":
		return ContentTypeHTML
	case "text/plain":
		return ContentTypeText
	case "image/png":
		return ContentTypePng
	default:
		return ContentTypeJSON
	}
}
