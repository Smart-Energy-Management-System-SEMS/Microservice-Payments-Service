// Package stripeadapter is the concrete adapter that talks to Stripe. It is the
// only place in the codebase that imports the Stripe SDK; the rest of the app
// depends on the PaymentProvider port instead. It also adds a circuit breaker
// for resilience when Stripe is unavailable.
package stripeadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	stripesdk "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/paymentmethod"
	"github.com/stripe/stripe-go/v82/webhook"

	"Microservice-Payments-Service/payments/application/outboundservices"
	paymentdomain "Microservice-Payments-Service/payments/domain"
)

// breakerThreshold/breakerCooldown configure a "circuit breaker". After this
// many consecutive failures, the adapter stops calling Stripe for the cooldown
// period. This protects us when Stripe is down: instead of hammering a failing
// service (and making every request hang), we fail fast for a while.
const (
	breakerThreshold = 5
	breakerCooldown  = 30 * time.Second
)

// Adapter is the infrastructure adapter that implements the PaymentProvider
// port using the real Stripe SDK. The last three fields make up the circuit
// breaker's state. Because many HTTP requests run concurrently, that shared
// state is guarded by a mutex (mu) to avoid data races.
type Adapter struct {
	secretKey     string
	webhookSecret string
	mu            sync.Mutex
	failures      int       // count of consecutive failures
	openedAt      time.Time // when the breaker "opened" (started blocking)
}

// NewAdapter configures the global Stripe API key and returns the adapter.
func NewAdapter(secretKey string, webhookSecret string) *Adapter {
	stripesdk.Key = secretKey
	return &Adapter{secretKey: secretKey, webhookSecret: webhookSecret}
}

func (a *Adapter) GetPaymentMethodDetails(ctx context.Context, stripePaymentMethodID string) (*outboundservices.PaymentMethodDetails, error) {
	_ = ctx
	if err := a.requireSecretKey(); err != nil {
		return nil, err
	}
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
	if err := a.requireSecretKey(); err != nil {
		return err
	}
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

// CreatePaymentIntent starts a charge in Stripe. Every public method here
// follows the same protective pattern: check the breaker (allowRequest), call
// Stripe, then record a success or failure to update the breaker. Errors are
// wrapped with ErrExternalProvider using %w, so callers can still detect the
// category with errors.Is while keeping Stripe's original message.
func (a *Adapter) CreatePaymentIntent(ctx context.Context, request outboundservices.CreatePaymentIntentRequest) (*outboundservices.PaymentIntentResult, error) {
	_ = ctx
	if err := a.requireSecretKey(); err != nil {
		return nil, err
	}
	if err := a.allowRequest(); err != nil {
		return nil, err
	}
	// Confirm: true tells Stripe to attempt the charge immediately. Disabling
	// redirect-based methods keeps this a simple server-to-server charge.
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
	// Attach our own ids as metadata so that when Stripe later sends a webhook,
	// we can tie its event back to our payment/subscription/user records.
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
	if err := a.requireSecretKey(); err != nil {
		return nil, err
	}
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

// ParseWebhookEvent verifies and decodes an incoming Stripe webhook.
// webhook.ConstructEvent checks the signature against our webhook secret; if it
// does not match, the request is rejected as a possible forgery. For the
// payment-intent events we also extract the intent id and status so the caller
// can update the matching payment. (Note: webhook verification is intentionally
// NOT behind the circuit breaker — it is local crypto, not a network call.)
func (a *Adapter) ParseWebhookEvent(ctx context.Context, payload []byte, signature string) (*outboundservices.ProviderWebhookEvent, error) {
	_ = ctx
	if strings.TrimSpace(a.webhookSecret) == "" {
		return nil, fmt.Errorf("%w: stripe webhook secret is not configured", paymentdomain.ErrExternalProvider)
	}
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

// toMinorUnits converts a decimal amount (e.g. 12.34) into the integer "minor
// units" Stripe expects (1234 cents). math.Round avoids floating-point errors
// like 12.34*100 = 1233.9999 turning into 1233.
func toMinorUnits(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func (a *Adapter) requireSecretKey() error {
	if strings.TrimSpace(a.secretKey) == "" {
		return fmt.Errorf("%w: stripe secret key is not configured", paymentdomain.ErrExternalProvider)
	}
	return nil
}

// The three methods below are the circuit breaker. They all lock the mutex so
// the shared state is updated safely under concurrency.

// allowRequest decides whether a call to Stripe is permitted right now. If the
// breaker is "open" and still cooling down, it blocks the call; once the
// cooldown has passed it resets the failure count and allows traffic again
// (a "half-open"-style retry).
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

// recordFailure increments the failure counter and, once the threshold is hit,
// stamps the time the breaker opened.
func (a *Adapter) recordFailure() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.failures++
	if a.failures >= breakerThreshold {
		a.openedAt = time.Now().UTC()
	}
}

// recordSuccess resets the breaker after any successful call, so a single good
// response clears the failure streak.
func (a *Adapter) recordSuccess() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.failures = 0
	a.openedAt = time.Time{}
}
