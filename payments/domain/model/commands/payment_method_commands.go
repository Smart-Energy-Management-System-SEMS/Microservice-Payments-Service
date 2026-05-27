package commands

type RegisterPaymentMethodCommand struct {
	UserID                string
	Type                  string
	StripePaymentMethodID string
	IsDefault             bool
}

type SetDefaultPaymentMethodCommand struct {
	PaymentMethodID string
}

type DeletePaymentMethodCommand struct {
	PaymentMethodID string
}
