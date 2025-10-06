package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	apiserver "github.com/dcm-project/service-provider-api/internal/api_server"
	"github.com/dcm-project/service-provider-api/internal/config"
	"github.com/dcm-project/service-provider-api/internal/store"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	err := runCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the planner api",
	RunE: func(cmd *cobra.Command, args []string) error {
		defer zap.S().Info("API service stopped")

		cfg, err := config.New()
		if err != nil {
			zap.S().Fatalw("reading configuration", "error", err)
		}

		zap.S().Info("Starting API service...")
		zap.S().Info("Initializing data store")
		db, err := store.InitDB(cfg)
		if err != nil {
			zap.S().Fatalw("initializing data store", "error", err)
		}

		store := store.NewStore(db)
		defer store.Close()

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)

		go func() {
			defer cancel()
			listener, err := newListener(cfg.Service.Address)
			if err != nil {
				zap.S().Fatalw("creating listener", "error", err)
			}

			server := apiserver.New(cfg, store, listener)
			if err := server.Run(ctx); err != nil {
				zap.S().Fatalw("Error running server", "error", err)
			}
		}()

		<-ctx.Done()

		return nil
	},
}

func newListener(address string) (net.Listener, error) {
	if address == "" {
		address = "localhost:0"
	}
	return net.Listen("tcp", address)
}
