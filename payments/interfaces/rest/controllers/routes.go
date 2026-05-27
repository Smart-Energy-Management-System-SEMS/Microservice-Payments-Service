package controllers

import "github.com/gin-gonic/gin"

type Controllers struct {
	PaymentMethods *PaymentMethodController
	Payments       *PaymentController
	Invoices       *InvoiceController
	Webhooks       *WebhookController
}

func RegisterRoutes(router *gin.Engine, basePath string, controllers Controllers) {
	api := router.Group(basePath)
	controllers.PaymentMethods.RegisterRoutes(api)
	controllers.Payments.RegisterRoutes(api)
	controllers.Invoices.RegisterRoutes(api)
	controllers.Webhooks.RegisterRoutes(api)
}
