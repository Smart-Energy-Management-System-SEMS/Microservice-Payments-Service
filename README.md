# Microservice-Payments-Service

Payments Service para Smart Energy Management System (SEMS). Procesa pagos, metodos de pago, invoices, webhooks de Stripe y eventos Kafka.

## Stack

- Go + Gin
- DDD + Clean Architecture
- PostgreSQL (Neon) con GORM
- Stripe SDK oficial para Go
- Apache Kafka con segmentio/kafka-go
- Docker + Docker Compose

## Configuracion centralizada

Este microservicio ahora soporta **Config Service** via `CONFIG_SERVICE_URL`.

Se consulta:

- `GET /api/v1/config/{service-name}` (ejemplo: `payments-service`)
- `GET /api/v1/config/kafka`
- (opcional) `GET /api/v1/config/services`

Orden de resolucion:

1. Carga `.env` (secretos y runtime local).
2. Si existe `CONFIG_SERVICE_URL`, obtiene configuracion centralizada y sobrescribe:
   - `SERVER_PORT`
   - `API_BASE_PATH`
   - `STRIPE_CURRENCY`
   - Kafka brokers, client id y topics
3. Si Config Service no responde, usa fallback local (`.env`).

## Variables de entorno del microservicio

Copia `.env.example` a `.env`.

### Requeridas (sensibles o de despliegue)

- `CONFIG_SERVICE_URL`
- `SERVER_PORT`
- `DATABASE_URL`
- `STRIPE_SECRET_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `APP_ENV`
- `GIN_MODE`
- `DB_AUTO_MIGRATE`

### Fallback local (compatibilidad)

Solo para desarrollo cuando Config Service no esta disponible:

- `API_BASE_PATH`
- `STRIPE_CURRENCY`
- `KAFKA_BROKERS`
- `KAFKA_CLIENT_ID`
- `KAFKA_PAYMENT_PROCESSED_TOPIC`
- `KAFKA_PAYMENT_FAILED_TOPIC`
- `KAFKA_INVOICE_GENERATED_TOPIC`
- `KAFKA_PAYMENT_METHOD_ADDED_TOPIC`
- `KAFKA_SUBSCRIPTION_CREATED_TOPIC`
- `KAFKA_SUBSCRIPTION_RENEWAL_REQUESTED_TOPIC`
- `KAFKA_SUBSCRIPTION_CANCELLED_TOPIC`

## Ejecucion local

```bash
go mod tidy
go run main.go
```

Compilar:

```bash
go build main.go
```

## Docker

```bash
docker build -t sems-payments-service .
docker run --rm --env-file .env -p 8085:8085 sems-payments-service
```

## Endpoints

Health:

- `GET /health`

Base path por defecto (fallback): `/api/v1`.

Payment Methods:

- `POST /api/v1/payment-methods`
- `GET /api/v1/payment-methods/user/{userId}`
- `PUT /api/v1/payment-methods/{paymentMethodId}/default`
- `DELETE /api/v1/payment-methods/{paymentMethodId}`

Payments:

- `POST /api/v1/payments/process`
- `GET /api/v1/payments/{paymentId}`
- `GET /api/v1/payments/user/{userId}`
- `GET /api/v1/payments/subscription/{subscriptionId}`

Invoices:

- `GET /api/v1/invoices/{invoiceId}`
- `GET /api/v1/invoices/payment/{paymentId}`

Webhook Stripe:

- `POST /api/v1/webhooks/stripe`

## Configuracion Kafka

El servicio publica:

- `payment.processed`
- `payment.failed`
- `invoice.generated`
- `payment.method.added`

El servicio consume:

- `subscription.created`
- `subscription.renewal.requested`
- `subscription.cancelled`

Consumer group: `{KAFKA_CLIENT_ID}-group`.

## API Gateway

- Publicos recomendados:
  - `GET /health`
  - `POST /api/v1/webhooks/stripe`
- Protegidos con JWT:
  - resto de endpoints `/api/v1/payment-methods`, `/api/v1/payments`, `/api/v1/invoices`

## Azure Container Apps

Recomendado para despliegue:

1. Variables en Container App:
   - `CONFIG_SERVICE_URL`
   - `SERVER_PORT`
   - `APP_ENV=production`
   - `GIN_MODE=release`
   - `DB_AUTO_MIGRATE=false`
2. Secretos en ACA Secrets + referencias en env:
   - `DATABASE_URL`
   - `STRIPE_SECRET_KEY`
   - `STRIPE_WEBHOOK_SECRET`
3. Networking:
   - Permitir salida a Config Service, Kafka y Stripe.
4. Health probe:
   - `GET /health`.

## Notas de arquitectura

- Controllers solo hacen binding HTTP y delegan a application.
- Application orquesta casos de uso y eventos.
- Domain mantiene entidades/value objects/contratos.
- Infrastructure integra Stripe, Kafka y GORM.
