package grpc

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"time"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/interfaces"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type SubmitValueHandler func(context.Context, string, string, string) error

type ConsensusHandler interface {
	AppendReplicaValue(context.Context, string, string, string, *desc.Ballot) error
	AcceptEntries(context.Context, *desc.AcceptEntriesRequest) (*desc.JournalResponse, error)
	CommitEntries(context.Context, *desc.CommitEntriesRequest) (*desc.JournalResponse, error)
	Prepare(context.Context, *desc.PrepareRequest) (*desc.PrepareResponse, error)
	AnnounceLeader(context.Context, *desc.AnnounceLeaderRequest) (*desc.AnnounceLeaderResponse, error)
	StartElection(context.Context, *desc.StartElectionRequest) (*desc.StartElectionResponse, error)
}

type Server struct {
	consensus   ConsensusHandler
	submitValue SubmitValueHandler
	repo        interfaces.Repository
	port        int
	gserv       *grpc.Server
	desc.UnimplementedNodeServiceServer
}

func NewServer(ctx context.Context, port int, repo interfaces.Repository, submitValue SubmitValueHandler, consensus ConsensusHandler) *Server {
	log := logger.GetLogger(ctx)
	server := &Server{consensus: consensus, submitValue: submitValue, repo: repo, port: port, gserv: grpc.NewServer(grpc.MaxRecvMsgSize(32<<20), grpc.MaxSendMsgSize(32<<20), grpc.ChainUnaryInterceptor(ContextWithLogger(log)))}
	reflection.Register(server.gserv)
	desc.RegisterNodeServiceServer(server.gserv, server)
	return server
}

func (s *Server) Start(ctx context.Context) error {
	log := logger.GetLogger(ctx)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Error(ctx, "grpc listener creation failed")
		return err
	}
	log.Info(ctx, fmt.Sprintf("grpc Server started on port %d", s.port))

	err = s.gserv.Serve(lis)
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}

func (s *Server) Stop() {
	timer := time.AfterFunc(5*time.Second, s.gserv.Stop)
	defer timer.Stop()
	s.gserv.GracefulStop()
}

func (s *Server) AppendValue(ctx context.Context, req *desc.AppendValueRequest) (*desc.AppendValueResponse, error) {
	if req == nil || req.User == "" || req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "user and key are required")
	}
	err := s.consensus.AppendReplicaValue(ctx, req.User, req.Key, req.Value, req.Ballot)
	if err != nil {
		return &desc.AppendValueResponse{
			Success: false,
		}, err
	}
	return &desc.AppendValueResponse{
		Success: true,
	}, nil
}

func (s *Server) SubmitValue(ctx context.Context, req *desc.AppendValueRequest) (*desc.AppendValueResponse, error) {
	if req == nil || req.User == "" || req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "user and key are required")
	}
	if s.submitValue == nil {
		return nil, status.Error(codes.Unavailable, "leader write handler unavailable")
	}
	if err := s.submitValue(ctx, req.User, req.Key, req.Value); err != nil {
		return nil, err
	}
	return &desc.AppendValueResponse{Success: true}, nil
}

func (s *Server) Prepare(ctx context.Context, req *desc.PrepareRequest) (*desc.PrepareResponse, error) {
	return s.consensus.Prepare(ctx, req)
}
func (s *Server) AnnounceLeader(ctx context.Context, req *desc.AnnounceLeaderRequest) (*desc.AnnounceLeaderResponse, error) {
	return s.consensus.AnnounceLeader(ctx, req)
}
func (s *Server) StartElection(ctx context.Context, req *desc.StartElectionRequest) (*desc.StartElectionResponse, error) {
	return s.consensus.StartElection(ctx, req)
}

func (s *Server) AcceptEntries(ctx context.Context, req *desc.AcceptEntriesRequest) (*desc.JournalResponse, error) {
	return s.consensus.AcceptEntries(ctx, req)
}
func (s *Server) CommitEntries(ctx context.Context, req *desc.CommitEntriesRequest) (*desc.JournalResponse, error) {
	return s.consensus.CommitEntries(ctx, req)
}
