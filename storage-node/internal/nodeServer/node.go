package nodeServer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/config"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/interfaces"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/models"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/grpc"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/rest"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Node struct {
	id                string
	ctx               context.Context
	restServer        *rest.Server
	grpcServer        *grpc.Server
	clients           *grpc.Multiclient
	mu                sync.RWMutex
	role              config.ROLE
	leaderGRPCAddress string
	repo              interfaces.Repository
	stopOnce          sync.Once
	stopErr           error
}

func New(ctx context.Context, cfg *config.Config, repo interfaces.Repository) (*Node, error) {
	if strings.TrimSpace(cfg.NodeID) == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	restServer, err := rest.New(ctx, cfg.RestPort, repo)
	if err != nil {
		return nil, err
	}
	clients := grpc.NewMulticlient()
	for _, address := range cfg.PeerGRPCAddresses {
		clients.AddClient(grpc.Client{Address: address})
	}
	node := &Node{id: cfg.NodeID, ctx: ctx, restServer: restServer, repo: repo, clients: clients}
	node.grpcServer = grpc.NewServer(ctx, cfg.GrpcPort, repo, node.submitValue)
	if err := node.SetRole(cfg.ROLE, cfg.LeaderGRPCAddress); err != nil {
		return nil, err
	}
	return node, nil
}

func (n *Node) SetRole(role config.ROLE, leaderGRPCAddress string) error {
	if role != config.MASTER && role != config.REPLICA {
		return fmt.Errorf("unknown role: %q", role)
	}
	if role == config.MASTER {
		leaderGRPCAddress = ""
	}
	if leaderGRPCAddress != "" {
		host, port, err := net.SplitHostPort(leaderGRPCAddress)
		if err != nil || host == "" {
			return fmt.Errorf("leader gRPC address must be host:port")
		}
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 {
			return fmt.Errorf("invalid leader gRPC port")
		}
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.role, n.leaderGRPCAddress = role, leaderGRPCAddress
	return nil
}

func (n *Node) Role() config.ROLE {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.role
}

func (n *Node) LeaderGRPCAddress() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.leaderGRPCAddress
}

func (n *Node) Start() error {
	ctx, cancel := context.WithCancel(n.ctx)
	defer cancel()
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		<-ctx.Done()
		_ = n.Stop()
	}()
	var group errgroup.Group
	group.Go(func() error { defer cancel(); return n.grpcServer.Start(ctx) })
	group.Go(func() error { defer cancel(); return n.restServer.Start(ctx, n.routeWrite) })
	err := group.Wait()
	<-stopped
	return err
}

func (n *Node) Stop() error {
	n.stopOnce.Do(func() {
		n.stopErr = n.restServer.Stop()
		n.grpcServer.Stop()
	})
	return n.stopErr
}

func (n *Node) routeWrite(_ func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		request, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		var data models.RestMessage
		if err := json.Unmarshal(request, &data); err != nil || data.User == "" || data.Key == "" {
			http.Error(w, "user and key are required", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), logger.LoggerKey, logger.GetLogger(n.ctx))
		n.mu.RLock()
		if n.role == config.MASTER {
			err = n.storeLeaderValue(ctx, data.User, data.Key, data.Value)
		} else if n.leaderGRPCAddress == "" {
			err = status.Error(codes.Unavailable, "leader unknown")
		} else {
			client := grpc.Client{Address: n.leaderGRPCAddress}
			err = client.SubmitValue(ctx, data.User, data.Key, data.Value)
		}
		n.mu.RUnlock()
		if err != nil {
			logger.GetLogger(n.ctx).Warn(n.ctx, "Write request failed")
			code := http.StatusBadGateway
			switch status.Code(err) {
			case codes.InvalidArgument:
				code = http.StatusBadRequest
			case codes.Unavailable, codes.FailedPrecondition:
				code = http.StatusServiceUnavailable
			case codes.DeadlineExceeded:
				code = http.StatusGatewayTimeout
			}
			http.Error(w, http.StatusText(code), code)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (n *Node) submitValue(ctx context.Context, user, key, value string) error {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.role != config.MASTER {
		return status.Error(codes.FailedPrecondition, "node is not the leader")
	}
	return n.storeLeaderValue(ctx, user, key, value)
}

func (n *Node) storeLeaderValue(ctx context.Context, user, key, value string) error {
	log := logger.GetLogger(n.ctx)
	ctx = context.WithValue(ctx, logger.LoggerKey, log)
	if err := n.clients.AppendValue(ctx, user, key, value); err != nil {
		log.Warn(ctx, "Failed to send request to replicas")
	}
	if err := n.repo.Add(ctx, user, key, value); err != nil {
		return status.Error(codes.Internal, "could not store value")
	}
	return nil
}
