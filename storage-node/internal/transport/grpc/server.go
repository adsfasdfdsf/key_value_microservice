package grpc

import (
	"context"
	"fmt"
	"net"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/interfaces"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct{
	repo interfaces.Repository
	port int
	gserv *grpc.Server
	desc.UnimplementedNodeServiceServer
}

func NewServer(ctx context.Context, port int, repo interfaces.Repository) *Server{
	return &Server{
		repo: repo,
		port: port,
	}
}

func (s *Server) Start(ctx context.Context) error {
	log := logger.GetLogger(ctx)
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			ContextWithLogger(log),
		),
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Error(ctx, "grpc listener creation failed")
		return err
	}
	s.gserv = grpc.NewServer(opts...)
	reflection.Register(s.gserv)
	desc.RegisterNodeServiceServer(s.gserv, s)
	log.Info(ctx, fmt.Sprintf("grpc Server started on port %d", s.port))

	return s.gserv.Serve(lis)
}

func (s *Server) Stop()  {
	s.gserv.GracefulStop()
}

func (s *Server) AppendValue(ctx context.Context, req *desc.AppendValueRequest) (*desc.AppendValueResponse, error){
	err := s.repo.Add(req.Key, req.Value)
	if (err != nil){
		return &desc.AppendValueResponse{
			Success: false,
		}, err
	}
	return &desc.AppendValueResponse{
		Success: true,
	}, nil
}