package aggregates

import "Microservice-Payments-Service/payments/domain/model/entities"

type PaymentAggregate struct {
	Payment *entities.Payment
	Invoice *entities.Invoice
}
