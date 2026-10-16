package domain

import "errors"

// ErrInvalidInput is returned by services when input fails validation.
var ErrInvalidInput = errors.New("invalid input")
