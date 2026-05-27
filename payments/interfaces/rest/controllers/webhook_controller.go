package controllers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"Microservice-Payments-Service/payments/application/commandservices"
	"Microservice-Payments-Service/payments/domain/model/commands"
	"Microservice-Payments-Service/payments/interfaces/rest/resources"
)

type WebhookController struct {
	commands *commandservices.WebhookCommandService
}

func NewWebhookController(commands *commandservices.WebhookCommandService) *WebhookController {
	return &WebhookController{commands: commands}
}

func (ctl *WebhookController) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/webhooks/stripe", ctl.HandleStripe)
}

func (ctl *WebhookController) HandleStripe(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid payload"})
		return
	}
	event, duplicate, err := ctl.commands.HandleStripe(c.Request.Context(), commands.HandleStripeWebhookCommand{
		Payload:   payload,
		Signature: c.GetHeader("Stripe-Signature"),
	})
	if duplicate {
		c.JSON(http.StatusOK, resources.WebhookResponse{Received: true, Duplicate: true})
		return
	}
	if err != nil {
		RespondError(c, err)
		return
	}
	response := resources.WebhookResponse{Received: true}
	if event != nil {
		response.EventID = event.EventID.String()
	}
	c.JSON(http.StatusOK, response)
}
