package main

import (
	"auth/internal/config"
	"auth/internal/storage/node_storage"
	"auth/internal/transport/UserServer"
	"auth/pkg/logger"
	"auth/pkg/storage/tokenrepo"
	"auth/pkg/storage/userrepo"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const (
	serviceName = "auth_service"
)

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.LoggerKey, logger.New(serviceName))
	mainLogger := logger.GetLogger(ctx)
	cfg, err := config.ReadFromFile()

	if err != nil {
		mainLogger.Error(ctx, "Error reading config")
		return
	}

	repo, err := userrepo.NewUserRepoPg(ctx, cfg.AuthPostgreConfig)
	if err != nil {
		mainLogger.Error(ctx, "Error making db")
		return
	}
	defer repo.Close()
	tokens, err := tokenrepo.New(ctx, repo.Pool())
	if err != nil {
		mainLogger.Error(ctx, "Error initializing token storage")
		return
	}
	s := userserver.New(ctx, "1128", repo, nodestorage.New(cfg.StorageURL), tokens)
	mainLogger.Info(ctx, "serverStarted")

	failures := make(chan error, 1)
	go func() { failures <- s.Run() }()
	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(graceCh)
	select {
	case <-graceCh:
	case runErr := <-failures:
		if runErr != nil {
			mainLogger.Error(ctx, "Server failed: "+runErr.Error())
		}
	}

	if err = s.Stop(); err != nil {
		mainLogger.Error(ctx, "Graceful shutdown failed")
	}
	fmt.Println("Server Stopped")
}
