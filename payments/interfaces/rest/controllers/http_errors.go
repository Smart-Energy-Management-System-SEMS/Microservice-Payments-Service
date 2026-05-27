package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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
	}
	c.JSON(status, ErrorResponse{Error: err.Error()})
}
