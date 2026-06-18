package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"Microservice-Payments-Service/payments/application/commandservices"
	"Microservice-Payments-Service/payments/application/eventhandlers"
	"Microservice-Payments-Service/payments/application/queryservices"
	appconfig "Microservice-Payments-Service/payments/infrastructure/configuration"
	kafkaadapter "Microservice-Payments-Service/payments/infrastructure/messaging/kafka"
	stripeadapter "Microservice-Payments-Service/payments/infrastructure/payments/stripe"
	gormconfig "Microservice-Payments-Service/payments/infrastructure/persistence/gorm/configuration"
	gormrepos "Microservice-Payments-Service/payments/infrastructure/persistence/gorm/repositories"
	"Microservice-Payments-Service/payments/interfaces/rest/controllers"
	restswagger "Microservice-Payments-Service/payments/interfaces/rest/swagger"
)

func main() {
	cfg := appconfig.Load()
	appconfig.ConfigureLogger(cfg.AppEnv)

	db, err := gormconfig.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	if cfg.AutoMigrate {
		if err := gormconfig.AutoMigrate(db); err != nil {
			log.Fatalf("database migration failed: %v", err)
		}
		log.Println("database automigration completed")
	}

	paymentMethodRepository := gormrepos.NewGormPaymentMethodRepository(db)
	paymentRepository := gormrepos.NewGormPaymentRepository(db)
	invoiceRepository := gormrepos.NewGormInvoiceRepository(db)
	webhookEventRepository := gormrepos.NewGormWebhookEventRepository(db)

	topics := kafkaadapter.Topics{
		PaymentsEvents:      cfg.KafkaPaymentsEventsTopic,
		BillingEvents:       cfg.KafkaBillingEventsTopic,
		SubscriptionsEvents: cfg.KafkaSubscriptionsEventsTopic,
	}
	kafkaConfig := kafkaadapter.ConnectionConfig{
		Brokers:          cfg.KafkaBrokers,
		SecurityProtocol: cfg.KafkaSecurityProtocol,
		SASLMechanism:    cfg.KafkaSASLMechanism,
		Username:         cfg.KafkaUsername,
		Password:         cfg.KafkaPassword,
		ClientID:         cfg.KafkaClientID,
		ConsumerGroup:    cfg.KafkaConsumerGroup,
	}
	log.Printf(
		"kafka configured brokers=%v produced_topics=[%s,%s] consumed_topics=[%s,%s]",
		cfg.KafkaBrokers,
		topics.PaymentsEvents,
		topics.BillingEvents,
		topics.SubscriptionsEvents,
		topics.BillingEvents,
	)
	log.Printf("kafka username=[%s]", os.Getenv("KAFKA_USERNAME"))
	log.Printf("kafka effective username=[%s]", cfg.KafkaUsername)
	log.Printf("kafka password starts Endpoint=%t", strings.HasPrefix(os.Getenv("KAFKA_PASSWORD"), "Endpoint=sb://sems-kafka-ns.servicebus.windows.net/;"))
	if cfg.KafkaEnsureTopics {
		if err := kafkaadapter.EnsureTopics(kafkaConfig, topics); err != nil {
			log.Fatalf("kafka topic ensure failed: %v", err)
		}
	} else {
		log.Printf("kafka topic ensure skipped: KAFKA_ENSURE_TOPICS=%t", cfg.KafkaEnsureTopics)
	}
	publisher := kafkaadapter.NewProducer(kafkaConfig, topics)
	paymentProvider := stripeadapter.NewAdapter(cfg.StripeSecretKey, cfg.StripeWebhookSecret)

	paymentMethodCommands := commandservices.NewPaymentMethodCommandService(paymentMethodRepository, paymentProvider, publisher)
	paymentCommands := commandservices.NewPaymentCommandService(paymentRepository, paymentMethodRepository, invoiceRepository, paymentProvider, publisher)
	webhookCommands := commandservices.NewWebhookCommandService(webhookEventRepository, paymentProvider, paymentCommands)

	paymentMethodQueries := queryservices.NewPaymentMethodQueryService(paymentMethodRepository)
	paymentQueries := queryservices.NewPaymentQueryService(paymentRepository)
	invoiceQueries := queryservices.NewInvoiceQueryService(invoiceRepository)

	subscriptionHandler := eventhandlers.NewSubscriptionEventsHandler(paymentCommands)
	consumer := kafkaadapter.NewConsumer(kafkaConfig, topics, subscriptionHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("kafka consumer startup failed: %v", err)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowCredentials: cfg.CORSAllowCredentials,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Stripe-Signature"},
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
	}))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "payments-service"})
	})
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "payments-service"})
	})
	if cfg.SwaggerEnabled {
		restswagger.RegisterRoutes(router, cfg.ServerPort, cfg.APIBasePath)
	}

	controllers.RegisterRoutes(router, cfg.APIBasePath, controllers.Controllers{
		PaymentMethods: controllers.NewPaymentMethodController(paymentMethodCommands, paymentMethodQueries),
		Payments:       controllers.NewPaymentController(paymentCommands, paymentQueries, cfg.StripeCurrency),
		Invoices:       controllers.NewInvoiceController(invoiceQueries),
		Webhooks:       controllers.NewWebhookController(webhookCommands),
	})

	server := &http.Server{Addr: ":" + cfg.ServerPort, Handler: router}
	go func() {
		log.Printf("payments service listening on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down payments service")

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	if err := consumer.Close(); err != nil {
		log.Printf("kafka consumer close error: %v", err)
	}
	if err := publisher.Close(); err != nil {
		log.Printf("kafka producer close error: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}
