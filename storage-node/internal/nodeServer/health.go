package nodeServer

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (n *Node) readAvailable() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.ready
}

func (n *Node) health(w http.ResponseWriter, r *http.Request) {
	n.mu.RLock()
	result := struct {
		NodeID    string `json:"node_id"`
		Role      string `json:"role"`
		Leader    string `json:"leader"`
		Ballot    any    `json:"ballot"`
		Committed uint64 `json:"committed_index"`
		Last      uint64 `json:"last_index"`
		Ready     bool   `json:"ready"`
	}{n.id, string(n.role), n.leaderGRPCAddress, cloneBallot(n.promisedBallot), n.committedIndex, uint64(len(n.entries)), n.ready}
	n.mu.RUnlock()
	code := http.StatusOK
	if r.URL.Path != "/health/live" {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if _, err := n.repo.GetAll(ctx, "__health_probe__"); err != nil {
			result.Ready = false
		}
		if !result.Ready && r.URL.Path != "/api/v1/status" {
			code = http.StatusServiceUnavailable
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(result)
}
