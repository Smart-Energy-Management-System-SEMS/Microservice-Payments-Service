package commandservices

import (
	"context"
	"errors"
	"log"
	"strings"

	"gorm.io/gorm"

	"Microservice-Payments-Service/payments/application/outboundservices"
	"Microservice-Payments-Service/payments/domain/model/commands"
	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/model/valueobjects"
	"Microservice-Payments-Service/payments/domain/repositories"
	"Microservice-Payments-Service/payments/domain/services"
	shareddomain "Microservice-Payments-Service/payments/shared/domain"
)

type PaymentCommandService struct {
	payments       repositories.PaymentRepository
	paymentMethods repositories.PaymentMethodRepository
	invoices       repositories.InvoiceRepository
	provider       outboundservices.PaymentProvider
	publisher      outboundservices.EventPublisher
	statusMapper   services.PaymentStatusMapper
}

func NewPaymentCommandService(payments repositories.PaymentRepository, paymentMethods repositories.PaymentMethodRepository, invoices repositories.InvoiceRepository, provider outboundservices.PaymentProvider, publisher outboundservices.EventPublisher) *PaymentCommandService {
	return &PaymentCommandService{payments: payments, paymentMethods: paymentMethods, invoices: invoices, provider: provider, publisher: publisher, statusMapper: services.PaymentStatusMapper{}}
}

func (s *PaymentCommandService) Process(ctx context.Context, command commands.ProcessPaymentCommand) (*entities.Payment, *entities.Invoice, error) {
	subscriptionID, err := shareddomain.ParseID(command.SubscriptionID)
	if err != nil {
		return nil, nil, shareddomain.ErrInvalidUUID
	}
	userID, err := shareddomain.ParseID(command.UserID)
	if err != nil {
		return nil, nil, shareddomain.ErrInvalidUUID
	}
	paymentMethodID, err := shareddomain.ParseID(command.PaymentMethodID)
	if err != nil {
		return nil, nil, shareddomain.ErrInvalidUUID
	}

	method, err := s.paymentMethods.FindByID(ctx, paymentMethodID)
	if err != nil {
		return nil, nil, err
	}
	if method.UserID != userID {
		return nil, nil, shareddomain.ErrUnauthorizedResource
	}

	paymentMethodName := strings.TrimSpace(command.PaymentMethod)
	if paymentMethodName == "" {
		paymentMethodName = method.Type
	}
	payment, err := entities.NewPayment(subscriptionID, userID, paymentMethodID, command.Amount, command.Currency, paymentMethodName)
	if err != nil {
		return nil, nil, err
	}

	intent, err := s.provider.CreatePaymentIntent(ctx, outboundservices.CreatePaymentIntentRequest{
		Amount:                payment.Amount,
		Currency:              payment.Currency,
		StripePaymentMethodID: method.StripePaymentMethodID,
		UserID:                userID.String(),
		SubscriptionID:        subscriptionID.String(),
		PaymentID:             payment.PaymentID.String(),
	})
	if err != nil {
		payment.MarkFailed("")
		if saveErr := s.payments.Save(ctx, &payment); saveErr != nil {
			log.Printf("could not persist failed payment: %v", saveErr)
		}
		_ = s.publisher.PublishPaymentFailed(ctx, payment)
		return &payment, nil, err
	}

	s.applyProviderStatus(&payment, intent.ID, intent.Status)
	if err := s.payments.Save(ctx, &payment); err != nil {
		return nil, nil, err
	}
	return s.afterPaymentStatusChanged(ctx, &payment)
}

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
