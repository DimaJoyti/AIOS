package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aios/aios/pkg/dataintegration"
	"github.com/aios/aios/pkg/dataintegration/connectors"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		ForceColors:      true,
	})

	fmt.Println("🔗 AIOS External Data Integration Demo")
	fmt.Println("=====================================")

	// Run the comprehensive demo
	if err := runExternalDataIntegrationDemo(logger); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n✅ External Data Integration Demo completed successfully!")
}

func runExternalDataIntegrationDemo(logger *logrus.Logger) error {
	// Step 1: Create Data Integration Engine
	fmt.Println("\n1. Creating Data Integration Engine...")
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
	fmt.Println("✓ Data Integration Engine created successfully")

	// Step 2: Register Data Connectors
	fmt.Println("\n2. Registering Data Connectors...")

	// Register Web Crawler connector
	webCrawler := connectors.NewWebCrawlerConnector()
	err := dataEngine.RegisterConnector(webCrawler)
	if err != nil {
		return fmt.Errorf("failed to register web crawler: %w", err)
	}
	fmt.Printf("   ✓ Web Crawler registered: %s v%s\n",
		webCrawler.GetName(), webCrawler.GetVersion())

	// Register REST API connector
	restAPI := connectors.NewRestAPIConnector()
	err = dataEngine.RegisterConnector(restAPI)
	if err != nil {
		return fmt.Errorf("failed to register REST API connector: %w", err)
	}
	fmt.Printf("   ✓ REST API Connector registered: %s v%s\n",
		restAPI.GetName(), restAPI.GetVersion())

	// List registered connectors
	connectorTypes := dataEngine.ListConnectors()
	fmt.Printf("   ✓ Total connectors registered: %d\n", len(connectorTypes))
	for _, connectorType := range connectorTypes {
		connector, _ := dataEngine.GetConnector(connectorType)
		fmt.Printf("     - %s: %s\n", connectorType, connector.GetDescription())
	}

	// Step 3: Create Web Crawler Data Source
	fmt.Println("\n3. Creating Web Crawler Data Source...")

	webCrawlerSource := &dataintegration.DataSource{
		Name:        "News Website Crawler",
		Description: "Web crawler for extracting news articles and content",
		Type:        "webcrawler",
		Provider:    "WebCrawler",
		Status:      dataintegration.DataSourceStatusConfiguring,
		Config: &dataintegration.DataSourceConfig{
			URL:     "https://example-news.com",
			Timeout: 30 * time.Second,
			RetryPolicy: &dataintegration.RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  1 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
			},
			RateLimit: &dataintegration.RateLimit{
				RequestsPerSecond: 2, // Respectful crawling
				BurstSize:         5,
				WindowSize:        time.Minute,
			},
			CrawlConfig: &dataintegration.CrawlConfig{
				MaxDepth:        3,
				MaxPages:        50,
				FollowRedirects: true,
				RespectRobots:   true,
				UserAgent:       "AIOS-DataCrawler/1.0",
				Delay:           2 * time.Second,
				Selectors:       []string{"article", "h1", "h2", "p", ".content"},
				ExcludePatterns: []string{"/admin", "/login", "/api"},
				IncludePatterns: []string{"/news", "/articles"},
			},
			Custom: map[string]interface{}{
				"base_url":         "https://example-news.com",
				"user_agent":       "AIOS-DataCrawler/1.0",
				"max_depth":        3,
				"max_pages":        50,
				"delay":            "2s",
				"selectors":        []string{"article", "h1", "h2", "p"},
				"exclude_patterns": []string{"/admin", "/login"},
				"include_patterns": []string{"/news", "/articles"},
				"respect_robots":   true,
			},
		},
		Settings: &dataintegration.DataSourceSettings{
			AutoSync:        true,
			SyncInterval:    6 * time.Hour, // Crawl every 6 hours
			EnableStreaming: false,
			EnableEvents:    true,
			DataRetention:   30 * 24 * time.Hour,
			MaxRecords:      10000,
			LogLevel:        "info",
			NotifyOnError:   true,
			ErrorRecipients: []string{"data-team@company.com"},
		},
		CreatedBy: "admin@company.com",
	}

	createdWebCrawlerSource, err := dataEngine.CreateDataSource(webCrawlerSource)
	if err != nil {
		return fmt.Errorf("failed to create web crawler data source: %w", err)
	}

	fmt.Printf("   ✓ Web Crawler Data Source Created: %s (ID: %s)\n",
		createdWebCrawlerSource.Name, createdWebCrawlerSource.ID)
	fmt.Printf("     - Type: %s\n", createdWebCrawlerSource.Type)
	fmt.Printf("     - Provider: %s\n", createdWebCrawlerSource.Provider)
	fmt.Printf("     - Status: %s\n", createdWebCrawlerSource.Status)
	fmt.Printf("     - Max Depth: %d\n", createdWebCrawlerSource.Config.CrawlConfig.MaxDepth)
	fmt.Printf("     - Max Pages: %d\n", createdWebCrawlerSource.Config.CrawlConfig.MaxPages)
	fmt.Printf("     - Sync Interval: %s\n", createdWebCrawlerSource.Settings.SyncInterval)

	// Step 4: Create REST API Data Source
	fmt.Println("\n4. Creating REST API Data Source...")

	restAPISource := &dataintegration.DataSource{
		Name:        "JSONPlaceholder API",
		Description: "REST API for fetching sample posts and user data",
		Type:        "restapi",
		Provider:    "JSONPlaceholder",
		Status:      dataintegration.DataSourceStatusConfiguring,
		Config: &dataintegration.DataSourceConfig{
			URL:     "https://jsonplaceholder.typicode.com",
			Timeout: 30 * time.Second,
			RetryPolicy: &dataintegration.RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  1 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
			},
			RateLimit: &dataintegration.RateLimit{
				RequestsPerSecond: 10,
				BurstSize:         20,
				WindowSize:        time.Minute,
			},
			Custom: map[string]interface{}{
				"base_url":         "https://jsonplaceholder.typicode.com",
				"timeout":          "30s",
				"rate_limit":       10,
				"pagination_type":  "offset",
				"pagination_limit": 100,
				"headers": map[string]interface{}{
					"Accept":     "application/json",
					"User-Agent": "AIOS-DataIntegration/1.0",
				},
			},
		},
		Credentials: &dataintegration.DataSourceCredentials{
			Type: dataintegration.CredentialTypeCustom,
			Custom: map[string]interface{}{
				"auth_type": "none", // Public API
			},
		},
		Settings: &dataintegration.DataSourceSettings{
			AutoSync:        true,
			SyncInterval:    1 * time.Hour, // Sync every hour
			EnableStreaming: false,
			EnableEvents:    true,
			DataRetention:   7 * 24 * time.Hour,
			MaxRecords:      50000,
			LogLevel:        "info",
			NotifyOnError:   true,
			ErrorRecipients: []string{"api-team@company.com"},
		},
		CreatedBy: "admin@company.com",
	}

	createdRestAPISource, err := dataEngine.CreateDataSource(restAPISource)
	if err != nil {
		return fmt.Errorf("failed to create REST API data source: %w", err)
	}

	fmt.Printf("   ✓ REST API Data Source Created: %s (ID: %s)\n",
		createdRestAPISource.Name, createdRestAPISource.ID)
	fmt.Printf("     - Type: %s\n", createdRestAPISource.Type)
	fmt.Printf("     - Provider: %s\n", createdRestAPISource.Provider)
	fmt.Printf("     - Status: %s\n", createdRestAPISource.Status)
	fmt.Printf("     - Base URL: %s\n", createdRestAPISource.Config.URL)
	fmt.Printf("     - Sync Interval: %s\n", createdRestAPISource.Settings.SyncInterval)

	// Step 5: Test Data Source Connections
	fmt.Println("\n5. Testing Data Source Connections...")

	// Test web crawler connection
	webCrawlerTestResult, err := dataEngine.TestDataSource(createdWebCrawlerSource.ID)
	if err != nil {
		return fmt.Errorf("failed to test web crawler: %w", err)
	}

	fmt.Printf("   ✓ Web Crawler Connection Test:\n")
	fmt.Printf("     - Success: %t\n", webCrawlerTestResult.Success)
	fmt.Printf("     - Message: %s\n", webCrawlerTestResult.Message)
	fmt.Printf("     - Duration: %s\n", webCrawlerTestResult.Duration)
	fmt.Printf("     - Capabilities: %v\n", webCrawlerTestResult.Capabilities)

	// Test REST API connection
	restAPITestResult, err := dataEngine.TestDataSource(createdRestAPISource.ID)
	if err != nil {
		return fmt.Errorf("failed to test REST API: %w", err)
	}

	fmt.Printf("   ✓ REST API Connection Test:\n")
	fmt.Printf("     - Success: %t\n", restAPITestResult.Success)
	fmt.Printf("     - Message: %s\n", restAPITestResult.Message)
	fmt.Printf("     - Duration: %s\n", restAPITestResult.Duration)
	fmt.Printf("     - Capabilities: %v\n", restAPITestResult.Capabilities)

	// Step 6: Create Data Processing Pipelines
	fmt.Println("\n6. Creating Data Processing Pipelines...")

	// Web crawler pipeline
	webCrawlerPipeline := &dataintegration.DataPipeline{
		Name:         "News Content Processing Pipeline",
		Description:  "Pipeline for processing crawled news content with NLP and categorization",
		DataSourceID: createdWebCrawlerSource.ID,
		Status:       dataintegration.PipelineStatusStopped,
		Config: &dataintegration.PipelineConfig{
			BatchSize:        100,
			Parallelism:      4,
			Timeout:          5 * time.Minute,
			ErrorHandling:    dataintegration.ErrorHandlingSkip,
			ValidationMode:   dataintegration.ValidationModeWarn,
			DeduplicationKey: "url",
		},
		Transformations: []*dataintegration.DataTransformation{
			{
				Name:        "Content Extraction",
				Type:        dataintegration.TransformationTypeMap,
				Order:       1,
				Enabled:     true,
				Description: "Extract and clean article content",
				Config: map[string]interface{}{
					"extract_fields": []string{"title", "content", "author", "date"},
					"clean_html":     true,
					"min_length":     100,
				},
			},
			{
				Name:        "Content Validation",
				Type:        dataintegration.TransformationTypeValidate,
				Order:       2,
				Enabled:     true,
				Description: "Validate article content quality",
				Config: map[string]interface{}{
					"required_fields":    []string{"title", "content"},
					"min_content_length": 200,
					"max_content_length": 50000,
				},
			},
			{
				Name:        "Content Enrichment",
				Type:        dataintegration.TransformationTypeEnrich,
				Order:       3,
				Enabled:     true,
				Description: "Add metadata and categorization",
				Config: map[string]interface{}{
					"add_timestamp":      true,
					"extract_keywords":   true,
					"categorize":         true,
					"sentiment_analysis": true,
				},
			},
		},
		Storage: &dataintegration.StorageConfig{
			Type:       dataintegration.StorageTypePostgreSQL,
			Connection: "postgresql://localhost:5432/aios_data",
			Database:   "aios_data",
			Table:      "news_articles",
			IndexConfig: &dataintegration.IndexConfig{
				Fields:     []string{"title", "url", "created_at"},
				Type:       "btree",
				Unique:     false,
				Background: true,
			},
			Compression: "gzip",
		},
		Schedule: &dataintegration.ScheduleConfig{
			Type:       dataintegration.ScheduleTypeCron,
			Expression: "0 */6 * * *", // Every 6 hours
			Timezone:   "UTC",
			Enabled:    true,
		},
		CreatedBy: "admin@company.com",
	}

	createdWebCrawlerPipeline, err := dataEngine.CreatePipeline(webCrawlerPipeline)
	if err != nil {
		return fmt.Errorf("failed to create web crawler pipeline: %w", err)
	}

	fmt.Printf("   ✓ Web Crawler Pipeline Created: %s (ID: %s)\n",
		createdWebCrawlerPipeline.Name, createdWebCrawlerPipeline.ID)
	fmt.Printf("     - Data Source: %s\n", createdWebCrawlerPipeline.DataSourceID)
	fmt.Printf("     - Status: %s\n", createdWebCrawlerPipeline.Status)
	fmt.Printf("     - Transformations: %d\n", len(createdWebCrawlerPipeline.Transformations))
	fmt.Printf("     - Storage Type: %s\n", createdWebCrawlerPipeline.Storage.Type)
	fmt.Printf("     - Schedule: %s\n", createdWebCrawlerPipeline.Schedule.Expression)

	// REST API pipeline
	restAPIPipeline := &dataintegration.DataPipeline{
		Name:         "API Data Processing Pipeline",
		Description:  "Pipeline for processing REST API data with validation and enrichment",
		DataSourceID: createdRestAPISource.ID,
		Status:       dataintegration.PipelineStatusStopped,
		Config: &dataintegration.PipelineConfig{
			BatchSize:        200,
			Parallelism:      2,
			Timeout:          3 * time.Minute,
			ErrorHandling:    dataintegration.ErrorHandlingRetry,
			ValidationMode:   dataintegration.ValidationModeStrict,
			DeduplicationKey: "id",
		},
		Transformations: []*dataintegration.DataTransformation{
			{
				Name:        "Data Normalization",
				Type:        dataintegration.TransformationTypeMap,
				Order:       1,
				Enabled:     true,
				Description: "Normalize API response data",
				Config: map[string]interface{}{
					"normalize_fields": true,
					"convert_types":    true,
					"trim_strings":     true,
				},
			},
			{
				Name:        "Data Validation",
				Type:        dataintegration.TransformationTypeValidate,
				Order:       2,
				Enabled:     true,
				Description: "Validate API data structure",
				Config: map[string]interface{}{
					"schema_validation": true,
					"required_fields":   []string{"id", "title"},
					"data_types": map[string]string{
						"id":     "integer",
						"title":  "string",
						"userId": "integer",
					},
				},
			},
		},
		Storage: &dataintegration.StorageConfig{
			Type:       dataintegration.StorageTypeMongoDB,
			Connection: "mongodb://localhost:27017",
			Database:   "aios_data",
			Collection: "api_data",
			IndexConfig: &dataintegration.IndexConfig{
				Fields:     []string{"id", "userId", "createdAt"},
				Type:       "compound",
				Unique:     true,
				Background: true,
			},
		},
		Schedule: &dataintegration.ScheduleConfig{
			Type:       dataintegration.ScheduleTypeInterval,
			Expression: "1h", // Every hour
			Timezone:   "UTC",
			Enabled:    true,
		},
		CreatedBy: "admin@company.com",
	}

	createdRestAPIPipeline, err := dataEngine.CreatePipeline(restAPIPipeline)
	if err != nil {
		return fmt.Errorf("failed to create REST API pipeline: %w", err)
	}

	fmt.Printf("   ✓ REST API Pipeline Created: %s (ID: %s)\n",
		createdRestAPIPipeline.Name, createdRestAPIPipeline.ID)
	fmt.Printf("     - Data Source: %s\n", createdRestAPIPipeline.DataSourceID)
	fmt.Printf("     - Status: %s\n", createdRestAPIPipeline.Status)
	fmt.Printf("     - Transformations: %d\n", len(createdRestAPIPipeline.Transformations))
	fmt.Printf("     - Storage Type: %s\n", createdRestAPIPipeline.Storage.Type)
	fmt.Printf("     - Schedule: %s\n", createdRestAPIPipeline.Schedule.Expression)

	// Step 7: Data Source Management
	fmt.Println("\n7. Managing Data Sources...")

	// List all data sources
	allDataSources, err := dataEngine.ListDataSources(&dataintegration.DataSourceFilter{
		Limit: 10,
	})
	if err != nil {
		return fmt.Errorf("failed to list data sources: %w", err)
	}

	fmt.Printf("   ✓ Active Data Sources (%d total):\n", len(allDataSources))
	for _, source := range allDataSources {
		fmt.Printf("     - %s (%s): %s\n",
			source.Name, source.Type, source.Status)
	}

	// List all pipelines
	allPipelines, err := dataEngine.ListPipelines(&dataintegration.PipelineFilter{
		Limit: 10,
	})
	if err != nil {
		return fmt.Errorf("failed to list pipelines: %w", err)
	}

	fmt.Printf("   ✓ Active Pipelines (%d total):\n", len(allPipelines))
	for _, pipeline := range allPipelines {
		fmt.Printf("     - %s: %s\n",
			pipeline.Name, pipeline.Status)
	}

	// Step 8: Health Monitoring
	fmt.Println("\n8. Monitoring Data Source Health...")

	// Get web crawler health
	webCrawlerHealth, err := dataEngine.GetDataSourceHealth(createdWebCrawlerSource.ID)
	if err != nil {
		return fmt.Errorf("failed to get web crawler health: %w", err)
	}

	fmt.Printf("   ✓ Web Crawler Health:\n")
	fmt.Printf("     - Status: %s\n", webCrawlerHealth.Status)
	fmt.Printf("     - Message: %s\n", webCrawlerHealth.Message)
	fmt.Printf("     - Last Check: %s\n", webCrawlerHealth.LastCheck.Format("15:04:05"))
	fmt.Printf("     - Uptime: %s\n", webCrawlerHealth.Uptime)
	fmt.Printf("     - Health Checks: %d\n", len(webCrawlerHealth.Checks))

	// Get REST API health
	restAPIHealth, err := dataEngine.GetDataSourceHealth(createdRestAPISource.ID)
	if err != nil {
		return fmt.Errorf("failed to get REST API health: %w", err)
	}

	fmt.Printf("   ✓ REST API Health:\n")
	fmt.Printf("     - Status: %s\n", restAPIHealth.Status)
	fmt.Printf("     - Message: %s\n", restAPIHealth.Message)
	fmt.Printf("     - Last Check: %s\n", restAPIHealth.LastCheck.Format("15:04:05"))
	fmt.Printf("     - Uptime: %s\n", restAPIHealth.Uptime)

	// Step 9: Performance Metrics
	fmt.Println("\n9. Data Integration Analytics and Metrics...")

	// Get data source metrics
	timeRange := &dataintegration.TimeRange{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}

	webCrawlerMetrics, err := dataEngine.GetDataSourceMetrics(createdWebCrawlerSource.ID, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get web crawler metrics: %w", err)
	}

	fmt.Printf("   ✓ Web Crawler Metrics (24h):\n")
	fmt.Printf("     - Records Extracted: %d\n", webCrawlerMetrics.RecordsExtracted)
	fmt.Printf("     - Records Processed: %d\n", webCrawlerMetrics.RecordsProcessed)
	fmt.Printf("     - Success Rate: %.1f%%\n", webCrawlerMetrics.SuccessRate*100)
	fmt.Printf("     - Average Latency: %s\n", webCrawlerMetrics.AverageLatency)
	fmt.Printf("     - Throughput: %.1f records/sec\n", webCrawlerMetrics.ThroughputPerSec)

	restAPIMetrics, err := dataEngine.GetDataSourceMetrics(createdRestAPISource.ID, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get REST API metrics: %w", err)
	}

	fmt.Printf("   ✓ REST API Metrics (24h):\n")
	fmt.Printf("     - Records Extracted: %d\n", restAPIMetrics.RecordsExtracted)
	fmt.Printf("     - Records Processed: %d\n", restAPIMetrics.RecordsProcessed)
	fmt.Printf("     - Success Rate: %.1f%%\n", restAPIMetrics.SuccessRate*100)
	fmt.Printf("     - Average Latency: %s\n", restAPIMetrics.AverageLatency)
	fmt.Printf("     - Throughput: %.1f records/sec\n", restAPIMetrics.ThroughputPerSec)

	// Get data source logs
	webCrawlerLogs, err := dataEngine.GetDataSourceLogs(createdWebCrawlerSource.ID, &dataintegration.LogFilter{
		Level: dataintegration.LogLevelInfo,
		Limit: 5,
	})
	if err != nil {
		return fmt.Errorf("failed to get web crawler logs: %w", err)
	}

	fmt.Printf("   ✓ Recent Web Crawler Logs (%d entries):\n", len(webCrawlerLogs))
	for i, logEntry := range webCrawlerLogs {
		if i >= 3 { // Show only first 3
			break
		}
		fmt.Printf("     %d. [%s] %s: %s\n",
			i+1, logEntry.Level, logEntry.Operation, logEntry.Message)
	}

	return nil
}
