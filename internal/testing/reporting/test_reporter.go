package reporting

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// TestReporter provides comprehensive test reporting capabilities
type TestReporter struct {
	config ReportConfig
	results []TestResult
}

// ReportConfig defines configuration for test reporting
type ReportConfig struct {
	OutputDir     string   `json:"output_dir"`
	Formats       []string `json:"formats"` // html, json, xml, junit
	IncludeCoverage bool   `json:"include_coverage"`
	IncludeMetrics  bool   `json:"include_metrics"`
	Theme           string `json:"theme"` // light, dark
}

// TestResult represents a comprehensive test result
type TestResult struct {
	Suite       string        `json:"suite"`
	Name        string        `json:"name"`
	Status      TestStatus    `json:"status"`
	Duration    time.Duration `json:"duration"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Error       string        `json:"error,omitempty"`
	Output      string        `json:"output,omitempty"`
	Assertions  int           `json:"assertions"`
	Coverage    *Coverage     `json:"coverage,omitempty"`
	Metrics     *TestMetrics  `json:"metrics,omitempty"`
	Tags        []string      `json:"tags"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// TestStatus represents the status of a test
type TestStatus string

const (
	StatusPassed  TestStatus = "passed"
	StatusFailed  TestStatus = "failed"
	StatusSkipped TestStatus = "skipped"
	StatusError   TestStatus = "error"
)

// Coverage represents test coverage information
type Coverage struct {
	Lines      float64 `json:"lines"`
	Functions  float64 `json:"functions"`
	Branches   float64 `json:"branches"`
	Statements float64 `json:"statements"`
}

// TestMetrics represents test performance metrics
type TestMetrics struct {
	MemoryUsage    int64         `json:"memory_usage"`
	CPUTime        time.Duration `json:"cpu_time"`
	Allocations    int64         `json:"allocations"`
	GoroutineCount int           `json:"goroutine_count"`
}

// TestSummary represents a summary of test results
type TestSummary struct {
	TotalTests    int           `json:"total_tests"`
	PassedTests   int           `json:"passed_tests"`
	FailedTests   int           `json:"failed_tests"`
	SkippedTests  int           `json:"skipped_tests"`
	ErrorTests    int           `json:"error_tests"`
	TotalDuration time.Duration `json:"total_duration"`
	SuccessRate   float64       `json:"success_rate"`
	Coverage      *Coverage     `json:"coverage,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
}

// NewTestReporter creates a new test reporter
func NewTestReporter(config ReportConfig) *TestReporter {
	if config.OutputDir == "" {
		config.OutputDir = "test-reports"
	}
	if len(config.Formats) == 0 {
		config.Formats = []string{"html", "json"}
	}
	if config.Theme == "" {
		config.Theme = "light"
	}
	
	return &TestReporter{
		config:  config,
		results: make([]TestResult, 0),
	}
}

// AddResult adds a test result
func (tr *TestReporter) AddResult(result TestResult) {
	tr.results = append(tr.results, result)
}

// AddResults adds multiple test results
func (tr *TestReporter) AddResults(results []TestResult) {
	tr.results = append(tr.results, results...)
}

// GenerateReports generates reports in all configured formats
func (tr *TestReporter) GenerateReports() error {
	// Ensure output directory exists
	if err := os.MkdirAll(tr.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	summary := tr.calculateSummary()
	
	for _, format := range tr.config.Formats {
		switch strings.ToLower(format) {
		case "html":
			if err := tr.generateHTMLReport(summary); err != nil {
				return fmt.Errorf("failed to generate HTML report: %w", err)
			}
		case "json":
			if err := tr.generateJSONReport(summary); err != nil {
				return fmt.Errorf("failed to generate JSON report: %w", err)
			}
		case "xml":
			if err := tr.generateXMLReport(summary); err != nil {
				return fmt.Errorf("failed to generate XML report: %w", err)
			}
		case "junit":
			if err := tr.generateJUnitReport(summary); err != nil {
				return fmt.Errorf("failed to generate JUnit report: %w", err)
			}
		default:
			return fmt.Errorf("unsupported report format: %s", format)
		}
	}
	
	return nil
}

// calculateSummary calculates test summary statistics
func (tr *TestReporter) calculateSummary() TestSummary {
	summary := TestSummary{
		TotalTests: len(tr.results),
		Timestamp:  time.Now(),
	}
	
	var totalDuration time.Duration
	var totalLines, totalFunctions, totalBranches, totalStatements float64
	var coverageCount int
	
	for _, result := range tr.results {
		totalDuration += result.Duration
		
		switch result.Status {
		case StatusPassed:
			summary.PassedTests++
		case StatusFailed:
			summary.FailedTests++
		case StatusSkipped:
			summary.SkippedTests++
		case StatusError:
			summary.ErrorTests++
		}
		
		if result.Coverage != nil {
			totalLines += result.Coverage.Lines
			totalFunctions += result.Coverage.Functions
			totalBranches += result.Coverage.Branches
			totalStatements += result.Coverage.Statements
			coverageCount++
		}
	}
	
	summary.TotalDuration = totalDuration
	
	if summary.TotalTests > 0 {
		summary.SuccessRate = float64(summary.PassedTests) / float64(summary.TotalTests) * 100
	}
	
	if coverageCount > 0 {
		summary.Coverage = &Coverage{
			Lines:      totalLines / float64(coverageCount),
			Functions:  totalFunctions / float64(coverageCount),
			Branches:   totalBranches / float64(coverageCount),
			Statements: totalStatements / float64(coverageCount),
		}
	}
	
	return summary
}

// generateHTMLReport generates an HTML report
func (tr *TestReporter) generateHTMLReport(summary TestSummary) error {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .summary { display: flex; gap: 20px; margin: 20px 0; }
        .metric { background: white; padding: 15px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .passed { color: #28a745; }
        .failed { color: #dc3545; }
        .skipped { color: #ffc107; }
        .error { color: #fd7e14; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #f8f9fa; }
        .status-passed { background: #d4edda; }
        .status-failed { background: #f8d7da; }
        .status-skipped { background: #fff3cd; }
        .status-error { background: #f5c6cb; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Test Report</h1>
        <p>Generated on {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
    </div>
    
    <div class="summary">
        <div class="metric">
            <h3>Total Tests</h3>
            <p>{{.TotalTests}}</p>
        </div>
        <div class="metric">
            <h3>Success Rate</h3>
            <p>{{printf "%.1f%%" .SuccessRate}}</p>
        </div>
        <div class="metric">
            <h3>Duration</h3>
            <p>{{.TotalDuration}}</p>
        </div>
        {{if .Coverage}}
        <div class="metric">
            <h3>Coverage</h3>
            <p>{{printf "%.1f%%" .Coverage.Lines}}</p>
        </div>
        {{end}}
    </div>
    
    <table>
        <thead>
            <tr>
                <th>Suite</th>
                <th>Test</th>
                <th>Status</th>
                <th>Duration</th>
                <th>Error</th>
            </tr>
        </thead>
        <tbody>
            {{range .Results}}
            <tr class="status-{{.Status}}">
                <td>{{.Suite}}</td>
                <td>{{.Name}}</td>
                <td class="{{.Status}}">{{.Status}}</td>
                <td>{{.Duration}}</td>
                <td>{{.Error}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</body>
</html>
`
	
	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return err
	}
	
	data := struct {
		TestSummary
		Results []TestResult
	}{
		TestSummary: summary,
		Results:     tr.results,
	}
	
	file, err := os.Create(filepath.Join(tr.config.OutputDir, "report.html"))
	if err != nil {
		return err
	}
	defer file.Close()
	
	return t.Execute(file, data)
}

// generateJSONReport generates a JSON report
func (tr *TestReporter) generateJSONReport(summary TestSummary) error {
	report := struct {
		Summary TestSummary  `json:"summary"`
		Results []TestResult `json:"results"`
	}{
		Summary: summary,
		Results: tr.results,
	}
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath.Join(tr.config.OutputDir, "report.json"), data, 0644)
}

// generateXMLReport generates an XML report
func (tr *TestReporter) generateXMLReport(summary TestSummary) error {
	// Simplified XML generation
	var xml strings.Builder
	
	xml.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	xml.WriteString(`<testsuites>` + "\n")
	
	// Group by suite
	suites := make(map[string][]TestResult)
	for _, result := range tr.results {
		suites[result.Suite] = append(suites[result.Suite], result)
	}
	
	for suiteName, suiteResults := range suites {
		xml.WriteString(fmt.Sprintf(`  <testsuite name="%s" tests="%d">`, suiteName, len(suiteResults)) + "\n")
		
		for _, result := range suiteResults {
			xml.WriteString(fmt.Sprintf(`    <testcase name="%s" time="%.3f">`,
				result.Name, result.Duration.Seconds()) + "\n")
			
			if result.Status == StatusFailed || result.Status == StatusError {
				xml.WriteString(fmt.Sprintf(`      <failure message="%s">%s</failure>`,
					result.Error, result.Output) + "\n")
			} else if result.Status == StatusSkipped {
				xml.WriteString(`      <skipped/>` + "\n")
			}
			
			xml.WriteString(`    </testcase>` + "\n")
		}
		
		xml.WriteString(`  </testsuite>` + "\n")
	}
	
	xml.WriteString(`</testsuites>` + "\n")
	
	return os.WriteFile(filepath.Join(tr.config.OutputDir, "report.xml"), []byte(xml.String()), 0644)
}

// generateJUnitReport generates a JUnit-compatible XML report
func (tr *TestReporter) generateJUnitReport(summary TestSummary) error {
	// JUnit format is similar to XML but with specific structure
	return tr.generateXMLReport(summary) // Simplified for now
}

// GetSummary returns the test summary
func (tr *TestReporter) GetSummary() TestSummary {
	return tr.calculateSummary()
}

// GetResults returns all test results
func (tr *TestReporter) GetResults() []TestResult {
	return tr.results
}

// GetResultsBySuite returns results grouped by suite
func (tr *TestReporter) GetResultsBySuite() map[string][]TestResult {
	suites := make(map[string][]TestResult)
	for _, result := range tr.results {
		suites[result.Suite] = append(suites[result.Suite], result)
	}
	return suites
}

// GetFailedResults returns only failed test results
func (tr *TestReporter) GetFailedResults() []TestResult {
	var failed []TestResult
	for _, result := range tr.results {
		if result.Status == StatusFailed || result.Status == StatusError {
			failed = append(failed, result)
		}
	}
	return failed
}

// GetSlowestTests returns the slowest test results
func (tr *TestReporter) GetSlowestTests(limit int) []TestResult {
	results := make([]TestResult, len(tr.results))
	copy(results, tr.results)
	
	sort.Slice(results, func(i, j int) bool {
		return results[i].Duration > results[j].Duration
	})
	
	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}
	
	return results
}

// Clear clears all test results
func (tr *TestReporter) Clear() {
	tr.results = make([]TestResult, 0)
}

// TestReportBuilder provides a fluent interface for building test results
type TestReportBuilder struct {
	result TestResult
}

// NewTestReportBuilder creates a new test report builder
func NewTestReportBuilder() *TestReportBuilder {
	return &TestReportBuilder{
		result: TestResult{
			StartTime:  time.Now(),
			Properties: make(map[string]interface{}),
		},
	}
}

// Suite sets the test suite name
func (b *TestReportBuilder) Suite(suite string) *TestReportBuilder {
	b.result.Suite = suite
	return b
}

// Name sets the test name
func (b *TestReportBuilder) Name(name string) *TestReportBuilder {
	b.result.Name = name
	return b
}

// Status sets the test status
func (b *TestReportBuilder) Status(status TestStatus) *TestReportBuilder {
	b.result.Status = status
	return b
}

// Duration sets the test duration
func (b *TestReportBuilder) Duration(duration time.Duration) *TestReportBuilder {
	b.result.Duration = duration
	return b
}

// Error sets the test error
func (b *TestReportBuilder) Error(err string) *TestReportBuilder {
	b.result.Error = err
	return b
}

// Output sets the test output
func (b *TestReportBuilder) Output(output string) *TestReportBuilder {
	b.result.Output = output
	return b
}

// Coverage sets the test coverage
func (b *TestReportBuilder) Coverage(coverage *Coverage) *TestReportBuilder {
	b.result.Coverage = coverage
	return b
}

// Metrics sets the test metrics
func (b *TestReportBuilder) Metrics(metrics *TestMetrics) *TestReportBuilder {
	b.result.Metrics = metrics
	return b
}

// Tags sets the test tags
func (b *TestReportBuilder) Tags(tags ...string) *TestReportBuilder {
	b.result.Tags = tags
	return b
}

// Property sets a custom property
func (b *TestReportBuilder) Property(key string, value interface{}) *TestReportBuilder {
	b.result.Properties[key] = value
	return b
}

// Build builds the test result
func (b *TestReportBuilder) Build() TestResult {
	b.result.EndTime = time.Now()
	if b.result.Duration == 0 {
		b.result.Duration = b.result.EndTime.Sub(b.result.StartTime)
	}
	return b.result
}
