# Microservice-Payments-Service

Microservicio de pagos de SEMS. Expone endpoints REST de payment methods, payments, invoices y webhook Stripe.

## Ejecucion local (Gateway + Config-Service)

- Config-Service local: `http://localhost:8090`
- API Gateway local: `http://localhost:8081`
- Puerto local del microservicio: `8085`
- Base URL local final: `http://localhost:8085`
- Route prefix: `/api/v1`

## Health check

Endpoint publico sin autenticacion:

- `GET /health` -> `200 OK`

## Configuracion por entorno

Copiar `.env.example` a `.env`.

### Variables no sensibles

- `APP_ENV`
- `GIN_MODE`
- `SERVER_PORT`
- `DB_AUTO_MIGRATE`
- `CONFIG_SERVICE_URL`
- `CORS_ALLOWED_ORIGINS`
- `CORS_ALLOW_CREDENTIALS`
- `API_BASE_PATH` (fallback)
- `STRIPE_CURRENCY` (fallback)
- `KAFKA_*` (fallback)

### Variables sensibles

- `DATABASE_URL`
- `STRIPE_SECRET_KEY`
- `STRIPE_WEBHOOK_SECRET`

## Config Service

Si `CONFIG_SERVICE_URL` esta definido, el servicio consulta:

- `GET /api/v1/config/{service-name}`
- `GET /api/v1/config/kafka`
- `GET /api/v1/config/services` (opcional)

`service-name` usado: `payments-service`.

Si Config-Service no responde, usa fallback del `.env`.

## CORS local

Por defecto permite:

- `http://localhost:3000`
- `http://localhost:5173`

Controlado por:

- `CORS_ALLOWED_ORIGINS`
- `CORS_ALLOW_CREDENTIALS`

## Endpoints reales

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

## Auth/JWT con Gateway

Este microservicio no valida JWT internamente.

- Si `API_GATEWAY_AUTH_REQUIRED=false`: se puede probar sin JWT.
- Endpoints publicos recomendados en Gateway:
  - `GET /health`
  - `POST /api/v1/webhooks/stripe`
- Endpoints protegidos recomendados:
  - resto de `/api/v1/payment-methods`, `/api/v1/payments`, `/api/v1/invoices`

## Dependencias locales

Kafka local (ya levantado en `localhost:9092`) y Postgres accesible desde `DATABASE_URL`.

## Pruebas rapidas

Health del MS:

```bash
curl -i http://localhost:8085/health
```

Endpoint principal del MS (ejemplo):

```bash
curl -i http://localhost:8085/api/v1/payments/user/test-user
```

Endpoint via Gateway (ejemplo proxied, ajusta path segun tu gateway):

```bash
curl -i http://localhost:8081/payments/api/v1/payments/user/test-user
```

## Azure Container Apps

- Definir env vars no sensibles en la app.
- Definir `DATABASE_URL`, `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET` como secretos de ACA.
- Health probe: `GET /health`.
- Permitir egress a Config-Service, Kafka y Stripe.
