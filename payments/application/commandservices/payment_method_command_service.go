package commandservices

import (
	"context"
	"strings"

	"Microservice-Payments-Service/payments/application/outboundservices"
	"Microservice-Payments-Service/payments/domain/model/commands"
	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/repositories"
	shareddomain "Microservice-Payments-Service/payments/shared/domain"
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
	userID, err := shareddomain.ParseID(command.UserID)
	if err != nil {
		return nil, shareddomain.ErrInvalidUUID
	}
	stripePaymentMethodID := strings.TrimSpace(command.StripePaymentMethodID)
	if stripePaymentMethodID == "" {
		return nil, shareddomain.ErrExternalProvider
	}

	details, err := s.provider.GetPaymentMethodDetails(ctx, stripePaymentMethodID)
	if err != nil {
		return nil, err
	}

	methodType := strings.TrimSpace(command.Type)
	if methodType == "" {
		methodType = details.Type
	}
	method := entities.NewPaymentMethod(userID, methodType, details.Brand, details.Last4, details.ExpMonth, details.ExpYear, stripePaymentMethodID, command.IsDefault)
	if method.IsDefault {
		if err := s.repository.ClearDefaultForUser(ctx, userID); err != nil {
			return nil, err
		}
	}
	if err := s.repository.Save(ctx, &method); err != nil {
		return nil, err
	}
	_ = s.publisher.PublishPaymentMethodAdded(ctx, method)
	return &method, nil
}

func (s *PaymentMethodCommandService) SetDefault(ctx context.Context, command commands.SetDefaultPaymentMethodCommand) (*entities.PaymentMethod, error) {
	paymentMethodID, err := shareddomain.ParseID(command.PaymentMethodID)
	if err != nil {
		return nil, shareddomain.ErrInvalidUUID
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
	paymentMethodID, err := shareddomain.ParseID(command.PaymentMethodID)
	if err != nil {
		return shareddomain.ErrInvalidUUID
	}
	return s.repository.Delete(ctx, paymentMethodID)
}
