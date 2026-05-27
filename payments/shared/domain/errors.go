package domain

import "errors"

var (
	ErrInvalidUUID          = errors.New("invalid uuid")
	ErrInvalidAmount        = errors.New("amount must be greater than zero")
	ErrInvalidCurrency      = errors.New("currency is required")
	ErrNotFound             = errors.New("resource not found")
	ErrUnauthorizedResource = errors.New("resource does not belong to owner")
	ErrDuplicateWebhook     = errors.New("webhook event already processed")
	ErrExternalProvider     = errors.New("external payment provider error")
)
