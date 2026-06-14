package commandservices

import (
	"context"
	"log"
	"strings"

	"Microservice-Payments-Service/payments/application/outboundservices"
	paymentdomain "Microservice-Payments-Service/payments/domain"
	"Microservice-Payments-Service/payments/domain/model/commands"
	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/repositories"
)

type PaymentMethodCommandService struct {
	repository repositories.PaymentMethodRepository
	provider   outboundservices.PaymentProvider
	publisher  outboundservices.EventPublisher
}

func NewPaymentMethodCommandService(repository repositories.PaymentMethodRepository, provider outboundservices.PaymentProvider, publisher outboundservices.EventPublisher) *PaymentMethodCommandService {
	return &PaymentMethodCommandService{repository: repository, provider: provider, publisher: publisher}
}

func (s *PaymentMethodCommandService) Register(ctx context.Context, command commands.RegisterPaymentMethodCommand) (*entities.PaymentMethod, error) {
	userID, err := paymentdomain.ParseID(command.UserID)
	if err != nil {
		return nil, paymentdomain.ErrInvalidUUID
	}
	stripePaymentMethodID := strings.TrimSpace(command.StripePaymentMethodID)
	if stripePaymentMethodID == "" {
		return nil, paymentdomain.ErrExternalProvider
	}

	details, err := s.provider.GetPaymentMethodDetails(ctx, stripePaymentMethodID)
	if err != nil {
		log.Printf("payment method lookup failed stripe_payment_method_id=%s: %v", stripePaymentMethodID, err)
		return nil, err
	}

	methodType := strings.TrimSpace(command.Type)
	if methodType == "" {
		methodType = details.Type
	}
	method := entities.NewPaymentMethod(userID, methodType, details.Brand, details.Last4, details.ExpMonth, details.ExpYear, stripePaymentMethodID, command.IsDefault)
	if method.IsDefault {
		if err := s.repository.ClearDefaultForUser(ctx, userID); err != nil {
			log.Printf("payment method default reset failed user_id=%s: %v", userID, err)
			return nil, err
		}
	}
	if err := s.repository.Save(ctx, &method); err != nil {
		log.Printf("payment method persistence failed payment_method_id=%s user_id=%s: %v", method.PaymentMethodID, method.UserID, err)
		return nil, err
	}
	if err := s.publisher.PublishPaymentMethodAdded(ctx, method); err != nil {
		log.Printf("payment method event publish failed payment_method_id=%s: %v", method.PaymentMethodID, err)
	}
	return &method, nil
}

func (s *PaymentMethodCommandService) SetDefault(ctx context.Context, command commands.SetDefaultPaymentMethodCommand) (*entities.PaymentMethod, error) {
	paymentMethodID, err := paymentdomain.ParseID(command.PaymentMethodID)
	if err != nil {
		return nil, paymentdomain.ErrInvalidUUID
	}
	method, err := s.repository.FindByID(ctx, paymentMethodID)
	if err != nil {
		return nil, err
	}
	if err := s.repository.ClearDefaultForUser(ctx, method.UserID); err != nil {
		return nil, err
	}
	method.MarkDefault()
	if err := s.repository.Save(ctx, method); err != nil {
		return nil, err
	}
	return method, nil
}

func (s *PaymentMethodCommandService) Delete(ctx context.Context, command commands.DeletePaymentMethodCommand) error {
	paymentMethodID, err := paymentdomain.ParseID(command.PaymentMethodID)
	if err != nil {
		return paymentdomain.ErrInvalidUUID
	}
	return s.repository.Delete(ctx, paymentMethodID)
}
