package commandservices

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"Microservice-Payments-Service/payments/application/outboundservices"
	paymentdomain "Microservice-Payments-Service/payments/domain"
	"Microservice-Payments-Service/payments/domain/model/commands"
	"Microservice-Payments-Service/payments/domain/model/entities"
	"Microservice-Payments-Service/payments/domain/model/valueobjects"
)

func TestPaymentMethodRegisterWithStripeTestMethodPersistsAndPublishes(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryPaymentMethodRepository()
	provider := &fakePaymentProvider{
		details: map[string]paymentMethodFixture{
			"pm_card_visa": {
				Type:     "card",
				Brand:    "visa",
				Last4:    "4242",
				ExpMonth: 12,
				ExpYear:  2030,
			},
		},
	}
	publisher := &fakeEventPublisher{}
	service := NewPaymentMethodCommandService(repo, provider, publisher)

	userID := uuid.New()
	method, err := service.Register(ctx, commands.RegisterPaymentMethodCommand{
		UserID:                userID.String(),
		StripePaymentMethodID: "pm_card_visa",
		IsDefault:             true,
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if method.UserID != userID {
		t.Fatalf("Register() user id = %s, want %s", method.UserID, userID)
	}
	if method.Brand != "visa" || method.Last4 != "4242" || method.Type != "card" {
		t.Fatalf("Register() returned unexpected method details: %+v", method)
	}
	if got := len(repo.methods); got != 1 {
		t.Fatalf("saved payment methods = %d, want 1", got)
	}
	if got := publisher.eventTypes(); len(got) != 1 || got[0] != "payment.method.added" {
		t.Fatalf("published events = %v, want [payment.method.added]", got)
	}
}

func TestPaymentMethodRegisterDoesNotPublishWhenPersistenceFails(t *testing.T) {
	ctx := context.Background()
	repo := newMemoryPaymentMethodRepository()
	repo.saveErr = errors.New("insert failed")
	provider := &fakePaymentProvider{
		details: map[string]paymentMethodFixture{
			"pm_card_visa": {Type: "card", Brand: "visa", Last4: "4242", ExpMonth: 12, ExpYear: 2030},
		},
	}
	publisher := &fakeEventPublisher{}
	service := NewPaymentMethodCommandService(repo, provider, publisher)

	_, err := service.Register(ctx, commands.RegisterPaymentMethodCommand{
		UserID:                uuid.New().String(),
		StripePaymentMethodID: "pm_card_visa",
	})
	if err == nil {
		t.Fatal("Register() error = nil, want persistence failure")
	}
	if got := publisher.eventTypes(); len(got) != 0 {
		t.Fatalf("published events = %v, want none", got)
	}
}

func TestProcessPaymentWithValidInternalMethodPersistsAndPublishes(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	subscriptionID := uuid.New()
	method := entities.NewPaymentMethod(userID, "card", "visa", "4242", 12, 2030, "pm_card_visa", true)

	paymentMethods := newMemoryPaymentMethodRepository()
	paymentMethods.methods[method.PaymentMethodID] = method
	payments := newMemoryPaymentRepository()
	invoices := newMemoryInvoiceRepository()
	provider := &fakePaymentProvider{
		paymentIntentResult: &paymentIntentFixture{ID: "pi_test_123", Status: "succeeded"},
	}
	publisher := &fakeEventPublisher{}
	service := NewPaymentCommandService(payments, paymentMethods, invoices, provider, publisher)

	payment, invoice, err := service.Process(ctx, commands.ProcessPaymentCommand{
		SubscriptionID:  subscriptionID.String(),
		UserID:          userID.String(),
		PaymentMethodID: method.PaymentMethodID.String(),
		Amount:          49.90,
		Currency:        "pen",
	})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if payment.Status != valueobjects.PaymentStatusProcessed {
		t.Fatalf("payment status = %s, want %s", payment.Status, valueobjects.PaymentStatusProcessed)
	}
	if invoice == nil {
		t.Fatal("invoice = nil, want generated invoice")
	}
	if got := len(payments.saved); got != 1 {
		t.Fatalf("saved payments = %d, want 1", got)
	}
	if got := len(invoices.saved); got != 1 {
		t.Fatalf("saved invoices = %d, want 1", got)
	}
	if got := publisher.eventTypes(); !containsInOrder(got, []string{"payment.processed", "invoice.generated"}) {
		t.Fatalf("published events = %v, want payment.processed then invoice.generated", got)
	}
}

func TestProcessPaymentFailurePersistsAndPublishesFailure(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	method := entities.NewPaymentMethod(userID, "card", "visa", "4242", 12, 2030, "pm_card_visa", true)

	paymentMethods := newMemoryPaymentMethodRepository()
	paymentMethods.methods[method.PaymentMethodID] = method
	payments := newMemoryPaymentRepository()
	invoices := newMemoryInvoiceRepository()
	provider := &fakePaymentProvider{paymentIntentErr: paymentdomain.ErrExternalProvider}
	publisher := &fakeEventPublisher{}
	service := NewPaymentCommandService(payments, paymentMethods, invoices, provider, publisher)

	payment, invoice, err := service.Process(ctx, commands.ProcessPaymentCommand{
		SubscriptionID:  uuid.New().String(),
		UserID:          userID.String(),
		PaymentMethodID: method.PaymentMethodID.String(),
		Amount:          19.90,
		Currency:        "pen",
	})
	if !errors.Is(err, paymentdomain.ErrExternalProvider) {
		t.Fatalf("Process() error = %v, want ErrExternalProvider", err)
	}
	if invoice != nil {
		t.Fatalf("invoice = %+v, want nil on failed payment", invoice)
	}
	if payment.Status != valueobjects.PaymentStatusFailed {
		t.Fatalf("payment status = %s, want %s", payment.Status, valueobjects.PaymentStatusFailed)
	}
	if got := publisher.eventTypes(); len(got) != 1 || got[0] != "payment.failed" {
		t.Fatalf("published events = %v, want [payment.failed]", got)
	}
}

type memoryPaymentMethodRepository struct {
	methods map[uuid.UUID]entities.PaymentMethod
	saveErr error
}

func newMemoryPaymentMethodRepository() *memoryPaymentMethodRepository {
	return &memoryPaymentMethodRepository{methods: make(map[uuid.UUID]entities.PaymentMethod)}
}

func (r *memoryPaymentMethodRepository) Save(_ context.Context, method *entities.PaymentMethod) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.methods[method.PaymentMethodID] = *method
	return nil
}

func (r *memoryPaymentMethodRepository) FindByID(_ context.Context, id uuid.UUID) (*entities.PaymentMethod, error) {
	method, ok := r.methods[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &method, nil
}

func (r *memoryPaymentMethodRepository) FindByUserID(_ context.Context, userID uuid.UUID) ([]entities.PaymentMethod, error) {
	var methods []entities.PaymentMethod
	for _, method := range r.methods {
		if method.UserID == userID {
			methods = append(methods, method)
		}
	}
	return methods, nil
}

func (r *memoryPaymentMethodRepository) ClearDefaultForUser(_ context.Context, userID uuid.UUID) error {
	for id, method := range r.methods {
		if method.UserID == userID {
			method.IsDefault = false
			r.methods[id] = method
		}
	}
	return nil
}

func (r *memoryPaymentMethodRepository) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.methods, id)
	return nil
}

type memoryPaymentRepository struct {
	saved map[uuid.UUID]entities.Payment
}

func newMemoryPaymentRepository() *memoryPaymentRepository {
	return &memoryPaymentRepository{saved: make(map[uuid.UUID]entities.Payment)}
}

func (r *memoryPaymentRepository) Save(_ context.Context, payment *entities.Payment) error {
	r.saved[payment.PaymentID] = *payment
	return nil
}

func (r *memoryPaymentRepository) Update(_ context.Context, payment *entities.Payment) error {
	r.saved[payment.PaymentID] = *payment
	return nil
}

func (r *memoryPaymentRepository) FindByID(_ context.Context, id uuid.UUID) (*entities.Payment, error) {
	payment, ok := r.saved[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &payment, nil
}

func (r *memoryPaymentRepository) FindByUserID(_ context.Context, userID uuid.UUID) ([]entities.Payment, error) {
	var payments []entities.Payment
	for _, payment := range r.saved {
		if payment.UserID == userID {
			payments = append(payments, payment)
		}
	}
	return payments, nil
}

func (r *memoryPaymentRepository) FindBySubscriptionID(_ context.Context, subscriptionID uuid.UUID) ([]entities.Payment, error) {
	var payments []entities.Payment
	for _, payment := range r.saved {
		if payment.SubscriptionID == subscriptionID {
			payments = append(payments, payment)
		}
	}
	return payments, nil
}

func (r *memoryPaymentRepository) FindByStripePaymentIntentID(_ context.Context, stripePaymentIntentID string) (*entities.Payment, error) {
	for _, payment := range r.saved {
		if payment.StripePaymentIntentID == stripePaymentIntentID {
			return &payment, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

type memoryInvoiceRepository struct {
	saved map[uuid.UUID]entities.Invoice
}

func newMemoryInvoiceRepository() *memoryInvoiceRepository {
	return &memoryInvoiceRepository{saved: make(map[uuid.UUID]entities.Invoice)}
}

func (r *memoryInvoiceRepository) Save(_ context.Context, invoice *entities.Invoice) error {
	r.saved[invoice.InvoiceID] = *invoice
	return nil
}

func (r *memoryInvoiceRepository) FindByID(_ context.Context, id uuid.UUID) (*entities.Invoice, error) {
	invoice, ok := r.saved[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &invoice, nil
}

func (r *memoryInvoiceRepository) FindByPaymentID(_ context.Context, paymentID uuid.UUID) (*entities.Invoice, error) {
	for _, invoice := range r.saved {
		if invoice.PaymentID == paymentID {
			return &invoice, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

type paymentMethodFixture struct {
	Type     string
	Brand    string
	Last4    string
	ExpMonth int
	ExpYear  int
}

type paymentIntentFixture struct {
	ID     string
	Status string
}

type fakePaymentProvider struct {
	details             map[string]paymentMethodFixture
	paymentIntentResult *paymentIntentFixture
	paymentIntentErr    error
}

func (p *fakePaymentProvider) GetPaymentMethodDetails(_ context.Context, stripePaymentMethodID string) (*outboundservices.PaymentMethodDetails, error) {
	details, ok := p.details[stripePaymentMethodID]
	if !ok {
		return nil, paymentdomain.ErrExternalProvider
	}
	return &outboundservices.PaymentMethodDetails{
		Type:     details.Type,
		Brand:    details.Brand,
		Last4:    details.Last4,
		ExpMonth: details.ExpMonth,
		ExpYear:  details.ExpYear,
	}, nil
}

func (p *fakePaymentProvider) AttachPaymentMethod(_ context.Context, _ string, _ string) error {
	return nil
}

func (p *fakePaymentProvider) CreatePaymentIntent(_ context.Context, _ outboundservices.CreatePaymentIntentRequest) (*outboundservices.PaymentIntentResult, error) {
	if p.paymentIntentErr != nil {
		return nil, p.paymentIntentErr
	}
	if p.paymentIntentResult == nil {
		return nil, paymentdomain.ErrExternalProvider
	}
	return &outboundservices.PaymentIntentResult{ID: p.paymentIntentResult.ID, Status: p.paymentIntentResult.Status}, nil
}

func (p *fakePaymentProvider) ConfirmPaymentIntent(_ context.Context, _ string) (*outboundservices.PaymentIntentResult, error) {
	return nil, nil
}

func (p *fakePaymentProvider) ParseWebhookEvent(_ context.Context, _ []byte, _ string) (*outboundservices.ProviderWebhookEvent, error) {
	return nil, nil
}

type fakeEventPublisher struct {
	events []string
}

func (p *fakeEventPublisher) PublishPaymentProcessed(_ context.Context, _ entities.Payment, _ entities.Invoice) error {
	p.events = append(p.events, "payment.processed")
	return nil
}

func (p *fakeEventPublisher) PublishPaymentFailed(_ context.Context, _ entities.Payment) error {
	p.events = append(p.events, "payment.failed")
	return nil
}

func (p *fakeEventPublisher) PublishInvoiceGenerated(_ context.Context, _ entities.Invoice) error {
	p.events = append(p.events, "invoice.generated")
	return nil
}

func (p *fakeEventPublisher) PublishPaymentMethodAdded(_ context.Context, _ entities.PaymentMethod) error {
	p.events = append(p.events, "payment.method.added")
	return nil
}

func (p *fakeEventPublisher) Close() error {
	return nil
}

func (p *fakeEventPublisher) eventTypes() []string {
	return append([]string(nil), p.events...)
}

func containsInOrder(got []string, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
