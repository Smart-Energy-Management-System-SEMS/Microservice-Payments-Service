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

type PaymentMethodController struct {
	commands *commandservices.PaymentMethodCommandService
	queries  *queryservices.PaymentMethodQueryService
}

func NewPaymentMethodController(commands *commandservices.PaymentMethodCommandService, queries *queryservices.PaymentMethodQueryService) *PaymentMethodController {
	return &PaymentMethodController{commands: commands, queries: queries}
}

func (ctl *PaymentMethodController) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/payment-methods", ctl.Register)
	group.GET("/payment-methods/user/:userId", ctl.FindByUser)
	group.PUT("/payment-methods/:paymentMethodId/default", ctl.SetDefault)
	group.DELETE("/payment-methods/:paymentMethodId", ctl.Delete)
}

func (ctl *PaymentMethodController) Register(c *gin.Context) {
	var request resources.RegisterPaymentMethodRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, sharedinterfaces.ErrorResponse{Error: err.Error()})
		return
	}
	method, err := ctl.commands.Register(c.Request.Context(), commands.RegisterPaymentMethodCommand{
		UserID: request.UserID,
		Type: request.Type,
		StripePaymentMethodID: request.StripePaymentMethodID,
		IsDefault: request.IsDefault,
	})
	if err != nil {
		sharedinterfaces.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, transform.ToPaymentMethodResponse(*method))
}

func (ctl *PaymentMethodController) FindByUser(c *gin.Context) {
	methods, err := ctl.queries.FindByUser(c.Request.Context(), queries.GetPaymentMethodsByUserQuery{UserID: c.Param("userId")})
	if err != nil {
		sharedinterfaces.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, transform.ToPaymentMethodResponses(methods))
}

func (ctl *PaymentMethodController) SetDefault(c *gin.Context) {
	method, err := ctl.commands.SetDefault(c.Request.Context(), commands.SetDefaultPaymentMethodCommand{PaymentMethodID: c.Param("paymentMethodId")})
	if err != nil {
		sharedinterfaces.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, transform.ToPaymentMethodResponse(*method))
}

func (ctl *PaymentMethodController) Delete(c *gin.Context) {
	if err := ctl.commands.Delete(c.Request.Context(), commands.DeletePaymentMethodCommand{PaymentMethodID: c.Param("paymentMethodId")}); err != nil {
		sharedinterfaces.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
