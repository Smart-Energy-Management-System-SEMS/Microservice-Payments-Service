package queries

type GetPaymentMethodsByUserQuery struct {
	UserID string
}

type GetPaymentsByUserQuery struct {
	UserID string
}

type GetPaymentsBySubscriptionQuery struct {
	SubscriptionID string
}
