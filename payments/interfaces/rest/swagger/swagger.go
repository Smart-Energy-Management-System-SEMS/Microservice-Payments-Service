package swagger

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes exposes a tiny Swagger UI backed by a local OpenAPI document.
func RegisterRoutes(router *gin.Engine, serverPort, apiBasePath string) {
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/")
	})
	router.GET("/swagger/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML()))
	})
	router.GET("/swagger/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(openAPISpec(serverPort, apiBasePath)))
	})
}

func swaggerHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Payments Service Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: './openapi.json',
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis],
      layout: 'BaseLayout'
    });
  </script>
</body>
</html>`
}

func openAPISpec(serverPort, apiBasePath string) string {
	return fmt.Sprintf(`{
  "openapi": "3.0.3",
  "info": {
    "title": "Payments Service API",
    "version": "1.0.0",
    "description": "Swagger local para probar el servicio de pagos con Stripe, Neon y Kafka/Event Hubs."
  },
  "servers": [
    {
      "url": "http://localhost:%s",
      "description": "Local"
    }
  ],
  "paths": {
    "/health": {
      "get": {
        "summary": "Health check",
        "responses": {
          "200": {
            "description": "Service healthy",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/HealthResponse"
                }
              }
            }
          }
        }
      }
    },
    "%s/health": {
      "get": {
        "summary": "Health check con prefijo API",
        "responses": {
          "200": {
            "description": "Service healthy",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/HealthResponse"
                }
              }
            }
          }
        }
      }
    },
    "%s/payment-methods": {
      "post": {
        "summary": "Registrar metodo de pago",
        "description": "Registra un metodo de pago de Stripe para un usuario.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/RegisterPaymentMethodRequest"
              },
              "example": {
                "user_id": "22222222-2222-2222-2222-222222222222",
                "type": "card",
                "stripe_payment_method_id": "pm_1TbSAMPLE123456789",
                "is_default": true
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Metodo de pago registrado",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/PaymentMethodResponse"
                }
              }
            }
          },
          "400": {
            "description": "Solicitud invalida",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "409": {
            "description": "Conflicto",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "502": {
            "description": "Fallo del proveedor externo",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "%s/payment-methods/user/{userId}": {
      "get": {
        "summary": "Listar metodos de pago por usuario",
        "parameters": [
          {
            "name": "userId",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Lista de metodos",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/PaymentMethodResponse"
                  }
                }
              }
            }
          },
          "400": {
            "description": "Solicitud invalida",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "%s/payment-methods/{paymentMethodId}/default": {
      "put": {
        "summary": "Marcar metodo de pago por defecto",
        "parameters": [
          {
            "name": "paymentMethodId",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Metodo actualizado",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/PaymentMethodResponse"
                }
              }
            }
          },
          "400": {
            "description": "Solicitud invalida",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "404": {
            "description": "Metodo no encontrado",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "%s/payment-methods/{paymentMethodId}": {
      "delete": {
        "summary": "Eliminar metodo de pago",
        "parameters": [
          {
            "name": "paymentMethodId",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          }
        ],
        "responses": {
          "204": {
            "description": "Metodo eliminado"
          },
          "400": {
            "description": "Solicitud invalida",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "404": {
            "description": "Metodo no encontrado",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "%s/payments/process": {
      "post": {
        "summary": "Procesar un pago",
        "description": "Crea y procesa un pago usando Stripe, persiste en Neon y publica eventos.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/ProcessPaymentRequest"
              },
              "example": {
                "subscription_id": "11111111-1111-1111-1111-111111111111",
                "user_id": "22222222-2222-2222-2222-222222222222",
                "payment_method_id": "33333333-3333-3333-3333-333333333333",
                "amount": 49.90,
                "currency": "pen",
                "payment_method": "card"
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Pago procesado",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ProcessPaymentResponse"
                }
              }
            }
          },
          "400": {
            "description": "Solicitud invalida",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "404": {
            "description": "Recurso no encontrado",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "409": {
            "description": "Conflicto",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "502": {
            "description": "Fallo del proveedor externo",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    }
    ,
    "%s/payments/{paymentId}": {
      "get": {
        "summary": "Buscar pago por ID",
        "parameters": [
          {
            "name": "paymentId",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Pago encontrado",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/PaymentResponse"
                }
              }
            }
          },
          "400": {
            "description": "Solicitud invalida",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "404": {
            "description": "Pago no encontrado",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "%s/payments/user/{userId}": {
      "get": {
        "summary": "Listar pagos por usuario",
        "parameters": [
          {
            "name": "userId",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Lista de pagos",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/PaymentResponse"
                  }
                }
              }
            }
          },
          "400": {
            "description": "Solicitud invalida",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "%s/payments/subscription/{subscriptionId}": {
      "get": {
        "summary": "Listar pagos por suscripcion",
        "parameters": [
          {
            "name": "subscriptionId",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Lista de pagos",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/PaymentResponse"
                  }
                }
              }
            }
          },
          "400": {
            "description": "Solicitud invalida",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "%s/invoices/{invoiceId}": {
      "get": {
        "summary": "Buscar factura por ID",
        "parameters": [
          {
            "name": "invoiceId",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Factura encontrada",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/InvoiceResponse"
                }
              }
            }
          },
          "400": {
            "description": "Solicitud invalida",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "404": {
            "description": "Factura no encontrada",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "%s/invoices/payment/{paymentId}": {
      "get": {
        "summary": "Buscar factura por payment ID",
        "parameters": [
          {
            "name": "paymentId",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Factura encontrada",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/InvoiceResponse"
                }
              }
            }
          },
          "400": {
            "description": "Solicitud invalida",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "404": {
            "description": "Factura no encontrada",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "%s/webhooks/stripe": {
      "post": {
        "summary": "Recibir webhook de Stripe",
        "description": "Endpoint para eventos firmados por Stripe. Para pruebas manuales puedes enviar un payload JSON y el header Stripe-Signature.",
        "parameters": [
          {
            "name": "Stripe-Signature",
            "in": "header",
            "required": false,
            "schema": {
              "type": "string"
            },
            "description": "Firma enviada por Stripe."
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/StripeWebhookPayload"
              },
              "example": {
                "id": "evt_test_webhook",
                "type": "payment_intent.succeeded",
                "data": {
                  "object": {
                    "id": "pi_123456789"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Webhook recibido",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/WebhookResponse"
                }
              }
            }
          },
          "400": {
            "description": "Payload invalido",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "409": {
            "description": "Evento duplicado",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/WebhookResponse"
                }
              }
            }
          },
          "502": {
            "description": "Fallo del proveedor externo",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "HealthResponse": {
        "type": "object",
        "properties": {
          "status": { "type": "string", "example": "ok" },
          "service": { "type": "string", "example": "payments-service" }
        }
      },
      "ProcessPaymentRequest": {
        "type": "object",
        "required": ["subscription_id", "user_id", "payment_method_id", "amount"],
        "properties": {
          "subscription_id": { "type": "string", "format": "uuid" },
          "user_id": { "type": "string", "format": "uuid" },
          "payment_method_id": { "type": "string", "format": "uuid" },
          "amount": { "type": "number", "format": "double", "example": 49.90 },
          "currency": { "type": "string", "example": "pen" },
          "payment_method": { "type": "string", "example": "card" }
        }
      },
      "RegisterPaymentMethodRequest": {
        "type": "object",
        "required": ["user_id", "stripe_payment_method_id"],
        "properties": {
          "user_id": { "type": "string", "format": "uuid" },
          "type": { "type": "string", "example": "card" },
          "stripe_payment_method_id": { "type": "string", "example": "pm_1TbSAMPLE123456789" },
          "is_default": { "type": "boolean", "example": true }
        }
      },
      "ProcessPaymentResponse": {
        "type": "object",
        "properties": {
          "payment": { "$ref": "#/components/schemas/PaymentResponse" },
          "invoice": { "$ref": "#/components/schemas/InvoiceResponse" }
        }
      },
      "PaymentResponse": {
        "type": "object",
        "properties": {
          "payment_id": { "type": "string", "format": "uuid" },
          "subscription_id": { "type": "string", "format": "uuid" },
          "user_id": { "type": "string", "format": "uuid" },
          "payment_method_id": { "type": "string", "format": "uuid" },
          "amount": { "type": "number", "format": "double" },
          "currency": { "type": "string", "example": "pen" },
          "status": { "type": "string", "example": "paid" },
          "payment_method": { "type": "string", "example": "card" },
          "stripe_payment_intent_id": { "type": "string", "example": "pi_123456789" },
          "paid_at": { "type": "string", "nullable": true, "example": "2026-06-18T18:10:00Z" },
          "created_at": { "type": "string", "example": "2026-06-18T18:09:00Z" }
        }
      },
      "PaymentMethodResponse": {
        "type": "object",
        "properties": {
          "payment_method_id": { "type": "string", "format": "uuid" },
          "user_id": { "type": "string", "format": "uuid" },
          "type": { "type": "string", "example": "card" },
          "brand": { "type": "string", "example": "visa" },
          "last4": { "type": "string", "example": "4242" },
          "exp_month": { "type": "integer", "example": 12 },
          "exp_year": { "type": "integer", "example": 2030 },
          "stripe_payment_method_id": { "type": "string", "example": "pm_1TbSAMPLE123456789" },
          "is_default": { "type": "boolean", "example": true },
          "created_at": { "type": "string", "example": "2026-06-18T18:00:00Z" }
        }
      },
      "InvoiceResponse": {
        "type": "object",
        "properties": {
          "invoice_id": { "type": "string", "format": "uuid" },
          "payment_id": { "type": "string", "format": "uuid" },
          "invoice_number": { "type": "string", "example": "INV-20260618-0001" },
          "issued_at": { "type": "string", "example": "2026-06-18T18:10:00Z" },
          "total_amount": { "type": "number", "format": "double", "example": 49.90 },
          "pdf_url": { "type": "string", "example": "https://example.com/invoices/INV-20260618-0001.pdf" }
        }
      },
      "WebhookResponse": {
        "type": "object",
        "properties": {
          "received": { "type": "boolean", "example": true },
          "duplicate": { "type": "boolean", "example": false },
          "event_id": { "type": "string", "example": "11111111-1111-1111-1111-111111111111" }
        }
      },
      "StripeWebhookPayload": {
        "type": "object",
        "additionalProperties": true,
        "properties": {
          "id": { "type": "string", "example": "evt_test_webhook" },
          "type": { "type": "string", "example": "payment_intent.succeeded" },
          "data": {
            "type": "object",
            "additionalProperties": true
          }
        }
      },
      "ErrorResponse": {
        "type": "object",
        "properties": {
          "error": { "type": "string", "example": "record not found" }
        }
      }
    }
  }
}`, serverPort, apiBasePath, apiBasePath, apiBasePath, apiBasePath, apiBasePath, apiBasePath, apiBasePath, apiBasePath, apiBasePath, apiBasePath, apiBasePath, apiBasePath)
}
