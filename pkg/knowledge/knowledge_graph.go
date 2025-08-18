package knowledge

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultKnowledgeGraph implements the KnowledgeGraph interface
type DefaultKnowledgeGraph struct {
	entities      map[string]*Entity
	relationships map[string]*Relationship
	entityIndex   map[string][]string // entity type -> entity IDs
	adjacencyList map[string][]string // entity ID -> connected entity IDs
	logger        *logrus.Logger
	tracer        trace.Tracer
	mu            sync.RWMutex
	config        *KnowledgeGraphConfig
}

// KnowledgeGraphConfig represents configuration for the knowledge graph
type KnowledgeGraphConfig struct {
	MaxEntities      int     `json:"max_entities"`
	MaxRelationships int     `json:"max_relationships"`
	SimilarityThreshold float32 `json:"similarity_threshold"`
	EnableCaching    bool    `json:"enable_caching"`
	CacheTTL         time.Duration `json:"cache_ttl"`
}

// Path represents a path between entities
type Path struct {
	Entities      []*Entity      `json:"entities"`
	Relationships []*Relationship `json:"relationships"`
	Length        int            `json:"length"`
	Weight        float64        `json:"weight"`
}

// GraphQuery represents a graph query
type GraphQuery struct {
	Type       string                 `json:"type"`       // "find_path", "neighbors", "subgraph"
	EntityID   string                 `json:"entity_id,omitempty"`
	FromEntity string                 `json:"from_entity,omitempty"`
	ToEntity   string                 `json:"to_entity,omitempty"`
	MaxDepth   int                    `json:"max_depth,omitempty"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
}

// GraphResult represents graph query results
type GraphResult struct {
	Entities      []*Entity       `json:"entities,omitempty"`
	Relationships []*Relationship `json:"relationships,omitempty"`
	Paths         []*Path         `json:"paths,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ProcessingTime time.Duration  `json:"processing_time"`
}

// NewDefaultKnowledgeGraph creates a new default knowledge graph
func NewDefaultKnowledgeGraph(logger *logrus.Logger) (KnowledgeGraph, error) {
	config := &KnowledgeGraphConfig{
		MaxEntities:         100000,
		MaxRelationships:    500000,
		SimilarityThreshold: 0.8,
		EnableCaching:       true,
		CacheTTL:           24 * time.Hour,
	}

	graph := &DefaultKnowledgeGraph{
		entities:      make(map[string]*Entity),
		relationships: make(map[string]*Relationship),
		entityIndex:   make(map[string][]string),
		adjacencyList: make(map[string][]string),
		logger:        logger,
		tracer:        otel.Tracer("knowledge.graph"),
		config:        config,
	}

	return graph, nil
}

// AddEntity adds an entity to the knowledge graph
func (kg *DefaultKnowledgeGraph) AddEntity(ctx context.Context, entity *Entity) error {
	ctx, span := kg.tracer.Start(ctx, "knowledge_graph.add_entity")
	defer span.End()

	span.SetAttributes(
		attribute.String("entity.id", entity.ID),
		attribute.String("entity.name", entity.Name),
		attribute.String("entity.type", entity.Type),
	)

	kg.mu.Lock()
	defer kg.mu.Unlock()

	// Check if entity already exists
	if existing, exists := kg.entities[entity.ID]; exists {
		// Update existing entity
		existing.Name = entity.Name
		existing.Type = entity.Type
		existing.Description = entity.Description
		existing.Properties = entity.Properties
		existing.Aliases = entity.Aliases
		existing.Embedding = entity.Embedding
		existing.Confidence = entity.Confidence
		existing.UpdatedAt = time.Now()
		
		kg.logger.WithField("entity_id", entity.ID).Debug("Entity updated")
		return nil
	}

	// Set timestamps
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = time.Now()
	}
	entity.UpdatedAt = time.Now()

	// Add entity
	kg.entities[entity.ID] = entity

	// Update entity index
	if kg.entityIndex[entity.Type] == nil {
		kg.entityIndex[entity.Type] = make([]string, 0)
	}
	kg.entityIndex[entity.Type] = append(kg.entityIndex[entity.Type], entity.ID)

	// Initialize adjacency list
	if kg.adjacencyList[entity.ID] == nil {
		kg.adjacencyList[entity.ID] = make([]string, 0)
	}

	kg.logger.WithFields(logrus.Fields{
		"entity_id":   entity.ID,
		"entity_name": entity.Name,
		"entity_type": entity.Type,
	}).Debug("Entity added to knowledge graph")

	return nil
}

// AddRelationship adds a relationship to the knowledge graph
func (kg *DefaultKnowledgeGraph) AddRelationship(ctx context.Context, rel *Relationship) error {
	ctx, span := kg.tracer.Start(ctx, "knowledge_graph.add_relationship")
	defer span.End()

	span.SetAttributes(
		attribute.String("relationship.id", rel.ID),
		attribute.String("relationship.type", rel.Type),
		attribute.String("from_entity", rel.FromEntity),
		attribute.String("to_entity", rel.ToEntity),
	)

	kg.mu.Lock()
	defer kg.mu.Unlock()

	// Validate entities exist
	if _, exists := kg.entities[rel.FromEntity]; !exists {
		return fmt.Errorf("from entity not found: %s", rel.FromEntity)
	}
	if _, exists := kg.entities[rel.ToEntity]; !exists {
		return fmt.Errorf("to entity not found: %s", rel.ToEntity)
	}

	// Set timestamps
	if rel.CreatedAt.IsZero() {
		rel.CreatedAt = time.Now()
	}
	rel.UpdatedAt = time.Now()

	// Add relationship
	kg.relationships[rel.ID] = rel

	// Update adjacency list
	kg.adjacencyList[rel.FromEntity] = append(kg.adjacencyList[rel.FromEntity], rel.ToEntity)
	kg.adjacencyList[rel.ToEntity] = append(kg.adjacencyList[rel.ToEntity], rel.FromEntity)

	kg.logger.WithFields(logrus.Fields{
		"relationship_id":   rel.ID,
		"relationship_type": rel.Type,
		"from_entity":       rel.FromEntity,
		"to_entity":         rel.ToEntity,
	}).Debug("Relationship added to knowledge graph")

	return nil
}

// GetEntity retrieves an entity by ID
func (kg *DefaultKnowledgeGraph) GetEntity(ctx context.Context, id string) (*Entity, error) {
	ctx, span := kg.tracer.Start(ctx, "knowledge_graph.get_entity")
	defer span.End()

	span.SetAttributes(attribute.String("entity.id", id))

	kg.mu.RLock()
	defer kg.mu.RUnlock()

	entity, exists := kg.entities[id]
	if !exists {
		return nil, fmt.Errorf("entity not found: %s", id)
	}

	return entity, nil
}

// GetRelationships retrieves relationships for an entity
func (kg *DefaultKnowledgeGraph) GetRelationships(ctx context.Context, entityID string) ([]*Relationship, error) {
	ctx, span := kg.tracer.Start(ctx, "knowledge_graph.get_relationships")
	defer span.End()

	span.SetAttributes(attribute.String("entity.id", entityID))

	kg.mu.RLock()
	defer kg.mu.RUnlock()

	var relationships []*Relationship
	for _, rel := range kg.relationships {
		if rel.FromEntity == entityID || rel.ToEntity == entityID {
			relationships = append(relationships, rel)
		}
	}

	return relationships, nil
}

// FindPath finds paths between two entities
func (kg *DefaultKnowledgeGraph) FindPath(ctx context.Context, fromID, toID string, maxDepth int) ([]*Path, error) {
	ctx, span := kg.tracer.Start(ctx, "knowledge_graph.find_path")
	defer span.End()

	span.SetAttributes(
		attribute.String("from_entity", fromID),
		attribute.String("to_entity", toID),
		attribute.Int("max_depth", maxDepth),
	)

	kg.mu.RLock()
	defer kg.mu.RUnlock()

	// Validate entities exist
	if _, exists := kg.entities[fromID]; !exists {
		return nil, fmt.Errorf("from entity not found: %s", fromID)
	}
	if _, exists := kg.entities[toID]; !exists {
		return nil, fmt.Errorf("to entity not found: %s", toID)
	}

	// Use BFS to find shortest paths
	paths := kg.findPathsBFS(fromID, toID, maxDepth)

	kg.logger.WithFields(logrus.Fields{
		"from_entity": fromID,
		"to_entity":   toID,
		"max_depth":   maxDepth,
		"paths_found": len(paths),
	}).Debug("Path search completed")

	return paths, nil
}

// findPathsBFS finds paths using breadth-first search
func (kg *DefaultKnowledgeGraph) findPathsBFS(fromID, toID string, maxDepth int) []*Path {
	if maxDepth <= 0 {
		maxDepth = 5 // Default max depth
	}

	type pathState struct {
		currentEntity string
		path          []string
		depth         int
	}

	queue := []*pathState{{currentEntity: fromID, path: []string{fromID}, depth: 0}}
	visited := make(map[string]bool)
	var foundPaths []*Path

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.depth > maxDepth {
			continue
		}

		if current.currentEntity == toID && current.depth > 0 {
			// Found a path
			path := kg.buildPath(current.path)
			if path != nil {
				foundPaths = append(foundPaths, path)
			}
			continue
		}

		// Mark as visited for this path
		stateKey := fmt.Sprintf("%s-%d", current.currentEntity, current.depth)
		if visited[stateKey] {
			continue
		}
		visited[stateKey] = true

		// Explore neighbors
		neighbors := kg.adjacencyList[current.currentEntity]
		for _, neighbor := range neighbors {
			// Avoid cycles
			alreadyInPath := false
			for _, entity := range current.path {
				if entity == neighbor {
					alreadyInPath = true
					break
				}
			}

			if !alreadyInPath {
				newPath := make([]string, len(current.path))
				copy(newPath, current.path)
				newPath = append(newPath, neighbor)

				queue = append(queue, &pathState{
					currentEntity: neighbor,
					path:          newPath,
					depth:         current.depth + 1,
				})
			}
		}
	}

	return foundPaths
}

// buildPath builds a Path object from entity IDs
func (kg *DefaultKnowledgeGraph) buildPath(entityIDs []string) *Path {
	if len(entityIDs) < 2 {
		return nil
	}

	var entities []*Entity
	var relationships []*Relationship
	weight := 0.0

	// Get entities
	for _, entityID := range entityIDs {
		if entity, exists := kg.entities[entityID]; exists {
			entities = append(entities, entity)
		}
	}

	// Get relationships
	for i := 0; i < len(entityIDs)-1; i++ {
		fromID := entityIDs[i]
		toID := entityIDs[i+1]

		// Find relationship between these entities
		for _, rel := range kg.relationships {
			if (rel.FromEntity == fromID && rel.ToEntity == toID) ||
				(rel.FromEntity == toID && rel.ToEntity == fromID) {
				relationships = append(relationships, rel)
				weight += float64(rel.Confidence)
				break
			}
		}
	}

	return &Path{
		Entities:      entities,
		Relationships: relationships,
		Length:        len(entityIDs) - 1,
		Weight:        weight / float64(len(relationships)),
	}
}

// QueryGraph executes a graph query
func (kg *DefaultKnowledgeGraph) QueryGraph(ctx context.Context, query *GraphQuery) (*GraphResult, error) {
	ctx, span := kg.tracer.Start(ctx, "knowledge_graph.query_graph")
	defer span.End()

	startTime := time.Now()
	span.SetAttributes(attribute.String("query.type", query.Type))

	result := &GraphResult{
		Metadata: make(map[string]interface{}),
	}

	switch query.Type {
	case "find_path":
		paths, err := kg.FindPath(ctx, query.FromEntity, query.ToEntity, query.MaxDepth)
		if err != nil {
			return nil, err
		}
		result.Paths = paths

	case "neighbors":
		neighbors, err := kg.GetNeighbors(ctx, query.EntityID, query.MaxDepth)
		if err != nil {
			return nil, err
		}
		result.Entities = neighbors

	case "subgraph":
		entities, relationships := kg.getSubgraph(query.EntityID, query.MaxDepth)
		result.Entities = entities
		result.Relationships = relationships

	default:
		return nil, fmt.Errorf("unsupported query type: %s", query.Type)
	}

	result.ProcessingTime = time.Since(startTime)
	return result, nil
}

// GetNeighbors gets neighboring entities
func (kg *DefaultKnowledgeGraph) GetNeighbors(ctx context.Context, entityID string, depth int) ([]*Entity, error) {
	ctx, span := kg.tracer.Start(ctx, "knowledge_graph.get_neighbors")
	defer span.End()

	kg.mu.RLock()
	defer kg.mu.RUnlock()

	visited := make(map[string]bool)
	var neighbors []*Entity

	kg.getNeighborsRecursive(entityID, depth, visited, &neighbors)

	return neighbors, nil
}

// getNeighborsRecursive recursively gets neighbors
func (kg *DefaultKnowledgeGraph) getNeighborsRecursive(entityID string, depth int, visited map[string]bool, neighbors *[]*Entity) {
	if depth <= 0 || visited[entityID] {
		return
	}

	visited[entityID] = true

	// Add current entity if it's not the starting entity
	if len(visited) > 1 {
		if entity, exists := kg.entities[entityID]; exists {
			*neighbors = append(*neighbors, entity)
		}
	}

	// Recurse to connected entities
	for _, neighborID := range kg.adjacencyList[entityID] {
		kg.getNeighborsRecursive(neighborID, depth-1, visited, neighbors)
	}
}

// getSubgraph gets a subgraph around an entity
func (kg *DefaultKnowledgeGraph) getSubgraph(entityID string, depth int) ([]*Entity, []*Relationship) {
	kg.mu.RLock()
	defer kg.mu.RUnlock()

	visited := make(map[string]bool)
	var entities []*Entity
	var relationships []*Relationship

	kg.getSubgraphRecursive(entityID, depth, visited, &entities, &relationships)

	return entities, relationships
}

// getSubgraphRecursive recursively builds a subgraph
func (kg *DefaultKnowledgeGraph) getSubgraphRecursive(entityID string, depth int, visited map[string]bool, entities *[]*Entity, relationships *[]*Relationship) {
	if depth < 0 || visited[entityID] {
		return
	}

	visited[entityID] = true

	// Add current entity
	if entity, exists := kg.entities[entityID]; exists {
		*entities = append(*entities, entity)
	}

	// Add relationships and recurse
	for _, rel := range kg.relationships {
		if rel.FromEntity == entityID || rel.ToEntity == entityID {
			*relationships = append(*relationships, rel)

			// Recurse to connected entity
			var nextEntity string
			if rel.FromEntity == entityID {
				nextEntity = rel.ToEntity
			} else {
				nextEntity = rel.FromEntity
			}

			kg.getSubgraphRecursive(nextEntity, depth-1, visited, entities, relationships)
		}
	}
}

// CalculateCentrality calculates centrality score for an entity
func (kg *DefaultKnowledgeGraph) CalculateCentrality(ctx context.Context, entityID string) (float64, error) {
	ctx, span := kg.tracer.Start(ctx, "knowledge_graph.calculate_centrality")
	defer span.End()

	kg.mu.RLock()
	defer kg.mu.RUnlock()

	// Simple degree centrality calculation
	connections := len(kg.adjacencyList[entityID])
	totalEntities := len(kg.entities)

	if totalEntities <= 1 {
		return 0.0, nil
	}

	centrality := float64(connections) / float64(totalEntities-1)
	return math.Min(centrality, 1.0), nil
}
