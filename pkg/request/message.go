package request

import "fmt"

// Message represents a message response.
type Message struct {
	Message string `json:"message" xml:"message"`
}

// NewMessage creates a new Message.
func NewMessage(message string, args ...any) *Message {
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(message, args...)
	} else {
		msg = message
	}
	return &Message{
		Message: msg,
	}
}
