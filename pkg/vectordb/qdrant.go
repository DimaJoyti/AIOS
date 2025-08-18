package vectordb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// QdrantDB implements VectorDB interface for Qdrant
type QdrantDB struct {
	config     *VectorDBConfig
	httpClient *http.Client
	baseURL    string
	connected  bool
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// QdrantFactory implements VectorDBFactory for Qdrant
type QdrantFactory struct{}

// NewQdrantFactory creates a new Qdrant factory
func NewQdrantFactory() *QdrantFactory {
	return &QdrantFactory{}
}

// Create creates a new Qdrant vector database instance
func (f *QdrantFactory) Create(config *VectorDBConfig) (VectorDB, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	logger := logrus.New()
	if config.Metadata != nil {
		if logLevel, exists := config.Metadata["log_level"]; exists {
			if level, ok := logLevel.(string); ok {
				if parsedLevel, err := logrus.ParseLevel(level); err == nil {
					logger.SetLevel(parsedLevel)
				}
			}
		}
	}

	baseURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)
	if config.TLS {
		baseURL = fmt.Sprintf("https://%s:%d", config.Host, config.Port)
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &QdrantDB{
		config:     config,
		httpClient: httpClient,
		baseURL:    baseURL,
		connected:  false,
		logger:     logger,
		tracer:     otel.Tracer("vectordb.qdrant"),
	}, nil
}

// GetProviderName returns the provider name
func (f *QdrantFactory) GetProviderName() string {
	return "qdrant"
}

// ValidateConfig validates the Qdrant configuration
func (f *QdrantFactory) ValidateConfig(config *VectorDBConfig) error {
	if config.Host == "" {
		return fmt.Errorf("host is required")
	}
	if config.Port <= 0 {
		return fmt.Errorf("port must be positive")
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries < 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = 1 * time.Second
	}
	return nil
}

// Connect establishes connection to Qdrant
func (q *QdrantDB) Connect(ctx context.Context) error {
	ctx, span := q.tracer.Start(ctx, "qdrant.connect")
	defer span.End()

	// Test connection by getting cluster info
	_, err := q.makeRequest(ctx, "GET", "/cluster", nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to connect to Qdrant: %w", err)
	}

	q.connected = true
	q.logger.WithField("base_url", q.baseURL).Info("Connected to Qdrant")

	return nil
}

// Disconnect closes the connection
func (q *QdrantDB) Disconnect(ctx context.Context) error {
	q.connected = false
	q.logger.Info("Disconnected from Qdrant")
	return nil
}

// IsConnected returns connection status
func (q *QdrantDB) IsConnected() bool {
	return q.connected
}

// CreateCollection creates a new collection
func (q *QdrantDB) CreateCollection(ctx context.Context, config *CollectionConfig) error {
	ctx, span := q.tracer.Start(ctx, "qdrant.create_collection")
	defer span.End()

	span.SetAttributes(
		attribute.String("collection.name", config.Name),
		attribute.Int("collection.dimension", config.Dimension),
		attribute.String("collection.metric", config.Metric),
	)

	// Map metric names to Qdrant format
	distance := "Cosine"
	switch strings.ToLower(config.Metric) {
	case "cosine":
		distance = "Cosine"
	case "euclidean":
		distance = "Euclid"
	case "dot_product":
		distance = "Dot"
	}

	payload := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     config.Dimension,
			"distance": distance,
		},
	}

	path := fmt.Sprintf("/collections/%s", config.Name)
	_, err := q.makeRequest(ctx, "PUT", path, payload)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create collection: %w", err)
	}

	q.logger.WithFields(logrus.Fields{
		"collection": config.Name,
		"dimension":  config.Dimension,
		"metric":     config.Metric,
	}).Info("Created collection")

	return nil
}

// DeleteCollection deletes a collection
func (q *QdrantDB) DeleteCollection(ctx context.Context, name string) error {
	ctx, span := q.tracer.Start(ctx, "qdrant.delete_collection")
	defer span.End()

	span.SetAttributes(attribute.String("collection.name", name))

	path := fmt.Sprintf("/collections/%s", name)
	_, err := q.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	q.logger.WithField("collection", name).Info("Deleted collection")
	return nil
}

// ListCollections returns all collections
func (q *QdrantDB) ListCollections(ctx context.Context) ([]string, error) {
	ctx, span := q.tracer.Start(ctx, "qdrant.list_collections")
	defer span.End()

	response, err := q.makeRequest(ctx, "GET", "/collections", nil)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	var result struct {
		Result struct {
			Collections []struct {
				Name string `json:"name"`
			} `json:"collections"`
		} `json:"result"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	collections := make([]string, len(result.Result.Collections))
	for i, col := range result.Result.Collections {
		collections[i] = col.Name
	}

	return collections, nil
}

// CollectionExists checks if a collection exists
func (q *QdrantDB) CollectionExists(ctx context.Context, name string) (bool, error) {
	collections, err := q.ListCollections(ctx)
	if err != nil {
		return false, err
	}

	for _, col := range collections {
		if col == name {
			return true, nil
		}
	}

	return false, nil
}

// Insert inserts vectors into a collection
func (q *QdrantDB) Insert(ctx context.Context, collection string, vectors []*Vector) error {
	ctx, span := q.tracer.Start(ctx, "qdrant.insert")
	defer span.End()

	span.SetAttributes(
		attribute.String("collection.name", collection),
		attribute.Int("vectors.count", len(vectors)),
	)

	points := make([]map[string]interface{}, len(vectors))
	for i, vector := range vectors {
		point := map[string]interface{}{
			"id":      vector.ID,
			"vector":  vector.Values,
			"payload": vector.Metadata,
		}
		points[i] = point
	}

	payload := map[string]interface{}{
		"points": points,
	}

	path := fmt.Sprintf("/collections/%s/points", collection)
	_, err := q.makeRequest(ctx, "PUT", path, payload)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to insert vectors: %w", err)
	}

	q.logger.WithFields(logrus.Fields{
		"collection": collection,
		"count":      len(vectors),
	}).Debug("Inserted vectors")

	return nil
}

// Update updates existing vectors
func (q *QdrantDB) Update(ctx context.Context, collection string, vectors []*Vector) error {
	// Qdrant uses the same endpoint for insert and update
	return q.Insert(ctx, collection, vectors)
}

// Delete deletes vectors by IDs
func (q *QdrantDB) Delete(ctx context.Context, collection string, ids []string) error {
	ctx, span := q.tracer.Start(ctx, "qdrant.delete")
	defer span.End()

	span.SetAttributes(
		attribute.String("collection.name", collection),
		attribute.Int("ids.count", len(ids)),
	)

	payload := map[string]interface{}{
		"points": ids,
	}

	path := fmt.Sprintf("/collections/%s/points/delete", collection)
	_, err := q.makeRequest(ctx, "POST", path, payload)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	q.logger.WithFields(logrus.Fields{
		"collection": collection,
		"count":      len(ids),
	}).Debug("Deleted vectors")

	return nil
}

// Search performs similarity search
func (q *QdrantDB) Search(ctx context.Context, request *SearchRequest) (*SearchResult, error) {
	ctx, span := q.tracer.Start(ctx, "qdrant.search")
	defer span.End()

	span.SetAttributes(
		attribute.String("collection.name", request.Collection),
		attribute.Int("search.top_k", request.TopK),
	)

	payload := map[string]interface{}{
		"vector":       request.Vector,
		"limit":        request.TopK,
		"with_vector":  contains(request.Include, "values"),
		"with_payload": contains(request.Include, "metadata"),
	}

	if request.Filter != nil {
		payload["filter"] = request.Filter
	}

	path := fmt.Sprintf("/collections/%s/points/search", request.Collection)
	response, err := q.makeRequest(ctx, "POST", path, payload)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	var result struct {
		Result []struct {
			ID      interface{}            `json:"id"`
			Score   float32                `json:"score"`
			Vector  []float32              `json:"vector,omitempty"`
			Payload map[string]interface{} `json:"payload,omitempty"`
		} `json:"result"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	matches := make([]Match, len(result.Result))
	for i, item := range result.Result {
		match := Match{
			ID:       fmt.Sprintf("%v", item.ID),
			Score:    item.Score,
			Values:   item.Vector,
			Metadata: item.Payload,
		}
		matches[i] = match
	}

	return &SearchResult{
		Matches: matches,
		Total:   int64(len(matches)),
	}, nil
}

// BatchSearch performs multiple searches
func (q *QdrantDB) BatchSearch(ctx context.Context, requests []*SearchRequest) ([]*SearchResult, error) {
	results := make([]*SearchResult, len(requests))
	for i, request := range requests {
		result, err := q.Search(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("batch search failed at index %d: %w", i, err)
		}
		results[i] = result
	}
	return results, nil
}

// GetVector retrieves a vector by ID
func (q *QdrantDB) GetVector(ctx context.Context, collection string, id string) (*Vector, error) {
	vectors, err := q.GetVectors(ctx, collection, []string{id})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("vector not found: %s", id)
	}
	return vectors[0], nil
}

// GetVectors retrieves multiple vectors by IDs
func (q *QdrantDB) GetVectors(ctx context.Context, collection string, ids []string) ([]*Vector, error) {
	ctx, span := q.tracer.Start(ctx, "qdrant.get_vectors")
	defer span.End()

	span.SetAttributes(
		attribute.String("collection.name", collection),
		attribute.Int("ids.count", len(ids)),
	)

	payload := map[string]interface{}{
		"ids":          ids,
		"with_vector":  true,
		"with_payload": true,
	}

	path := fmt.Sprintf("/collections/%s/points", collection)
	response, err := q.makeRequest(ctx, "POST", path, payload)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get vectors: %w", err)
	}

	var result struct {
		Result []struct {
			ID      interface{}            `json:"id"`
			Vector  []float32              `json:"vector"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	vectors := make([]*Vector, len(result.Result))
	for i, item := range result.Result {
		vector := &Vector{
			ID:       fmt.Sprintf("%v", item.ID),
			Values:   item.Vector,
			Metadata: item.Payload,
		}
		vectors[i] = vector
	}

	return vectors, nil
}

// Count returns the number of vectors in a collection
func (q *QdrantDB) Count(ctx context.Context, collection string) (int64, error) {
	ctx, span := q.tracer.Start(ctx, "qdrant.count")
	defer span.End()

	span.SetAttributes(attribute.String("collection.name", collection))

	path := fmt.Sprintf("/collections/%s", collection)
	response, err := q.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to get collection info: %w", err)
	}

	var result struct {
		Result struct {
			PointsCount int64 `json:"points_count"`
		} `json:"result"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Result.PointsCount, nil
}

// CreateIndex creates an index (Qdrant handles indexing automatically)
func (q *QdrantDB) CreateIndex(ctx context.Context, collection string, config *IndexConfig) error {
	// Qdrant handles indexing automatically, so this is a no-op
	q.logger.WithField("collection", collection).Debug("Index creation not needed for Qdrant")
	return nil
}

// DeleteIndex deletes an index (Qdrant handles indexing automatically)
func (q *QdrantDB) DeleteIndex(ctx context.Context, collection string, indexName string) error {
	// Qdrant handles indexing automatically, so this is a no-op
	q.logger.WithField("collection", collection).Debug("Index deletion not needed for Qdrant")
	return nil
}

// GetCollectionInfo returns collection information
func (q *QdrantDB) GetCollectionInfo(ctx context.Context, collection string) (*CollectionInfo, error) {
	ctx, span := q.tracer.Start(ctx, "qdrant.get_collection_info")
	defer span.End()

	span.SetAttributes(attribute.String("collection.name", collection))

	path := fmt.Sprintf("/collections/%s", collection)
	response, err := q.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get collection info: %w", err)
	}

	var result struct {
		Result struct {
			Config struct {
				Params struct {
					Vectors struct {
						Size     int    `json:"size"`
						Distance string `json:"distance"`
					} `json:"vectors"`
				} `json:"params"`
			} `json:"config"`
			PointsCount int64 `json:"points_count"`
		} `json:"result"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Map Qdrant distance to standard metric
	metric := "cosine"
	switch result.Result.Config.Params.Vectors.Distance {
	case "Cosine":
		metric = "cosine"
	case "Euclid":
		metric = "euclidean"
	case "Dot":
		metric = "dot_product"
	}

	info := &CollectionInfo{
		Name:        collection,
		Dimension:   result.Result.Config.Params.Vectors.Size,
		Metric:      metric,
		VectorCount: result.Result.PointsCount,
		IndexCount:  1,          // Qdrant always has an index
		CreatedAt:   time.Now(), // Qdrant doesn't provide creation time
		UpdatedAt:   time.Now(),
	}

	return info, nil
}

// Health returns the health status
func (q *QdrantDB) Health(ctx context.Context) (*HealthStatus, error) {
	ctx, span := q.tracer.Start(ctx, "qdrant.health")
	defer span.End()

	response, err := q.makeRequest(ctx, "GET", "/", nil)
	if err != nil {
		span.RecordError(err)
		return &HealthStatus{
			Status: "unhealthy",
		}, nil
	}

	var result struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return &HealthStatus{
			Status: "degraded",
		}, nil
	}

	// Get collections count
	collections, _ := q.ListCollections(ctx)

	return &HealthStatus{
		Status:      "healthy",
		Version:     result.Version,
		Collections: len(collections),
	}, nil
}

// Helper methods

func (q *QdrantDB) makeRequest(ctx context.Context, method, path string, payload interface{}) ([]byte, error) {
	url := q.baseURL + path

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if q.config.APIKey != "" {
		req.Header.Set("api-key", q.config.APIKey)
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
