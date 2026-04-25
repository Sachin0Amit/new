package mesh

import (
	"crypto/sha256"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	tamperAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tamper_attempts_total",
		Help: "Total number of gossip messages with invalid signatures",
	})
)

// SignMessage appends an ED25519 signature to the gossip message.
func (n *Node) SignMessage(msg *GossipMessage) error {
	// Create payload for signing (NodeID + Timestamp + Payload)
	payload := fmt.Sprintf("%s%d%s", msg.NodeId, msg.Timestamp, string(msg.SummaryPayload))
	hash := sha256.Sum256([]byte(payload))
	msg.SummaryHash = hash[:]

	sig, err := n.Priv.Sign(hash[:])
	if err != nil {
		return err
	}
	msg.Signature = sig
	return nil
}

// VerifyMessage validates the ED25519 signature against the peer's public key.
func (n *Node) VerifyMessage(msg *GossipMessage) (bool, error) {
	pid, err := peer.Decode(msg.NodeId)
	if err != nil {
		return false, err
	}

	pubKey, err := pid.ExtractPublicKey()
	if err != nil {
		return false, err
	}

	payload := fmt.Sprintf("%s%d%s", msg.NodeId, msg.Timestamp, string(msg.SummaryPayload))
	hash := sha256.Sum256([]byte(payload))

	ok, err := pubKey.Verify(hash[:], msg.Signature)
	if !ok || err != nil {
		tamperAttempts.Inc()
		return false, err
	}

	return true, nil
}
