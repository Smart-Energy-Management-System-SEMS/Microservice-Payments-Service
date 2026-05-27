package resources

type ProcessPaymentRequest struct {
	SubscriptionID  string  `json:"subscription_id" binding:"required"`
	UserID          string  `json:"user_id" binding:"required"`
	PaymentMethodID string  `json:"payment_method_id" binding:"required"`
	Amount          float64 `json:"amount" binding:"required"`
	Currency        string  `json:"currency"`
	PaymentMethod   string  `json:"payment_method"`
}

type PaymentResponse struct {
	PaymentID             string  `json:"payment_id"`
	SubscriptionID        string  `json:"subscription_id"`
	UserID                string  `json:"user_id"`
	PaymentMethodID       string  `json:"payment_method_id"`
	Amount                float64 `json:"amount"`
	Currency              string  `json:"currency"`
	Status                string  `json:"status"`
	PaymentMethod         string  `json:"payment_method"`
	StripePaymentIntentID string  `json:"stripe_payment_intent_id"`
	PaidAt                *string `json:"paid_at"`
	CreatedAt             string  `json:"created_at"`
}

type ProcessPaymentResponse struct {
	Payment *PaymentResponse `json:"payment"`
	Invoice *InvoiceResponse `json:"invoice,omitempty"`
}
