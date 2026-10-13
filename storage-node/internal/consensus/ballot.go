package consensus

import desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"

// Compare returns -1, 0, or 1 when a is smaller, equal, or greater than b.
// A nil ballot means no promise has been made and precedes any non-nil ballot.
func Compare(a, b *desc.Ballot) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}
	if a.Round < b.Round {
		return -1
	}
	if a.Round > b.Round {
		return 1
	}
	if a.NodeId < b.NodeId {
		return -1
	}
	if a.NodeId > b.NodeId {
		return 1
	}
	return 0
}
