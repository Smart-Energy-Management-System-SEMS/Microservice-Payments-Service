package interfaces

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	shareddomain "Microservice-Payments-Service/payments/shared/domain"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func RespondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := err.Error()

	switch {
	case errors.Is(err, shareddomain.ErrInvalidUUID),
		errors.Is(err, shareddomain.ErrInvalidAmount),
		errors.Is(err, shareddomain.ErrInvalidCurrency):
		status = http.StatusBadRequest
	case errors.Is(err, shareddomain.ErrNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		status = http.StatusNotFound
		message = shareddomain.ErrNotFound.Error()
	case errors.Is(err, shareddomain.ErrUnauthorizedResource):
		status = http.StatusForbidden
	case errors.Is(err, shareddomain.ErrDuplicateWebhook):
		status = http.StatusOK
	case errors.Is(err, shareddomain.ErrExternalProvider):
		status = http.StatusBadGateway
	}

	c.JSON(status, ErrorResponse{Error: message})
}
