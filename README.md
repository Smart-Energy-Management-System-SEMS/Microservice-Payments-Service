# Microservice-Payments-Service

Microservicio de pagos de SEMS. Expone endpoints REST de payment methods, payments, invoices y webhook Stripe.

## Health checks

- `GET /health`
- `GET /api/v1/health`

Ambos endpoints devuelven `200 OK`.

## Swagger local

- `GET /swagger/`
- `GET /swagger/openapi.json`

Desde Swagger puedes probar `POST /api/v1/payments/process` con un body de ejemplo ya cargado.

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

## Contrato de eventos Kafka

Todos los eventos consumidos por este servicio esperan un envelope con esta forma:

```json
{
  "eventId": "uuid",
  "eventType": "subscription.created",
  "occurredAt": "2026-06-12T22:30:00Z",
  "data": {
    "subscription_id": "uuid",
    "user_id": "uuid"
  }
}
```

Campos requeridos por evento:

- `subscription.created`: `data.subscription_id`, `data.user_id`
- `subscription.cancelled`: `data.subscription_id`, `data.user_id`
- `subscription.renewal.requested`: `data.subscription_id`, `data.user_id`, `data.payment_method_id`, `data.amount`
- `billing.payment.requested`: `data.subscription_id`, `data.user_id`, `data.payment_method_id`, `data.amount`

Notas importantes:

- El servicio lee `eventType` desde el nivel raíz del mensaje.
- El payload funcional del evento debe venir dentro de `data`.
- Los nombres de campos esperados son `snake_case`.
- Si falta `data` o faltan campos obligatorios, el consumer registrará un `kafka handler error`.
- Cuando el handler falla, el mensaje no se confirma (`commit`) y puede volver a procesarse en el siguiente intento del mismo consumer group.

## Diagnóstico rápido de consumo Kafka

Si el servicio arranca bien pero luego aparecen logs como:

```text
kafka handler error topic=subscriptions.events partition=0 offset=1: subscription.created payload requires subscription_id and user_id
```

significa que el proceso ya se conectó correctamente a Kafka, pero recibió un mensaje cuyo `data` no cumple el contrato esperado.

Puntos a revisar:

- El productor realmente está enviando `subscription_id` y `user_id` dentro de `data`.
- El mensaje no pertenece a un formato legacy o a otro servicio con otro contrato.
- El `KAFKA_CONSUMER_GROUP` no está releyendo mensajes antiguos incompatibles.

