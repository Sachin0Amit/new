package mesh

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Node represents a libp2p host in the Sovereign mesh.
type Node struct {
	Host host.Host
	Priv crypto.PrivKey
	registry *PeerRegistry
}

// discoveryNotifee is used to handle mDNS peer discovery events.
type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("mDNS: Peer found: %s\n", pi.ID.String())
	if err := n.h.Connect(context.Background(), pi); err != nil {
		fmt.Printf("mDNS: Connection failed: %v\n", err)
	}
}

// NewNode initializes a libp2p host with Noise security and Yamux multiplexing.
func NewNode(ctx context.Context, listenPort int) (*Node, error) {
	// Generate ED25519 keypair
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		return nil, err
	}

	h, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
		libp2p.DefaultTransports,
		libp2p.DefaultSecurity,
		libp2p.DefaultMuxers,
	)
	if err != nil {
		return nil, err
	}

	node := &Node{
		Host: h,
		Priv: priv,
		registry: NewPeerRegistry(),
	}
	return node, nil
}

// Start initializes background tasks like gossip loops and discovery.
func (n *Node) Start(ctx context.Context) error {
	// Initialize mDNS discovery
	dn := &discoveryNotifee{h: n.Host}
	ser := mdns.NewMdnsService(n.Host, "sovereign-mesh", dn)
	if err := ser.Start(); err != nil {
		return err
	}
	fmt.Println("mDNS discovery started")
	return nil
}

func (n *Node) Close() error {
	return n.Host.Close()
}
