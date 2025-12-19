package usecase

import "errors"

var (
	// ErrInvalidKey will be returned is key is empty
	ErrInvalidKey = errors.New("input key is invalid")
	// ErrInvalidLimit will be returned if limit <= 0
	ErrInvalidLimit = errors.New("input limit is invalid")
	// ErrInvalidWindow will be returned if window <= 0
	ErrInvalidWindow = errors.New("input window is invalid")
	// ErrInvalidCost will be returned if cost <= 0
	ErrInvalidCost = errors.New("input cost is invalid")
)
