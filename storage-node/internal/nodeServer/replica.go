package nodeServer

import (
	"context"
	"net/http"

	"golang.org/x/sync/errgroup"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/interfaces"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/grpc"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/rest"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
)


type Replica struct {
	ctx context.Context
	grpcServer *grpc.Server
	restServer *rest.Server
}

func NewReplica(ctx context.Context, grpcServer *grpc.Server, restServer *rest.Server, repo interfaces.Repository) *Replica{
	return &Replica{
		ctx: ctx,
		grpcServer: grpcServer,
		restServer: restServer,
	}
}

func (s *Replica) Start() error{
	eg := errgroup.Group{}
	
	eg.Go(func() error {
		logger.GetLogger(s.ctx).Info(s.ctx, "starting gRPC server")
		return s.grpcServer.Start(s.ctx)
	})

	eg.Go(func() error {
		logger.GetLogger(s.ctx).Info(s.ctx, "starting rest server")
		interceptor := func(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request){
				handler(w, r)
			}
		}
		return s.restServer.Start(s.ctx, interceptor)
	})

	return eg.Wait()
}

func (s *Replica) Stop() error{
	s.grpcServer.Stop()
	l := logger.GetLogger(s.ctx)
	if l != nil {
		l.Info(s.ctx, "gRPC server stopped")
	}
	return s.restServer.Stop()
}