package controllers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"Microservice-Payments-Service/payments/application/commandservices"
	"Microservice-Payments-Service/payments/domain/model/commands"
	"Microservice-Payments-Service/payments/interfaces/rest/resources"
)

// WebhookController exposes the endpoint Stripe calls to notify us of events.
type WebhookController struct {
	commands *commandservices.WebhookCommandService
}

func NewWebhookController(commands *commandservices.WebhookCommandService) *WebhookController {
	return &WebhookController{commands: commands}
}

func (ctl *WebhookController) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/webhooks/stripe", ctl.HandleStripe)
}

// HandleStripe handles "POST /webhooks/stripe". Two details are specific to
// webhooks:
//   - We read the RAW request body with io.ReadAll. The exact bytes matter
//     because Stripe's signature is computed over them; parsing into a struct
//     first would break verification.
//   - We pass the "Stripe-Signature" header along so the service can verify the
//     request truly came from Stripe.
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
	// A duplicate is NOT an error: we answer 200 OK so Stripe knows we received
	// it and stops retrying, but we flag it as already handled.
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
