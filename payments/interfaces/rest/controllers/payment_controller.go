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
)

// PaymentController is the REST entry point for payments. In the interfaces
// layer a controller only handles HTTP: it reads the request, calls the
// application services, and shapes the response. It holds both a command service
// (writes) and a query service (reads), matching the CQRS split.
type PaymentController struct {
	commands        *commandservices.PaymentCommandService
	queries         *queryservices.PaymentQueryService
	defaultCurrency string
}

func NewPaymentController(commands *commandservices.PaymentCommandService, queries *queryservices.PaymentQueryService, defaultCurrency string) *PaymentController {
	return &PaymentController{commands: commands, queries: queries, defaultCurrency: defaultCurrency}
}

// RegisterRoutes maps URLs and HTTP verbs to handler methods, following REST
// conventions (POST to create, GET to read). ":paymentId" etc. are path
// parameters Gin extracts from the URL.
func (ctl *PaymentController) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/payments/process", ctl.Process)
	group.GET("/payments/:paymentId", ctl.FindByID)
	group.GET("/payments/user/:userId", ctl.FindByUser)
	group.GET("/payments/subscription/:subscriptionId", ctl.FindBySubscription)
}

// Process handles "POST /payments/process". It shows the typical controller
// flow: bind+validate the JSON body, fall back to the default currency if none
// was sent, delegate to the command service, and translate the result (or error)
// into an HTTP response.
func (ctl *PaymentController) Process(c *gin.Context) {
	var request resources.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if request.Currency == "" {
		request.Currency = ctl.defaultCurrency
	}
	// c.Request.Context() forwards the request's context so a cancelled request
	// stops the work downstream too.
	payment, invoice, err := ctl.commands.Process(c.Request.Context(), commands.ProcessPaymentCommand{
		SubscriptionID:  request.SubscriptionID,
		UserID:          request.UserID,
		PaymentMethodID: request.PaymentMethodID,
		Amount:          request.Amount,
		Currency:        request.Currency,
		PaymentMethod:   request.PaymentMethod,
	})
	if err != nil {
		// RespondError centralises turning a domain error into the right HTTP
		// status code, so each handler stays short.
		RespondError(c, err)
		return
	}
	// We never expose domain entities directly; transform builds response DTOs.
	// The invoice is only attached when one was actually created (it may be nil).
	response := resources.ProcessPaymentResponse{Payment: responsePointer(transform.ToPaymentResponse(*payment))}
	if invoice != nil {
		response.Invoice = invoiceResponsePointer(transform.ToInvoiceResponse(*invoice))
	}
	c.JSON(http.StatusCreated, response)
}

// FindByID handles "GET /payments/:paymentId". c.Param reads the id from the URL.
func (ctl *PaymentController) FindByID(c *gin.Context) {
	payment, err := ctl.queries.FindByID(c.Request.Context(), c.Param("paymentId"))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, transform.ToPaymentResponse(*payment))
}

// FindByUser handles "GET /payments/user/:userId" — all payments of one user.
func (ctl *PaymentController) FindByUser(c *gin.Context) {
	payments, err := ctl.queries.FindByUser(c.Request.Context(), queries.GetPaymentsByUserQuery{UserID: c.Param("userId")})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, transform.ToPaymentResponses(payments))
}

// FindBySubscription handles "GET /payments/subscription/:subscriptionId".
func (ctl *PaymentController) FindBySubscription(c *gin.Context) {
	payments, err := ctl.queries.FindBySubscription(c.Request.Context(), queries.GetPaymentsBySubscriptionQuery{SubscriptionID: c.Param("subscriptionId")})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, transform.ToPaymentResponses(payments))
}

// These two tiny helpers exist because Go does not let you take the address of a
// function's return value directly (e.g. &transform.ToPaymentResponse(...)).
// Passing the value through a helper gives it an addressable local variable, so
// we can store a pointer to it in the optional response fields.
func responsePointer(value resources.PaymentResponse) *resources.PaymentResponse {
	return &value
}

func invoiceResponsePointer(value resources.InvoiceResponse) *resources.InvoiceResponse {
	return &value
}
