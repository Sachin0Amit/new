package p2p

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/papi-ai/sovereign-core/internal/models"
	"github.com/papi-ai/sovereign-core/pkg/logger"
)

const (
	MulticastAddr     = "239.0.0.1:9999"
	DiscoveryInterval = 5 * time.Second
)

// DiscoveryEngine manages peer-to-peer node discovery via UDP multicast.
type DiscoveryEngine struct {
	nodeID   uuid.UUID
	logger   logger.Logger
	peers    sync.Map // Map[uuid.UUID]*models.NodePeer
	port     int
}

// NewDiscoveryEngine creates a new P2P discovery manager.
func NewDiscoveryEngine(nodeID uuid.UUID, port int) *DiscoveryEngine {
	return &DiscoveryEngine{
		nodeID: nodeID,
		logger: logger.New(),
		port:   port,
	}
}

// Start initiates the discovery broadcasting and listening loops.
func (d *DiscoveryEngine) Start(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp4", MulticastAddr)
	if err != nil {
		return err
	}

	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		return err
	}

	d.logger.Info("Sovereign Discovery Engine active", logger.String("addr", MulticastAddr))

	go d.broadcastLoop(ctx, addr)
	go d.listenLoop(ctx, conn)

	return nil
}

func (d *DiscoveryEngine) broadcastLoop(ctx context.Context, addr *net.UDPAddr) {
	ticker := time.NewTicker(DiscoveryInterval)
	defer ticker.Stop()

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		d.logger.Error("Failed to dial multicast", logger.ErrorF(err))
		return
	}
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Announcement: ID|PORT
			msg := fmt.Sprintf("%s|%d", d.nodeID.String(), d.port)
			_, err = conn.Write([]byte(msg))
			if err != nil {
				d.logger.Warn("Multicast broadcast failed", logger.ErrorF(err))
			}
		}
	}
}

func (d *DiscoveryEngine) listenLoop(ctx context.Context, conn *net.UDPConn) {
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn.SetReadDeadline(time.Now().Add(time.Second))
			n, src, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}

			d.handleMessage(src.IP.String(), string(buf[:n]))
		}
	}
}

func (d *DiscoveryEngine) handleMessage(ip, msg string) {
	// Parse ID|PORT
	var idStr string
	var port int
	n, err := fmt.Sscanf(msg, "%s|%d", &idStr, &port)
	if err != nil || n != 2 {
		return
	}

	peerID, err := uuid.Parse(idStr)
	if err != nil || peerID == d.nodeID {
		return
	}

	address := fmt.Sprintf("%s:%d", ip, port)
	
	if val, ok := d.peers.Load(peerID); ok {
		peer := val.(*models.NodePeer)
		peer.LastSeen = time.Now()
		peer.IsActive = true
	} else {
		d.logger.Info("Discovered new Sovereign peer", logger.String("id", idStr), logger.String("addr", address))
		d.peers.Store(peerID, &models.NodePeer{
			ID:       peerID,
			Address:  address,
			LastSeen: time.Now(),
			IsActive: true,
		})
	}
}

// GetPeers returns a list of all active discovered peers.
func (d *DiscoveryEngine) GetPeers() []*models.NodePeer {
	var peers []*models.NodePeer
	d.peers.Range(func(key, value interface{}) bool {
		peer := value.(*models.NodePeer)
		if time.Since(peer.LastSeen) < DiscoveryInterval*3 {
			peers = append(peers, peer)
		} else {
			peer.IsActive = false
		}
		return true
	})
	return peers
}
