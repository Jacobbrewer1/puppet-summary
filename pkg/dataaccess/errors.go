package dataaccess

import "errors"

var (
	// ErrDuplicate is the error for a duplicate record.
	ErrDuplicate = errors.New("duplicate record")

	// ErrNotFound is the error for a record not found.
	ErrNotFound = errors.New("record not found")
)
