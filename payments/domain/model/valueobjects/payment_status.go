package valueobjects

// PaymentStatus is a named string type used to model the lifecycle of a payment.
// Defining a dedicated type (instead of a plain string) lets the compiler catch
// mistakes: a function that wants a PaymentStatus will not accept any random
// text by accident.
type PaymentStatus string

// The set of valid payment statuses, written as typed constants. Go has no
// built-in enum, so a block of typed constants is the idiomatic way to express
// "one of these fixed values". The normal happy path is:
// pending -> processing -> processed, while failed/cancelled are end states.
const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusProcessed  PaymentStatus = "processed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
)

// String makes PaymentStatus satisfy the fmt.Stringer interface and gives us a
// clean way to get the underlying text (e.g. when saving to the DB or building
// a JSON response).
func (s PaymentStatus) String() string {
	return string(s)
}
