package coding

import (
	"context"
	"io"
	"time"
)

// CodingAssistant is the main interface for AI coding assistance
type CodingAssistant interface {
	// Code Analysis
	AnalyzeCode(ctx context.Context, request *CodeAnalysisRequest) (*CodeAnalysisResult, error)
	AnalyzeProject(ctx context.Context, projectPath string, options *ProjectAnalysisOptions) (*ProjectAnalysisResult, error)
	DetectIssues(ctx context.Context, code string, language string) ([]*CodeIssue, error)

	// Code Generation
	GenerateCode(ctx context.Context, request *CodeGenerationRequest) (*CodeGenerationResult, error)
	CompleteCode(ctx context.Context, request *CodeCompletionRequest) (*CodeCompletionResult, error)
	GenerateFromTemplate(ctx context.Context, template string, params map[string]interface{}) (string, error)

	// Code Refactoring
	SuggestRefactoring(ctx context.Context, code string, language string) ([]*RefactoringSuggestion, error)
	ApplyRefactoring(ctx context.Context, request *RefactoringRequest) (*RefactoringResult, error)
	OptimizeCode(ctx context.Context, code string, language string) (*OptimizationResult, error)

	// Documentation
	GenerateDocumentation(ctx context.Context, request *DocumentationRequest) (*DocumentationResult, error)
	GenerateComments(ctx context.Context, code string, language string) (string, error)
	GenerateREADME(ctx context.Context, projectPath string) (string, error)

	// Testing
	GenerateTests(ctx context.Context, request *TestGenerationRequest) (*TestGenerationResult, error)
	AnalyzeTestCoverage(ctx context.Context, projectPath string) (*TestCoverageResult, error)
	SuggestTestCases(ctx context.Context, code string, language string) ([]*TestCase, error)

	// Project Management
	AnalyzeDependencies(ctx context.Context, projectPath string) (*DependencyGraph, error)
	SuggestProjectStructure(ctx context.Context, projectType string, language string) (*ProjectStructure, error)
	GenerateBuildConfig(ctx context.Context, projectPath string, buildSystem string) (string, error)
}

// CodeAnalyzer handles code analysis and understanding
type CodeAnalyzer interface {
	ParseCode(ctx context.Context, code string, language string) (*ParsedCode, error)
	AnalyzeComplexity(ctx context.Context, code string, language string) (*ComplexityAnalysis, error)
	DetectPatterns(ctx context.Context, code string, language string) ([]*CodePattern, error)
	AnalyzeSecurity(ctx context.Context, code string, language string) ([]*SecurityIssue, error)
	AnalyzePerformance(ctx context.Context, code string, language string) ([]*PerformanceIssue, error)
	ExtractMetrics(ctx context.Context, code string, language string) (*CodeMetrics, error)
}

// CodeGenerator handles code generation and completion
type CodeGenerator interface {
	GenerateFunction(ctx context.Context, request *FunctionGenerationRequest) (string, error)
	GenerateClass(ctx context.Context, request *ClassGenerationRequest) (string, error)
	GenerateInterface(ctx context.Context, request *InterfaceGenerationRequest) (string, error)
	GenerateBoilerplate(ctx context.Context, template string, language string) (string, error)
	CompleteStatement(ctx context.Context, partialCode string, language string) ([]string, error)
	SuggestImports(ctx context.Context, code string, language string) ([]string, error)
}

// CodeRefactorer handles code refactoring and optimization
type CodeRefactorer interface {
	ExtractMethod(ctx context.Context, code string, selection *CodeSelection) (*RefactoringResult, error)
	ExtractVariable(ctx context.Context, code string, selection *CodeSelection) (*RefactoringResult, error)
	RenameSymbol(ctx context.Context, code string, oldName string, newName string) (*RefactoringResult, error)
	InlineMethod(ctx context.Context, code string, methodName string) (*RefactoringResult, error)
	SimplifyExpression(ctx context.Context, code string, language string) (*RefactoringResult, error)
	RemoveDeadCode(ctx context.Context, code string, language string) (*RefactoringResult, error)
}

// DocumentationGenerator handles documentation generation
type DocumentationGenerator interface {
	GenerateFunctionDoc(ctx context.Context, function *ParsedFunction) (string, error)
	GenerateClassDoc(ctx context.Context, class *ParsedClass) (string, error)
	GenerateAPIDoc(ctx context.Context, code string, language string) (string, error)
	GenerateInlineComments(ctx context.Context, code string, language string) (string, error)
	GenerateChangelog(ctx context.Context, commits []*GitCommit) (string, error)
}

// TestGenerator handles test generation and analysis
type TestGenerator interface {
	GenerateUnitTests(ctx context.Context, function *ParsedFunction) ([]*TestCase, error)
	GenerateIntegrationTests(ctx context.Context, module *ParsedModule) ([]*TestCase, error)
	GenerateMockObjects(ctx context.Context, interfaces []*ParsedInterface) (string, error)
	GenerateTestData(ctx context.Context, dataType string, count int) (interface{}, error)
	AnalyzeCoverage(ctx context.Context, testResults *TestResults) (*CoverageReport, error)
}

// ProjectAnalyzer handles project-level analysis
type ProjectAnalyzer interface {
	AnalyzeStructure(ctx context.Context, projectPath string) (*ProjectStructure, error)
	AnalyzeDependencies(ctx context.Context, projectPath string) (*DependencyGraph, error)
	DetectProjectType(ctx context.Context, projectPath string) (*ProjectType, error)
	AnalyzeBuildSystem(ctx context.Context, projectPath string) (*BuildSystemInfo, error)
	ScanForIssues(ctx context.Context, projectPath string) ([]*ProjectIssue, error)
	GenerateMetrics(ctx context.Context, projectPath string) (*ProjectMetrics, error)
}

// CodeFormatter handles code formatting and style
type CodeFormatter interface {
	FormatCode(ctx context.Context, code string, language string, style *FormattingStyle) (string, error)
	CheckStyle(ctx context.Context, code string, language string, rules *StyleRules) ([]*StyleViolation, error)
	FixStyleIssues(ctx context.Context, code string, language string) (string, error)
	ConvertStyle(ctx context.Context, code string, fromStyle string, toStyle string) (string, error)
}

// LanguageServer provides language-specific services
type LanguageServer interface {
	GetCompletions(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
	GetHover(ctx context.Context, request *HoverRequest) (*HoverResponse, error)
	GetDefinition(ctx context.Context, request *DefinitionRequest) (*DefinitionResponse, error)
	GetReferences(ctx context.Context, request *ReferencesRequest) (*ReferencesResponse, error)
	GetSymbols(ctx context.Context, request *SymbolsRequest) (*SymbolsResponse, error)
	GetDiagnostics(ctx context.Context, request *DiagnosticsRequest) (*DiagnosticsResponse, error)
}

// CodeSearcher handles code search and navigation
type CodeSearcher interface {
	SearchCode(ctx context.Context, query string, options *SearchOptions) (*SearchResults, error)
	FindSymbol(ctx context.Context, symbol string, projectPath string) ([]*SymbolLocation, error)
	FindUsages(ctx context.Context, symbol string, projectPath string) ([]*Usage, error)
	FindSimilarCode(ctx context.Context, code string, projectPath string) ([]*SimilarCodeMatch, error)
	SearchByPattern(ctx context.Context, pattern string, language string) ([]*PatternMatch, error)
}

// CodeReviewer handles code review and quality assessment
type CodeReviewer interface {
	ReviewCode(ctx context.Context, request *CodeReviewRequest) (*CodeReviewResult, error)
	CheckQuality(ctx context.Context, code string, language string) (*QualityReport, error)
	SuggestImprovements(ctx context.Context, code string, language string) ([]*Improvement, error)
	CompareVersions(ctx context.Context, oldCode string, newCode string) (*CodeComparison, error)
	AnalyzeChanges(ctx context.Context, diff string) (*ChangeAnalysis, error)
}

// AIModelManager handles AI model interactions
type AIModelManager interface {
	GenerateCompletion(ctx context.Context, prompt string, options *CompletionOptions) (string, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	ClassifyCode(ctx context.Context, code string) (*CodeClassification, error)
	ExplainCode(ctx context.Context, code string, language string) (string, error)
	TranslateCode(ctx context.Context, code string, fromLang string, toLang string) (string, error)
}

// ContextManager handles context and workspace management
type ContextManager interface {
	GetWorkspaceContext(ctx context.Context, workspacePath string) (*WorkspaceContext, error)
	GetFileContext(ctx context.Context, filePath string) (*FileContext, error)
	GetSymbolContext(ctx context.Context, symbol string, filePath string) (*SymbolContext, error)
	UpdateContext(ctx context.Context, changes []*ContextChange) error
	InvalidateContext(ctx context.Context, filePath string) error
}

// ConfigurationManager handles configuration and settings
type ConfigurationManager interface {
	GetConfiguration(ctx context.Context, scope string) (*Configuration, error)
	UpdateConfiguration(ctx context.Context, config *Configuration) error
	GetLanguageConfig(ctx context.Context, language string) (*LanguageConfiguration, error)
	GetFormattingConfig(ctx context.Context, language string) (*FormattingConfiguration, error)
	GetLintingConfig(ctx context.Context, language string) (*LintingConfiguration, error)
}

// PluginManager handles plugin and extension management
type PluginManager interface {
	LoadPlugin(ctx context.Context, pluginPath string) (*Plugin, error)
	UnloadPlugin(ctx context.Context, pluginID string) error
	ListPlugins(ctx context.Context) ([]*Plugin, error)
	ExecutePlugin(ctx context.Context, pluginID string, command string, args map[string]interface{}) (interface{}, error)
	GetPluginCapabilities(ctx context.Context, pluginID string) (*PluginCapabilities, error)
}

// CacheManager handles caching for performance optimization
type CacheManager interface {
	Get(ctx context.Context, key string) (interface{}, bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context, pattern string) error
	GetStats(ctx context.Context) (*CacheStats, error)
}

// EventManager handles events and notifications
type EventManager interface {
	Subscribe(ctx context.Context, eventType string, handler EventHandler) error
	Unsubscribe(ctx context.Context, eventType string, handlerID string) error
	Publish(ctx context.Context, event *Event) error
	GetEventHistory(ctx context.Context, filter *EventFilter) ([]*Event, error)
}

// EventHandler handles specific events
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
	GetID() string
	GetEventTypes() []string
}

// MetricsCollector handles metrics collection and reporting
type MetricsCollector interface {
	RecordMetric(ctx context.Context, metric *Metric) error
	GetMetrics(ctx context.Context, filter *MetricFilter) ([]*Metric, error)
	GetAggregatedMetrics(ctx context.Context, aggregation *MetricAggregation) (*AggregatedMetrics, error)
	ExportMetrics(ctx context.Context, format string, writer io.Writer) error
}

// SecurityScanner handles security analysis
type SecurityScanner interface {
	ScanForVulnerabilities(ctx context.Context, code string, language string) ([]*Vulnerability, error)
	CheckDependencies(ctx context.Context, dependencies []*Dependency) ([]*SecurityIssue, error)
	AnalyzeSecrets(ctx context.Context, code string) ([]*SecretLeak, error)
	ValidatePermissions(ctx context.Context, code string, language string) ([]*PermissionIssue, error)
	GenerateSecurityReport(ctx context.Context, projectPath string) (*SecurityReport, error)
}
