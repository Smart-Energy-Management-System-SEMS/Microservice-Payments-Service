package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"Microservice-Payments-Service/payments/application/queryservices"
	"Microservice-Payments-Service/payments/interfaces/rest/transform"
)

type InvoiceController struct {
	queries *queryservices.InvoiceQueryService
}

func NewInvoiceController(queries *queryservices.InvoiceQueryService) *InvoiceController {
	return &InvoiceController{queries: queries}
}

func (ctl *InvoiceController) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/invoices/:invoiceId", ctl.FindByID)
	group.GET("/invoices/payment/:paymentId", ctl.FindByPaymentID)
}

func (ctl *InvoiceController) FindByID(c *gin.Context) {
	invoice, err := ctl.queries.FindByID(c.Request.Context(), c.Param("invoiceId"))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, transform.ToInvoiceResponse(*invoice))
}

func (ctl *InvoiceController) FindByPaymentID(c *gin.Context) {
	invoice, err := ctl.queries.FindByPaymentID(c.Request.Context(), c.Param("paymentId"))
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, transform.ToInvoiceResponse(*invoice))
}
