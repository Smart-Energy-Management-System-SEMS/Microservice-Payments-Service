// Package domain declares the shared "sentinel errors" of the service. A
// sentinel error is a predefined error value the rest of the code can compare
// against with errors.Is(...). Centralising them here gives every layer a common
// vocabulary, and lets the HTTP layer later map each one to the right status
// code (400, 401, 409, 502, ...).
package domain

import "errors"

var (
	ErrInvalidUUID          = errors.New("invalid uuid")          // a provided id is not a valid UUID
	ErrInvalidAmount        = errors.New("invalid amount")        // amount is zero or negative
	ErrInvalidCurrency      = errors.New("invalid currency")      // currency is missing/blank
	ErrUnauthorizedResource = errors.New("unauthorized resource") // caller does not own the resource
	ErrExternalProvider     = errors.New("external provider error") // Stripe (or similar) failed
	ErrDuplicateWebhook     = errors.New("duplicate webhook")     // this webhook event was already handled
)
