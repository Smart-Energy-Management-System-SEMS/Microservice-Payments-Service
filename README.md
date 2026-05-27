# Microservice-Payments-Service

Payments Service para Smart Energy Management System (SEMS). El servicio procesa pagos, administra metodos de pago, registra invoices, recibe webhooks de Stripe y publica/consume eventos Kafka.

## Stack

- Go + Gin
- DDD + Clean Architecture
- PostgreSQL Neon con GORM
- Stripe SDK oficial para Go
- Apache Kafka con segmentio/kafka-go
- Variables de entorno con `.env`

## Estructura

```text
payments/
??? application/
?   ??? commandservices/
?   ??? eventhandlers/
?   ??? outboundservices/
?   ??? queryservices/
??? domain/
?   ??? model/
?   ??? repositories/
?   ??? services/
??? infrastructure/
?   ??? configuration/
?   ??? messaging/kafka/
?   ??? payments/stripe/
?   ??? persistence/gorm/
??? interfaces/
?   ??? acl/
?   ??? rest/
??? shared/
```

## Variables de entorno

Copia `.env.example` a `.env` y configura los valores reales:

```bash
SERVER_PORT=8085
API_BASE_PATH=/api/v1
DB_AUTO_MIGRATE=true
DATABASE_URL=postgresql://USER:PASSWORD@HOST/DB?sslmode=require
STRIPE_SECRET_KEY=sk_test_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx
STRIPE_CURRENCY=pen
KAFKA_BROKERS=localhost:9092
KAFKA_CLIENT_ID=payments-service
```

## Instalacion y ejecucion

```bash
go mod tidy
go run main.go
```

Compilar:

```bash
go build main.go
```

## Endpoints

Base path por defecto: `/api/v1`.

### Payment Methods

Registrar metodo de pago:

```http
POST /api/v1/payment-methods
Content-Type: application/json

{
  "user_id": "9d78e8e6-7f6d-4aa6-a4b0-3d6c44cb84f1",
  "type": "card",
  "stripe_payment_method_id": "pm_123",
  "is_default": true
}
```

Listar por usuario:

```http
GET /api/v1/payment-methods/user/{userId}
```

Marcar como default:

```http
PUT /api/v1/payment-methods/{paymentMethodId}/default
```

Eliminar:

```http
DELETE /api/v1/payment-methods/{paymentMethodId}
```

### Payments

Procesar pago:

```http
POST /api/v1/payments/process
Content-Type: application/json

{
  "subscription_id": "86a73829-a1dd-4228-9791-ea2b95ab308a",
  "user_id": "9d78e8e6-7f6d-4aa6-a4b0-3d6c44cb84f1",
  "payment_method_id": "8cb43113-13cc-4478-8582-4113516cf432",
  "amount": 49.90,
  "currency": "pen",
  "payment_method": "card"
}
```

Consultar:

```http
GET /api/v1/payments/{paymentId}
GET /api/v1/payments/user/{userId}
GET /api/v1/payments/subscription/{subscriptionId}
```

### Invoices

```http
GET /api/v1/invoices/{invoiceId}
GET /api/v1/invoices/payment/{paymentId}
```

### Stripe Webhook

Configura en Stripe el endpoint:

```http
POST /api/v1/webhooks/stripe
```

El servicio valida `Stripe-Signature`, guarda el evento en `payment_webhook_events` y evita procesarlo dos veces usando `provider + provider_event_id`.

## Eventos Kafka

Publica:

- `payment.processed`
- `payment.failed`
- `invoice.generated`
- `payment.method.added`

Consume:

- `subscription.created`
- `subscription.renewal.requested`
- `subscription.cancelled`

`subscription.renewal.requested` dispara el procesamiento de pago si el evento incluye `subscription_id`, `user_id`, `payment_method_id`, `amount` y `currency`.

## Notas de arquitectura

- Controllers solo hacen binding HTTP y llaman servicios de aplicacion.
- Application orquesta casos de uso y eventos.
- Domain contiene entidades, value objects, comandos, queries y contratos.
- Infrastructure contiene Stripe, Kafka y GORM.
- No hay foreign keys hacia otros microservicios; `subscription_id` y `user_id` son referencias externas.
