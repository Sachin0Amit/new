package mesh

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

const GossipProtocolID = "/sovereign/gossip/1.0.0"

// StartGossipLoop initializes the 30s lazy gossip cycle.
func (n *Node) StartGossipLoop(ctx context.Context) {
	// 1. Set stream handler
	n.Host.SetStreamHandler(GossipProtocolID, n.handleGossipStream)

	// 2. Start broadcast ticker
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				n.BroadcastGossip(ctx)
			}
		}
	}()
}

func (n *Node) BroadcastGossip(ctx context.Context) {
	msg := &GossipMessage{
		NodeId:           n.Host.ID().String(),
		Timestamp:        time.Now().Unix(),
		SummaryPayload:   []byte("Sovereign Node Summary: Healthy"),
	}

	// Sign the message
	if err := n.SignMessage(msg); err != nil {
		fmt.Printf("Gossip: Signing failed: %v\n", err)
		return
	}

	data, _ := proto.Marshal(msg)

	for _, pid := range n.Host.Network().Peers() {
		go func(p peer.ID) {
			s, err := n.Host.NewStream(ctx, p, GossipProtocolID)
			if err != nil {
				return
			}
			defer s.Close()
			s.Write(data)
		}(pid)
	}
}

func (n *Node) handleGossipStream(s network.Stream) {
	data, err := ioutil.ReadAll(s)
	if err != nil {
		s.Reset()
		return
	}
	defer s.Close()

	var msg GossipMessage
	if err := proto.Unmarshal(data, &msg); err != nil {
		return
	}

	// Verify signature
	if ok, err := n.VerifyMessage(&msg); !ok || err != nil {
		fmt.Printf("Gossip: Invalid signature from %s\n", msg.NodeId)
		// Increment Prometheus counter (handled in verify.go)
		return
	}

	fmt.Printf("Gossip: Received verified summary from %s: %s\n", msg.NodeId, string(msg.SummaryPayload))
	
	// Update registry
	pid, _ := peer.Decode(msg.NodeId)
	n.registry.UpdatePeer(&PeerInfo{
		ID:            pid,
		LastSeen:      time.Now(),
		ResourceScore: 0.1, // Mock score
	})
}
