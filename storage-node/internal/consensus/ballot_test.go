package consensus

import (
	"testing"

	desc "gitlab.com/adsfasdfdsf-group/key-value-service/pkg/node_v1"
)

func TestCompare(t *testing.T) {
	a := &desc.Ballot{Round: 5, NodeId: "node-1"}
	for _, tc := range []struct {
		name string
		a, b *desc.Ballot
		want int
	}{
		{"no promises", nil, nil, 0},
		{"first promise", a, nil, 1},
		{"same ballot", a, &desc.Ballot{Round: 5, NodeId: "node-1"}, 0},
		{"round takes priority", &desc.Ballot{Round: 6, NodeId: "node-0"}, a, 1},
		{"node breaks tie", &desc.Ballot{Round: 5, NodeId: "node-2"}, a, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Compare(tc.a, tc.b); got != tc.want {
				t.Fatalf("Compare = %d, want %d", got, tc.want)
			}
			if got := Compare(tc.b, tc.a); got != -tc.want {
				t.Fatalf("reverse Compare = %d, want %d", got, -tc.want)
			}
		})
	}
}
