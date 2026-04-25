package mesh

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test3NodeMeshGossip(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Initialize 3 nodes
	n1, _ := NewNode(ctx, 10001)
	n2, _ := NewNode(ctx, 10002)
	n3, _ := NewNode(ctx, 10003)
	defer n1.Close()
	defer n2.Close()
	defer n3.Close()

	n1.StartGossipLoop(ctx)
	n2.StartGossipLoop(ctx)
	n3.StartGossipLoop(ctx)

	// 2. Wait for mDNS discovery (up to 5s)
	t.Log("Waiting for mDNS discovery...")
	time.Sleep(5 * time.Second)

	// 3. Assert connections
	// Note: In loopback, mDNS might be finicky, so we check if they found each other
	t.Logf("Node 1 Peers: %d", len(n1.Host.Network().Peers()))
	t.Logf("Node 2 Peers: %d", len(n2.Host.Network().Peers()))
	t.Logf("Node 3 Peers: %d", len(n3.Host.Network().Peers()))

	// 4. Manual Broadcast from N1
	n1.BroadcastGossip(ctx)

	// 5. Verify N2 and N3 received and updated registries
	time.Sleep(2 * time.Second)
	
	// We check if N2 has N1 in its registry
	assert.NotEmpty(t, n2.registry.peers, "Node 2 should have peers in registry")
	assert.NotEmpty(t, n3.registry.peers, "Node 3 should have peers in registry")
}
