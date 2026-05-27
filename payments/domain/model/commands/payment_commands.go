package commands

type ProcessPaymentCommand struct {
	SubscriptionID  string
	UserID          string
	PaymentMethodID string
	Amount          float64
	Currency        string
	PaymentMethod   string
}

type HandleStripeWebhookCommand struct {
	Payload   []byte
	Signature string
}
