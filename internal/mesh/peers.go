package mesh

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// PeerInfo holds resource metrics and status of a mesh node.
type PeerInfo struct {
	ID            peer.ID
	LastSeen      time.Time
	ResourceScore float64 // CPU + Memory load factor (0.0 to 1.0)
	ActiveTasks   int
}

// PeerRegistry manages a thread-safe map of known peers.
type PeerRegistry struct {
	peers map[peer.ID]*PeerInfo
	mu    sync.RWMutex
}

func NewPeerRegistry() *PeerRegistry {
	return &PeerRegistry{
		peers: make(map[peer.ID]*PeerInfo),
	}
}

func (pr *PeerRegistry) UpdatePeer(info *PeerInfo) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.peers[info.ID] = info
}

// SelectBestPeer returns the peer with the lowest resource gravity.
func (pr *PeerRegistry) SelectBestPeer(taskComplexity float64) peer.ID {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	var bestPeer peer.ID
	lowestScore := 2.0 // Score is 0-1, so 2 is infinity

	for id, info := range pr.peers {
		// Ignore stale peers (> 2 min)
		if time.Since(info.LastSeen) > 2*time.Minute {
			continue
		}
		
		// Gravity Score = Load Factor + (ActiveTasks * 0.1)
		gravity := info.ResourceScore + (float64(info.ActiveTasks) * 0.1)
		if gravity < lowestScore {
			lowestScore = gravity
			bestPeer = id
		}
	}
	return bestPeer
}
