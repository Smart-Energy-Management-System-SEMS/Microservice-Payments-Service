package resources

type RegisterPaymentMethodRequest struct {
	UserID                string `json:"user_id" binding:"required"`
	Type                  string `json:"type"`
	StripePaymentMethodID string `json:"stripe_payment_method_id" binding:"required"`
	IsDefault             bool   `json:"is_default"`
}

type PaymentMethodResponse struct {
	PaymentMethodID       string `json:"payment_method_id"`
	UserID                string `json:"user_id"`
	Type                  string `json:"type"`
	Brand                 string `json:"brand"`
	Last4                 string `json:"last4"`
	ExpMonth              int    `json:"exp_month"`
	ExpYear               int    `json:"exp_year"`
	StripePaymentMethodID string `json:"stripe_payment_method_id"`
	IsDefault             bool   `json:"is_default"`
	CreatedAt             string `json:"created_at"`
}
