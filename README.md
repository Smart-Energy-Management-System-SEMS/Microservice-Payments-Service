# Microservice-Payments-Service

Microservicio de pagos de SEMS. Expone endpoints REST de payment methods, payments, invoices y webhook Stripe.

## Health checks

- `GET /health`
- `GET /api/v1/health`

Ambos endpoints devuelven `200 OK`.

## Variables de entorno requeridas

Base (Azure/local):

- `PORT` (ejemplo: `8080`)
- `CONFIG_SERVICE_URL`
- `KAFKA_BROKERS`
- `KAFKA_SECURITY_PROTOCOL`
- `KAFKA_SASL_MECHANISM`
- `KAFKA_USERNAME`
- `KAFKA_PASSWORD`
- `DATABASE_URL`
- `GIN_MODE` (recomendado en Azure: `release`)

Adicionales del servicio:

- `APP_ENV`
- `DB_AUTO_MIGRATE`
- `CORS_ALLOWED_ORIGINS`
- `CORS_ALLOW_CREDENTIALS`
- `API_BASE_PATH`
- `STRIPE_SECRET_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `STRIPE_CURRENCY`
- `KAFKA_CLIENT_ID`
- `KAFKA_CONSUMER_GROUP`
- `KAFKA_TOPIC_PAYMENTS_EVENTS`
- `KAFKA_TOPIC_BILLING_EVENTS`
- `KAFKA_TOPIC_SUBSCRIPTIONS_EVENTS`

Compatibilidad:

- El servicio prioriza `PORT`; si no existe, usa `SERVER_PORT`.

## Ejemplo Event Hubs / Azure (.env)

```env
PORT=8080
CONFIG_SERVICE_URL=https://<config-service-url>
KAFKA_BROKERS=<eventhubs-namespace>.servicebus.windows.net:9093
KAFKA_SECURITY_PROTOCOL=SASL_SSL
KAFKA_SASL_MECHANISM=PLAIN
KAFKA_USERNAME=$ConnectionString
KAFKA_PASSWORD=Endpoint=sb://<eventhubs-namespace>.servicebus.windows.net/;SharedAccessKeyName=<policy>;SharedAccessKey=<key>
KAFKA_CLIENT_ID=payments-service
KAFKA_CONSUMER_GROUP=payments-service-group
KAFKA_TOPIC_PAYMENTS_EVENTS=payments.events
KAFKA_TOPIC_BILLING_EVENTS=billing.events
KAFKA_TOPIC_SUBSCRIPTIONS_EVENTS=subscriptions.events
DATABASE_URL=postgresql://USER:PASSWORD@<postgres-host>:5432/payments?sslmode=require
GIN_MODE=release
```

## Docker

Build de imagen:

```bash
docker build -t sems-payments-service:local .
```

Run con archivo `.env`:

```bash
docker run --rm -p 8080:8080 --env-file .env --name sems-payments-service sems-payments-service:local
```

Prueba rápida:

```bash
curl -i http://localhost:8080/api/v1/health
```

## Azure Container Apps

El contenedor no debe usar `localhost` para servicios externos. Define variables en ACA con hosts reales para Config Service, Postgres y Azure Event Hubs.

Variables mínimas recomendadas en ACA:

```text
PORT=8080
GIN_MODE=release
CONFIG_SERVICE_URL=https://<config-service-url>
KAFKA_BROKERS=<eventhubs-namespace>.servicebus.windows.net:9093
KAFKA_SECURITY_PROTOCOL=SASL_SSL
KAFKA_SASL_MECHANISM=PLAIN
KAFKA_USERNAME=$ConnectionString
KAFKA_PASSWORD=Endpoint=sb://<eventhubs-namespace>.servicebus.windows.net/;SharedAccessKeyName=<policy>;SharedAccessKey=<key>
KAFKA_CLIENT_ID=payments-service
KAFKA_CONSUMER_GROUP=payments-service-group
KAFKA_TOPIC_PAYMENTS_EVENTS=payments.events
KAFKA_TOPIC_BILLING_EVENTS=billing.events
KAFKA_TOPIC_SUBSCRIPTIONS_EVENTS=subscriptions.events
DATABASE_URL=<conexion-postgres>
```

Checklist de despliegue ACA:

- Exponer puerto objetivo `8080`.
- Configurar secretos para `DATABASE_URL`, `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `KAFKA_PASSWORD`.
- Configurar egress a Kafka, Config Service y Stripe.
- Configurar health probe sobre `GET /api/v1/health` (o `GET /health`).

## Endpoints funcionales

Con prefijo `/api/v1`:

- `POST /payment-methods`
- `GET /payment-methods/user/:userId`
- `PUT /payment-methods/:paymentMethodId/default`
- `DELETE /payment-methods/:paymentMethodId`
- `POST /payments/process`
- `GET /payments/:paymentId`
- `GET /payments/user/:userId`
- `GET /payments/subscription/:subscriptionId`
- `GET /invoices/:invoiceId`
- `GET /invoices/payment/:paymentId`
- `POST /webhooks/stripe`

## Kafka topics agrupados

Payments publica:

- `payments.events`
- `billing.events`

Payments consume:

- `subscriptions.events`
- `billing.events`

Eventos publicados en `payments.events`:

- `payment.method.added`
- `payment.processed`
- `payment.failed`

Eventos publicados o consumidos en `billing.events`:

- `invoice.generated` se publica cuando el pago se confirma y la factura queda generada.
- `billing.payment.requested` puede consumirse como disparador de cobro si trae `subscription_id`, `user_id`, `payment_method_id`, `amount` y `currency` dentro de `data`.

Eventos consumidos en `subscriptions.events`:

- `subscription.created`
- `subscription.cancelled`
- `subscription.renewal.requested`

Envelope esperado:

```json
{
  "eventId": "uuid",
  "eventType": "payment.processed",
  "occurredAt": "2026-06-12T22:30:00Z",
  "data": {
    "payment_id": "uuid",
    "subscription_id": "uuid",
    "user_id": "uuid"
  }
}
```

