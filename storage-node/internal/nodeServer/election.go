package nodeServer

import (
	"context"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/config"
	transport "gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/grpc"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validateGRPCAddress(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil || host == "" {
		return fmt.Errorf("gRPC address must be host:port")
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return fmt.Errorf("invalid gRPC port")
	}
	return nil
}

func validBallot(ballot *desc.Ballot) bool {
	return ballot != nil && ballot.Round > 0 && strings.TrimSpace(ballot.NodeId) != ""
}

func (n *Node) Prepare(ctx context.Context, req *desc.PrepareRequest) (*desc.PrepareResponse, error) {
	if req == nil || !validBallot(req.Ballot) {
		return nil, status.Error(codes.InvalidArgument, "valid ballot required")
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	accepted, current, err := n.consensusStore.TryPromise(ctx, req.Ballot)
	if err != nil {
		if compareBallots(req.Ballot, n.promisedBallot) > 0 {
			n.role = config.REPLICA
			n.ready = false
			n.leaderGRPCAddress = ""
		}
		return nil, status.Error(codes.Internal, "could not persist promise")
	}
	if compareBallots(current, n.promisedBallot) > 0 {
		n.role = config.REPLICA
		n.ready = false
		n.leaderGRPCAddress = ""
		n.resetElectionTimeout()
	}
	n.promisedBallot = cloneBallot(current)
	if current != nil && current.Round > n.observedRound {
		n.observedRound = current.Round
	}
	reply := &desc.PrepareResponse{Promised: accepted, PromisedBallot: cloneBallot(current), NodeId: n.id}
	if accepted {
		if err := n.reloadJournal(ctx); err != nil {
			return nil, status.Error(codes.Internal, "read journal after promise")
		}
		reply.Entries, reply.CommittedIndex = n.entries, n.committedIndex
		if len(n.entries) == 0 {
			records, err := n.journalStore.ReadRecords(ctx)
			if err != nil {
				return nil, status.Error(codes.Internal, "read bootstrap state")
			}
			reply.Records = records
		}
	}
	return reply, nil
}

func (n *Node) AnnounceLeader(ctx context.Context, req *desc.AnnounceLeaderRequest) (*desc.AnnounceLeaderResponse, error) {
	if req == nil || !validBallot(req.Ballot) {
		return nil, status.Error(codes.InvalidArgument, "valid ballot required")
	}
	if err := validateGRPCAddress(req.LeaderGrpcAddress); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	accepted := compareBallots(req.Ballot, n.promisedBallot) == 0
	if req.Ballot.NodeId == n.id {
		accepted = false
	}
	known := false
	for _, peer := range n.peers {
		if peer == req.LeaderGrpcAddress {
			known = true
		}
	}
	accepted = accepted && known
	if accepted {
		n.role = config.REPLICA
		n.ready = false
		n.leaderGRPCAddress = req.LeaderGrpcAddress
		n.leaderCommitIndex = req.CommittedIndex
		n.ready = n.committedIndex >= req.CommittedIndex
		n.resetElectionTimeout()
	}
	return &desc.AnnounceLeaderResponse{Accepted: accepted, PromisedBallot: cloneBallot(n.promisedBallot), NodeId: n.id, CommittedIndex: n.committedIndex}, nil
}

func (n *Node) observeBallot(ballot *desc.Ballot) {
	if !validBallot(ballot) {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if ballot.Round > n.observedRound {
		n.observedRound = ballot.Round
	}
	if compareBallots(ballot, n.promisedBallot) > 0 {
		n.role = config.REPLICA
		n.ready = false
		n.leaderGRPCAddress = ""
	}
}

func (n *Node) StartElection(ctx context.Context, _ *desc.StartElectionRequest) (*desc.StartElectionResponse, error) {
	return n.startElection(ctx, false)
}

func (n *Node) startElection(ctx context.Context, automatic bool) (*desc.StartElectionResponse, error) {
	n.electionMu.Lock()
	defer n.electionMu.Unlock()
	n.proposalMu.Lock()
	defer n.proposalMu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	n.mu.Lock()
	if automatic && ((n.role == config.MASTER && validBallot(n.promisedBallot)) ||
		(n.role != config.MASTER && time.Now().Before(n.electionDeadline))) {
		n.mu.Unlock()
		return &desc.StartElectionResponse{}, nil
	}
	n.resetElectionTimeout()
	round := n.observedRound
	if n.promisedBallot != nil && n.promisedBallot.Round > round {
		round = n.promisedBallot.Round
	}
	n.mu.Unlock()
	if round == math.MaxUint64 {
		return nil, status.Error(codes.ResourceExhausted, "ballot round exhausted")
	}
	ballot := &desc.Ballot{Round: round + 1, NodeId: n.id}
	self, err := n.Prepare(ctx, &desc.PrepareRequest{Ballot: ballot})
	if err != nil {
		return nil, err
	}
	result := &desc.StartElectionResponse{Ballot: cloneBallot(ballot)}
	if !self.Promised {
		return result, nil
	}
	type promiseResult struct {
		reply *desc.PrepareResponse
		err   error
	}
	promises := make(chan promiseResult, len(n.peers))
	for _, address := range n.peers {
		go func(address string) {
			client := transport.Client{Address: address}
			reply, err := client.Prepare(ctx, ballot)
			promises <- promiseResult{reply, err}
		}(address)
	}
	voters := map[string]bool{n.id: true}
	recovery := []*desc.PrepareResponse{self}
	higher := false
	for range n.peers {
		entry := <-promises
		if entry.err != nil || entry.reply == nil {
			continue
		}
		reply := entry.reply
		n.observeBallot(reply.PromisedBallot)
		if compareBallots(reply.PromisedBallot, ballot) > 0 {
			higher = true
		}
		if reply.Promised && compareBallots(reply.PromisedBallot, ballot) == 0 && strings.TrimSpace(reply.NodeId) != "" {
			if !voters[reply.NodeId] {
				recovery = append(recovery, reply)
			}
			voters[reply.NodeId] = true
		}
	}
	if len(voters) < 2 || higher {
		return result, nil
	}
	entries, recoveryErr := recoverEntries(ballot, recovery)
	if recoveryErr != nil {
		return nil, status.Error(codes.FailedPrecondition, recoveryErr.Error())
	}
	if err := n.chooseEntries(ctx, ballot, entries, true, uint64(len(entries))); err != nil {
		return result, err
	}
	type announceResult struct {
		reply *desc.AnnounceLeaderResponse
		err   error
	}
	announcements := make(chan announceResult, len(n.peers))
	for _, address := range n.peers {
		go func(address string) {
			client := transport.Client{Address: address}
			reply, err := client.AnnounceLeader(ctx, ballot, n.ownGRPCAddress, uint64(len(entries)))
			announcements <- announceResult{reply, err}
		}(address)
	}
	confirmed := map[string]bool{n.id: true}
	for range n.peers {
		entry := <-announcements
		if entry.err != nil || entry.reply == nil {
			continue
		}
		reply := entry.reply
		n.observeBallot(reply.PromisedBallot)
		if compareBallots(reply.PromisedBallot, ballot) > 0 {
			higher = true
		}
		if reply.Accepted && compareBallots(reply.PromisedBallot, ballot) == 0 && voters[reply.NodeId] {
			confirmed[reply.NodeId] = true
		}
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}
	if len(confirmed) < 2 || higher || compareBallots(n.promisedBallot, ballot) != 0 {
		return result, nil
	}
	n.role = config.MASTER
	n.ready = true
	n.leaderCommitIndex = n.committedIndex
	n.leaderGRPCAddress = ""
	n.resetElectionTimeout()
	result.Elected = true
	return result, nil
}
