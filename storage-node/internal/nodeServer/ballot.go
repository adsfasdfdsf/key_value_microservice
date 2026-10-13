package nodeServer

import (
	"gitlab.com/adsfasdfdsf-group/key-value-service/internal/consensus"
	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
)

func compareBallots(a, b *desc.Ballot) int { return consensus.Compare(a, b) }

func cloneBallot(ballot *desc.Ballot) *desc.Ballot {
	if ballot == nil {
		return nil
	}
	return &desc.Ballot{Round: ballot.Round, NodeId: ballot.NodeId}
}
