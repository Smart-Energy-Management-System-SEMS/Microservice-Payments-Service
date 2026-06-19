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
// PaymentMethodCommandService gestiona las operaciones de comandos relacionadas con los métodos de pago
// Implementa la lógica de aplicación para registrar, establecer como predeterminado y eliminar métodos de pago
// Interactúa con el repositorio, el proveedor de pago externo (Stripe) y el publicador de eventos
type PaymentMethodCommandService struct {
	repository repositories.PaymentMethodRepository
	provider   outboundservices.PaymentProvider
	publisher  outboundservices.EventPublisher
}
// NewPaymentMethodCommandService crea una nueva instancia del servicio de comandos de métodos de pago
// Parámetros:
//   - repository: repositorio para persistencia de datos
//   - provider: proveedor de servicios de pago externo
//   - publisher: publicador de eventos de dominio
// Retorna:
//   - *PaymentMethodCommandService: instancia inicializada del servicio
func NewPaymentMethodCommandService(repository repositories.PaymentMethodRepository, provider outboundservices.PaymentProvider, publisher outboundservices.EventPublisher) *PaymentMethodCommandService {
	return &PaymentMethodCommandService{repository: repository, provider: provider, publisher: publisher}
}
// Register registra un nuevo método de pago para un usuario
// Procesa un comando de registro validando los datos y consultando el proveedor de pago
// Parámetros:
//   - ctx: contexto para cancelación y timeouts
//   - command: comando que contiene userID, stripePaymentMethodID, tipo de método e indicador de predeterminado
// Retorna:
//   - *entities.PaymentMethod: el método de pago creado
//   - error: si ocurre alguna validación o error de persistencia
// Lógica:
//   1. Valida que el userID sea un UUID válido
//   2. Valida que el stripePaymentMethodID no esté vacío
//   3. Obtiene los detalles del método de pago desde el proveedor externo (Stripe)
//   4. Si el tipo no se especifica, lo obtiene del proveedor
//   5. Crea la entidad PaymentMethod con los datos validados
//   6. Si se marca como predeterminado, limpia el flag en otros métodos del usuario
//   7. Persiste el nuevo método de pago
//   8. Publica un evento de "PaymentMethodAdded"
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
// SetDefault establece un método de pago como el predeterminado para el usuario
// Procesa un comando para cambiar el método de pago predeterminado
// Parámetros:
//   - ctx: contexto para cancelación y timeouts
//   - command: comando que contiene el ID del método de pago a establecer como predeterminado
// Retorna:
//   - *entities.PaymentMethod: el método de pago actualizado como predeterminado
//   - error: si ocurre alguna validación o error de persistencia
// Lógica:
//   1. Valida que el paymentMethodID sea un UUID válido
//   2. Busca el método de pago por su ID
//   3. Limpia el flag de predeterminado en todos los métodos del usuario
//   4. Marca el método actual como predeterminado
//   5. Persiste el cambio
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
// Delete elimina un método de pago del usuario
// Procesa un comando para eliminar un método de pago específico
// Parámetros:
//   - ctx: contexto para cancelación y timeouts
//   - command: comando que contiene el ID del método de pago a eliminar
// Retorna:
//   - error: si ocurre alguna validación o error de persistencia
// Lógica:
//   1. Valida que el paymentMethodID sea un UUID válido
//   2. Elimina el método de pago del repositorio
func (s *PaymentMethodCommandService) Delete(ctx context.Context, command commands.DeletePaymentMethodCommand) error {
	paymentMethodID, err := paymentdomain.ParseID(command.PaymentMethodID)
	if err != nil {
		return paymentdomain.ErrInvalidUUID
	}
	return s.repository.Delete(ctx, paymentMethodID)
}
