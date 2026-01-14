package domain

import "errors"

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrNotFound          = errors.New("not_found")
	ErrInvalid           = errors.New("invalid")
	ErrInsufficientStock = errors.New("insufficient stock")
)
