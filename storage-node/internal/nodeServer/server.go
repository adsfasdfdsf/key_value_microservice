package nodeServer

import (
	"context"

	"golang.org/x/sync/errgroup"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/interfaces"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/grpc"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/rest"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
)


type Server struct {
	grpcServer *grpc.Server
	restServer *rest.Server
}

func New(ctx context.Context, grpcServer *grpc.Server, restServer *rest.Server, repo interfaces.Repository) *Server{
	return &Server{
		grpcServer: grpcServer,
		restServer: restServer,
	}
}

func (s *Server) Start(ctx context.Context) error{
	eg := errgroup.Group{}
	
	eg.Go(func() error {
		logger.GetLogger(ctx).Info(ctx, "starting gRPC server")
		return s.grpcServer.Start(ctx)
	})

	eg.Go(func() error {
		logger.GetLogger(ctx).Info(ctx, "starting rest server")
		return s.restServer.Start(ctx)
	})

	return eg.Wait()
}

func (s *Server) Stop(ctx context.Context) error{
	s.grpcServer.Stop()
	l := logger.GetLogger(ctx)
	if l != nil {
		l.Info(ctx, "gRPC server stopped")
	}
	return s.restServer.Stop()
}