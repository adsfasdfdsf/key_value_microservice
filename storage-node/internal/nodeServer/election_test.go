package nodeServer

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/config"
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/storage/map_storage"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type promiseStore struct {
	*map_storage.MapStorage
	mu      sync.Mutex
	promise *desc.Ballot
	err     error
}

func (s *promiseStore) LoadPromisedBallot(context.Context) (*desc.Ballot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneBallot(s.promise), s.err
}

func (s *promiseStore) TryPromise(_ context.Context, ballot *desc.Ballot) (bool, *desc.Ballot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	accepted := compareBallots(ballot, s.promise) >= 0 && s.err == nil
	if accepted {
		s.promise = cloneBallot(ballot)
	}
	return accepted, cloneBallot(s.promise), s.err
}

func testCluster(t *testing.T) ([]*Node, []*grpc.Server) {
	t.Helper()
	ctx := context.WithValue(context.Background(), logger.LoggerKey, logger.New("election-test"))
	listeners := make([]net.Listener, 3)
	for i := range listeners {
		var err error
		listeners[i], err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { listeners[i].Close() })
	}
	nodes, servers := make([]*Node, 3), make([]*grpc.Server, 3)
	for i := range nodes {
		peers := []string{}
		for j, listener := range listeners {
			if i != j {
				peers = append(peers, listener.Addr().String())
			}
		}
		cfg := &config.Config{NodeID: []string{"node-1", "node-2", "node-3"}[i],
			NodeGRPCAddress: listeners[i].Addr().String(), PeerGRPCAddresses: peers, ROLE: config.REPLICA}
		store := &promiseStore{MapStorage: map_storage.New()}
		var err error
		nodes[i], err = New(ctx, cfg, store)
		if err != nil {
			t.Fatal(err)
		}
		server := grpc.NewServer()
		servers[i] = server
		desc.RegisterNodeServiceServer(server, nodes[i].grpcServer)
		done := make(chan struct{})
		go func() { defer close(done); _ = server.Serve(listeners[i]) }()
		t.Cleanup(func() { server.Stop(); <-done })
	}
	return nodes, servers
}

func TestPrepare(t *testing.T) {
	store := &promiseStore{MapStorage: map_storage.New()}
	n := &Node{id: "node-1", role: config.MASTER, consensusStore: store}
	ctx := context.Background()
	ballot := &desc.Ballot{Round: 7, NodeId: "node-2"}
	for _, tc := range []struct {
		ballot *desc.Ballot
		want   bool
	}{{ballot, true}, {cloneBallot(ballot), true}, {&desc.Ballot{Round: 6, NodeId: "node-3"}, false}} {
		reply, err := n.Prepare(ctx, &desc.PrepareRequest{Ballot: tc.ballot})
		if err != nil || reply.Promised != tc.want || compareBallots(reply.PromisedBallot, ballot) != 0 {
			t.Fatalf("unexpected Prepare response: %v, %v", reply, err)
		}
	}
	if n.Role() != config.REPLICA {
		t.Fatal("higher promise must demote the old master")
	}
	if saved, _ := store.LoadPromisedBallot(ctx); compareBallots(saved, ballot) != 0 {
		t.Fatal("positive response did not save the promise")
	}
	if err := n.AppendReplicaValue(ctx, "user", "key", "stale", &desc.Ballot{Round: 6, NodeId: "node-3"}); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("old replication was not rejected: %v", err)
	}
	store.err = errors.New("storage unavailable")
	n.role = config.MASTER
	if _, err := n.Prepare(ctx, &desc.PrepareRequest{Ballot: &desc.Ballot{Round: 8, NodeId: "node-2"}}); status.Code(err) != codes.Internal {
		t.Fatalf("persistence failure must not return a promise: %v", err)
	}
	if n.Role() != config.REPLICA || compareBallots(n.promisedBallot, ballot) != 0 {
		t.Fatal("failed persistence must demote without publishing a new promise")
	}
}

func TestElection(t *testing.T) {
	for _, name := range []string{"two of three", "no majority", "higher ballot"} {
		t.Run(name, func(t *testing.T) {
			nodes, servers := testCluster(t)
			n := nodes[0]
			if name == "higher ballot" {
				if _, err := nodes[1].Prepare(n.ctx, &desc.PrepareRequest{Ballot: &desc.Ballot{Round: 9, NodeId: "node-2"}}); err != nil {
					t.Fatal(err)
				}
			} else {
				servers[2].Stop()
				if name == "no majority" {
					servers[1].Stop()
				}
			}
			reply, err := n.StartElection(n.ctx, &desc.StartElectionRequest{})
			if err != nil {
				t.Fatal(err)
			}
			if reply.Elected != (name == "two of three") || (n.Role() == config.MASTER) != reply.Elected {
				t.Fatalf("wrong election result: %v, role %s", reply, n.Role())
			}
			if name == "higher ballot" {
				reply, err = n.StartElection(n.ctx, &desc.StartElectionRequest{})
				if err != nil || !reply.Elected || reply.Ballot.Round != 10 {
					t.Fatalf("retry must exceed observed round: %v, %v", reply, err)
				}
			}
			if reply.Elected && nodes[1].LeaderGRPCAddress() != n.ownGRPCAddress {
				t.Fatal("replica did not learn the elected leader")
			}
		})
	}
}

func TestHeartbeat(t *testing.T) {
	nodes, servers := testCluster(t)
	leader, follower := nodes[0], nodes[1]
	if reply, err := leader.StartElection(leader.ctx, &desc.StartElectionRequest{}); err != nil || !reply.Elected {
		t.Fatalf("initial election failed: %v, %v", reply, err)
	}
	follower.mu.Lock()
	follower.electionDeadline = time.Now().Add(-time.Second)
	follower.mu.Unlock()
	leader.mu.RLock()
	ballot := cloneBallot(leader.promisedBallot)
	leader.mu.RUnlock()
	leader.sendHeartbeat(leader.ctx, ballot)
	follower.mu.RLock()
	deadline := follower.electionDeadline
	follower.mu.RUnlock()
	if !deadline.After(time.Now()) {
		t.Fatal("heartbeat did not renew the follower timeout")
	}
	if reply, err := follower.startElection(follower.ctx, true); err != nil || reply.Elected {
		t.Fatalf("healthy leader must prevent automatic election: %v, %v", reply, err)
	}
	old := &desc.Ballot{Round: 1, NodeId: "aaa"}
	reply, err := follower.AnnounceLeader(follower.ctx, &desc.AnnounceLeaderRequest{Ballot: old, LeaderGrpcAddress: leader.ownGRPCAddress})
	follower.mu.RLock()
	unchanged := follower.electionDeadline.Equal(deadline)
	follower.mu.RUnlock()
	if err != nil || reply.Accepted || !unchanged {
		t.Fatalf("stale heartbeat must not reset the timeout: %v, %v", reply, err)
	}
	follower.mu.Lock()
	follower.electionDeadline = time.Now().Add(-time.Second)
	follower.mu.Unlock()
	ctx, cancel := context.WithCancel(follower.ctx)
	done := make(chan struct{})
	go func() { defer close(done); follower.runHeartbeat(ctx) }()
	t.Cleanup(func() { cancel(); <-done })
	until := time.Now().Add(5 * time.Second)
	for follower.Role() != config.MASTER && time.Now().Before(until) {
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	<-done
	if follower.Role() != config.MASTER || leader.Role() != config.REPLICA {
		t.Fatal("heartbeat timeout did not replace the leader")
	}
	servers[0].Stop()
	servers[2].Stop()
	follower.mu.Lock()
	follower.electionDeadline = time.Now().Add(-time.Second)
	ballot = cloneBallot(follower.promisedBallot)
	follower.mu.Unlock()
	follower.sendHeartbeat(follower.ctx, ballot)
	if follower.Role() != config.REPLICA {
		t.Fatal("leader kept its role without a heartbeat majority")
	}
}
