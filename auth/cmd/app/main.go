package main

import (
	"auth/internal/storagenode"
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
	s := userserver.New("1128", userrepo.New(), storagenode.NewSimpleStorage())
	mainLogger.Info(ctx, "serverStarted")
	if s.Run(ctx) != nil {
	}
}
