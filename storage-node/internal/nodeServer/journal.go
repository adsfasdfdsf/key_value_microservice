package nodeServer

import (
	"context"
	"fmt"
	"sort"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/config"
	transport "gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/grpc"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (n *Node) reloadJournal(ctx context.Context) error {
	entries, committed, err := n.journalStore.ReadJournal(ctx)
	if err == nil {
		n.entries, n.committedIndex = entries, committed
	}
	return err
}

func (n *Node) journalReply(accepted bool) *desc.JournalResponse {
	return &desc.JournalResponse{Accepted: accepted, PromisedBallot: cloneBallot(n.promisedBallot), NodeId: n.id, CommittedIndex: n.committedIndex}
}

func (n *Node) AcceptEntries(ctx context.Context, req *desc.AcceptEntriesRequest) (*desc.JournalResponse, error) {
	if req == nil || !validBallot(req.Ballot) {
		return nil, status.Error(codes.InvalidArgument, "valid ballot required")
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if compareBallots(req.Ballot, n.promisedBallot) != 0 {
		return n.journalReply(false), nil
	}
	if err := n.journalStore.AcceptEntries(ctx, req); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	if err := n.reloadJournal(ctx); err != nil {
		n.ready = false
		return nil, status.Error(codes.Internal, "load accepted journal")
	}
	return n.journalReply(true), nil
}

func (n *Node) CommitEntries(ctx context.Context, req *desc.CommitEntriesRequest) (*desc.JournalResponse, error) {
	if req == nil || !validBallot(req.Ballot) {
		return nil, status.Error(codes.InvalidArgument, "valid ballot required")
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if compareBallots(req.Ballot, n.promisedBallot) != 0 {
		return n.journalReply(false), nil
	}
	if err := n.journalStore.CommitEntries(ctx, req); err != nil {
		n.ready = false
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	if err := n.reloadJournal(ctx); err != nil {
		n.ready = false
		return nil, status.Error(codes.Internal, "load committed journal")
	}
	n.ready = (n.role == config.MASTER || n.leaderGRPCAddress != "") && n.committedIndex >= n.leaderCommitIndex
	return n.journalReply(true), nil
}

func (n *Node) chooseEntries(ctx context.Context, ballot *desc.Ballot, entries []*desc.LogEntry, replace bool, last uint64) error {
	req := &desc.AcceptEntriesRequest{Ballot: ballot, Entries: entries, ReplaceTail: replace, LastIndex: last}
	self, err := n.AcceptEntries(ctx, req)
	if err != nil {
		return err
	}
	if !self.Accepted {
		return status.Error(codes.Unavailable, "proposal superseded")
	}
	type result struct {
		response *desc.JournalResponse
		err      error
		address  string
	}
	replies := make(chan result, len(n.peers))
	for _, address := range n.peers {
		go func(address string) {
			client := transport.Client{Address: address}
			response, err := client.AcceptEntries(ctx, req)
			replies <- result{response, err, address}
		}(address)
	}
	voters := map[string]bool{n.id: true}
	accepted := []string{}
	higher := false
	for range n.peers {
		reply := <-replies
		if reply.err != nil || reply.response == nil {
			continue
		}
		response := reply.response
		n.observeBallot(response.PromisedBallot)
		if compareBallots(response.PromisedBallot, ballot) > 0 {
			higher = true
		}
		if response.Accepted && compareBallots(response.PromisedBallot, ballot) == 0 && response.NodeId != "" {
			voters[response.NodeId] = true
			accepted = append(accepted, reply.address)
		}
	}
	if len(voters) < 2 || higher {
		return status.Error(codes.Unavailable, "write requires two durable acceptances")
	}
	commit := &desc.CommitEntriesRequest{Ballot: ballot, ThroughIndex: last}
	local, err := n.CommitEntries(ctx, commit)
	if err != nil {
		return err
	}
	if !local.Accepted {
		return status.Error(codes.Unavailable, "commit superseded")
	}
	done := make(chan struct{}, len(accepted))
	for _, address := range accepted {
		go func(address string) {
			client := transport.Client{Address: address}
			response, err := client.CommitEntries(ctx, commit)
			if err == nil && response != nil {
				n.observeBallot(response.PromisedBallot)
			}
			done <- struct{}{}
		}(address)
	}
	for range accepted {
		<-done
	}
	n.mu.Lock()
	if compareBallots(n.promisedBallot, ballot) == 0 && n.role == config.MASTER {
		n.ready = true
	}
	n.mu.Unlock()
	return nil
}

func recoverEntries(ballot *desc.Ballot, replies []*desc.PrepareResponse) ([]*desc.LogEntry, error) {
	selected := map[uint64]*desc.LogEntry{}
	var last uint64
	for _, reply := range replies {
		for _, entry := range reply.Entries {
			if entry == nil || entry.Index == 0 || !validBallot(entry.Ballot) {
				return nil, fmt.Errorf("invalid accepted journal")
			}
			if entry.Index > last {
				last = entry.Index
			}
			old := selected[entry.Index]
			if old == nil || compareBallots(entry.Ballot, old.Ballot) > 0 {
				selected[entry.Index] = entry
			} else if compareBallots(entry.Ballot, old.Ballot) == 0 {
				a, b := proto.Clone(old).(*desc.LogEntry), proto.Clone(entry).(*desc.LogEntry)
				a.Ballot, b.Ballot = nil, nil
				if !proto.Equal(a, b) {
					return nil, fmt.Errorf("conflicting accepted values")
				}
			}
		}
	}
	if last == 0 {
		records := map[string]map[string]*desc.Record{}
		for _, reply := range replies {
			for _, record := range reply.Records {
				if records[record.User] == nil {
					records[record.User] = map[string]*desc.Record{}
				}
				if old := records[record.User][record.Key]; old != nil && old.Value != record.Value {
					return nil, fmt.Errorf("legacy databases disagree for user %q key %q; reconcile before bootstrap", record.User, record.Key)
				}
				records[record.User][record.Key] = record
			}
		}
		seed := []*desc.Record{}
		for _, values := range records {
			for _, record := range values {
				seed = append(seed, proto.Clone(record).(*desc.Record))
			}
		}
		sort.Slice(seed, func(i, j int) bool {
			if seed[i].User == seed[j].User {
				return seed[i].Key < seed[j].Key
			}
			return seed[i].User < seed[j].User
		})
		return []*desc.LogEntry{{Index: 1, Ballot: cloneBallot(ballot), Snapshot: true, Records: seed}}, nil
	}
	entries := make([]*desc.LogEntry, 0, last)
	for index := uint64(1); index <= last; index++ {
		entry := selected[index]
		if entry == nil {
			entry = &desc.LogEntry{Index: index, Noop: true}
		} else {
			entry = proto.Clone(entry).(*desc.LogEntry)
		}
		entry.Ballot = cloneBallot(ballot)
		entries = append(entries, entry)
	}
	return entries, nil
}

func (n *Node) syncPeer(ctx context.Context, address string, ballot *desc.Ballot, committed uint64) {
	n.mu.RLock()
	target := n.committedIndex
	entries := []*desc.LogEntry{}
	for _, entry := range n.entries {
		if entry.Index > committed && entry.Index <= target {
			copy := proto.Clone(entry).(*desc.LogEntry)
			copy.Ballot = cloneBallot(ballot)
			entries = append(entries, copy)
		}
	}
	n.mu.RUnlock()
	if committed > target {
		return
	}
	client := transport.Client{Address: address}
	reply, err := client.AcceptEntries(ctx, &desc.AcceptEntriesRequest{Ballot: ballot, Entries: entries, ReplaceTail: true, LastIndex: target})
	if err != nil {
		return
	}
	n.observeBallot(reply.PromisedBallot)
	if !reply.Accepted {
		return
	}
	response, err := client.CommitEntries(ctx, &desc.CommitEntriesRequest{Ballot: ballot, ThroughIndex: target})
	if err == nil {
		n.observeBallot(response.PromisedBallot)
	}
}
