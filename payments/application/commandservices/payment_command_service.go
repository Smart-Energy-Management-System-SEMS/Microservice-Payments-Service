package commandservices

import (
	"context"
	"errors"
	"log"
	"strings"

	"gorm.io/gorm"

	"Microservice-Payments-Service/payments/application/outboundservices"
	paymentdomain "Microservice-Payments-Service/payments/domain"
	"Microservice-Payments-Service/payments/domain/model/commands"
	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/model/valueobjects"
	"Microservice-Payments-Service/payments/domain/repositories"
	"Microservice-Payments-Service/payments/domain/services"
)

// PaymentCommandService holds the write use cases for payments (the "command"
// side of CQRS, where commands change state). It lives in the application layer:
// it orchestrates repositories, the external payment provider and the event
// publisher, but the actual business rules live in the domain entities.
//
// Every dependency is an interface (repositories, provider, publisher), which is
// dependency injection: the service depends on abstractions, so the real Stripe
// adapter or database can be swapped for fakes in tests.
type PaymentCommandService struct {
	payments       repositories.PaymentRepository
	paymentMethods repositories.PaymentMethodRepository
	invoices       repositories.InvoiceRepository
	provider       outboundservices.PaymentProvider
	publisher      outboundservices.EventPublisher
	statusMapper   services.PaymentStatusMapper
}

// NewPaymentCommandService wires up the dependencies. The statusMapper is a
// small stateless domain helper, so we just create it inline.
func NewPaymentCommandService(payments repositories.PaymentRepository, paymentMethods repositories.PaymentMethodRepository, invoices repositories.InvoiceRepository, provider outboundservices.PaymentProvider, publisher outboundservices.EventPublisher) *PaymentCommandService {
	return &PaymentCommandService{payments: payments, paymentMethods: paymentMethods, invoices: invoices, provider: provider, publisher: publisher, statusMapper: services.PaymentStatusMapper{}}
}

// Process is the main use case: charge a customer for a subscription. It returns
// three values (payment, invoice, error) which is idiomatic Go. The invoice may
// be nil when the charge did not succeed.
func (s *PaymentCommandService) Process(ctx context.Context, command commands.ProcessPaymentCommand) (*entities.Payment, *entities.Invoice, error) {
	// The incoming ids are strings, so first parse and validate them as UUIDs.
	subscriptionID, err := paymentdomain.ParseID(command.SubscriptionID)
	if err != nil {
		return nil, nil, paymentdomain.ErrInvalidUUID
	}
	userID, err := paymentdomain.ParseID(command.UserID)
	if err != nil {
		return nil, nil, paymentdomain.ErrInvalidUUID
	}
	paymentMethodID, err := paymentdomain.ParseID(command.PaymentMethodID)
	if err != nil {
		return nil, nil, paymentdomain.ErrInvalidUUID
	}

	// Load the payment method and verify it belongs to this user. This is an
	// authorization check: it stops a user from charging someone else's card.
	method, err := s.paymentMethods.FindByID(ctx, paymentMethodID)
	if err != nil {
		return nil, nil, err
	}
	if method.UserID != userID {
		return nil, nil, paymentdomain.ErrUnauthorizedResource
	}

	// Use the provided method label, or fall back to the stored type.
	paymentMethodName := strings.TrimSpace(command.PaymentMethod)
	if paymentMethodName == "" {
		paymentMethodName = method.Type
	}
	// Build the payment via the domain factory (which validates amount/currency).
	payment, err := entities.NewPayment(subscriptionID, userID, paymentMethodID, command.Amount, command.Currency, paymentMethodName)
	if err != nil {
		return nil, nil, err
	}

	// Ask the external provider (Stripe) to actually create the charge.
	intent, err := s.provider.CreatePaymentIntent(ctx, outboundservices.CreatePaymentIntentRequest{
		Amount:                payment.Amount,
		Currency:              payment.Currency,
		StripePaymentMethodID: method.StripePaymentMethodID,
		UserID:                userID.String(),
		SubscriptionID:        subscriptionID.String(),
		PaymentID:             payment.PaymentID.String(),
	})
	// If the provider call fails, we still record the attempt as a failed
	// payment and announce it, then return the error. We log (but don't fail on)
	// a save error here because the original provider error is the real problem
	// to report. The "_ =" deliberately ignores the publish result: emitting the
	// event is best-effort and must not mask the provider error.
	if err != nil {
		payment.MarkFailed("")
		if saveErr := s.payments.Save(ctx, &payment); saveErr != nil {
			log.Printf("could not persist failed payment: %v", saveErr)
		}
		_ = s.publisher.PublishPaymentFailed(ctx, payment)
		return &payment, nil, err
	}

	// Translate the provider's status into our own status, persist, then run the
	// follow-up side effects (invoice + events).
	s.applyProviderStatus(&payment, intent.ID, intent.Status)
	if err := s.payments.Save(ctx, &payment); err != nil {
		return nil, nil, err
	}
	return s.afterPaymentStatusChanged(ctx, &payment)
}

// MarkFromProvider updates an existing payment when the provider later tells us
// its real outcome (typically via a webhook). It finds the payment by its Stripe
// intent id, re-applies the status, saves, and runs the same follow-up logic.
func (s *PaymentCommandService) MarkFromProvider(ctx context.Context, stripePaymentIntentID string, providerStatus string) (*entities.Payment, *entities.Invoice, error) {
	payment, err := s.payments.FindByStripePaymentIntentID(ctx, stripePaymentIntentID)
	if err != nil {
		return nil, nil, err
	}
	s.applyProviderStatus(payment, stripePaymentIntentID, providerStatus)
	if err := s.payments.Update(ctx, payment); err != nil {
		return nil, nil, err
	}
	return s.afterPaymentStatusChanged(ctx, payment)
}

// applyProviderStatus maps the provider's raw status string to one of our domain
// statuses and calls the matching method on the payment. Anything we don't
// explicitly recognise is treated as "still processing" (the default case),
// which is the safe assumption.
func (s *PaymentCommandService) applyProviderStatus(payment *entities.Payment, stripePaymentIntentID string, providerStatus string) {
	switch s.statusMapper.FromStripe(providerStatus) {
	case valueobjects.PaymentStatusProcessed:
		payment.MarkProcessed(stripePaymentIntentID)
	case valueobjects.PaymentStatusFailed:
		payment.MarkFailed(stripePaymentIntentID)
	case valueobjects.PaymentStatusCancelled:
		payment.MarkCancelled(stripePaymentIntentID)
	default:
		payment.MarkProcessing(stripePaymentIntentID)
	}
}

// afterPaymentStatusChanged centralises the side effects that depend on the new
// status, so both Process and MarkFromProvider behave identically:
//   - processed  -> make sure an invoice exists and publish the success events.
//   - failed/cancelled -> publish a failure event.
// Keeping this in one place avoids duplicating the rules and risking the two
// flows drifting apart.
func (s *PaymentCommandService) afterPaymentStatusChanged(ctx context.Context, payment *entities.Payment) (*entities.Payment, *entities.Invoice, error) {
	if payment.Status == valueobjects.PaymentStatusProcessed {
		invoice, err := s.ensureInvoice(ctx, payment)
		if err != nil {
			return payment, nil, err
		}
		_ = s.publisher.PublishPaymentProcessed(ctx, *payment, *invoice)
		_ = s.publisher.PublishInvoiceGenerated(ctx, *invoice)
		return payment, invoice, nil
	}
	if payment.Status == valueobjects.PaymentStatusFailed || payment.Status == valueobjects.PaymentStatusCancelled {
		_ = s.publisher.PublishPaymentFailed(ctx, *payment)
	}
	return payment, nil, nil
}

// ensureInvoice returns the existing invoice for a payment, or creates one if it
// is missing. This makes the operation "idempotent": if a webhook arrives twice
// for the same successful payment, we will not generate a second invoice.
//
// The error handling is the key part: a successful lookup means the invoice
// already exists; a gorm.ErrRecordNotFound means we must create it; any OTHER
// error is a real database problem and is returned.
func (s *PaymentCommandService) ensureInvoice(ctx context.Context, payment *entities.Payment) (*entities.Invoice, error) {
	existing, err := s.invoices.FindByPaymentID(ctx, payment.PaymentID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	invoice := entities.NewInvoice(payment.PaymentID, payment.Amount, "")
	if err := s.invoices.Save(ctx, &invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}
