package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	paymentdomain "Microservice-Payments-Service/payments/domain"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func RespondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, paymentdomain.ErrInvalidUUID),
		errors.Is(err, paymentdomain.ErrInvalidAmount),
		errors.Is(err, paymentdomain.ErrInvalidCurrency):
		status = http.StatusBadRequest
	case errors.Is(err, paymentdomain.ErrUnauthorizedResource):
		status = http.StatusForbidden
	case errors.Is(err, gorm.ErrRecordNotFound):
		status = http.StatusNotFound
	case errors.Is(err, paymentdomain.ErrDuplicateWebhook):
		status = http.StatusConflict
	case errors.Is(err, paymentdomain.ErrExternalProvider):
		status = http.StatusBadGateway
	case isBadRequestDatabaseError(err):
		status = http.StatusBadRequest
	case isConflictDatabaseError(err):
		status = http.StatusConflict
	}
	c.JSON(status, ErrorResponse{Error: err.Error()})
}

func isConflictDatabaseError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505", "23503":
			return true
		}
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "foreign key constraint")
}

func isBadRequestDatabaseError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23502", "23514", "22P02":
			return true
		}
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "null value") ||
		strings.Contains(message, "violates check constraint") ||
		strings.Contains(message, "invalid input syntax")
}
