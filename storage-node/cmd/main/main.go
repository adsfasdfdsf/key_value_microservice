package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/config"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/nodeServer"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/storage/postgres"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
)

const (
	serviceName = "node storage"
)

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.LoggerKey, logger.New(serviceName))
	cfg := config.New(ctx)
	repo, err := postgres.New(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, err.Error())
		return
	}
	defer repo.Close()
	mainLogger := logger.GetLogger(ctx)
	node, err := nodeServer.New(ctx, cfg, repo)
	if err != nil {
		mainLogger.Error(ctx, err.Error())
		return
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan error, 1)
	go func() {
		mainLogger.Info(ctx, fmt.Sprintf("Starting %v node with REST and gRPC", node.Role()))
		done <- node.Start()
	}()
	finished := false
	select {
	case <-graceCh:
	case err := <-done:
		finished = true
		if err != nil {
			mainLogger.Error(ctx, err.Error())
		}
	}

	if err = node.Stop(); err != nil {
		mainLogger.Error(ctx, "Graceful shutdown failed")
	}
	if !finished {
		<-done
	}
	fmt.Println("Server stopped")
}
