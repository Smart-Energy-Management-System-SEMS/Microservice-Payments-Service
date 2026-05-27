package domain

import "errors"

var (
	ErrInvalidUUID          = errors.New("invalid uuid")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrInvalidCurrency      = errors.New("invalid currency")
	ErrUnauthorizedResource = errors.New("unauthorized resource")
	ErrExternalProvider     = errors.New("external provider error")
	ErrDuplicateWebhook     = errors.New("duplicate webhook")
)
