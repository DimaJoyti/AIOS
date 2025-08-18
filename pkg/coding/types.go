package coding

import (
	"time"
)

// Core request and response types

// CodeAnalysisRequest represents a request for code analysis
type CodeAnalysisRequest struct {
	Code     string                 `json:"code"`
	Language string                 `json:"language"`
	FilePath string                 `json:"file_path,omitempty"`
	Options  *AnalysisOptions       `json:"options,omitempty"`
	Context  *AnalysisContext       `json:"context,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CodeAnalysisResult represents the result of code analysis
type CodeAnalysisResult struct {
	ParsedCode        *ParsedCode            `json:"parsed_code"`
	Issues            []*CodeIssue           `json:"issues"`
	Metrics           *CodeMetrics           `json:"metrics"`
	Complexity        *ComplexityAnalysis    `json:"complexity"`
	SecurityIssues    []*SecurityIssue       `json:"security_issues"`
	PerformanceIssues []*PerformanceIssue    `json:"performance_issues"`
	Suggestions       []*Suggestion          `json:"suggestions"`
	ProcessingTime    time.Duration          `json:"processing_time"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// CodeGenerationRequest represents a request for code generation
type CodeGenerationRequest struct {
	Prompt      string                 `json:"prompt"`
	Language    string                 `json:"language"`
	Context     *GenerationContext     `json:"context,omitempty"`
	Style       *CodingStyle           `json:"style,omitempty"`
	Constraints *GenerationConstraints `json:"constraints,omitempty"`
	Examples    []string               `json:"examples,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CodeGenerationResult represents the result of code generation
type CodeGenerationResult struct {
	GeneratedCode  string                 `json:"generated_code"`
	Explanation    string                 `json:"explanation,omitempty"`
	Alternatives   []string               `json:"alternatives,omitempty"`
	Confidence     float32                `json:"confidence"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CodeCompletionRequest represents a request for code completion
type CodeCompletionRequest struct {
	Code       string                 `json:"code"`
	Position   *Position              `json:"position"`
	Language   string                 `json:"language"`
	Context    *CompletionContext     `json:"context,omitempty"`
	MaxResults int                    `json:"max_results,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// CodeCompletionResult represents the result of code completion
type CodeCompletionResult struct {
	Completions    []*Completion          `json:"completions"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Core data structures

// ParsedCode represents parsed and analyzed code
type ParsedCode struct {
	Language   string                 `json:"language"`
	AST        *ASTNode               `json:"ast"`
	Symbols    []*Symbol              `json:"symbols"`
	Functions  []*ParsedFunction      `json:"functions"`
	Classes    []*ParsedClass         `json:"classes"`
	Interfaces []*ParsedInterface     `json:"interfaces"`
	Variables  []*ParsedVariable      `json:"variables"`
	Imports    []*Import              `json:"imports"`
	Comments   []*Comment             `json:"comments"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ASTNode represents a node in the Abstract Syntax Tree
type ASTNode struct {
	Type       string                 `json:"type"`
	Value      string                 `json:"value,omitempty"`
	Position   *Position              `json:"position"`
	Children   []*ASTNode             `json:"children,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// Position represents a position in source code
type Position struct {
	Line      int `json:"line"`
	Column    int `json:"column"`
	Offset    int `json:"offset"`
	EndLine   int `json:"end_line,omitempty"`
	EndColumn int `json:"end_column,omitempty"`
	EndOffset int `json:"end_offset,omitempty"`
}

// Symbol represents a code symbol
type Symbol struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Kind          SymbolKind             `json:"kind"`
	Position      *Position              `json:"position"`
	Scope         string                 `json:"scope"`
	Visibility    string                 `json:"visibility"`
	Signature     string                 `json:"signature,omitempty"`
	Documentation string                 `json:"documentation,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ParsedFunction represents a parsed function
type ParsedFunction struct {
	Name          string                 `json:"name"`
	Signature     string                 `json:"signature"`
	Parameters    []*Parameter           `json:"parameters"`
	ReturnType    string                 `json:"return_type,omitempty"`
	Body          string                 `json:"body"`
	Position      *Position              `json:"position"`
	Visibility    string                 `json:"visibility"`
	IsAsync       bool                   `json:"is_async"`
	IsStatic      bool                   `json:"is_static"`
	Decorators    []string               `json:"decorators,omitempty"`
	Documentation string                 `json:"documentation,omitempty"`
	Complexity    *ComplexityMetrics     `json:"complexity,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ParsedClass represents a parsed class
type ParsedClass struct {
	Name          string                 `json:"name"`
	BaseClasses   []string               `json:"base_classes,omitempty"`
	Interfaces    []string               `json:"interfaces,omitempty"`
	Methods       []*ParsedFunction      `json:"methods"`
	Properties    []*ParsedProperty      `json:"properties"`
	Position      *Position              `json:"position"`
	Visibility    string                 `json:"visibility"`
	IsAbstract    bool                   `json:"is_abstract"`
	IsInterface   bool                   `json:"is_interface"`
	Decorators    []string               `json:"decorators,omitempty"`
	Documentation string                 `json:"documentation,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ParsedInterface represents a parsed interface
type ParsedInterface struct {
	Name          string                 `json:"name"`
	Methods       []*MethodSignature     `json:"methods"`
	Properties    []*PropertySignature   `json:"properties"`
	Extends       []string               `json:"extends,omitempty"`
	Position      *Position              `json:"position"`
	Documentation string                 `json:"documentation,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ParsedVariable represents a parsed variable
type ParsedVariable struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Value         string                 `json:"value,omitempty"`
	Position      *Position              `json:"position"`
	Scope         string                 `json:"scope"`
	Visibility    string                 `json:"visibility"`
	IsConstant    bool                   `json:"is_constant"`
	IsStatic      bool                   `json:"is_static"`
	Documentation string                 `json:"documentation,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Parameter represents a function parameter
type Parameter struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	DefaultValue string `json:"default_value,omitempty"`
	IsOptional   bool   `json:"is_optional"`
	IsVariadic   bool   `json:"is_variadic"`
}

// ParsedProperty represents a parsed property
type ParsedProperty struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Value         string                 `json:"value,omitempty"`
	Position      *Position              `json:"position"`
	Visibility    string                 `json:"visibility"`
	IsStatic      bool                   `json:"is_static"`
	IsReadOnly    bool                   `json:"is_read_only"`
	Getter        *ParsedFunction        `json:"getter,omitempty"`
	Setter        *ParsedFunction        `json:"setter,omitempty"`
	Documentation string                 `json:"documentation,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// MethodSignature represents a method signature
type MethodSignature struct {
	Name       string       `json:"name"`
	Parameters []*Parameter `json:"parameters"`
	ReturnType string       `json:"return_type,omitempty"`
	IsAsync    bool         `json:"is_async"`
}

// PropertySignature represents a property signature
type PropertySignature struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsOptional bool   `json:"is_optional"`
	IsReadOnly bool   `json:"is_read_only"`
}

// Import represents an import statement
type Import struct {
	Module     string    `json:"module"`
	Alias      string    `json:"alias,omitempty"`
	Items      []string  `json:"items,omitempty"`
	Position   *Position `json:"position"`
	IsWildcard bool      `json:"is_wildcard"`
}

// Comment represents a code comment
type Comment struct {
	Text     string    `json:"text"`
	Type     string    `json:"type"` // line, block, doc
	Position *Position `json:"position"`
}

// CodeIssue represents a code issue or problem
type CodeIssue struct {
	ID          string                 `json:"id"`
	Type        IssueType              `json:"type"`
	Severity    IssueSeverity          `json:"severity"`
	Message     string                 `json:"message"`
	Description string                 `json:"description,omitempty"`
	Position    *Position              `json:"position"`
	Rule        string                 `json:"rule,omitempty"`
	Category    string                 `json:"category"`
	Suggestion  *Suggestion            `json:"suggestion,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Suggestion represents a code improvement suggestion
type Suggestion struct {
	ID          string                 `json:"id"`
	Type        SuggestionType         `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Code        string                 `json:"code,omitempty"`
	Position    *Position              `json:"position,omitempty"`
	Confidence  float32                `json:"confidence"`
	Impact      ImpactLevel            `json:"impact"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CodeMetrics represents code metrics and statistics
type CodeMetrics struct {
	LinesOfCode          int                    `json:"lines_of_code"`
	LinesOfComments      int                    `json:"lines_of_comments"`
	LinesBlank           int                    `json:"lines_blank"`
	FunctionCount        int                    `json:"function_count"`
	ClassCount           int                    `json:"class_count"`
	VariableCount        int                    `json:"variable_count"`
	CyclomaticComplexity int                    `json:"cyclomatic_complexity"`
	CognitiveComplexity  int                    `json:"cognitive_complexity"`
	Maintainability      float32                `json:"maintainability"`
	TechnicalDebt        time.Duration          `json:"technical_debt"`
	TestCoverage         float32                `json:"test_coverage,omitempty"`
	Duplication          float32                `json:"duplication"`
	Dependencies         int                    `json:"dependencies"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// ComplexityAnalysis represents complexity analysis results
type ComplexityAnalysis struct {
	Overall   *ComplexityMetrics            `json:"overall"`
	Functions map[string]*ComplexityMetrics `json:"functions"`
	Classes   map[string]*ComplexityMetrics `json:"classes"`
	Modules   map[string]*ComplexityMetrics `json:"modules"`
	Hotspots  []*ComplexityHotspot          `json:"hotspots"`
	Metadata  map[string]interface{}        `json:"metadata,omitempty"`
}

// ComplexityMetrics represents complexity metrics
type ComplexityMetrics struct {
	Cyclomatic int     `json:"cyclomatic"`
	Cognitive  int     `json:"cognitive"`
	Halstead   float32 `json:"halstead"`
	Nesting    int     `json:"nesting"`
	Parameters int     `json:"parameters"`
	Variables  int     `json:"variables"`
	Branches   int     `json:"branches"`
	Statements int     `json:"statements"`
}

// ComplexityHotspot represents a complexity hotspot
type ComplexityHotspot struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Position   *Position              `json:"position"`
	Complexity *ComplexityMetrics     `json:"complexity"`
	Severity   string                 `json:"severity"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Enums and constants

type SymbolKind string

const (
	SymbolKindFunction  SymbolKind = "function"
	SymbolKindClass     SymbolKind = "class"
	SymbolKindInterface SymbolKind = "interface"
	SymbolKindVariable  SymbolKind = "variable"
	SymbolKindConstant  SymbolKind = "constant"
	SymbolKindProperty  SymbolKind = "property"
	SymbolKindMethod    SymbolKind = "method"
	SymbolKindEnum      SymbolKind = "enum"
	SymbolKindNamespace SymbolKind = "namespace"
	SymbolKindModule    SymbolKind = "module"
)

type IssueType string

const (
	IssueTypeSyntax          IssueType = "syntax"
	IssueTypeLogic           IssueType = "logic"
	IssueTypeSecurity        IssueType = "security"
	IssueTypePerformance     IssueType = "performance"
	IssueTypeStyle           IssueType = "style"
	IssueTypeMaintainability IssueType = "maintainability"
	IssueTypeComplexity      IssueType = "complexity"
	IssueTypeDuplication     IssueType = "duplication"
	IssueTypeDeprecated      IssueType = "deprecated"
	IssueTypeUnused          IssueType = "unused"
)

type IssueSeverity string

const (
	IssueSeverityError   IssueSeverity = "error"
	IssueSeverityWarning IssueSeverity = "warning"
	IssueSeverityInfo    IssueSeverity = "info"
	IssueSeverityHint    IssueSeverity = "hint"
)

type SuggestionType string

const (
	SuggestionTypeRefactor  SuggestionType = "refactor"
	SuggestionTypeOptimize  SuggestionType = "optimize"
	SuggestionTypeSimplify  SuggestionType = "simplify"
	SuggestionTypeModernize SuggestionType = "modernize"
	SuggestionTypeSecure    SuggestionType = "secure"
	SuggestionTypeDocument  SuggestionType = "document"
	SuggestionTypeTest      SuggestionType = "test"
	SuggestionTypeFormat    SuggestionType = "format"
)

type ImpactLevel string

const (
	ImpactLevelLow      ImpactLevel = "low"
	ImpactLevelMedium   ImpactLevel = "medium"
	ImpactLevelHigh     ImpactLevel = "high"
	ImpactLevelCritical ImpactLevel = "critical"
)

// Additional types for comprehensive coding assistance

// SecurityIssue represents a security-related issue
type SecurityIssue struct {
	ID          string                 `json:"id"`
	Type        SecurityIssueType      `json:"type"`
	Severity    IssueSeverity          `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Position    *Position              `json:"position"`
	CWE         string                 `json:"cwe,omitempty"`
	CVSS        float32                `json:"cvss,omitempty"`
	Remediation string                 `json:"remediation,omitempty"`
	References  []string               `json:"references,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PerformanceIssue represents a performance-related issue
type PerformanceIssue struct {
	ID          string                 `json:"id"`
	Type        PerformanceIssueType   `json:"type"`
	Severity    IssueSeverity          `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Position    *Position              `json:"position"`
	Impact      string                 `json:"impact"`
	Suggestion  string                 `json:"suggestion"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RefactoringSuggestion represents a refactoring suggestion
type RefactoringSuggestion struct {
	ID          string                 `json:"id"`
	Type        RefactoringType        `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Position    *Position              `json:"position"`
	OldCode     string                 `json:"old_code"`
	NewCode     string                 `json:"new_code"`
	Confidence  float32                `json:"confidence"`
	Benefits    []string               `json:"benefits"`
	Risks       []string               `json:"risks,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RefactoringRequest represents a refactoring request
type RefactoringRequest struct {
	Code       string                 `json:"code"`
	Language   string                 `json:"language"`
	Type       RefactoringType        `json:"type"`
	Selection  *CodeSelection         `json:"selection,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Options    *RefactoringOptions    `json:"options,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RefactoringResult represents the result of a refactoring operation
type RefactoringResult struct {
	Success        bool                   `json:"success"`
	RefactoredCode string                 `json:"refactored_code"`
	Changes        []*CodeChange          `json:"changes"`
	Warnings       []string               `json:"warnings,omitempty"`
	Errors         []string               `json:"errors,omitempty"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CodeSelection represents a selection of code
type CodeSelection struct {
	Start    *Position `json:"start"`
	End      *Position `json:"end"`
	Text     string    `json:"text"`
	FilePath string    `json:"file_path,omitempty"`
}

// CodeChange represents a change to code
type CodeChange struct {
	Type        ChangeType `json:"type"`
	Position    *Position  `json:"position"`
	OldText     string     `json:"old_text"`
	NewText     string     `json:"new_text"`
	Description string     `json:"description"`
}

// OptimizationResult represents code optimization results
type OptimizationResult struct {
	OptimizedCode   string                 `json:"optimized_code"`
	Improvements    []*Improvement         `json:"improvements"`
	PerformanceGain float32                `json:"performance_gain"`
	ProcessingTime  time.Duration          `json:"processing_time"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Improvement represents a code improvement
type Improvement struct {
	Type        ImprovementType        `json:"type"`
	Description string                 `json:"description"`
	Position    *Position              `json:"position"`
	Impact      ImpactLevel            `json:"impact"`
	Confidence  float32                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentationRequest represents a documentation generation request
type DocumentationRequest struct {
	Code     string                 `json:"code"`
	Language string                 `json:"language"`
	Type     DocumentationType      `json:"type"`
	Style    *DocumentationStyle    `json:"style,omitempty"`
	Options  *DocumentationOptions  `json:"options,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentationResult represents documentation generation results
type DocumentationResult struct {
	Documentation  string                 `json:"documentation"`
	Format         string                 `json:"format"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TestGenerationRequest represents a test generation request
type TestGenerationRequest struct {
	Code      string                 `json:"code"`
	Language  string                 `json:"language"`
	TestType  TestType               `json:"test_type"`
	Framework string                 `json:"framework,omitempty"`
	Options   *TestGenerationOptions `json:"options,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TestGenerationResult represents test generation results
type TestGenerationResult struct {
	Tests          []*TestCase            `json:"tests"`
	TestCode       string                 `json:"test_code"`
	Framework      string                 `json:"framework"`
	Coverage       float32                `json:"coverage"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TestCase represents a test case
type TestCase struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Code        string                 `json:"code"`
	Type        TestCaseType           `json:"type"`
	Inputs      []interface{}          `json:"inputs,omitempty"`
	Expected    interface{}            `json:"expected,omitempty"`
	Setup       string                 `json:"setup,omitempty"`
	Teardown    string                 `json:"teardown,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Additional enum types

type SecurityIssueType string

const (
	SecurityIssueTypeInjection           SecurityIssueType = "injection"
	SecurityIssueTypeXSS                 SecurityIssueType = "xss"
	SecurityIssueTypeCSRF                SecurityIssueType = "csrf"
	SecurityIssueTypeInsecureStorage     SecurityIssueType = "insecure_storage"
	SecurityIssueTypeWeakCrypto          SecurityIssueType = "weak_crypto"
	SecurityIssueTypeAuthBypass          SecurityIssueType = "auth_bypass"
	SecurityIssueTypePrivilegeEscalation SecurityIssueType = "privilege_escalation"
	SecurityIssueTypeDataExposure        SecurityIssueType = "data_exposure"
)

type PerformanceIssueType string

const (
	PerformanceIssueTypeMemoryLeak    PerformanceIssueType = "memory_leak"
	PerformanceIssueTypeInefficient   PerformanceIssueType = "inefficient"
	PerformanceIssueTypeBlocking      PerformanceIssueType = "blocking"
	PerformanceIssueTypeResourceWaste PerformanceIssueType = "resource_waste"
	PerformanceIssueTypeSlowAlgorithm PerformanceIssueType = "slow_algorithm"
	PerformanceIssueTypeExcessiveIO   PerformanceIssueType = "excessive_io"
)

type RefactoringType string

const (
	RefactoringTypeExtractMethod   RefactoringType = "extract_method"
	RefactoringTypeExtractVariable RefactoringType = "extract_variable"
	RefactoringTypeInlineMethod    RefactoringType = "inline_method"
	RefactoringTypeInlineVariable  RefactoringType = "inline_variable"
	RefactoringTypeRename          RefactoringType = "rename"
	RefactoringTypeMoveMethod      RefactoringType = "move_method"
	RefactoringTypeExtractClass    RefactoringType = "extract_class"
	RefactoringTypeSimplify        RefactoringType = "simplify"
)

type ChangeType string

const (
	ChangeTypeInsert  ChangeType = "insert"
	ChangeTypeDelete  ChangeType = "delete"
	ChangeTypeReplace ChangeType = "replace"
	ChangeTypeMove    ChangeType = "move"
)

type ImprovementType string

const (
	ImprovementTypePerformance     ImprovementType = "performance"
	ImprovementTypeReadability     ImprovementType = "readability"
	ImprovementTypeMaintainability ImprovementType = "maintainability"
	ImprovementTypeSecurity        ImprovementType = "security"
	ImprovementTypeModernization   ImprovementType = "modernization"
	ImprovementTypeSimplification  ImprovementType = "simplification"
)

type DocumentationType string

const (
	DocumentationTypeAPI      DocumentationType = "api"
	DocumentationTypeFunction DocumentationType = "function"
	DocumentationTypeClass    DocumentationType = "class"
	DocumentationTypeModule   DocumentationType = "module"
	DocumentationTypeProject  DocumentationType = "project"
	DocumentationTypeInline   DocumentationType = "inline"
)

type TestType string

const (
	TestTypeUnit        TestType = "unit"
	TestTypeIntegration TestType = "integration"
	TestTypeEnd2End     TestType = "e2e"
	TestTypePerformance TestType = "performance"
	TestTypeSecurity    TestType = "security"
)

type TestCaseType string

const (
	TestCaseTypePositive TestCaseType = "positive"
	TestCaseTypeNegative TestCaseType = "negative"
	TestCaseTypeEdge     TestCaseType = "edge"
	TestCaseTypeBoundary TestCaseType = "boundary"
)

// Configuration and option types

// RefactoringOptions represents options for refactoring operations
type RefactoringOptions struct {
	PreserveComments bool                   `json:"preserve_comments"`
	UpdateReferences bool                   `json:"update_references"`
	ValidateChanges  bool                   `json:"validate_changes"`
	DryRun           bool                   `json:"dry_run"`
	Scope            string                 `json:"scope,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentationStyle represents documentation style preferences
type DocumentationStyle struct {
	Format          string                 `json:"format"` // markdown, html, plain
	IncludeExamples bool                   `json:"include_examples"`
	IncludeTypes    bool                   `json:"include_types"`
	IncludeParams   bool                   `json:"include_params"`
	IncludeReturns  bool                   `json:"include_returns"`
	IncludeThrows   bool                   `json:"include_throws"`
	Language        string                 `json:"language,omitempty"`
	Template        string                 `json:"template,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentationOptions represents options for documentation generation
type DocumentationOptions struct {
	IncludePrivate  bool                   `json:"include_private"`
	IncludeInternal bool                   `json:"include_internal"`
	GenerateIndex   bool                   `json:"generate_index"`
	GenerateTOC     bool                   `json:"generate_toc"`
	OutputFormat    string                 `json:"output_format"`
	OutputPath      string                 `json:"output_path,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// TestGenerationOptions represents options for test generation
type TestGenerationOptions struct {
	IncludeEdgeCases     bool                   `json:"include_edge_cases"`
	IncludeNegativeCases bool                   `json:"include_negative_cases"`
	GenerateMocks        bool                   `json:"generate_mocks"`
	GenerateSetup        bool                   `json:"generate_setup"`
	GenerateTeardown     bool                   `json:"generate_teardown"`
	TestFramework        string                 `json:"test_framework,omitempty"`
	MockFramework        string                 `json:"mock_framework,omitempty"`
	CoverageTarget       float32                `json:"coverage_target,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// AnalysisOptions represents options for code analysis
type AnalysisOptions struct {
	IncludeSecurity    bool                   `json:"include_security"`
	IncludePerformance bool                   `json:"include_performance"`
	IncludeComplexity  bool                   `json:"include_complexity"`
	IncludeStyle       bool                   `json:"include_style"`
	IncludeDuplication bool                   `json:"include_duplication"`
	IncludeMetrics     bool                   `json:"include_metrics"`
	Depth              int                    `json:"depth,omitempty"`
	Rules              []string               `json:"rules,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// AnalysisContext represents context for code analysis
type AnalysisContext struct {
	ProjectPath     string                 `json:"project_path,omitempty"`
	WorkspacePath   string                 `json:"workspace_path,omitempty"`
	Dependencies    []string               `json:"dependencies,omitempty"`
	ImportPaths     []string               `json:"import_paths,omitempty"`
	LanguageVersion string                 `json:"language_version,omitempty"`
	BuildTags       []string               `json:"build_tags,omitempty"`
	Environment     map[string]string      `json:"environment,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// GenerationContext represents context for code generation
type GenerationContext struct {
	ProjectPath     string                 `json:"project_path,omitempty"`
	PackageName     string                 `json:"package_name,omitempty"`
	Imports         []string               `json:"imports,omitempty"`
	ExistingCode    string                 `json:"existing_code,omitempty"`
	Dependencies    []string               `json:"dependencies,omitempty"`
	LanguageVersion string                 `json:"language_version,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// CodingStyle represents coding style preferences
type CodingStyle struct {
	IndentSize   int                    `json:"indent_size"`
	IndentType   string                 `json:"indent_type"` // spaces, tabs
	LineLength   int                    `json:"line_length"`
	NamingStyle  string                 `json:"naming_style"` // camelCase, snake_case, etc.
	BraceStyle   string                 `json:"brace_style"`  // K&R, Allman, etc.
	CommentStyle string                 `json:"comment_style"`
	ImportStyle  string                 `json:"import_style"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// GenerationConstraints represents constraints for code generation
type GenerationConstraints struct {
	MaxLines          int                    `json:"max_lines,omitempty"`
	MaxComplexity     int                    `json:"max_complexity,omitempty"`
	RequiredImports   []string               `json:"required_imports,omitempty"`
	ForbiddenImports  []string               `json:"forbidden_imports,omitempty"`
	RequiredPatterns  []string               `json:"required_patterns,omitempty"`
	ForbiddenPatterns []string               `json:"forbidden_patterns,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// CompletionContext represents context for code completion
type CompletionContext struct {
	TriggerCharacter string                 `json:"trigger_character,omitempty"`
	TriggerKind      string                 `json:"trigger_kind"`
	IncludeSnippets  bool                   `json:"include_snippets"`
	IncludeKeywords  bool                   `json:"include_keywords"`
	IncludeSymbols   bool                   `json:"include_symbols"`
	FilterText       string                 `json:"filter_text,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// Completion represents a code completion item
type Completion struct {
	Label         string                 `json:"label"`
	Kind          CompletionKind         `json:"kind"`
	Detail        string                 `json:"detail,omitempty"`
	Documentation string                 `json:"documentation,omitempty"`
	InsertText    string                 `json:"insert_text"`
	FilterText    string                 `json:"filter_text,omitempty"`
	SortText      string                 `json:"sort_text,omitempty"`
	Snippet       bool                   `json:"snippet"`
	Priority      int                    `json:"priority"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type CompletionKind string

const (
	CompletionKindText        CompletionKind = "text"
	CompletionKindMethod      CompletionKind = "method"
	CompletionKindFunction    CompletionKind = "function"
	CompletionKindConstructor CompletionKind = "constructor"
	CompletionKindField       CompletionKind = "field"
	CompletionKindVariable    CompletionKind = "variable"
	CompletionKindClass       CompletionKind = "class"
	CompletionKindInterface   CompletionKind = "interface"
	CompletionKindModule      CompletionKind = "module"
	CompletionKindProperty    CompletionKind = "property"
	CompletionKindUnit        CompletionKind = "unit"
	CompletionKindValue       CompletionKind = "value"
	CompletionKindEnum        CompletionKind = "enum"
	CompletionKindKeyword     CompletionKind = "keyword"
	CompletionKindSnippet     CompletionKind = "snippet"
	CompletionKindColor       CompletionKind = "color"
	CompletionKindFile        CompletionKind = "file"
	CompletionKindReference   CompletionKind = "reference"
	CompletionKindFolder      CompletionKind = "folder"
)

// Additional types for language server protocol and other components

// CompletionRequest represents a request for code completions
type CompletionRequest struct {
	Code     string             `json:"code"`
	Position *Position          `json:"position"`
	Language string             `json:"language"`
	Context  *CompletionContext `json:"context,omitempty"`
}

// CompletionResponse represents a response with code completions
type CompletionResponse struct {
	Completions  []*Completion `json:"completions"`
	IsIncomplete bool          `json:"is_incomplete"`
}

// HoverRequest represents a request for hover information
type HoverRequest struct {
	Code     string    `json:"code"`
	Position *Position `json:"position"`
	Language string    `json:"language"`
}

// HoverResponse represents hover information
type HoverResponse struct {
	Contents string    `json:"contents"`
	Range    *Position `json:"range,omitempty"`
}

// DefinitionRequest represents a request for symbol definition
type DefinitionRequest struct {
	Code     string    `json:"code"`
	Position *Position `json:"position"`
	Language string    `json:"language"`
}

// DefinitionResponse represents symbol definition information
type DefinitionResponse struct {
	Locations []*Location `json:"locations"`
}

// ReferencesRequest represents a request for symbol references
type ReferencesRequest struct {
	Code               string    `json:"code"`
	Position           *Position `json:"position"`
	Language           string    `json:"language"`
	IncludeDeclaration bool      `json:"include_declaration"`
}

// ReferencesResponse represents symbol references
type ReferencesResponse struct {
	References []*Location `json:"references"`
}

// SymbolsRequest represents a request for document symbols
type SymbolsRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

// SymbolsResponse represents document symbols
type SymbolsResponse struct {
	Symbols []*DocumentSymbol `json:"symbols"`
}

// DiagnosticsRequest represents a request for diagnostics
type DiagnosticsRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

// DiagnosticsResponse represents diagnostics information
type DiagnosticsResponse struct {
	Diagnostics []*Diagnostic `json:"diagnostics"`
}

// Location represents a location in source code
type Location struct {
	URI   string    `json:"uri"`
	Range *Position `json:"range"`
}

// DocumentSymbol represents a symbol in a document
type DocumentSymbol struct {
	Name           string            `json:"name"`
	Detail         string            `json:"detail,omitempty"`
	Kind           SymbolKind        `json:"kind"`
	Range          *Position         `json:"range"`
	SelectionRange *Position         `json:"selection_range"`
	Children       []*DocumentSymbol `json:"children,omitempty"`
}

// Diagnostic represents a diagnostic message
type Diagnostic struct {
	Range    *Position          `json:"range"`
	Severity DiagnosticSeverity `json:"severity"`
	Code     string             `json:"code,omitempty"`
	Source   string             `json:"source,omitempty"`
	Message  string             `json:"message"`
	Tags     []DiagnosticTag    `json:"tags,omitempty"`
}

// CompletionOptions represents options for AI completion
type CompletionOptions struct {
	Language    string   `json:"language,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature float32  `json:"temperature,omitempty"`
	TopP        float32  `json:"top_p,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// CodeClassification represents code classification results
type CodeClassification struct {
	Language   string                 `json:"language"`
	Framework  string                 `json:"framework,omitempty"`
	Purpose    string                 `json:"purpose,omitempty"`
	Complexity string                 `json:"complexity"`
	Quality    string                 `json:"quality"`
	Confidence float32                `json:"confidence"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// CodePattern represents a detected code pattern
type CodePattern struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Description   string                 `json:"description"`
	Position      *Position              `json:"position"`
	IsAntiPattern bool                   `json:"is_anti_pattern"`
	Confidence    float32                `json:"confidence"`
	Examples      []string               `json:"examples,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ProjectAnalysisOptions represents options for project analysis
type ProjectAnalysisOptions struct {
	IncludeDependencies bool     `json:"include_dependencies"`
	IncludeMetrics      bool     `json:"include_metrics"`
	IncludeIssues       bool     `json:"include_issues"`
	IncludeStructure    bool     `json:"include_structure"`
	ExcludePatterns     []string `json:"exclude_patterns,omitempty"`
	MaxDepth            int      `json:"max_depth,omitempty"`
}

// ProjectAnalysisResult represents project analysis results
type ProjectAnalysisResult struct {
	ProjectInfo    *ProjectInfo      `json:"project_info"`
	Structure      *ProjectStructure `json:"structure,omitempty"`
	Dependencies   *DependencyGraph  `json:"dependencies,omitempty"`
	Metrics        *ProjectMetrics   `json:"metrics,omitempty"`
	Issues         []*ProjectIssue   `json:"issues,omitempty"`
	Suggestions    []*Suggestion     `json:"suggestions,omitempty"`
	ProcessingTime time.Duration     `json:"processing_time"`
}

// ProjectInfo represents basic project information
type ProjectInfo struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Language    string   `json:"language"`
	Framework   string   `json:"framework,omitempty"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Authors     []string `json:"authors,omitempty"`
	License     string   `json:"license,omitempty"`
}

// ProjectStructure represents project structure
type ProjectStructure struct {
	Root      *DirectoryNode `json:"root"`
	FileCount int            `json:"file_count"`
	DirCount  int            `json:"dir_count"`
	TotalSize int64          `json:"total_size"`
	Languages map[string]int `json:"languages"`
}

// DirectoryNode represents a directory in the project structure
type DirectoryNode struct {
	Name     string           `json:"name"`
	Path     string           `json:"path"`
	Type     string           `json:"type"` // file, directory
	Size     int64            `json:"size,omitempty"`
	Language string           `json:"language,omitempty"`
	Children []*DirectoryNode `json:"children,omitempty"`
}

// DependencyGraph represents project dependencies
type DependencyGraph struct {
	Direct     []*Dependency `json:"direct"`
	Transitive []*Dependency `json:"transitive"`
	Conflicts  []*Conflict   `json:"conflicts,omitempty"`
	Outdated   []*Dependency `json:"outdated,omitempty"`
}

// Dependency represents a project dependency
type Dependency struct {
	Name            string           `json:"name"`
	Version         string           `json:"version"`
	LatestVersion   string           `json:"latest_version,omitempty"`
	Type            string           `json:"type"` // runtime, dev, peer, etc.
	Source          string           `json:"source,omitempty"`
	License         string           `json:"license,omitempty"`
	Vulnerabilities []*Vulnerability `json:"vulnerabilities,omitempty"`
}

// Conflict represents a dependency conflict
type Conflict struct {
	Package  string   `json:"package"`
	Versions []string `json:"versions"`
	Reason   string   `json:"reason"`
	Severity string   `json:"severity"`
}

// ProjectMetrics represents project-level metrics
type ProjectMetrics struct {
	LinesOfCode     int                    `json:"lines_of_code"`
	FileCount       int                    `json:"file_count"`
	FunctionCount   int                    `json:"function_count"`
	ClassCount      int                    `json:"class_count"`
	TestCoverage    float32                `json:"test_coverage,omitempty"`
	TechnicalDebt   time.Duration          `json:"technical_debt"`
	Maintainability float32                `json:"maintainability"`
	Complexity      *ComplexityMetrics     `json:"complexity"`
	Dependencies    int                    `json:"dependencies"`
	Languages       map[string]int         `json:"languages"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ProjectIssue represents a project-level issue
type ProjectIssue struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    IssueSeverity          `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	File        string                 `json:"file,omitempty"`
	Position    *Position              `json:"position,omitempty"`
	Category    string                 `json:"category"`
	Suggestion  string                 `json:"suggestion,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Enums for language server protocol

type DiagnosticSeverity int

const (
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

type DiagnosticTag int

const (
	DiagnosticTagUnnecessary DiagnosticTag = 1
	DiagnosticTagDeprecated  DiagnosticTag = 2
)

// Security-related types

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    VulnerabilitySeverity  `json:"severity"`
	CVSS        float32                `json:"cvss,omitempty"`
	CWE         string                 `json:"cwe,omitempty"`
	CVE         string                 `json:"cve,omitempty"`
	Package     string                 `json:"package,omitempty"`
	Version     string                 `json:"version,omitempty"`
	FixedIn     string                 `json:"fixed_in,omitempty"`
	References  []string               `json:"references,omitempty"`
	PublishedAt time.Time              `json:"published_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecretLeak represents a detected secret leak
type SecretLeak struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // api_key, password, token, etc.
	Value       string                 `json:"value,omitempty"`
	Position    *Position              `json:"position"`
	Confidence  float32                `json:"confidence"`
	Severity    IssueSeverity          `json:"severity"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PermissionIssue represents a permission-related issue
type PermissionIssue struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Permission  string                 `json:"permission"`
	Position    *Position              `json:"position"`
	Severity    IssueSeverity          `json:"severity"`
	Description string                 `json:"description"`
	Suggestion  string                 `json:"suggestion,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityReport represents a comprehensive security report
type SecurityReport struct {
	ProjectPath      string                 `json:"project_path"`
	Vulnerabilities  []*Vulnerability       `json:"vulnerabilities"`
	SecurityIssues   []*SecurityIssue       `json:"security_issues"`
	SecretLeaks      []*SecretLeak          `json:"secret_leaks"`
	PermissionIssues []*PermissionIssue     `json:"permission_issues"`
	Score            float32                `json:"score"`
	Grade            string                 `json:"grade"`
	GeneratedAt      time.Time              `json:"generated_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type VulnerabilitySeverity string

const (
	VulnerabilitySeverityLow      VulnerabilitySeverity = "low"
	VulnerabilitySeverityMedium   VulnerabilitySeverity = "medium"
	VulnerabilitySeverityHigh     VulnerabilitySeverity = "high"
	VulnerabilitySeverityCritical VulnerabilitySeverity = "critical"
)

// Additional types for comprehensive coding assistance

// TestCoverageResult represents test coverage analysis results
type TestCoverageResult struct {
	OverallCoverage  float32                  `json:"overall_coverage"`
	LineCoverage     float32                  `json:"line_coverage"`
	BranchCoverage   float32                  `json:"branch_coverage"`
	FunctionCoverage float32                  `json:"function_coverage"`
	FileCoverage     map[string]*FileCoverage `json:"file_coverage"`
	UncoveredLines   []*UncoveredLine         `json:"uncovered_lines"`
	Metadata         map[string]interface{}   `json:"metadata,omitempty"`
}

// FileCoverage represents coverage for a specific file
type FileCoverage struct {
	FilePath       string                 `json:"file_path"`
	LineCoverage   float32                `json:"line_coverage"`
	BranchCoverage float32                `json:"branch_coverage"`
	CoveredLines   []int                  `json:"covered_lines"`
	UncoveredLines []int                  `json:"uncovered_lines"`
	TotalLines     int                    `json:"total_lines"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// UncoveredLine represents an uncovered line of code
type UncoveredLine struct {
	FilePath string    `json:"file_path"`
	Line     int       `json:"line"`
	Content  string    `json:"content"`
	Position *Position `json:"position"`
}

// TestResults represents test execution results
type TestResults struct {
	TotalTests   int                    `json:"total_tests"`
	PassedTests  int                    `json:"passed_tests"`
	FailedTests  int                    `json:"failed_tests"`
	SkippedTests int                    `json:"skipped_tests"`
	Duration     time.Duration          `json:"duration"`
	Coverage     *TestCoverageResult    `json:"coverage,omitempty"`
	Failures     []*TestFailure         `json:"failures,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TestFailure represents a test failure
type TestFailure struct {
	TestName   string    `json:"test_name"`
	Message    string    `json:"message"`
	StackTrace string    `json:"stack_trace,omitempty"`
	Position   *Position `json:"position,omitempty"`
	Expected   string    `json:"expected,omitempty"`
	Actual     string    `json:"actual,omitempty"`
}

// CoverageReport represents a coverage report
type CoverageReport struct {
	Coverage    *TestCoverageResult    `json:"coverage"`
	Threshold   float32                `json:"threshold"`
	Passed      bool                   `json:"passed"`
	Suggestions []*Suggestion          `json:"suggestions,omitempty"`
	GeneratedAt time.Time              `json:"generated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BuildSystemInfo represents build system information
type BuildSystemInfo struct {
	Type        string                 `json:"type"` // maven, gradle, npm, cargo, etc.
	Version     string                 `json:"version,omitempty"`
	ConfigFiles []string               `json:"config_files"`
	Scripts     map[string]string      `json:"scripts,omitempty"`
	Targets     []string               `json:"targets,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ProjectType represents detected project type
type ProjectType struct {
	Language   string                 `json:"language"`
	Framework  string                 `json:"framework,omitempty"`
	Type       string                 `json:"type"` // web, cli, library, etc.
	Confidence float32                `json:"confidence"`
	Indicators []string               `json:"indicators"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// FunctionGenerationRequest represents a request to generate a function
type FunctionGenerationRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  []*Parameter           `json:"parameters,omitempty"`
	ReturnType  string                 `json:"return_type,omitempty"`
	Language    string                 `json:"language"`
	Style       *CodingStyle           `json:"style,omitempty"`
	Examples    []string               `json:"examples,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ClassGenerationRequest represents a request to generate a class
type ClassGenerationRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	BaseClass   string                 `json:"base_class,omitempty"`
	Interfaces  []string               `json:"interfaces,omitempty"`
	Properties  []*PropertySpec        `json:"properties,omitempty"`
	Methods     []*MethodSpec          `json:"methods,omitempty"`
	Language    string                 `json:"language"`
	Style       *CodingStyle           `json:"style,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// InterfaceGenerationRequest represents a request to generate an interface
type InterfaceGenerationRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Methods     []*MethodSpec          `json:"methods"`
	Properties  []*PropertySpec        `json:"properties,omitempty"`
	Language    string                 `json:"language"`
	Style       *CodingStyle           `json:"style,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PropertySpec represents a property specification
type PropertySpec struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
	IsStatic    bool   `json:"is_static"`
	IsReadOnly  bool   `json:"is_read_only"`
}

// MethodSpec represents a method specification
type MethodSpec struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Parameters  []*Parameter `json:"parameters,omitempty"`
	ReturnType  string       `json:"return_type,omitempty"`
	Visibility  string       `json:"visibility,omitempty"`
	IsStatic    bool         `json:"is_static"`
	IsAsync     bool         `json:"is_async"`
}

// Metrics and monitoring types

// Metric represents a metric data point
type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Type      MetricType        `json:"type"`
	Unit      string            `json:"unit,omitempty"`
}

// MetricFilter represents filters for metric queries
type MetricFilter struct {
	Names     []string          `json:"names,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	StartTime *time.Time        `json:"start_time,omitempty"`
	EndTime   *time.Time        `json:"end_time,omitempty"`
	Limit     int               `json:"limit,omitempty"`
}

// MetricAggregation represents metric aggregation parameters
type MetricAggregation struct {
	Function  AggregationFunction `json:"function"`
	GroupBy   []string            `json:"group_by,omitempty"`
	Interval  time.Duration       `json:"interval,omitempty"`
	TimeRange *TimeRange          `json:"time_range,omitempty"`
}

// AggregatedMetrics represents aggregated metric results
type AggregatedMetrics struct {
	Metrics   []*AggregatedMetric `json:"metrics"`
	TimeRange *TimeRange          `json:"time_range"`
	GroupBy   []string            `json:"group_by,omitempty"`
}

// AggregatedMetric represents a single aggregated metric
type AggregatedMetric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
}

// EventFilter represents filters for event queries
type EventFilter struct {
	Types     []string   `json:"types,omitempty"`
	Sources   []string   `json:"sources,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	SessionID string     `json:"session_id,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     int        `json:"limit,omitempty"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	Size      int           `json:"size"`
	MaxSize   int           `json:"max_size"`
	HitRate   float32       `json:"hit_rate"`
	MissRate  float32       `json:"miss_rate"`
	Hits      int64         `json:"hits"`
	Misses    int64         `json:"misses"`
	Evictions int64         `json:"evictions"`
	TTL       time.Duration `json:"ttl"`
}

// Plugin represents a plugin
type Plugin struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author,omitempty"`
	Enabled      bool                   `json:"enabled"`
	Capabilities *PluginCapabilities    `json:"capabilities"`
	Config       map[string]interface{} `json:"config,omitempty"`
	LoadedAt     time.Time              `json:"loaded_at"`
}

// PluginCapabilities represents plugin capabilities
type PluginCapabilities struct {
	Languages    []string `json:"languages,omitempty"`
	Commands     []string `json:"commands,omitempty"`
	Events       []string `json:"events,omitempty"`
	Permissions  []string `json:"permissions,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// Configuration represents system configuration
type Configuration struct {
	Scope     string                 `json:"scope"`
	Settings  map[string]interface{} `json:"settings"`
	Version   string                 `json:"version,omitempty"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// LanguageConfiguration represents language-specific configuration
type LanguageConfiguration struct {
	Language   string                 `json:"language"`
	Version    string                 `json:"version,omitempty"`
	Compiler   string                 `json:"compiler,omitempty"`
	Runtime    string                 `json:"runtime,omitempty"`
	Extensions []string               `json:"extensions"`
	Keywords   []string               `json:"keywords,omitempty"`
	Patterns   map[string]string      `json:"patterns,omitempty"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
}

// FormattingConfiguration represents formatting configuration
type FormattingConfiguration struct {
	Language   string                 `json:"language"`
	IndentSize int                    `json:"indent_size"`
	IndentType string                 `json:"indent_type"`
	LineLength int                    `json:"line_length"`
	Rules      map[string]interface{} `json:"rules,omitempty"`
	Enabled    bool                   `json:"enabled"`
}

// LintingConfiguration represents linting configuration
type LintingConfiguration struct {
	Language string                 `json:"language"`
	Rules    map[string]interface{} `json:"rules"`
	Severity map[string]string      `json:"severity,omitempty"`
	Enabled  bool                   `json:"enabled"`
}

// WorkspaceContext represents workspace context
type WorkspaceContext struct {
	Path         string                 `json:"path"`
	Language     string                 `json:"language,omitempty"`
	Framework    string                 `json:"framework,omitempty"`
	Dependencies []*Dependency          `json:"dependencies,omitempty"`
	Files        []*FileInfo            `json:"files,omitempty"`
	Symbols      []*Symbol              `json:"symbols,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// FileContext represents file context
type FileContext struct {
	Path      string                 `json:"path"`
	Language  string                 `json:"language"`
	Content   string                 `json:"content,omitempty"`
	Symbols   []*Symbol              `json:"symbols,omitempty"`
	Imports   []*Import              `json:"imports,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// SymbolContext represents symbol context
type SymbolContext struct {
	Symbol     *Symbol                `json:"symbol"`
	File       string                 `json:"file"`
	References []*Location            `json:"references,omitempty"`
	Definition *Location              `json:"definition,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Scope      string                 `json:"scope,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ContextChange represents a context change
type ContextChange struct {
	Type      ChangeType             `json:"type"`
	File      string                 `json:"file"`
	Position  *Position              `json:"position,omitempty"`
	Content   string                 `json:"content,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// FileInfo represents file information
type FileInfo struct {
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Language string    `json:"language,omitempty"`
	Modified time.Time `json:"modified"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Enums for metrics and events

type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

type AggregationFunction string

const (
	AggregationFunctionSum   AggregationFunction = "sum"
	AggregationFunctionAvg   AggregationFunction = "avg"
	AggregationFunctionMin   AggregationFunction = "min"
	AggregationFunctionMax   AggregationFunction = "max"
	AggregationFunctionCount AggregationFunction = "count"
)

// Additional types for search, review, and other components

// SearchOptions represents search options
type SearchOptions struct {
	MaxResults      int      `json:"max_results,omitempty"`
	IncludeComments bool     `json:"include_comments"`
	CaseSensitive   bool     `json:"case_sensitive"`
	WholeWord       bool     `json:"whole_word"`
	Regex           bool     `json:"regex"`
	FileTypes       []string `json:"file_types,omitempty"`
	ExcludePaths    []string `json:"exclude_paths,omitempty"`
}

// SearchResults represents search results
type SearchResults struct {
	Results        []*SearchResult `json:"results"`
	TotalCount     int             `json:"total_count"`
	Query          string          `json:"query"`
	ProcessingTime time.Duration   `json:"processing_time"`
}

// SearchResult represents a single search result
type SearchResult struct {
	File      string   `json:"file"`
	Line      int      `json:"line"`
	Column    int      `json:"column"`
	Content   string   `json:"content"`
	Context   []string `json:"context,omitempty"`
	Relevance float32  `json:"relevance"`
}

// SymbolLocation represents a symbol location
type SymbolLocation struct {
	Symbol   *Symbol   `json:"symbol"`
	File     string    `json:"file"`
	Position *Position `json:"position"`
}

// Usage represents a symbol usage
type Usage struct {
	Symbol   *Symbol   `json:"symbol"`
	File     string    `json:"file"`
	Position *Position `json:"position"`
	Type     string    `json:"type"` // read, write, call, etc.
	Context  string    `json:"context,omitempty"`
}

// SimilarCodeMatch represents a similar code match
type SimilarCodeMatch struct {
	File       string    `json:"file"`
	Position   *Position `json:"position"`
	Code       string    `json:"code"`
	Similarity float32   `json:"similarity"`
	Reason     string    `json:"reason"`
}

// PatternMatch represents a pattern match
type PatternMatch struct {
	Pattern  string    `json:"pattern"`
	File     string    `json:"file"`
	Position *Position `json:"position"`
	Match    string    `json:"match"`
	Context  string    `json:"context,omitempty"`
}

// CodeReviewRequest represents a code review request
type CodeReviewRequest struct {
	Code     string                 `json:"code"`
	Language string                 `json:"language"`
	Context  *ReviewContext         `json:"context,omitempty"`
	Options  *ReviewOptions         `json:"options,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CodeReviewResult represents code review results
type CodeReviewResult struct {
	Score          float32                `json:"score"`
	Grade          string                 `json:"grade,omitempty"`
	Issues         []*CodeIssue           `json:"issues"`
	Suggestions    []*Suggestion          `json:"suggestions"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// QualityReport represents a code quality report
type QualityReport struct {
	Score       float32                `json:"score"`
	Grade       string                 `json:"grade"`
	Issues      []*CodeIssue           `json:"issues"`
	Suggestions []*Suggestion          `json:"suggestions"`
	Metrics     *CodeMetrics           `json:"metrics,omitempty"`
	GeneratedAt time.Time              `json:"generated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CodeComparison represents a comparison between code versions
type CodeComparison struct {
	Additions   int                    `json:"additions"`
	Deletions   int                    `json:"deletions"`
	Changes     []*CodeChange          `json:"changes"`
	Similarity  float32                `json:"similarity"`
	GeneratedAt time.Time              `json:"generated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ChangeAnalysis represents analysis of code changes
type ChangeAnalysis struct {
	Type        string                 `json:"type"`
	Impact      ImpactLevel            `json:"impact"`
	Risk        string                 `json:"risk"`
	Suggestions []*Suggestion          `json:"suggestions"`
	GeneratedAt time.Time              `json:"generated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ReviewContext represents context for code review
type ReviewContext struct {
	ProjectPath string   `json:"project_path,omitempty"`
	FilePath    string   `json:"file_path,omitempty"`
	PullRequest string   `json:"pull_request,omitempty"`
	Branch      string   `json:"branch,omitempty"`
	Reviewer    string   `json:"reviewer,omitempty"`
	Guidelines  []string `json:"guidelines,omitempty"`
}

// ReviewOptions represents options for code review
type ReviewOptions struct {
	CheckSecurity      bool     `json:"check_security"`
	CheckPerformance   bool     `json:"check_performance"`
	CheckStyle         bool     `json:"check_style"`
	CheckComplexity    bool     `json:"check_complexity"`
	CheckDuplication   bool     `json:"check_duplication"`
	IncludeSuggestions bool     `json:"include_suggestions"`
	Severity           []string `json:"severity,omitempty"`
}
