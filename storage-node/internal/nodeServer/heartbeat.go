package nodeServer

import (
	"context"
	"math/rand/v2"
	"time"

	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/config"
	transport "gitlab.com/adsfasdfdsf-group/key-value-service/internal/transport/grpc"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
)

const heartbeatTimeout = 5 * time.Second

func (n *Node) resetElectionTimeout() {
	n.electionDeadline = time.Now().Add(heartbeatTimeout + time.Duration(rand.Int64N(int64(3*time.Second))))
}

func (n *Node) runHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		n.mu.RLock()
		role, ballot, deadline := n.role, cloneBallot(n.promisedBallot), n.electionDeadline
		n.mu.RUnlock()
		if role == config.MASTER && validBallot(ballot) {
			n.sendHeartbeat(ctx, ballot)
		} else if role == config.MASTER || time.Now().After(deadline) {
			result, err := n.startElection(ctx, true)
			if err != nil && ctx.Err() == nil {
				logger.GetLogger(n.ctx).Warn(n.ctx, "Automatic election failed: "+err.Error())
			} else if result != nil && result.Elected {
				logger.GetLogger(n.ctx).Info(n.ctx, "Elected leader after heartbeat timeout")
			}
		}
	}
}

func (n *Node) sendHeartbeat(ctx context.Context, ballot *desc.Ballot) {
	n.proposalMu.Lock()
	defer n.proposalMu.Unlock()
	probeCtx, probeCancel := context.WithTimeout(ctx, time.Second)
	_, storageErr := n.repo.GetAll(probeCtx, "__health_probe__")
	probeCancel()
	if storageErr != nil {
		n.mu.Lock()
		n.role, n.leaderGRPCAddress, n.ready = config.REPLICA, "", false
		n.resetElectionTimeout()
		n.mu.Unlock()
		return
	}
	n.mu.RLock()
	committed := n.committedIndex
	n.mu.RUnlock()
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	replies := make(chan *desc.AnnounceLeaderResponse, len(n.peers))
	for _, address := range n.peers {
		go func(address string) {
			client := transport.Client{Address: address}
			reply, err := client.AnnounceLeader(ctx, ballot, n.ownGRPCAddress, committed)
			if err == nil && !reply.Accepted && compareBallots(reply.PromisedBallot, ballot) < 0 {
				if promise, prepareErr := client.Prepare(ctx, ballot); prepareErr == nil {
					n.observeBallot(promise.PromisedBallot)
					if promise.Promised {
						reply, err = client.AnnounceLeader(ctx, ballot, n.ownGRPCAddress, committed)
					}
				}
			}
			if err == nil && reply != nil && reply.Accepted {
				n.syncPeer(ctx, address, ballot, reply.CommittedIndex)
			}
			if err != nil {
				reply = nil
			}
			replies <- reply
		}(address)
	}
	confirmed := map[string]bool{n.id: true}
	for range n.peers {
		if reply := <-replies; reply != nil {
			n.observeBallot(reply.PromisedBallot)
			if reply.Accepted && reply.NodeId != "" && compareBallots(reply.PromisedBallot, ballot) == 0 {
				confirmed[reply.NodeId] = true
			}
		}
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.role != config.MASTER || compareBallots(n.promisedBallot, ballot) != 0 {
		return
	}
	if len(confirmed) >= 2 {
		n.electionDeadline = time.Now().Add(heartbeatTimeout)
	} else if time.Now().After(n.electionDeadline) {
		n.role, n.leaderGRPCAddress, n.ready = config.REPLICA, "", false
		n.resetElectionTimeout()
		logger.GetLogger(n.ctx).Warn(n.ctx, "Stepped down after losing heartbeat majority")
	}
}
