package nodeServer

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/config"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
	"google.golang.org/grpc"
)

func restartTestNode(t *testing.T, n *Node) *grpc.Server {
	t.Helper()
	listener, err := net.Listen("tcp", n.ownGRPCAddress)
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	desc.RegisterNodeServiceServer(server, n.grpcServer)
	done := make(chan struct{})
	go func() { defer close(done); _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); <-done })
	return server
}

func TestQuorumAndCatchup(t *testing.T) {
	nodes, servers := testCluster(t)
	leader := nodes[0]
	if result, err := leader.StartElection(leader.ctx, &desc.StartElectionRequest{}); err != nil || !result.Elected {
		t.Fatalf("election: %v %v", result, err)
	}
	servers[2].Stop()
	if err := leader.submitValue(leader.ctx, "user", "key", "confirmed"); err != nil {
		t.Fatal(err)
	}
	if _, err := nodes[2].repo.Get(leader.ctx, "user", "key"); err == nil {
		t.Fatal("offline node unexpectedly received write")
	}
	restartTestNode(t, nodes[2])
	servers[0].Stop()
	recovered, err := nodes[2].StartElection(leader.ctx, &desc.StartElectionRequest{})
	if err != nil || !recovered.Elected {
		t.Fatalf("lagging candidate failed recovery: %v %v", recovered, err)
	}
	value, err := nodes[2].repo.Get(leader.ctx, "user", "key")
	if err != nil || value != "confirmed" {
		t.Fatalf("confirmed write lost after leader failure: %q %v", value, err)
	}
	servers[1].Stop()
	if err := nodes[2].submitValue(leader.ctx, "user", "isolated", "not-confirmed"); err == nil {
		t.Fatal("isolated leader acknowledged a write")
	}
	if _, err := nodes[2].repo.Get(leader.ctx, "user", "isolated"); err == nil {
		t.Fatal("unconfirmed write became visible")
	}
	if nodes[2].Role() == config.MASTER {
		t.Fatal("isolated writer did not step down")
	}
	restartTestNode(t, nodes[1])
	result, err := nodes[1].StartElection(leader.ctx, &desc.StartElectionRequest{})
	if err != nil || !result.Elected {
		t.Fatalf("recovery election failed: %v %v", result, err)
	}
	value, err = nodes[1].repo.Get(leader.ctx, "user", "key")
	if err != nil || value != "confirmed" {
		t.Fatalf("recovery lost committed value: %q %v", value, err)
	}
}

func TestConcurrentWritesAndDiff(t *testing.T) {
	nodes, servers := testCluster(t)
	leader := nodes[0]
	if result, err := leader.StartElection(leader.ctx, &desc.StartElectionRequest{}); err != nil || !result.Elected {
		t.Fatalf("election: %v %v", result, err)
	}
	servers[2].Stop()
	var group sync.WaitGroup
	failures := make(chan error, 8)
	for i := 0; i < 8; i++ {
		group.Go(func() {
			failures <- leader.submitValue(context.Background(), "user", fmt.Sprintf("key-%d", i), "value")
		})
	}
	group.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	restartTestNode(t, nodes[2])
	leader.mu.RLock()
	ballot := cloneBallot(leader.promisedBallot)
	leader.mu.RUnlock()
	leader.sendHeartbeat(leader.ctx, ballot)
	values, err := nodes[2].repo.GetAll(leader.ctx, "user")
	if err != nil || len(values) != 8 {
		t.Fatalf("diff did not restore missing writes: %d %v", len(values), err)
	}
	if nodes[2].committedIndex != leader.committedIndex || !nodes[2].readAvailable() {
		t.Fatal("replica did not become ready after catchup")
	}
}

func TestConcurrentCandidates(t *testing.T) {
	nodes, _ := testCluster(t)
	done := make(chan struct{}, 2)
	for _, node := range nodes[:2] {
		go func(n *Node) { _, _ = n.StartElection(n.ctx, &desc.StartElectionRequest{}); done <- struct{}{} }(node)
	}
	<-done
	<-done
	var elected *Node
	count := 0
	for _, node := range nodes {
		if node.Role() == config.MASTER {
			elected = node
			count++
		}
	}
	if count > 1 {
		t.Fatal("connected candidates retained conflicting leadership")
	}
	if elected == nil {
		for attempt := 0; attempt < 3; attempt++ {
			result, err := nodes[2].StartElection(nodes[2].ctx, &desc.StartElectionRequest{})
			if err == nil && result.Elected {
				elected = nodes[2]
				break
			}
		}
	}
	if elected == nil {
		t.Fatal("cluster failed to elect a leader after contention")
	}
	if err := elected.submitValue(elected.ctx, "user", "after-contention", "confirmed"); err != nil {
		t.Fatal(err)
	}
	for _, node := range nodes {
		value, err := node.repo.Get(node.ctx, "user", "after-contention")
		if err != nil || value != "confirmed" {
			t.Fatalf("inconsistent state after contention: %s %v", value, err)
		}
	}
}
