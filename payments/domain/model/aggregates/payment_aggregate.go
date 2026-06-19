// Package aggregates holds the aggregate roots of the domain. In Domain-Driven
// Design (DDD), an aggregate groups together objects that must stay consistent
// as a unit and are treated as a single whole when loaded or saved.
package aggregates

import "Microservice-Payments-Service/payments/domain/model/entities"

// PaymentAggregate bundles a Payment together with the Invoice that may be
// generated for it. Grouping them in one aggregate signals that they belong
// together: when a payment is processed, its invoice is part of the same
// consistency boundary. The fields are pointers because the Invoice is optional
// (it only exists once the payment actually succeeds).
type PaymentAggregate struct {
	Payment *entities.Payment
	Invoice *entities.Invoice
}
