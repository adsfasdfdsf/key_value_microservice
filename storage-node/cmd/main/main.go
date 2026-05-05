package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/config"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/rest"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/storage/map_storage"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
)

const (
	serviceName = "node storage"
)

func main() {
	os.Setenv("REST_PORT", "7070")
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.LoggerKey, logger.New(serviceName))
	cfg := config.New(ctx)
	repo := map_storage.New()
	mainLogger := logger.GetLogger(ctx)
	restServer, err := rest.New(ctx, cfg.RestPort, repo)
	if err != nil || restServer == nil {
		mainLogger.Error(ctx, "Server creation failed")
		return
	}

	go func() {
		if err = restServer.Start(ctx); err != nil {
			mainLogger.Error(ctx, "Server start failed")
			return
		}
	}()

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	<- graceCh
	
	if err = restServer.Stop(); err != nil {
		mainLogger.Error(ctx, "Graceful shutdown failed")
	}
	fmt.Println("Server stopped")
}