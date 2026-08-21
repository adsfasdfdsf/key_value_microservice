package main

import (
	userserver "auth/internal/transport/UserServer"
	"auth/pkg/logger"
	"auth/pkg/storage/userrepo"
	"context"
)

const (
	serviceName = "auth_service"
)

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.LoggerKey, logger.New(serviceName))
	mainLogger := logger.GetLogger(ctx)
	s := userserver.New("1128", userrepo.New())
	mainLogger.Info(ctx, "serverStarted")
	if s.Run(ctx) != nil {
	}
}
