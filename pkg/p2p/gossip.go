package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/Sachin0Amit/new/internal/models"
	"github.com/Sachin0Amit/new/pkg/logger"
)

// GossipNode implements a lightweight synchronization protocol.
type GossipNode struct {
	discovery *DiscoveryEngine
	logger    logger.Logger
}

type GossipMessageType string

const (
	MsgHeartbeat      GossipMessageType = "HEARTBEAT"
	MsgTaskDelegation GossipMessageType = "TASK_DELEGATION"
	MsgKnowledgeSync   GossipMessageType = "KNOWLEDGE_SYNC"
	MsgRemoteSearch    GossipMessageType = "REMOTE_SEARCH"
)

// GossipStatus represents the health data shared between peers.
type GossipStatus struct {
	Type      GossipMessageType `json:"type"`
	Payload   interface{}       `json:"payload"`
	NodeID    string            `json:"node_id"`
	Load      float64           `json:"load"`
	TaskCount int               `json:"tasks"`
}

// KnowledgePacket represents a shared fragment of semantic memory.
type KnowledgePacket struct {
	ChunkID   string    `json:"chunk_id"`
	Vector    []float32 `json:"vector"`
	Summary   string    `json:"summary"`
	Timestamp int64     `json:"timestamp"`
}

// SearchRequest represents a federated query to the mesh.
type SearchRequest struct {
	QueryVec []float32 `json:"query_vec"`
	TopK     int       `json:"top_k"`
	RequestID string    `json:"request_id"`
}

// DelegationPacket carries a task state for remote execution.
type DelegationPacket struct {
	Task      *models.Task `json:"task"`
	Requester string       `json:"requester_id"`
}

// NewGossipNode creates a new Gossip state manager.
func NewGossipNode(discovery *DiscoveryEngine) *GossipNode {
	return &GossipNode{
		discovery: discovery,
		logger:    logger.New(),
	}
}

// Start initiates the periodic gossip synchronization.
func (g *GossipNode) Start(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.round()
		}
	}
}

func (g *GossipNode) round() {
	peers := g.discovery.GetPeers()
	if len(peers) == 0 {
		return
	}

	// Pick a random peer to gossip with
	target := peers[rand.Intn(len(peers))]
	
	g.logger.Debug("Initiating gossip round", logger.String("target", target.Address))
	
	err := g.syncWith(target)
	if err != nil {
		g.logger.Warn("Gossip sync failed", logger.String("peer", target.Address), logger.ErrorF(err))
	}
}

func (g *GossipNode) syncWith(peer *models.NodePeer) error {
	conn, err := net.DialTimeout("tcp", peer.Address, 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 1. Send our status
	status := GossipStatus{
		NodeID: g.discovery.nodeID.String(),
		Load:   0.25, // Mock load for now
	}
	
	encoder := json.NewEncoder(conn)
	return encoder.Encode(status)
}

// Listen starts a TCP server to receive gossip from other peers.
func (g *GossipNode) Listen(ctx context.Context, port int) error {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				continue
			}
		}
		go g.handleSync(conn)
	}
}

func (g *GossipNode) handleSync(conn net.Conn) {
	defer conn.Close()
	
	var status GossipStatus
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&status); err != nil {
		return
	}

	g.logger.Debug("Received gossip update", logger.String("node", status.NodeID), logger.Float64("load", status.Load))

	// If this was a task delegation, route it
	if status.Type == MsgTaskDelegation {
		var del DelegationPacket
		b, _ := json.Marshal(status.Payload)
		json.Unmarshal(b, &del)
		g.logger.Info("Received Task Delegation from mesh", logger.String("from", del.Requester), logger.String("task", del.Task.ID.String()))
		// In production, this would be pushed into an internal channel for the orchestrator
	}
}

// DelegateTask attempts to migrate a task to a remote peer.
func (g *GossipNode) DelegateTask(peer *models.NodePeer, task *models.Task) error {
	conn, err := net.DialTimeout("tcp", peer.Address, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	packet := GossipStatus{
		NodeID: g.discovery.nodeID.String(),
		Type:   MsgTaskDelegation,
		Payload: DelegationPacket{
			Task:      task,
			Requester: g.discovery.nodeID.String(),
		},
	}

	return json.NewEncoder(conn).Encode(packet)
}
