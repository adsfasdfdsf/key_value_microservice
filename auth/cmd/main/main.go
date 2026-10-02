package main

import (
	"auth/internal/config"
	"auth/internal/storagenode"
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
	s := userserver.New(ctx, "1128", repo, storagenode.NewSimpleStorage(), tokens)
	mainLogger.Info(ctx, "serverStarted")

	go func() {
		if err = s.Run(); err != nil {
			mainLogger.Error(ctx, "Server start failed")
			return
		}
	}()

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	<-graceCh

	if err = s.Stop(); err != nil {
		mainLogger.Error(ctx, "Graceful shutdown failed")
	}
	fmt.Println("Server Stopped")
}
