package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	apiserver "github.com/dcm-project/service-provider-manager/internal/api_server"
	"github.com/dcm-project/service-provider-manager/internal/config"
	"github.com/dcm-project/service-provider-manager/internal/consumer"
	"github.com/dcm-project/service-provider-manager/internal/handlers"
	rmhandlers "github.com/dcm-project/service-provider-manager/internal/handlers/resource_manager"
	"github.com/dcm-project/service-provider-manager/internal/healthcheck"
	"github.com/dcm-project/service-provider-manager/internal/service"
	rmsvc "github.com/dcm-project/service-provider-manager/internal/service/resource_manager"
	"github.com/dcm-project/service-provider-manager/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := store.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize store, service, and handler
	dataStore := store.NewStore(db)
	defer dataStore.Close()

	// Provider API
	providerService := service.NewProviderService(dataStore)
	providerHandler := handlers.NewHandler(providerService)

	// Resource Manager API
	instanceService := rmsvc.NewInstanceService(dataStore)
	rmHandler := rmhandlers.NewHandler(instanceService)

	// Initialize StatusConsumer
	statusConsumer, err := consumer.New(cfg.NATS.URL, cfg.NATS.Subject, dataStore,
		consumer.SetStreamName(cfg.NATS.StreamName),
		consumer.SetConsumerName(cfg.NATS.ConsumerName),
	)
	if err != nil {
		log.Fatalf("Failed to initialize status consumer: %v", err)
	}
	log.Printf("Status consumer initialized with NATS URL: %s (stream=%s, consumer=%s)",
		cfg.NATS.URL, cfg.NATS.StreamName, cfg.NATS.ConsumerName)

	// Start server
	listener, err := net.Listen("tcp", cfg.Service.Address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	srv := apiserver.New(cfg, listener, providerHandler, rmHandler)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start StatusConsumer
	if err := statusConsumer.Start(ctx); err != nil {
		log.Fatalf("Failed to start status consumer: %v", err)
	}
	defer statusConsumer.Stop()

	// Start health check monitor
	healthMonitor := healthcheck.NewMonitor(dataStore.Provider(), cfg.HealthCheck)
	healthMonitor.Start(ctx)
	defer healthMonitor.Stop()
	log.Printf("Health check monitor started (interval: 10s)")

	log.Printf("Starting server on %s", listener.Addr().String())
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
