package main

import (
	userserver "auth/internal/transport/UserServer"
	"context"
)

const (
	serviceName = "auth_service"
)

func main() {
	//ctx := context.Background()
	//ctx = context.WithValue(ctx, logger.LoggerKey, logger.New(serviceName))
	//mainLogger := ctx.Value(logger.LoggerKey)
	s := userserver.New("1128")
	if s.Run(context.Background()) != nil {
	}
}
