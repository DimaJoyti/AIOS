# AIOS External Data Integration

## Overview

The AIOS External Data Integration system provides a comprehensive platform for connecting with and ingesting data from various external sources. Built with enterprise-grade capabilities, it offers web crawling, API integration, real-time data processing, ETL pipelines, and comprehensive monitoring to enable seamless data integration from any external source.

## 🏗️ Architecture

### Core Components

```
External Data Integration
├── Data Integration Engine (core orchestration, lifecycle management)
├── Connector Framework (pluggable data connectors)
├── Web Crawler (configurable web scraping with politeness)
├── API Connectors (REST, GraphQL, streaming APIs)
├── Data Processing Pipelines (ETL, validation, transformation)
├── Storage Management (multi-backend storage support)
├── Real-time Processing (streaming data, event-driven)
├── Schema Management (data validation, evolution)
├── Health Monitoring (source health, performance metrics)
└── Security Layer (authentication, rate limiting, compliance)
```

### Key Features

- **🕷️ Web Crawling**: Intelligent web crawler with configurable rules and politeness policies
- **🔌 API Integration**: REST and GraphQL connectors with authentication and pagination
- **⚡ Real-time Processing**: Streaming data support with event-driven architecture
- **🔄 ETL Pipelines**: Comprehensive data transformation and validation pipelines
- **💾 Multi-Storage**: Support for PostgreSQL, MongoDB, Elasticsearch, Redis, S3, and more
- **📊 Schema Management**: Automatic schema inference and validation
- **🛡️ Security & Compliance**: Rate limiting, authentication, and audit logging
- **📈 Monitoring & Analytics**: Real-time health checks and performance metrics

## 🚀 Quick Start

### Basic Data Integration Setup

```go
package main

import (
    "time"
    "github.com/aios/aios/pkg/dataintegration"
    "github.com/aios/aios/pkg/dataintegration/connectors"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create data integration engine
    config := &dataintegration.DataIntegrationEngineConfig{
        MaxDataSources:      1000,
        MaxPipelines:        5000,
        DefaultTimeout:      30 * time.Second,
        LogRetention:        7 * 24 * time.Hour,
        MetricsRetention:    30 * 24 * time.Hour,
        HealthCheckInterval: 5 * time.Minute,
        EnableMetrics:       true,
        EnableHealthChecks:  true,
        MaxConcurrentJobs:   10,
    }
    
    dataEngine := dataintegration.NewDefaultDataIntegrationEngine(config, logger)
    
    // Register connectors
    webCrawler := connectors.NewWebCrawlerConnector()
    dataEngine.RegisterConnector(webCrawler)
    
    restAPI := connectors.NewRestAPIConnector()
    dataEngine.RegisterConnector(restAPI)
    
    // Create data source
    dataSource := &dataintegration.DataSource{
        Name:        "News Website",
        Description: "Web crawler for news articles",
        Type:        "webcrawler",
        Provider:    "WebCrawler",
        Config: &dataintegration.DataSourceConfig{
            URL: "https://example-news.com",
            CrawlConfig: &dataintegration.CrawlConfig{
                MaxDepth:        3,
                MaxPages:        100,
                RespectRobots:   true,
                UserAgent:       "AIOS-Crawler/1.0",
                Delay:           2 * time.Second,
                Selectors:       []string{"article", "h1", "p"},
                ExcludePatterns: []string{"/admin", "/login"},
            },
        },
        Settings: &dataintegration.DataSourceSettings{
            AutoSync:     true,
            SyncInterval: 6 * time.Hour,
            EnableEvents: true,
        },
        CreatedBy: "admin@company.com",
    }
    
    createdSource, err := dataEngine.CreateDataSource(dataSource)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Data source created: %s\n", createdSource.ID)
}
```

## 🕷️ Web Crawling

### Web Crawler Configuration

The web crawler provides intelligent, respectful web scraping with configurable rules:

```go
// Create web crawler data source
webCrawlerSource := &dataintegration.DataSource{
    Name:     "E-commerce Product Crawler",
    Type:     "webcrawler",
    Provider: "WebCrawler",
    Config: &dataintegration.DataSourceConfig{
        URL: "https://example-store.com",
        CrawlConfig: &dataintegration.CrawlConfig{
            MaxDepth:        5,              // Maximum crawling depth
            MaxPages:        1000,           // Maximum pages to crawl
            FollowRedirects: true,           // Follow HTTP redirects
            RespectRobots:   true,           // Respect robots.txt
            UserAgent:       "AIOS-Crawler/1.0",
            Delay:           2 * time.Second, // Delay between requests
            Selectors:       []string{       // CSS selectors for content
                ".product-title",
                ".product-price", 
                ".product-description",
                ".product-specs",
            },
            ExcludePatterns: []string{       // URL patterns to exclude
                "/admin", "/login", "/cart", "/checkout",
            },
            IncludePatterns: []string{       // URL patterns to include
                "/products", "/categories",
            },
        },
        RateLimit: &dataintegration.RateLimit{
            RequestsPerSecond: 2,            // Respectful crawling rate
            BurstSize:         5,
            WindowSize:        time.Minute,
        },
    },
    Settings: &dataintegration.DataSourceSettings{
        AutoSync:      true,
        SyncInterval:  12 * time.Hour,     // Crawl twice daily
        DataRetention: 30 * 24 * time.Hour,
        MaxRecords:    100000,
    },
}

createdSource, err := dataEngine.CreateDataSource(webCrawlerSource)

// Test the crawler
testResult, err := dataEngine.TestDataSource(createdSource.ID)
fmt.Printf("Crawler test: %t - %s\n", testResult.Success, testResult.Message)

// Available crawler operations:
// - crawl_website: Full website crawling
// - extract_page: Single page extraction
// - get_links: Extract all links from a page
// - extract_content: Extract specific content using selectors
// - check_robots: Check robots.txt compliance
// - get_sitemap: Extract sitemap URLs
```

### Advanced Crawler Features

```go
// Custom crawler with advanced configuration
advancedCrawlerConfig := &dataintegration.CrawlConfig{
    MaxDepth:        3,
    MaxPages:        500,
    FollowRedirects: true,
    RespectRobots:   true,
    UserAgent:       "AIOS-Advanced-Crawler/1.0",
    Delay:           1 * time.Second,
    
    // Advanced selectors for structured data extraction
    Selectors: []string{
        "article h1",                    // Article titles
        "article .content",              // Article content
        "article .author",               // Author information
        "article .publish-date",         // Publication date
        "article .tags a",               // Article tags
        ".breadcrumb a",                 // Navigation breadcrumbs
        ".related-articles a",           // Related content
    },
    
    // Sophisticated filtering
    ExcludePatterns: []string{
        "/admin", "/login", "/register", "/cart",
        "/api/", "/.well-known/", "/assets/",
        "?print=", "&print=", "#comment",
    },
    
    IncludePatterns: []string{
        "/articles/", "/news/", "/blog/",
        "/category/", "/tag/", "/author/",
    },
}

// Apply configuration to data source
webCrawlerSource.Config.CrawlConfig = advancedCrawlerConfig
```

## 🔌 API Integration

### REST API Connector

Comprehensive REST API integration with authentication and pagination:

```go
// Create REST API data source
restAPISource := &dataintegration.DataSource{
    Name:     "Customer API",
    Type:     "restapi",
    Provider: "CustomerService",
    Config: &dataintegration.DataSourceConfig{
        URL: "https://api.customer-service.com",
        Headers: map[string]string{
            "Accept":     "application/json",
            "User-Agent": "AIOS-Integration/1.0",
        },
        Custom: map[string]interface{}{
            "base_url":         "https://api.customer-service.com",
            "auth_type":        "bearer",
            "auth_header":      "Authorization",
            "timeout":          "30s",
            "rate_limit":       100,
            "pagination_type":  "offset",
            "pagination_limit": 200,
        },
    },
    Credentials: &dataintegration.DataSourceCredentials{
        Type:  dataintegration.CredentialTypeBearer,
        Token: "your-api-token-here",
    },
    Settings: &dataintegration.DataSourceSettings{
        AutoSync:     true,
        SyncInterval: 30 * time.Minute,
        EnableEvents: true,
    },
}

createdAPISource, err := dataEngine.CreateDataSource(restAPISource)

// Extract data from API endpoints
extractionParams := &dataintegration.ExtractionParams{
    Query: "/customers",
    Filters: map[string]interface{}{
        "status": "active",
        "type":   "premium",
    },
    Limit:     100,
    Offset:    0,
    SortBy:    "created_at",
    SortOrder: "desc",
}

connector, _ := dataEngine.GetConnector("restapi")
extraction, err := connector.ExtractData(context.Background(), extractionParams)

fmt.Printf("Extracted %d records\n", len(extraction.Records))

// Available API operations:
// - get_data: GET requests with query parameters
// - post_data: POST requests with JSON body
// - put_data: PUT requests for updates
// - delete_data: DELETE requests
// - list_endpoints: Discover available endpoints
// - get_schema: Retrieve API schema/documentation
// - paginated_fetch: Handle paginated responses
// - batch_request: Batch multiple requests
```

### Authentication Methods

Support for multiple authentication methods:

```go
// API Key Authentication
apiKeyCredentials := &dataintegration.DataSourceCredentials{
    Type:   dataintegration.CredentialTypeAPIKey,
    APIKey: "your-api-key",
    Custom: map[string]interface{}{
        "auth_type":   "api_key",
        "auth_header": "X-API-Key",
    },
}

// Bearer Token Authentication
bearerCredentials := &dataintegration.DataSourceCredentials{
    Type:  dataintegration.CredentialTypeBearer,
    Token: "your-bearer-token",
    Custom: map[string]interface{}{
        "auth_type":   "bearer",
        "auth_header": "Authorization",
    },
}

// Basic Authentication
basicCredentials := &dataintegration.DataSourceCredentials{
    Type:     dataintegration.CredentialTypeBasic,
    Username: "your-username",
    Password: "your-password",
    Custom: map[string]interface{}{
        "auth_type": "basic",
    },
}

// OAuth 2.0 Authentication
oauthCredentials := &dataintegration.DataSourceCredentials{
    Type: dataintegration.CredentialTypeOAuth2,
    OAuth: &dataintegration.OAuthCredentials{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        AccessToken:  "access-token",
        RefreshToken: "refresh-token",
        TokenType:    "Bearer",
        Scope:        []string{"read", "write"},
        ExpiresAt:    time.Now().Add(time.Hour),
    },
}
```

## 🔄 Data Processing Pipelines

### ETL Pipeline Configuration

Create comprehensive data processing pipelines with transformations:

```go
// Create data processing pipeline
pipeline := &dataintegration.DataPipeline{
    Name:         "Customer Data Processing Pipeline",
    Description:  "ETL pipeline for customer data with validation and enrichment",
    DataSourceID: dataSourceID,
    Status:       dataintegration.PipelineStatusStopped,
    Config: &dataintegration.PipelineConfig{
        BatchSize:        500,                                    // Process in batches
        Parallelism:      8,                                      // Parallel workers
        Timeout:          10 * time.Minute,                       // Pipeline timeout
        ErrorHandling:    dataintegration.ErrorHandlingSkip,      // Skip errors
        ValidationMode:   dataintegration.ValidationModeStrict,   // Strict validation
        DeduplicationKey: "customer_id",                          // Dedup field
    },
    Transformations: []*dataintegration.DataTransformation{
        {
            Name:        "Data Cleaning",
            Type:        dataintegration.TransformationTypeMap,
            Order:       1,
            Enabled:     true,
            Description: "Clean and normalize customer data",
            Config: map[string]interface{}{
                "trim_strings":      true,
                "normalize_emails":  true,
                "standardize_phone": true,
                "remove_duplicates": true,
                "convert_types":     true,
            },
        },
        {
            Name:        "Data Validation",
            Type:        dataintegration.TransformationTypeValidate,
            Order:       2,
            Enabled:     true,
            Description: "Validate customer data quality",
            Config: map[string]interface{}{
                "required_fields": []string{"customer_id", "email", "name"},
                "email_validation": true,
                "phone_validation": true,
                "min_name_length":  2,
                "max_name_length":  100,
            },
        },
        {
            Name:        "Data Enrichment",
            Type:        dataintegration.TransformationTypeEnrich,
            Order:       3,
            Enabled:     true,
            Description: "Enrich customer data with additional information",
            Config: map[string]interface{}{
                "add_timestamp":     true,
                "geocode_address":   true,
                "lookup_company":    true,
                "calculate_score":   true,
                "add_segments":      true,
            },
        },
        {
            Name:        "Data Aggregation",
            Type:        dataintegration.TransformationTypeAggregate,
            Order:       4,
            Enabled:     true,
            Description: "Aggregate customer metrics",
            Config: map[string]interface{}{
                "group_by":     []string{"segment", "region"},
                "metrics":      []string{"count", "avg_score", "total_value"},
                "time_window":  "daily",
            },
        },
    },
    Storage: &dataintegration.StorageConfig{
        Type:       dataintegration.StorageTypePostgreSQL,
        Connection: "postgresql://localhost:5432/customer_db",
        Database:   "customer_db",
        Table:      "customers_processed",
        IndexConfig: &dataintegration.IndexConfig{
            Fields:     []string{"customer_id", "email", "created_at"},
            Type:       "btree",
            Unique:     true,
            Background: true,
        },
        Partitioning: &dataintegration.PartitionConfig{
            Strategy: "time",
            Field:    "created_at",
            Options: map[string]interface{}{
                "interval": "monthly",
            },
        },
        Compression: "lz4",
    },
    Schedule: &dataintegration.ScheduleConfig{
        Type:       dataintegration.ScheduleTypeCron,
        Expression: "0 2 * * *",  // Daily at 2 AM
        Timezone:   "UTC",
        Enabled:    true,
    },
    CreatedBy: "data-engineer@company.com",
}

createdPipeline, err := dataEngine.CreatePipeline(pipeline)

// Start the pipeline
err = dataEngine.StartPipeline(createdPipeline.ID)

// Monitor pipeline status
status, err := dataEngine.GetPipelineStatus(createdPipeline.ID)
fmt.Printf("Pipeline status: %s\n", *status)

// Get pipeline metrics
metrics, err := dataEngine.GetPipelineMetrics(createdPipeline.ID, timeRange)
fmt.Printf("Pipeline processed %d records\n", metrics.RecordsProcessed)
```

### Data Transformation Types

```go
// Map Transformation - Data mapping and conversion
mapTransformation := &dataintegration.DataTransformation{
    Name: "Field Mapping",
    Type: dataintegration.TransformationTypeMap,
    Config: map[string]interface{}{
        "field_mappings": map[string]string{
            "customer_name": "name",
            "customer_email": "email",
            "customer_phone": "phone",
        },
        "type_conversions": map[string]string{
            "age": "integer",
            "score": "float",
            "active": "boolean",
        },
        "default_values": map[string]interface{}{
            "status": "active",
            "created_at": "now()",
        },
    },
}

// Filter Transformation - Data filtering and selection
filterTransformation := &dataintegration.DataTransformation{
    Name: "Data Filtering",
    Type: dataintegration.TransformationTypeFilter,
    Config: map[string]interface{}{
        "conditions": []map[string]interface{}{
            {"field": "age", "operator": ">=", "value": 18},
            {"field": "email", "operator": "contains", "value": "@"},
            {"field": "status", "operator": "in", "value": []string{"active", "pending"}},
        },
        "logic": "AND",
    },
}

// Validation Transformation - Data quality validation
validationTransformation := &dataintegration.DataTransformation{
    Name: "Quality Validation",
    Type: dataintegration.TransformationTypeValidate,
    Config: map[string]interface{}{
        "schema": map[string]interface{}{
            "customer_id": map[string]interface{}{
                "type": "string",
                "required": true,
                "pattern": "^CUST[0-9]{6}$",
            },
            "email": map[string]interface{}{
                "type": "string",
                "required": true,
                "format": "email",
            },
            "age": map[string]interface{}{
                "type": "integer",
                "minimum": 0,
                "maximum": 150,
            },
        },
        "error_handling": "skip_invalid",
        "validation_level": "strict",
    },
}

// Enrichment Transformation - Data enhancement
enrichmentTransformation := &dataintegration.DataTransformation{
    Name: "Data Enrichment",
    Type: dataintegration.TransformationTypeEnrich,
    Config: map[string]interface{}{
        "enrichments": []map[string]interface{}{
            {
                "type": "lookup",
                "source": "company_database",
                "key": "email_domain",
                "fields": ["company_name", "industry", "size"],
            },
            {
                "type": "geocoding",
                "address_field": "address",
                "output_fields": ["latitude", "longitude", "country", "region"],
            },
            {
                "type": "calculation",
                "formula": "score * weight",
                "output_field": "weighted_score",
            },
        ],
    },
}
```

## 💾 Storage Management

### Multi-Backend Storage Support

Support for various storage backends with automatic configuration:

```go
// PostgreSQL Storage
postgresStorage := &dataintegration.StorageConfig{
    Type:       dataintegration.StorageTypePostgreSQL,
    Connection: "postgresql://user:pass@localhost:5432/dbname",
    Database:   "analytics_db",
    Table:      "processed_data",
    IndexConfig: &dataintegration.IndexConfig{
        Fields:     []string{"id", "timestamp", "category"},
        Type:       "btree",
        Unique:     false,
        Background: true,
    },
    Partitioning: &dataintegration.PartitionConfig{
        Strategy: "time",
        Field:    "timestamp",
        Options: map[string]interface{}{
            "interval": "monthly",
            "retention": "12 months",
        },
    },
    Compression: "lz4",
}

// MongoDB Storage
mongoStorage := &dataintegration.StorageConfig{
    Type:       dataintegration.StorageTypeMongoDB,
    Connection: "mongodb://localhost:27017",
    Database:   "analytics_db",
    Collection: "processed_data",
    IndexConfig: &dataintegration.IndexConfig{
        Fields:     []string{"id", "timestamp", "category"},
        Type:       "compound",
        Unique:     true,
        Background: true,
    },
    Custom: map[string]interface{}{
        "write_concern": "majority",
        "read_preference": "secondaryPreferred",
        "max_pool_size": 100,
    },
}

// Elasticsearch Storage
elasticStorage := &dataintegration.StorageConfig{
    Type:       dataintegration.StorageTypeElastic,
    Connection: "http://localhost:9200",
    Database:   "analytics",
    Collection: "processed_data",
    IndexConfig: &dataintegration.IndexConfig{
        Fields: []string{"timestamp", "category", "content"},
        Type:   "text_search",
        Custom: map[string]interface{}{
            "mappings": map[string]interface{}{
                "timestamp": map[string]string{"type": "date"},
                "category":  map[string]string{"type": "keyword"},
                "content":   map[string]string{"type": "text", "analyzer": "standard"},
            },
        },
    },
}

// Redis Storage (for caching and real-time data)
redisStorage := &dataintegration.StorageConfig{
    Type:       dataintegration.StorageTypeRedis,
    Connection: "redis://localhost:6379",
    Database:   "0",
    Custom: map[string]interface{}{
        "key_prefix": "aios:data:",
        "ttl":        "24h",
        "max_memory": "2gb",
        "eviction_policy": "allkeys-lru",
    },
}

// S3 Storage (for data archiving)
s3Storage := &dataintegration.StorageConfig{
    Type:       dataintegration.StorageTypeS3,
    Connection: "s3://my-data-bucket",
    Custom: map[string]interface{}{
        "region":     "us-west-2",
        "access_key": "your-access-key",
        "secret_key": "your-secret-key",
        "encryption": "AES256",
        "storage_class": "STANDARD_IA",
        "lifecycle_policy": map[string]interface{}{
            "transition_to_ia": "30 days",
            "transition_to_glacier": "90 days",
            "expiration": "7 years",
        },
    },
}
```

## 📊 Monitoring and Analytics

### Health Monitoring

Comprehensive health monitoring for all data sources and pipelines:

```go
// Monitor data source health
health, err := dataEngine.GetDataSourceHealth(dataSourceID)
fmt.Printf("Data Source Health:\n")
fmt.Printf("  Status: %s\n", health.Status)
fmt.Printf("  Message: %s\n", health.Message)
fmt.Printf("  Last Check: %s\n", health.LastCheck)
fmt.Printf("  Uptime: %s\n", health.Uptime)

// Detailed health checks
for _, check := range health.Checks {
    fmt.Printf("  Check %s: %s - %s\n", check.Name, check.Status, check.Message)
}

// Set up health-based alerts
if health.Status == dataintegration.HealthStatusUnhealthy {
    // Send alert to operations team
    fmt.Printf("ALERT: Data source %s is unhealthy\n", dataSourceID)
    
    // Automatically disable problematic source
    err := dataEngine.DisableDataSource(dataSourceID)
    if err != nil {
        fmt.Printf("Failed to disable unhealthy data source: %v\n", err)
    }
}
```

### Performance Metrics

Detailed performance analytics and monitoring:

```go
// Get comprehensive metrics
timeRange := &dataintegration.TimeRange{
    Start: time.Now().Add(-7 * 24 * time.Hour), // Last 7 days
    End:   time.Now(),
}

metrics, err := dataEngine.GetDataSourceMetrics(dataSourceID, timeRange)
fmt.Printf("Data Source Metrics (7 days):\n")
fmt.Printf("  Records Extracted: %d\n", metrics.RecordsExtracted)
fmt.Printf("  Records Processed: %d\n", metrics.RecordsProcessed)
fmt.Printf("  Records Stored: %d\n", metrics.RecordsStored)
fmt.Printf("  Success Rate: %.2f%%\n", metrics.SuccessRate*100)
fmt.Printf("  Average Latency: %s\n", metrics.AverageLatency)
fmt.Printf("  P95 Latency: %s\n", metrics.P95Latency)
fmt.Printf("  P99 Latency: %s\n", metrics.P99Latency)
fmt.Printf("  Throughput: %.1f records/sec\n", metrics.ThroughputPerSec)

// Error analysis
fmt.Printf("Errors by type:\n")
for errorType, count := range metrics.ErrorsByType {
    fmt.Printf("  %s: %d\n", errorType, count)
}

// Performance alerts
if metrics.SuccessRate < 0.95 {
    fmt.Printf("ALERT: Success rate below 95%% for data source %s\n", dataSourceID)
}

if metrics.P95Latency > 10*time.Second {
    fmt.Printf("ALERT: High latency detected for data source %s\n", dataSourceID)
}

if metrics.ThroughputPerSec < 1.0 {
    fmt.Printf("ALERT: Low throughput for data source %s\n", dataSourceID)
}
```

### Logging and Audit

Comprehensive logging and audit capabilities:

```go
// Get detailed logs
logs, err := dataEngine.GetDataSourceLogs(dataSourceID, &dataintegration.LogFilter{
    Level:  dataintegration.LogLevelError,
    Since:  &[]time.Time{time.Now().Add(-24 * time.Hour)}[0],
    Limit:  100,
})

fmt.Printf("Error logs from last 24 hours:\n")
for _, log := range logs {
    fmt.Printf("[%s] %s: %s\n", log.Level, log.Operation, log.Message)
    if log.Error != "" {
        fmt.Printf("  Error: %s\n", log.Error)
    }
    if log.Duration > 0 {
        fmt.Printf("  Duration: %s\n", log.Duration)
    }
}

// Search logs by operation
operationLogs, err := dataEngine.GetDataSourceLogs(dataSourceID, &dataintegration.LogFilter{
    Operation: "data_extraction",
    Level:     dataintegration.LogLevelInfo,
    Limit:     50,
})

// Search logs by content
searchLogs, err := dataEngine.GetDataSourceLogs(dataSourceID, &dataintegration.LogFilter{
    Search: "timeout",
    Limit:  25,
})
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all data integration tests
go test ./pkg/dataintegration/...

# Run with race detection
go test -race ./pkg/dataintegration/...

# Run integration tests with external services
go test -tags=integration ./pkg/dataintegration/...

# Run external data integration example
go run examples/external_data_integration_example.go
```

## 📖 Examples

See the complete example in `examples/external_data_integration_example.go` for a comprehensive demonstration including:

- Data integration engine setup and configuration
- Web crawler and REST API connector registration
- Data source creation with various configurations
- Data processing pipeline creation with transformations
- Health monitoring and performance metrics
- Error handling and logging patterns
- Storage configuration for multiple backends

## 🤝 Contributing

1. Follow established patterns and interfaces
2. Add comprehensive tests for new connectors and features
3. Update documentation for API changes
4. Ensure proper error handling and logging
5. Use OpenTelemetry for observability
6. Implement proper security and rate limiting

## 📄 License

This External Data Integration system is part of the AIOS project and follows the same licensing terms.
