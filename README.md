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
- `KAFKA_TOPIC_PAYMENT_PROCESSED`
- `KAFKA_TOPIC_PAYMENT_FAILED`
- `KAFKA_TOPIC_INVOICE_GENERATED`
- `KAFKA_TOPIC_PAYMENT_METHOD_ADDED`
- `KAFKA_TOPIC_SUBSCRIPTION_CREATED`
- `KAFKA_TOPIC_SUBSCRIPTION_RENEWAL_REQUESTED`
- `KAFKA_TOPIC_SUBSCRIPTION_CANCELLED`

Nota de compatibilidad: el servicio prioriza `PORT`; si no existe, usa `SERVER_PORT`.

## Ejemplo local (.env)

```env
PORT=8080
CONFIG_SERVICE_URL=http://localhost:8090
KAFKA_BROKERS=localhost:9092
KAFKA_SECURITY_PROTOCOL=
KAFKA_SASL_MECHANISM=
KAFKA_USERNAME=
KAFKA_PASSWORD=
KAFKA_TOPIC_PAYMENT_PROCESSED=payment.processed
KAFKA_TOPIC_PAYMENT_FAILED=payment.failed
KAFKA_TOPIC_INVOICE_GENERATED=invoice.generated
KAFKA_TOPIC_PAYMENT_METHOD_ADDED=payment.method.added
KAFKA_TOPIC_SUBSCRIPTION_CREATED=subscription.created
KAFKA_TOPIC_SUBSCRIPTION_RENEWAL_REQUESTED=subscription.renewal.requested
KAFKA_TOPIC_SUBSCRIPTION_CANCELLED=subscription.cancelled
DATABASE_URL=postgresql://USER:PASSWORD@localhost:5432/payments?sslmode=disable
GIN_MODE=debug
```

## Docker

Build de imagen:

```bash
docker build -t sems-payments-service:local .
```

Run local con archivo `.env`:

```bash
docker run --rm -p 8080:8080 --env-file .env --name sems-payments-service sems-payments-service:local
```

Prueba rápida:

```bash
curl -i http://localhost:8080/api/v1/health
```

## Ejemplo Azure Container Apps

El contenedor no debe usar `localhost` para servicios externos. Define variables en ACA con hosts reales (Kafka/Config Service/DB).

Variables mínimas recomendadas en ACA:

```text
PORT=8080
GIN_MODE=release
CONFIG_SERVICE_URL=https://<config-service-url>
KAFKA_BROKERS=<broker1:9092,broker2:9092>
KAFKA_SECURITY_PROTOCOL=SASL_SSL
KAFKA_SASL_MECHANISM=PLAIN
KAFKA_USERNAME=<usuario>
KAFKA_PASSWORD=<password>
KAFKA_TOPIC_PAYMENT_PROCESSED=payment.processed
KAFKA_TOPIC_PAYMENT_FAILED=payment.failed
KAFKA_TOPIC_INVOICE_GENERATED=invoice.generated
KAFKA_TOPIC_PAYMENT_METHOD_ADDED=payment.method.added
KAFKA_TOPIC_SUBSCRIPTION_CREATED=subscription.created
KAFKA_TOPIC_SUBSCRIPTION_RENEWAL_REQUESTED=subscription.renewal.requested
KAFKA_TOPIC_SUBSCRIPTION_CANCELLED=subscription.cancelled
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

## Kafka topics

Payments publica:

- `payment.processed`
- `payment.failed`
- `invoice.generated`
- `payment.method.added`

Payments consume:

- `subscription.created`
- `subscription.cancelled`
- `subscription.renewal.requested`

Nota: `subscription.renewal.requested` queda soportado por el consumer y por la configuracion del micro, pero hoy depende de que otro microservicio realmente lo publique. Si nadie lo produce, no rompe el arranque; simplemente no llegaran eventos de ese topic.

