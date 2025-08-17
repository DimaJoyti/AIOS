package coverage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CoverageAnalyzer analyzes test coverage data
type CoverageAnalyzer struct {
	profiles        []*CoverageProfile
	thresholds      *CoverageThresholds
	excludePatterns []*regexp.Regexp
}

// CoverageProfile represents coverage data for a file
type CoverageProfile struct {
	FileName     string
	Mode         string
	Blocks       []*CoverageBlock
	TotalLines   int
	CoveredLines int
	Percentage   float64
}

// CoverageBlock represents a coverage block
type CoverageBlock struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	NumStmt   int
	Count     int
}

// CoverageThresholds defines minimum coverage requirements
type CoverageThresholds struct {
	Overall  float64 `json:"overall"`
	Package  float64 `json:"package"`
	File     float64 `json:"file"`
	Function float64 `json:"function"`
	Line     float64 `json:"line"`
	Branch   float64 `json:"branch"`
}

// CoverageReport represents a comprehensive coverage report
type CoverageReport struct {
	Timestamp  time.Time                   `json:"timestamp"`
	Overall    *CoverageSummary            `json:"overall"`
	Packages   map[string]*CoverageSummary `json:"packages"`
	Files      map[string]*CoverageSummary `json:"files"`
	Violations []*CoverageViolation        `json:"violations"`
	Trends     *CoverageTrends             `json:"trends,omitempty"`
}

// CoverageSummary provides summary statistics
type CoverageSummary struct {
	TotalLines    int     `json:"total_lines"`
	CoveredLines  int     `json:"covered_lines"`
	Percentage    float64 `json:"percentage"`
	TotalBlocks   int     `json:"total_blocks"`
	CoveredBlocks int     `json:"covered_blocks"`
	Functions     int     `json:"functions"`
	CoveredFuncs  int     `json:"covered_functions"`
}

// CoverageViolation represents a coverage threshold violation
type CoverageViolation struct {
	Type     string  `json:"type"`
	Target   string  `json:"target"`
	Actual   float64 `json:"actual"`
	Required float64 `json:"required"`
	Severity string  `json:"severity"`
	Message  string  `json:"message"`
}

// CoverageTrends tracks coverage changes over time
type CoverageTrends struct {
	Previous *CoverageSummary `json:"previous"`
	Current  *CoverageSummary `json:"current"`
	Change   float64          `json:"change"`
	Trend    string           `json:"trend"` // "improving", "declining", "stable"
}

// NewCoverageAnalyzer creates a new coverage analyzer
func NewCoverageAnalyzer() *CoverageAnalyzer {
	return &CoverageAnalyzer{
		profiles:        make([]*CoverageProfile, 0),
		excludePatterns: make([]*regexp.Regexp, 0),
	}
}

// SetThresholds sets coverage thresholds
func (c *CoverageAnalyzer) SetThresholds(thresholds *CoverageThresholds) {
	c.thresholds = thresholds
}

// AddExcludePattern adds a pattern to exclude from coverage analysis
func (c *CoverageAnalyzer) AddExcludePattern(pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid exclude pattern %s: %w", pattern, err)
	}
	c.excludePatterns = append(c.excludePatterns, regex)
	return nil
}

// LoadCoverageProfile loads coverage data from a Go coverage profile
func (c *CoverageAnalyzer) LoadCoverageProfile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open coverage file: %w", err)
	}
	defer file.Close()

	return c.ParseCoverageProfile(file)
}

// ParseCoverageProfile parses coverage data from a reader
func (c *CoverageAnalyzer) ParseCoverageProfile(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)

	// Skip the mode line
	if !scanner.Scan() {
		return fmt.Errorf("empty coverage file")
	}

	mode := strings.TrimPrefix(scanner.Text(), "mode: ")

	profiles := make(map[string]*CoverageProfile)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		block, err := c.parseCoverageBlock(line)
		if err != nil {
			return fmt.Errorf("failed to parse coverage line: %w", err)
		}

		// Skip excluded files
		if c.isExcluded(block.FileName) {
			continue
		}

		profile, exists := profiles[block.FileName]
		if !exists {
			profile = &CoverageProfile{
				FileName: block.FileName,
				Mode:     mode,
				Blocks:   make([]*CoverageBlock, 0),
			}
			profiles[block.FileName] = profile
		}

		profile.Blocks = append(profile.Blocks, &CoverageBlock{
			StartLine: block.StartLine,
			StartCol:  block.StartCol,
			EndLine:   block.EndLine,
			EndCol:    block.EndCol,
			NumStmt:   block.NumStmt,
			Count:     block.Count,
		})
	}

	// Calculate coverage percentages
	for _, profile := range profiles {
		c.calculateCoverage(profile)
		c.profiles = append(c.profiles, profile)
	}

	return scanner.Err()
}

// parseCoverageBlock parses a single coverage block line
func (c *CoverageAnalyzer) parseCoverageBlock(line string) (*struct {
	FileName  string
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	NumStmt   int
	Count     int
}, error) {
	// Format: filename:startLine.startCol,endLine.endCol numStmt count
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid coverage line format: %s", line)
	}

	// Parse filename and positions
	fileAndPos := parts[0]
	colonIndex := strings.LastIndex(fileAndPos, ":")
	if colonIndex == -1 {
		return nil, fmt.Errorf("invalid file:position format: %s", fileAndPos)
	}

	fileName := fileAndPos[:colonIndex]
	positions := fileAndPos[colonIndex+1:]

	// Parse positions: startLine.startCol,endLine.endCol
	commaIndex := strings.Index(positions, ",")
	if commaIndex == -1 {
		return nil, fmt.Errorf("invalid position format: %s", positions)
	}

	startPos := positions[:commaIndex]
	endPos := positions[commaIndex+1:]

	startParts := strings.Split(startPos, ".")
	endParts := strings.Split(endPos, ".")

	if len(startParts) != 2 || len(endParts) != 2 {
		return nil, fmt.Errorf("invalid position format: %s", positions)
	}

	startLine, _ := strconv.Atoi(startParts[0])
	startCol, _ := strconv.Atoi(startParts[1])
	endLine, _ := strconv.Atoi(endParts[0])
	endCol, _ := strconv.Atoi(endParts[1])

	numStmt, _ := strconv.Atoi(parts[1])
	count, _ := strconv.Atoi(parts[2])

	return &struct {
		FileName  string
		StartLine int
		StartCol  int
		EndLine   int
		EndCol    int
		NumStmt   int
		Count     int
	}{
		FileName:  fileName,
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
		NumStmt:   numStmt,
		Count:     count,
	}, nil
}

// calculateCoverage calculates coverage percentage for a profile
func (c *CoverageAnalyzer) calculateCoverage(profile *CoverageProfile) {
	totalStmts := 0
	coveredStmts := 0

	for _, block := range profile.Blocks {
		totalStmts += block.NumStmt
		if block.Count > 0 {
			coveredStmts += block.NumStmt
		}
	}

	profile.TotalLines = totalStmts
	profile.CoveredLines = coveredStmts

	if totalStmts > 0 {
		profile.Percentage = float64(coveredStmts) / float64(totalStmts) * 100
	}
}

// isExcluded checks if a file should be excluded from coverage analysis
func (c *CoverageAnalyzer) isExcluded(filename string) bool {
	for _, pattern := range c.excludePatterns {
		if pattern.MatchString(filename) {
			return true
		}
	}
	return false
}

// GenerateReport generates a comprehensive coverage report
func (c *CoverageAnalyzer) GenerateReport() *CoverageReport {
	report := &CoverageReport{
		Timestamp:  time.Now(),
		Packages:   make(map[string]*CoverageSummary),
		Files:      make(map[string]*CoverageSummary),
		Violations: make([]*CoverageViolation, 0),
	}

	// Calculate overall coverage
	report.Overall = c.calculateOverallCoverage()

	// Calculate package coverage
	packages := c.groupByPackage()
	for pkg, profiles := range packages {
		report.Packages[pkg] = c.calculatePackageCoverage(profiles)
	}

	// Calculate file coverage
	for _, profile := range c.profiles {
		report.Files[profile.FileName] = &CoverageSummary{
			TotalLines:    profile.TotalLines,
			CoveredLines:  profile.CoveredLines,
			Percentage:    profile.Percentage,
			TotalBlocks:   len(profile.Blocks),
			CoveredBlocks: c.countCoveredBlocks(profile),
		}
	}

	// Check for threshold violations
	if c.thresholds != nil {
		report.Violations = c.checkThresholds(report)
	}

	return report
}

// calculateOverallCoverage calculates overall coverage statistics
func (c *CoverageAnalyzer) calculateOverallCoverage() *CoverageSummary {
	totalLines := 0
	coveredLines := 0
	totalBlocks := 0
	coveredBlocks := 0

	for _, profile := range c.profiles {
		totalLines += profile.TotalLines
		coveredLines += profile.CoveredLines
		totalBlocks += len(profile.Blocks)
		coveredBlocks += c.countCoveredBlocks(profile)
	}

	percentage := 0.0
	if totalLines > 0 {
		percentage = float64(coveredLines) / float64(totalLines) * 100
	}

	return &CoverageSummary{
		TotalLines:    totalLines,
		CoveredLines:  coveredLines,
		Percentage:    percentage,
		TotalBlocks:   totalBlocks,
		CoveredBlocks: coveredBlocks,
	}
}

// groupByPackage groups profiles by package
func (c *CoverageAnalyzer) groupByPackage() map[string][]*CoverageProfile {
	packages := make(map[string][]*CoverageProfile)

	for _, profile := range c.profiles {
		pkg := c.extractPackage(profile.FileName)
		packages[pkg] = append(packages[pkg], profile)
	}

	return packages
}

// extractPackage extracts package name from filename
func (c *CoverageAnalyzer) extractPackage(filename string) string {
	dir := filepath.Dir(filename)
	if dir == "." {
		return "main"
	}
	return dir
}

// calculatePackageCoverage calculates coverage for a package
func (c *CoverageAnalyzer) calculatePackageCoverage(profiles []*CoverageProfile) *CoverageSummary {
	totalLines := 0
	coveredLines := 0
	totalBlocks := 0
	coveredBlocks := 0

	for _, profile := range profiles {
		totalLines += profile.TotalLines
		coveredLines += profile.CoveredLines
		totalBlocks += len(profile.Blocks)
		coveredBlocks += c.countCoveredBlocks(profile)
	}

	percentage := 0.0
	if totalLines > 0 {
		percentage = float64(coveredLines) / float64(totalLines) * 100
	}

	return &CoverageSummary{
		TotalLines:    totalLines,
		CoveredLines:  coveredLines,
		Percentage:    percentage,
		TotalBlocks:   totalBlocks,
		CoveredBlocks: coveredBlocks,
	}
}

// countCoveredBlocks counts the number of covered blocks in a profile
func (c *CoverageAnalyzer) countCoveredBlocks(profile *CoverageProfile) int {
	count := 0
	for _, block := range profile.Blocks {
		if block.Count > 0 {
			count++
		}
	}
	return count
}

// checkThresholds checks for coverage threshold violations
func (c *CoverageAnalyzer) checkThresholds(report *CoverageReport) []*CoverageViolation {
	violations := make([]*CoverageViolation, 0)

	// Check overall threshold
	if report.Overall.Percentage < c.thresholds.Overall {
		violations = append(violations, &CoverageViolation{
			Type:     "overall",
			Target:   "project",
			Actual:   report.Overall.Percentage,
			Required: c.thresholds.Overall,
			Severity: "error",
			Message:  fmt.Sprintf("Overall coverage %.2f%% is below threshold %.2f%%", report.Overall.Percentage, c.thresholds.Overall),
		})
	}

	// Check package thresholds
	for pkg, summary := range report.Packages {
		if summary.Percentage < c.thresholds.Package {
			violations = append(violations, &CoverageViolation{
				Type:     "package",
				Target:   pkg,
				Actual:   summary.Percentage,
				Required: c.thresholds.Package,
				Severity: "warning",
				Message:  fmt.Sprintf("Package %s coverage %.2f%% is below threshold %.2f%%", pkg, summary.Percentage, c.thresholds.Package),
			})
		}
	}

	// Check file thresholds
	for file, summary := range report.Files {
		if summary.Percentage < c.thresholds.File {
			violations = append(violations, &CoverageViolation{
				Type:     "file",
				Target:   file,
				Actual:   summary.Percentage,
				Required: c.thresholds.File,
				Severity: "info",
				Message:  fmt.Sprintf("File %s coverage %.2f%% is below threshold %.2f%%", file, summary.Percentage, c.thresholds.File),
			})
		}
	}

	return violations
}

// SaveReport saves the coverage report to a file
func (c *CoverageAnalyzer) SaveReport(report *CoverageReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// LoadPreviousReport loads a previous coverage report for trend analysis
func (c *CoverageAnalyzer) LoadPreviousReport(filename string) (*CoverageReport, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read previous report: %w", err)
	}

	var report CoverageReport
	err = json.Unmarshal(data, &report)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal previous report: %w", err)
	}

	return &report, nil
}

// CompareTrends compares current coverage with previous report
func (c *CoverageAnalyzer) CompareTrends(current, previous *CoverageReport) *CoverageTrends {
	change := current.Overall.Percentage - previous.Overall.Percentage

	trend := "stable"
	if change > 1.0 {
		trend = "improving"
	} else if change < -1.0 {
		trend = "declining"
	}

	return &CoverageTrends{
		Previous: previous.Overall,
		Current:  current.Overall,
		Change:   change,
		Trend:    trend,
	}
}

// GetLowCoverageFiles returns files with coverage below a threshold
func (c *CoverageAnalyzer) GetLowCoverageFiles(threshold float64) []*CoverageProfile {
	var lowCoverage []*CoverageProfile

	for _, profile := range c.profiles {
		if profile.Percentage < threshold {
			lowCoverage = append(lowCoverage, profile)
		}
	}

	// Sort by coverage percentage (lowest first)
	sort.Slice(lowCoverage, func(i, j int) bool {
		return lowCoverage[i].Percentage < lowCoverage[j].Percentage
	})

	return lowCoverage
}

// GetUncoveredLines returns uncovered lines for a file
func (c *CoverageAnalyzer) GetUncoveredLines(filename string) []int {
	var uncoveredLines []int

	for _, profile := range c.profiles {
		if profile.FileName == filename {
			for _, block := range profile.Blocks {
				if block.Count == 0 {
					for line := block.StartLine; line <= block.EndLine; line++ {
						uncoveredLines = append(uncoveredLines, line)
					}
				}
			}
			break
		}
	}

	// Remove duplicates and sort
	lineMap := make(map[int]bool)
	for _, line := range uncoveredLines {
		lineMap[line] = true
	}

	uncoveredLines = uncoveredLines[:0]
	for line := range lineMap {
		uncoveredLines = append(uncoveredLines, line)
	}

	sort.Ints(uncoveredLines)
	return uncoveredLines
}
