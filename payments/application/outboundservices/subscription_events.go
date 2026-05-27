package outboundservices

type SubscriptionCreatedEvent struct {
	SubscriptionID string `json:"subscription_id"`
	UserID         string `json:"user_id"`
}

type SubscriptionRenewalRequestedEvent struct {
	SubscriptionID  string  `json:"subscription_id"`
	UserID          string  `json:"user_id"`
	PaymentMethodID string  `json:"payment_method_id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
}

type SubscriptionCancelledEvent struct {
	SubscriptionID string `json:"subscription_id"`
	UserID         string `json:"user_id"`
	Reason         string `json:"reason"`
}
