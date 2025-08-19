package devtools

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// CodeAnalyzer handles static code analysis
type CodeAnalyzer struct {
	logger   *logrus.Logger
	tracer   trace.Tracer
	config   CodeAnalysisConfig
	analyses map[string]*models.CodeAnalysis
	mu       sync.RWMutex
	running  bool
	stopCh   chan struct{}
}

// NewCodeAnalyzer creates a new code analyzer
func NewCodeAnalyzer(logger *logrus.Logger, config CodeAnalysisConfig) (*CodeAnalyzer, error) {
	tracer := otel.Tracer("code-analyzer")

	return &CodeAnalyzer{
		logger:   logger,
		tracer:   tracer,
		config:   config,
		analyses: make(map[string]*models.CodeAnalysis),
		stopCh:   make(chan struct{}),
	}, nil
}

// Start initializes the code analyzer
func (ca *CodeAnalyzer) Start(ctx context.Context) error {
	ctx, span := ca.tracer.Start(ctx, "codeAnalyzer.Start")
	defer span.End()

	ca.mu.Lock()
	defer ca.mu.Unlock()

	if ca.running {
		return fmt.Errorf("code analyzer is already running")
	}

	if !ca.config.Enabled {
		ca.logger.Info("Code analyzer is disabled")
		return nil
	}

	ca.logger.Info("Starting code analyzer")

	// Start continuous analysis
	go ca.continuousAnalysis()

	ca.running = true
	ca.logger.Info("Code analyzer started successfully")

	return nil
}

// Stop shuts down the code analyzer
func (ca *CodeAnalyzer) Stop(ctx context.Context) error {
	ctx, span := ca.tracer.Start(ctx, "codeAnalyzer.Stop")
	defer span.End()

	ca.mu.Lock()
	defer ca.mu.Unlock()

	if !ca.running {
		return nil
	}

	ca.logger.Info("Stopping code analyzer")

	close(ca.stopCh)
	ca.running = false
	ca.logger.Info("Code analyzer stopped")

	return nil
}

// GetStatus returns the current code analyzer status
func (ca *CodeAnalyzer) GetStatus(ctx context.Context) (*models.CodeAnalyzerStatus, error) {
	ctx, span := ca.tracer.Start(ctx, "codeAnalyzer.GetStatus")
	defer span.End()

	ca.mu.RLock()
	defer ca.mu.RUnlock()

	analyses := make([]*models.CodeAnalysis, 0, len(ca.analyses))
	for _, analysis := range ca.analyses {
		analyses = append(analyses, analysis)
	}

	return &models.CodeAnalyzerStatus{
		Enabled:         ca.config.Enabled,
		Running:         ca.running,
		StaticAnalysis:  ca.config.StaticAnalysis,
		SecurityScan:    ca.config.SecurityScan,
		QualityMetrics:  ca.config.QualityMetrics,
		DependencyCheck: ca.config.DependencyCheck,
		Analyses:        analyses,
		Timestamp:       time.Now(),
	}, nil
}

// AnalyzeProject analyzes the entire project
func (ca *CodeAnalyzer) AnalyzeProject(ctx context.Context, projectPath string) (*models.CodeAnalysis, error) {
	ctx, span := ca.tracer.Start(ctx, "codeAnalyzer.AnalyzeProject")
	defer span.End()

	ca.logger.WithField("project_path", projectPath).Info("Starting project analysis")

	analysisID := fmt.Sprintf("project-%d", time.Now().Unix())
	analysis := &models.CodeAnalysis{
		ID:        analysisID,
		Type:      "project",
		Path:      projectPath,
		StartTime: time.Now(),
		Status:    "running",
		Issues:    []models.CodeIssue{},
		Metrics:   models.CodeMetrics{},
	}

	ca.mu.Lock()
	ca.analyses[analysisID] = analysis
	ca.mu.Unlock()

	// Perform analysis
	go func() {
		defer func() {
			analysis.EndTime = time.Now()
			analysis.Duration = analysis.EndTime.Sub(analysis.StartTime)
			analysis.Status = "completed"
		}()

		// Static analysis
		if ca.config.StaticAnalysis {
			issues, err := ca.performStaticAnalysis(ctx, projectPath)
			if err != nil {
				ca.logger.WithError(err).Error("Static analysis failed")
				analysis.Status = "failed"
				return
			}
			analysis.Issues = append(analysis.Issues, issues...)
		}

		// Security scan
		if ca.config.SecurityScan {
			securityIssues, err := ca.performSecurityScan(ctx, projectPath)
			if err != nil {
				ca.logger.WithError(err).Error("Security scan failed")
			} else {
				analysis.Issues = append(analysis.Issues, securityIssues...)
			}
		}

		// Quality metrics
		if ca.config.QualityMetrics {
			metrics, err := ca.calculateQualityMetrics(ctx, projectPath)
			if err != nil {
				ca.logger.WithError(err).Error("Quality metrics calculation failed")
			} else {
				analysis.Metrics = metrics
			}
		}

		// Dependency check
		if ca.config.DependencyCheck {
			depIssues, err := ca.checkDependencies(ctx, projectPath)
			if err != nil {
				ca.logger.WithError(err).Error("Dependency check failed")
			} else {
				analysis.Issues = append(analysis.Issues, depIssues...)
			}
		}

		ca.logger.WithFields(logrus.Fields{
			"analysis_id":  analysisID,
			"duration":     analysis.Duration,
			"issues_found": len(analysis.Issues),
		}).Info("Project analysis completed")
	}()

	return analysis, nil
}

// AnalyzeFile analyzes a single file
func (ca *CodeAnalyzer) AnalyzeFile(ctx context.Context, filePath string) (*models.CodeAnalysis, error) {
	ctx, span := ca.tracer.Start(ctx, "codeAnalyzer.AnalyzeFile")
	defer span.End()

	ca.logger.WithField("file_path", filePath).Info("Starting file analysis")

	analysisID := fmt.Sprintf("file-%d", time.Now().Unix())
	analysis := &models.CodeAnalysis{
		ID:        analysisID,
		Type:      "file",
		Path:      filePath,
		StartTime: time.Now(),
		Status:    "running",
		Issues:    []models.CodeIssue{},
		Metrics:   models.CodeMetrics{},
	}

	ca.mu.Lock()
	ca.analyses[analysisID] = analysis
	ca.mu.Unlock()

	// Perform file analysis
	issues, metrics, err := ca.analyzeGoFile(ctx, filePath)
	if err != nil {
		analysis.Status = "failed"
		return analysis, fmt.Errorf("failed to analyze file: %w", err)
	}

	analysis.Issues = issues
	analysis.Metrics = metrics
	analysis.EndTime = time.Now()
	analysis.Duration = analysis.EndTime.Sub(analysis.StartTime)
	analysis.Status = "completed"

	ca.logger.WithFields(logrus.Fields{
		"analysis_id":  analysisID,
		"duration":     analysis.Duration,
		"issues_found": len(analysis.Issues),
	}).Info("File analysis completed")

	return analysis, nil
}

// GetAnalysis returns a specific analysis
func (ca *CodeAnalyzer) GetAnalysis(ctx context.Context, analysisID string) (*models.CodeAnalysis, error) {
	ctx, span := ca.tracer.Start(ctx, "codeAnalyzer.GetAnalysis")
	defer span.End()

	ca.mu.RLock()
	defer ca.mu.RUnlock()

	analysis, exists := ca.analyses[analysisID]
	if !exists {
		return nil, fmt.Errorf("analysis %s not found", analysisID)
	}

	return analysis, nil
}

// ListAnalyses returns all analyses
func (ca *CodeAnalyzer) ListAnalyses(ctx context.Context) ([]*models.CodeAnalysis, error) {
	ctx, span := ca.tracer.Start(ctx, "codeAnalyzer.ListAnalyses")
	defer span.End()

	ca.mu.RLock()
	defer ca.mu.RUnlock()

	analyses := make([]*models.CodeAnalysis, 0, len(ca.analyses))
	for _, analysis := range ca.analyses {
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

// Helper methods

func (ca *CodeAnalyzer) performStaticAnalysis(ctx context.Context, projectPath string) ([]models.CodeIssue, error) {
	var issues []models.CodeIssue

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded paths
		for _, exclude := range ca.config.ExcludePaths {
			if strings.Contains(path, exclude) {
				return nil
			}
		}

		// Only analyze Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fileIssues, _, err := ca.analyzeGoFile(ctx, path)
		if err != nil {
			ca.logger.WithError(err).Warn("Failed to analyze file")
			return nil
		}

		issues = append(issues, fileIssues...)
		return nil
	})

	return issues, err
}

func (ca *CodeAnalyzer) analyzeGoFile(ctx context.Context, filePath string) ([]models.CodeIssue, models.CodeMetrics, error) {
	var issues []models.CodeIssue
	var metrics models.CodeMetrics

	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, metrics, fmt.Errorf("failed to parse file: %w", err)
	}

	// Calculate metrics
	metrics.LinesOfCode = ca.countLines(filePath)
	metrics.Functions = ca.countFunctions(node)
	metrics.Complexity = ca.calculateComplexity(node)
	metrics.TestCoverage = 0.0 // TODO: Calculate actual test coverage

	// Find issues
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			// Check for long functions
			if ca.countFunctionLines(fset, x) > 50 {
				issues = append(issues, models.CodeIssue{
					Type:       "complexity",
					Severity:   "warning",
					Message:    "Function is too long (>50 lines)",
					File:       filePath,
					Line:       fset.Position(x.Pos()).Line,
					Column:     fset.Position(x.Pos()).Column,
					Rule:       "function-length",
					Suggestion: "Consider breaking this function into smaller functions",
				})
			}

			// Check for missing comments on exported functions
			if x.Name.IsExported() && x.Doc == nil {
				issues = append(issues, models.CodeIssue{
					Type:       "documentation",
					Severity:   "info",
					Message:    "Exported function missing documentation",
					File:       filePath,
					Line:       fset.Position(x.Pos()).Line,
					Column:     fset.Position(x.Pos()).Column,
					Rule:       "missing-doc",
					Suggestion: "Add a comment starting with the function name",
				})
			}
		}
		return true
	})

	return issues, metrics, nil
}

func (ca *CodeAnalyzer) performSecurityScan(ctx context.Context, projectPath string) ([]models.CodeIssue, error) {
	var issues []models.CodeIssue

	// TODO: Implement actual security scanning
	// For now, return mock security issues
	issues = append(issues, models.CodeIssue{
		Type:       "security",
		Severity:   "high",
		Message:    "Potential SQL injection vulnerability",
		File:       filepath.Join(projectPath, "example.go"),
		Line:       42,
		Column:     10,
		Rule:       "sql-injection",
		Suggestion: "Use parameterized queries",
	})

	return issues, nil
}

func (ca *CodeAnalyzer) calculateQualityMetrics(ctx context.Context, projectPath string) (models.CodeMetrics, error) {
	var metrics models.CodeMetrics

	// TODO: Implement actual quality metrics calculation
	metrics.LinesOfCode = 1000
	metrics.Functions = 50
	metrics.Complexity = 15.5
	metrics.TestCoverage = 85.2
	metrics.Maintainability = 7.8
	metrics.Duplication = 2.1

	return metrics, nil
}

func (ca *CodeAnalyzer) checkDependencies(ctx context.Context, projectPath string) ([]models.CodeIssue, error) {
	var issues []models.CodeIssue

	// TODO: Implement actual dependency checking
	// For now, return mock dependency issues
	issues = append(issues, models.CodeIssue{
		Type:       "dependency",
		Severity:   "warning",
		Message:    "Outdated dependency detected",
		File:       filepath.Join(projectPath, "go.mod"),
		Line:       5,
		Column:     1,
		Rule:       "outdated-dependency",
		Suggestion: "Update to the latest version",
	})

	return issues, nil
}

func (ca *CodeAnalyzer) countLines(filePath string) int {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0
	}
	return len(strings.Split(string(content), "\n"))
}

func (ca *CodeAnalyzer) countFunctions(node *ast.File) int {
	count := 0
	ast.Inspect(node, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncDecl); ok {
			count++
		}
		return true
	})
	return count
}

func (ca *CodeAnalyzer) calculateComplexity(node *ast.File) float64 {
	// Simple cyclomatic complexity calculation
	complexity := 0
	ast.Inspect(node, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			complexity++
		}
		return true
	})
	return float64(complexity)
}

func (ca *CodeAnalyzer) countFunctionLines(fset *token.FileSet, fn *ast.FuncDecl) int {
	start := fset.Position(fn.Pos()).Line
	end := fset.Position(fn.End()).Line
	return end - start + 1
}

func (ca *CodeAnalyzer) continuousAnalysis() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Perform periodic analysis of the current project
			ca.logger.Debug("Running periodic code analysis")
			// TODO: Implement periodic analysis

		case <-ca.stopCh:
			ca.logger.Debug("Continuous analysis stopped")
			return
		}
	}
}
