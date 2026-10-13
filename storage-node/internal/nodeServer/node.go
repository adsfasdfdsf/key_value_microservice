package nodeServer

import (
	"context"
	"encoding/json"
	"fmt"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

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
	ownGRPCAddress    string
	peers             []string
	consensusStore    interfaces.ConsensusStore
	promisedBallot    *desc.Ballot
	observedRound     uint64
	electionMu        sync.Mutex
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
	if err := validateGRPCAddress(cfg.NodeGRPCAddress); err != nil {
		return nil, err
	}
	if len(cfg.PeerGRPCAddresses) != 2 {
		return nil, fmt.Errorf("three-node cluster requires exactly two peers")
	}
	seen := map[string]bool{cfg.NodeGRPCAddress: true}
	for _, address := range cfg.PeerGRPCAddresses {
		if err := validateGRPCAddress(address); err != nil {
			return nil, err
		}
		if seen[address] {
			return nil, fmt.Errorf("duplicate or self peer address")
		}
		seen[address] = true
	}
	store, ok := repo.(interfaces.ConsensusStore)
	if !ok {
		return nil, fmt.Errorf("repository must persist consensus promises")
	}
	loadCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	promised, err := store.LoadPromisedBallot(loadCtx)
	cancel()
	if err != nil {
		return nil, fmt.Errorf("load promised ballot: %w", err)
	}
	restServer, err := rest.New(ctx, cfg.RestPort, repo)
	if err != nil {
		return nil, err
	}
	clients := grpc.NewMulticlient()
	for _, address := range cfg.PeerGRPCAddresses {
		clients.AddClient(grpc.Client{Address: address})
	}
	node := &Node{id: cfg.NodeID, ctx: ctx, restServer: restServer, repo: repo, clients: clients,
		ownGRPCAddress: cfg.NodeGRPCAddress, peers: append([]string(nil), cfg.PeerGRPCAddresses...), consensusStore: store, promisedBallot: cloneBallot(promised)}
	node.grpcServer = grpc.NewServer(ctx, cfg.GrpcPort, repo, node.submitValue, node)
	if err := node.SetRole(cfg.ROLE, cfg.LeaderGRPCAddress); err != nil {
		return nil, err
	}
	if promised != nil {
		node.role = config.REPLICA
		node.leaderGRPCAddress = ""
		node.observedRound = promised.Round
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
			n.mu.RUnlock()
		} else {
			address := n.leaderGRPCAddress
			n.mu.RUnlock()
			// Release the role lock before forwarding: the returning replication RPC needs it.
			if address == "" {
				err = status.Error(codes.Unavailable, "leader unknown")
			} else {
				client := grpc.Client{Address: address}
				err = client.SubmitValue(ctx, data.User, data.Key, data.Value)
			}
		}
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
	if err := n.clients.AppendValue(ctx, user, key, value, cloneBallot(n.promisedBallot)); err != nil {
		log.Warn(ctx, "Failed to send request to replicas")
	}
	if err := n.repo.Add(ctx, user, key, value); err != nil {
		return status.Error(codes.Internal, "could not store value")
	}
	return nil
}

// AppendReplicaValue enforces the saved promise on replicated writes.
// Empty ballots are accepted only before the first election, for initial bootstrap.
func (n *Node) AppendReplicaValue(ctx context.Context, user, key, value string, ballot *desc.Ballot) error {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if compareBallots(ballot, n.promisedBallot) != 0 {
		return status.Error(codes.FailedPrecondition, "replication ballot does not match promised ballot")
	}
	if ballot != nil && !validBallot(ballot) {
		return status.Error(codes.InvalidArgument, "invalid replication ballot")
	}
	if err := n.repo.Add(ctx, user, key, value); err != nil {
		return status.Error(codes.Internal, "could not store replicated value")
	}
	return nil
}
