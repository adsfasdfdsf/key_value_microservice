package main

import (
	"auth/pkg/logger"
	"context"
)


const (
	serviceName = "auth_service"
)


func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.LoggerKey, logger.New(serviceName))
	mainLogger := ctx.Value(logger.LoggerKey)
	
}