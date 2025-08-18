package langgraph

import (
	"fmt"
	"sync"
)

// DefaultGraph implements the Graph interface
type DefaultGraph struct {
	id          string
	nodes       map[string]Node
	edges       map[string][]Edge
	entryPoints []string
	exitPoints  []string
	metadata    map[string]interface{}
	mu          sync.RWMutex
}

// DefaultEdge implements the Edge interface
type DefaultEdge struct {
	from      string
	to        string
	condition Condition
	weight    float64
	metadata  map[string]interface{}
}

// NewGraph creates a new graph
func NewGraph(id string) Graph {
	return &DefaultGraph{
		id:          id,
		nodes:       make(map[string]Node),
		edges:       make(map[string][]Edge),
		entryPoints: make([]string, 0),
		exitPoints:  make([]string, 0),
		metadata:    make(map[string]interface{}),
	}
}

// NewEdge creates a new edge
func NewEdge(from, to string, condition Condition, weight float64) Edge {
	return &DefaultEdge{
		from:      from,
		to:        to,
		condition: condition,
		weight:    weight,
		metadata:  make(map[string]interface{}),
	}
}

// AddNode adds a node to the graph
func (g *DefaultGraph) AddNode(node Node) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	nodeID := node.GetID()
	if nodeID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	if _, exists := g.nodes[nodeID]; exists {
		return fmt.Errorf("node with ID %s already exists", nodeID)
	}

	if err := node.Validate(); err != nil {
		return fmt.Errorf("node validation failed: %w", err)
	}

	g.nodes[nodeID] = node
	g.edges[nodeID] = make([]Edge, 0)

	return nil
}

// AddEdge adds an edge to the graph
func (g *DefaultGraph) AddEdge(edge Edge) error {
	if edge == nil {
		return fmt.Errorf("edge cannot be nil")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	from := edge.GetFrom()
	to := edge.GetTo()

	if from == "" || to == "" {
		return fmt.Errorf("edge from and to cannot be empty")
	}

	if _, exists := g.nodes[from]; !exists {
		return fmt.Errorf("source node %s does not exist", from)
	}

	if _, exists := g.nodes[to]; !exists {
		return fmt.Errorf("target node %s does not exist", to)
	}

	// Check for duplicate edges
	for _, existingEdge := range g.edges[from] {
		if existingEdge.GetTo() == to {
			return fmt.Errorf("edge from %s to %s already exists", from, to)
		}
	}

	g.edges[from] = append(g.edges[from], edge)

	return nil
}

// RemoveNode removes a node from the graph
func (g *DefaultGraph) RemoveNode(nodeID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.nodes[nodeID]; !exists {
		return fmt.Errorf("node %s does not exist", nodeID)
	}

	// Remove the node
	delete(g.nodes, nodeID)
	delete(g.edges, nodeID)

	// Remove all edges pointing to this node
	for fromNode, edges := range g.edges {
		newEdges := make([]Edge, 0)
		for _, edge := range edges {
			if edge.GetTo() != nodeID {
				newEdges = append(newEdges, edge)
			}
		}
		g.edges[fromNode] = newEdges
	}

	// Remove from entry and exit points
	g.entryPoints = g.removeFromSlice(g.entryPoints, nodeID)
	g.exitPoints = g.removeFromSlice(g.exitPoints, nodeID)

	return nil
}

// RemoveEdge removes an edge from the graph
func (g *DefaultGraph) RemoveEdge(from, to string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	edges, exists := g.edges[from]
	if !exists {
		return fmt.Errorf("source node %s does not exist", from)
	}

	newEdges := make([]Edge, 0)
	found := false
	for _, edge := range edges {
		if edge.GetTo() != to {
			newEdges = append(newEdges, edge)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("edge from %s to %s does not exist", from, to)
	}

	g.edges[from] = newEdges

	return nil
}

// GetNode returns a node by ID
func (g *DefaultGraph) GetNode(nodeID string) (Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s does not exist", nodeID)
	}

	return node, nil
}

// GetNodes returns all nodes in the graph
func (g *DefaultGraph) GetNodes() map[string]Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodes := make(map[string]Node)
	for id, node := range g.nodes {
		nodes[id] = node
	}

	return nodes
}

// GetEdges returns all edges from a node
func (g *DefaultGraph) GetEdges(nodeID string) []Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges, exists := g.edges[nodeID]
	if !exists {
		return []Edge{}
	}

	// Return a copy to prevent external modification
	edgesCopy := make([]Edge, len(edges))
	copy(edgesCopy, edges)

	return edgesCopy
}

// GetEntryPoints returns the entry point nodes
func (g *DefaultGraph) GetEntryPoints() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.entryPoints) > 0 {
		entryPoints := make([]string, len(g.entryPoints))
		copy(entryPoints, g.entryPoints)
		return entryPoints
	}

	// If no explicit entry points, find nodes with no incoming edges
	incomingEdges := make(map[string]bool)
	for _, edges := range g.edges {
		for _, edge := range edges {
			incomingEdges[edge.GetTo()] = true
		}
	}

	var entryPoints []string
	for nodeID := range g.nodes {
		if !incomingEdges[nodeID] {
			entryPoints = append(entryPoints, nodeID)
		}
	}

	return entryPoints
}

// GetExitPoints returns the exit point nodes
func (g *DefaultGraph) GetExitPoints() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.exitPoints) > 0 {
		exitPoints := make([]string, len(g.exitPoints))
		copy(exitPoints, g.exitPoints)
		return exitPoints
	}

	// If no explicit exit points, find nodes with no outgoing edges
	var exitPoints []string
	for nodeID := range g.nodes {
		if len(g.edges[nodeID]) == 0 {
			exitPoints = append(exitPoints, nodeID)
		}
	}

	return exitPoints
}

// Validate validates the graph structure
func (g *DefaultGraph) Validate() error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.nodes) == 0 {
		return fmt.Errorf("graph must have at least one node")
	}

	// Validate all nodes
	for nodeID, node := range g.nodes {
		if err := node.Validate(); err != nil {
			return fmt.Errorf("node %s validation failed: %w", nodeID, err)
		}
	}

	// Check for cycles (simple DFS-based cycle detection)
	if g.hasCycles() {
		return fmt.Errorf("graph contains cycles")
	}

	// Validate that all nodes are reachable from entry points
	entryPoints := g.GetEntryPoints()
	if len(entryPoints) == 0 {
		return fmt.Errorf("graph must have at least one entry point")
	}

	reachableNodes := g.getReachableNodes(entryPoints)
	if len(reachableNodes) != len(g.nodes) {
		return fmt.Errorf("not all nodes are reachable from entry points")
	}

	return nil
}

// Clone creates a copy of the graph
func (g *DefaultGraph) Clone() Graph {
	g.mu.RLock()
	defer g.mu.RUnlock()

	clone := &DefaultGraph{
		id:          g.id + "_clone",
		nodes:       make(map[string]Node),
		edges:       make(map[string][]Edge),
		entryPoints: make([]string, len(g.entryPoints)),
		exitPoints:  make([]string, len(g.exitPoints)),
		metadata:    make(map[string]interface{}),
	}

	// Copy entry and exit points
	copy(clone.entryPoints, g.entryPoints)
	copy(clone.exitPoints, g.exitPoints)

	// Copy metadata
	for k, v := range g.metadata {
		clone.metadata[k] = v
	}

	// Copy nodes (shallow copy - nodes themselves are not cloned)
	for id, node := range g.nodes {
		clone.nodes[id] = node
	}

	// Copy edges
	for nodeID, edges := range g.edges {
		clone.edges[nodeID] = make([]Edge, len(edges))
		copy(clone.edges[nodeID], edges)
	}

	return clone
}

// SetEntryPoints sets the entry points
func (g *DefaultGraph) SetEntryPoints(entryPoints []string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Validate that all entry points exist
	for _, nodeID := range entryPoints {
		if _, exists := g.nodes[nodeID]; !exists {
			return fmt.Errorf("entry point node %s does not exist", nodeID)
		}
	}

	g.entryPoints = make([]string, len(entryPoints))
	copy(g.entryPoints, entryPoints)

	return nil
}

// SetExitPoints sets the exit points
func (g *DefaultGraph) SetExitPoints(exitPoints []string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Validate that all exit points exist
	for _, nodeID := range exitPoints {
		if _, exists := g.nodes[nodeID]; !exists {
			return fmt.Errorf("exit point node %s does not exist", nodeID)
		}
	}

	g.exitPoints = make([]string, len(exitPoints))
	copy(g.exitPoints, exitPoints)

	return nil
}

// Helper methods

func (g *DefaultGraph) removeFromSlice(slice []string, item string) []string {
	result := make([]string, 0)
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

func (g *DefaultGraph) hasCycles() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeID := range g.nodes {
		if !visited[nodeID] {
			if g.hasCyclesUtil(nodeID, visited, recStack) {
				return true
			}
		}
	}

	return false
}

func (g *DefaultGraph) hasCyclesUtil(nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	for _, edge := range g.edges[nodeID] {
		to := edge.GetTo()
		if !visited[to] {
			if g.hasCyclesUtil(to, visited, recStack) {
				return true
			}
		} else if recStack[to] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

func (g *DefaultGraph) getReachableNodes(startNodes []string) map[string]bool {
	reachable := make(map[string]bool)
	queue := make([]string, 0)

	// Initialize queue with start nodes
	for _, nodeID := range startNodes {
		if _, exists := g.nodes[nodeID]; exists {
			queue = append(queue, nodeID)
			reachable[nodeID] = true
		}
	}

	// BFS to find all reachable nodes
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, edge := range g.edges[current] {
			to := edge.GetTo()
			if !reachable[to] {
				reachable[to] = true
				queue = append(queue, to)
			}
		}
	}

	return reachable
}

// DefaultEdge methods

// GetFrom returns the source node ID
func (e *DefaultEdge) GetFrom() string {
	return e.from
}

// GetTo returns the target node ID
func (e *DefaultEdge) GetTo() string {
	return e.to
}

// GetCondition returns the condition for this edge
func (e *DefaultEdge) GetCondition() Condition {
	return e.condition
}

// GetWeight returns the weight of this edge
func (e *DefaultEdge) GetWeight() float64 {
	return e.weight
}

// SetCondition sets the condition for this edge
func (e *DefaultEdge) SetCondition(condition Condition) {
	e.condition = condition
}

// SetWeight sets the weight of this edge
func (e *DefaultEdge) SetWeight(weight float64) {
	e.weight = weight
}

// GetMetadata returns the edge metadata
func (e *DefaultEdge) GetMetadata() map[string]interface{} {
	return e.metadata
}

// SetMetadata sets edge metadata
func (e *DefaultEdge) SetMetadata(key string, value interface{}) {
	e.metadata[key] = value
}
