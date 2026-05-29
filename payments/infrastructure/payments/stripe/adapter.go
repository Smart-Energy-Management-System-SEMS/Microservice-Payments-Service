package stripeadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	stripesdk "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/paymentmethod"
	"github.com/stripe/stripe-go/v82/webhook"

	"Microservice-Payments-Service/payments/application/outboundservices"
	paymentdomain "Microservice-Payments-Service/payments/domain"
)

const (
	breakerThreshold = 5
	breakerCooldown  = 30 * time.Second
)

type Adapter struct {
	webhookSecret string
	mu            sync.Mutex
	failures      int
	openedAt      time.Time
}

func NewAdapter(secretKey string, webhookSecret string) *Adapter {
	stripesdk.Key = secretKey
	return &Adapter{webhookSecret: webhookSecret}
}

func (a *Adapter) GetPaymentMethodDetails(ctx context.Context, stripePaymentMethodID string) (*outboundservices.PaymentMethodDetails, error) {
	_ = ctx
	if err := a.allowRequest(); err != nil {
		return nil, err
	}
	pm, err := paymentmethod.Get(stripePaymentMethodID, nil)
	if err != nil {
		a.recordFailure()
		return nil, fmt.Errorf("%w: %v", paymentdomain.ErrExternalProvider, err)
	}
	a.recordSuccess()

	details := &outboundservices.PaymentMethodDetails{Type: string(pm.Type)}
	if pm.Card != nil {
		details.Brand = string(pm.Card.Brand)
		details.Last4 = pm.Card.Last4
		details.ExpMonth = int(pm.Card.ExpMonth)
		details.ExpYear = int(pm.Card.ExpYear)
	}
	return details, nil
}

func (a *Adapter) AttachPaymentMethod(ctx context.Context, stripePaymentMethodID string, customerID string) error {
	_ = ctx
	if err := a.allowRequest(); err != nil {
		return err
	}
	_, err := paymentmethod.Attach(stripePaymentMethodID, &stripesdk.PaymentMethodAttachParams{Customer: stripesdk.String(customerID)})
	if err != nil {
		a.recordFailure()
		return fmt.Errorf("%w: %v", paymentdomain.ErrExternalProvider, err)
	}
	a.recordSuccess()
	return nil
}

func (a *Adapter) CreatePaymentIntent(ctx context.Context, request outboundservices.CreatePaymentIntentRequest) (*outboundservices.PaymentIntentResult, error) {
	_ = ctx
	if err := a.allowRequest(); err != nil {
		return nil, err
	}
	params := &stripesdk.PaymentIntentParams{
		Amount:        stripesdk.Int64(toMinorUnits(request.Amount)),
		Currency:      stripesdk.String(request.Currency),
		PaymentMethod: stripesdk.String(request.StripePaymentMethodID),
		Confirm:       stripesdk.Bool(true),
		AutomaticPaymentMethods: &stripesdk.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled:        stripesdk.Bool(true),
			AllowRedirects: stripesdk.String("never"),
		},
	}
	params.AddMetadata("payment_id", request.PaymentID)
	params.AddMetadata("subscription_id", request.SubscriptionID)
	params.AddMetadata("user_id", request.UserID)

	intent, err := paymentintent.New(params)
	if err != nil {
		a.recordFailure()
		return nil, fmt.Errorf("%w: %v", paymentdomain.ErrExternalProvider, err)
	}
	a.recordSuccess()
	return &outboundservices.PaymentIntentResult{ID: intent.ID, Status: string(intent.Status)}, nil
}

func (a *Adapter) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string) (*outboundservices.PaymentIntentResult, error) {
	_ = ctx
	if err := a.allowRequest(); err != nil {
		return nil, err
	}
	intent, err := paymentintent.Confirm(paymentIntentID, nil)
	if err != nil {
		a.recordFailure()
		return nil, fmt.Errorf("%w: %v", paymentdomain.ErrExternalProvider, err)
	}
	a.recordSuccess()
	return &outboundservices.PaymentIntentResult{ID: intent.ID, Status: string(intent.Status)}, nil
}

func (a *Adapter) ParseWebhookEvent(ctx context.Context, payload []byte, signature string) (*outboundservices.ProviderWebhookEvent, error) {
	_ = ctx
	event, err := webhook.ConstructEvent(payload, signature, a.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid stripe webhook signature", paymentdomain.ErrExternalProvider)
	}

	providerEvent := &outboundservices.ProviderWebhookEvent{
		ProviderEventID: event.ID,
		EventType:       string(event.Type),
		Payload:         json.RawMessage(payload),
	}

	switch string(event.Type) {
	case "payment_intent.succeeded", "payment_intent.payment_failed", "payment_intent.canceled":
		var intent stripesdk.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &intent); err != nil {
			return nil, fmt.Errorf("%w: %v", paymentdomain.ErrExternalProvider, err)
		}
		providerEvent.StripePaymentIntentID = intent.ID
		providerEvent.PaymentStatus = string(intent.Status)
	}
	return providerEvent, nil
}

func toMinorUnits(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func (a *Adapter) allowRequest() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.failures >= breakerThreshold && time.Since(a.openedAt) < breakerCooldown {
		return fmt.Errorf("%w: stripe circuit breaker open", paymentdomain.ErrExternalProvider)
	}
	if a.failures >= breakerThreshold && time.Since(a.openedAt) >= breakerCooldown {
		a.failures = 0
	}
	return nil
}

func (a *Adapter) recordFailure() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.failures++
	if a.failures >= breakerThreshold {
		a.openedAt = time.Now().UTC()
	}
}

func (a *Adapter) recordSuccess() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.failures = 0
	a.openedAt = time.Time{}
}
