package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"Microservice-Payments-Service/payments/application/commandservices"
	"Microservice-Payments-Service/payments/application/queryservices"
	"Microservice-Payments-Service/payments/domain/model/commands"
	"Microservice-Payments-Service/payments/domain/model/queries"
	"Microservice-Payments-Service/payments/interfaces/rest/resources"
	"Microservice-Payments-Service/payments/interfaces/rest/transform"
	sharedinterfaces "Microservice-Payments-Service/payments/shared/interfaces"
)

type PaymentController struct {
	commands        *commandservices.PaymentCommandService
	queries         *queryservices.PaymentQueryService
	defaultCurrency string
}

func NewPaymentController(commands *commandservices.PaymentCommandService, queries *queryservices.PaymentQueryService, defaultCurrency string) *PaymentController {
	return &PaymentController{commands: commands, queries: queries, defaultCurrency: defaultCurrency}
}

func (ctl *PaymentController) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/payments/process", ctl.Process)
	group.GET("/payments/:paymentId", ctl.FindByID)
	group.GET("/payments/user/:userId", ctl.FindByUser)
	group.GET("/payments/subscription/:subscriptionId", ctl.FindBySubscription)
}

func (ctl *PaymentController) Process(c *gin.Context) {
	var request resources.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, sharedinterfaces.ErrorResponse{Error: err.Error()})
		return
	}
	if request.Currency == "" {
		request.Currency = ctl.defaultCurrency
	}
	payment, invoice, err := ctl.commands.Process(c.Request.Context(), commands.ProcessPaymentCommand{
		SubscriptionID:  request.SubscriptionID,
		UserID:          request.UserID,
		PaymentMethodID: request.PaymentMethodID,
		Amount:          request.Amount,
		Currency:        request.Currency,
		PaymentMethod:   request.PaymentMethod,
	})
	if err != nil {
		sharedinterfaces.RespondError(c, err)
		return
	}
	response := resources.ProcessPaymentResponse{Payment: responsePointer(transform.ToPaymentResponse(*payment))}
	if invoice != nil {
		response.Invoice = invoiceResponsePointer(transform.ToInvoiceResponse(*invoice))
	}
	c.JSON(http.StatusCreated, response)
}

func (ctl *PaymentController) FindByID(c *gin.Context) {
	payment, err := ctl.queries.FindByID(c.Request.Context(), c.Param("paymentId"))
	if err != nil {
		sharedinterfaces.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, transform.ToPaymentResponse(*payment))
}

func (ctl *PaymentController) FindByUser(c *gin.Context) {
	payments, err := ctl.queries.FindByUser(c.Request.Context(), queries.GetPaymentsByUserQuery{UserID: c.Param("userId")})
	if err != nil {
		sharedinterfaces.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, transform.ToPaymentResponses(payments))
}

func (ctl *PaymentController) FindBySubscription(c *gin.Context) {
	payments, err := ctl.queries.FindBySubscription(c.Request.Context(), queries.GetPaymentsBySubscriptionQuery{SubscriptionID: c.Param("subscriptionId")})
	if err != nil {
		sharedinterfaces.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, transform.ToPaymentResponses(payments))
}

func responsePointer(value resources.PaymentResponse) *resources.PaymentResponse {
	return &value
}

func invoiceResponsePointer(value resources.InvoiceResponse) *resources.InvoiceResponse {
	return &value
}
