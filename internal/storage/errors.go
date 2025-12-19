package storage

import "errors"

var (
	// ErrKeyNotFound will be returned when a given identifier does not have a corresponding value
	ErrKeyNotFound = errors.New("key not found")
)
