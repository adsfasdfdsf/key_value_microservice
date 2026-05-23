package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/config"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/nodeServer"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/storage/map_storage"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/grpc"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/rest"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
)

const (
	serviceName = "node storage"
)

type Node interface {
	Start() error
	Stop() error
}

func main() {
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
	var node Node

	switch cfg.ROLE {
	case config.MASTER:
		mc := grpc.NewMulticlient()
		mc.AddClient(grpc.Client{Address: "replica1:4040"})
		mc.AddClient(grpc.Client{Address: "replica2:4040"})
		node = nodeServer.NewMaster(ctx, restServer, mc)
	case config.REPLICA:
		grpcServer := grpc.NewServer(ctx, cfg.GrpcPort, repo)
		node = nodeServer.NewReplica(ctx, grpcServer, restServer, repo)
	}

	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)


	go func(){
		mainLogger.Info(ctx, fmt.Sprintf("Starting %v server", cfg.ROLE))
		if err := node.Start(); err != nil {
			mainLogger.Error(ctx, err.Error())
		}
	}()

	<- graceCh
	
	if err = node.Stop(); err != nil {
		mainLogger.Error(ctx, "Graceful shutdown failed")
	}
	fmt.Println("Server stopped")
}