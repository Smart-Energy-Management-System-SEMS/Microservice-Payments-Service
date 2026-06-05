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
// PaymentCommandService gestiona los comandos de pago en la capa de aplicación.
// Coordina las transacciones de pago, genera facturas e interactúa con proveedores externos.

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
// Process procesa un comando de pago.
// Valida los IDs, verifica la autorización del usuario, crea una intención de pago
// con el proveedor externo y maneja los cambios de estado resultantes.
// Retorna el pago, la factura generada (si aplica) y un error si ocurre.

func (s *PaymentCommandService) Process(ctx context.Context, command commands.ProcessPaymentCommand) (*entities.Payment, *entities.Invoice, error) {
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

	method, err := s.paymentMethods.FindByID(ctx, paymentMethodID)
	if err != nil {
		return nil, nil, err
	}
	if method.UserID != userID {
		return nil, nil, paymentdomain.ErrUnauthorizedResource
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
// MarkFromProvider actualiza el estado de un pago basado en la información del proveedor.
// Se utiliza para webhook callbacks del proveedor de pago.
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
// applyProviderStatus mapea el estado del proveedor externo al estado interno del pago.
// Actualiza la entidad de pago con el estado y la intención de pago correspondiente.

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
// afterPaymentStatusChanged maneja las acciones que deben ocurrir después de un cambio de estado.
// Si el pago se procesó correctamente, genera una factura y publica eventos.
// Si falló o se canceló, publica un evento de fallo.

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
// ensureInvoice garantiza que existe una factura para un pago específico.
// Retorna la factura existente o crea una nueva si no existe.

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
