package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// NodeType categorizes entities in the knowledge graph.
type NodeType string

const (
	NodeTypeConcept  NodeType = "CONCEPT"
	NodeTypeEntity   NodeType = "ENTITY"
	NodeTypeFact     NodeType = "FACT"
	NodeTypeRelation NodeType = "RELATION"
	NodeTypeQuery    NodeType = "QUERY"
)

// EdgeType defines the semantic relationship between two nodes.
type EdgeType string

const (
	EdgeIsA        EdgeType = "IS_A"
	EdgeHasPart    EdgeType = "HAS_PART"
	EdgeRelatedTo  EdgeType = "RELATED_TO"
	EdgeCauses     EdgeType = "CAUSES"
	EdgeDerivedFrom EdgeType = "DERIVED_FROM"
	EdgeUsedIn     EdgeType = "USED_IN"
	EdgeContains   EdgeType = "CONTAINS"
	EdgeDependsOn  EdgeType = "DEPENDS_ON"
)

// Node represents a vertex in the knowledge graph.
type Node struct {
	ID         string                 `json:"id"`
	Label      string                 `json:"label"`
	Type       NodeType               `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Embedding  []float32              `json:"embedding,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	AccessCount int                   `json:"access_count"`
}

// Edge represents a directed relationship between two nodes.
type Edge struct {
	ID         string                 `json:"id"`
	Source     string                 `json:"source"`
	Target     string                 `json:"target"`
	Type       EdgeType               `json:"type"`
	Weight     float64                `json:"weight"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Graph implements a thread-safe in-memory knowledge graph with traversal
// and semantic search capabilities.
type Graph struct {
	nodes    map[string]*Node
	edges    map[string]*Edge
	outEdges map[string][]*Edge // source ID -> edges
	inEdges  map[string][]*Edge // target ID -> edges
	labelIdx map[string]string  // lowercase label -> node ID
	mu       sync.RWMutex
}

// NewGraph creates a new empty knowledge graph.
func NewGraph() *Graph {
	return &Graph{
		nodes:    make(map[string]*Node),
		edges:    make(map[string]*Edge),
		outEdges: make(map[string][]*Edge),
		inEdges:  make(map[string][]*Edge),
		labelIdx: make(map[string]string),
	}
}

// AddNode inserts a new node or returns the existing one if the label matches.
func (g *Graph) AddNode(label string, nodeType NodeType, props map[string]interface{}) *Node {
	g.mu.Lock()
	defer g.mu.Unlock()

	key := strings.ToLower(strings.TrimSpace(label))
	if existingID, ok := g.labelIdx[key]; ok {
		n := g.nodes[existingID]
		n.AccessCount++
		n.UpdatedAt = time.Now()
		return n
	}

	n := &Node{
		ID:         uuid.New().String(),
		Label:      label,
		Type:       nodeType,
		Properties: props,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	g.nodes[n.ID] = n
	g.labelIdx[key] = n.ID
	return n
}

// AddEdge creates a directed edge between two nodes.
func (g *Graph) AddEdge(sourceID, targetID string, edgeType EdgeType, weight float64) (*Edge, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, ok := g.nodes[sourceID]; !ok {
		return nil, fmt.Errorf("source node %s not found", sourceID)
	}
	if _, ok := g.nodes[targetID]; !ok {
		return nil, fmt.Errorf("target node %s not found", targetID)
	}

	// Prevent duplicate edges
	for _, e := range g.outEdges[sourceID] {
		if e.Target == targetID && e.Type == edgeType {
			e.Weight = (e.Weight + weight) / 2 // Rolling average
			return e, nil
		}
	}

	e := &Edge{
		ID:        uuid.New().String(),
		Source:    sourceID,
		Target:   targetID,
		Type:     edgeType,
		Weight:   weight,
		CreatedAt: time.Now(),
	}
	g.edges[e.ID] = e
	g.outEdges[sourceID] = append(g.outEdges[sourceID], e)
	g.inEdges[targetID] = append(g.inEdges[targetID], e)
	return e, nil
}

// GetNode retrieves a node by ID.
func (g *Graph) GetNode(id string) *Node {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.nodes[id]
}

// FindByLabel looks up a node by its label (case-insensitive).
func (g *Graph) FindByLabel(label string) *Node {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if id, ok := g.labelIdx[strings.ToLower(strings.TrimSpace(label))]; ok {
		return g.nodes[id]
	}
	return nil
}

// Neighbors returns all nodes connected to the given node (in either direction).
func (g *Graph) Neighbors(nodeID string, depth int) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visited := make(map[string]bool)
	result := make([]*Node, 0)
	g.bfs(nodeID, depth, visited, &result)
	return result
}

func (g *Graph) bfs(startID string, maxDepth int, visited map[string]bool, result *[]*Node) {
	type item struct {
		id    string
		depth int
	}
	queue := []item{{startID, 0}}
	visited[startID] = true

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if cur.depth > 0 {
			if n, ok := g.nodes[cur.id]; ok {
				*result = append(*result, n)
			}
		}

		if cur.depth >= maxDepth {
			continue
		}

		for _, e := range g.outEdges[cur.id] {
			if !visited[e.Target] {
				visited[e.Target] = true
				queue = append(queue, item{e.Target, cur.depth + 1})
			}
		}
		for _, e := range g.inEdges[cur.id] {
			if !visited[e.Source] {
				visited[e.Source] = true
				queue = append(queue, item{e.Source, cur.depth + 1})
			}
		}
	}
}

// ShortestPath finds the shortest path between two nodes using BFS.
func (g *Graph) ShortestPath(fromID, toID string) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	type pathItem struct {
		id   string
		path []string
	}
	visited := map[string]bool{fromID: true}
	queue := []pathItem{{fromID, []string{fromID}}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if cur.id == toID {
			result := make([]*Node, len(cur.path))
			for i, nid := range cur.path {
				result[i] = g.nodes[nid]
			}
			return result
		}

		for _, e := range g.outEdges[cur.id] {
			if !visited[e.Target] {
				visited[e.Target] = true
				newPath := make([]string, len(cur.path)+1)
				copy(newPath, cur.path)
				newPath[len(cur.path)] = e.Target
				queue = append(queue, pathItem{e.Target, newPath})
			}
		}
	}
	return nil // No path found
}

// TopNodes returns the most-accessed nodes, sorted by access count.
func (g *Graph) TopNodes(limit int) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	all := make([]*Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		all = append(all, n)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].AccessCount > all[j].AccessCount
	})
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}

// Stats returns graph statistics.
func (g *Graph) Stats() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	typeCounts := make(map[NodeType]int)
	for _, n := range g.nodes {
		typeCounts[n.Type]++
	}
	edgeTypeCounts := make(map[EdgeType]int)
	for _, e := range g.edges {
		edgeTypeCounts[e.Type]++
	}

	return map[string]interface{}{
		"total_nodes":      len(g.nodes),
		"total_edges":      len(g.edges),
		"node_types":       typeCounts,
		"edge_types":       edgeTypeCounts,
	}
}

// ExportJSON serializes the entire graph to JSON.
func (g *Graph) ExportJSON(_ context.Context) ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	data := struct {
		Nodes []*Node `json:"nodes"`
		Edges []*Edge `json:"edges"`
	}{
		Nodes: make([]*Node, 0, len(g.nodes)),
		Edges: make([]*Edge, 0, len(g.edges)),
	}
	for _, n := range g.nodes {
		data.Nodes = append(data.Nodes, n)
	}
	for _, e := range g.edges {
		data.Edges = append(data.Edges, e)
	}
	return json.MarshalIndent(data, "", "  ")
}

// IngestTriple adds a subject-predicate-object triple to the graph.
// This is the primary ingestion API for building the graph from natural language.
func (g *Graph) IngestTriple(subject, predicate, object string) error {
	sNode := g.AddNode(subject, NodeTypeEntity, nil)
	oNode := g.AddNode(object, NodeTypeEntity, nil)

	edgeType := mapPredicateToEdge(predicate)
	_, err := g.AddEdge(sNode.ID, oNode.ID, edgeType, 1.0)
	return err
}

func mapPredicateToEdge(predicate string) EdgeType {
	p := strings.ToLower(predicate)
	switch {
	case strings.Contains(p, "is a") || strings.Contains(p, "is an"):
		return EdgeIsA
	case strings.Contains(p, "has") || strings.Contains(p, "contains"):
		return EdgeHasPart
	case strings.Contains(p, "cause") || strings.Contains(p, "lead"):
		return EdgeCauses
	case strings.Contains(p, "derive") || strings.Contains(p, "from"):
		return EdgeDerivedFrom
	case strings.Contains(p, "use") || strings.Contains(p, "apply"):
		return EdgeUsedIn
	case strings.Contains(p, "depend"):
		return EdgeDependsOn
	default:
		return EdgeRelatedTo
	}
}
