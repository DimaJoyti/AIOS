package coding

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultCodeAnalyzer implements the CodeAnalyzer interface
type DefaultCodeAnalyzer struct {
	logger *logrus.Logger
	tracer trace.Tracer
	config *CodeAnalyzerConfig
}

// CodeAnalyzerConfig represents configuration for the code analyzer
type CodeAnalyzerConfig struct {
	MaxComplexity     int                    `json:"max_complexity"`
	EnablePatterns    bool                   `json:"enable_patterns"`
	EnableSecurity    bool                   `json:"enable_security"`
	EnablePerformance bool                   `json:"enable_performance"`
	LanguageRules     map[string]interface{} `json:"language_rules"`
}

// NewDefaultCodeAnalyzer creates a new default code analyzer
func NewDefaultCodeAnalyzer(logger *logrus.Logger) (CodeAnalyzer, error) {
	config := &CodeAnalyzerConfig{
		MaxComplexity:     10,
		EnablePatterns:    true,
		EnableSecurity:    true,
		EnablePerformance: true,
		LanguageRules:     make(map[string]interface{}),
	}

	analyzer := &DefaultCodeAnalyzer{
		logger: logger,
		tracer: otel.Tracer("coding.analyzer"),
		config: config,
	}

	return analyzer, nil
}

// ParseCode parses code and returns a structured representation
func (ca *DefaultCodeAnalyzer) ParseCode(ctx context.Context, code string, language string) (*ParsedCode, error) {
	ctx, span := ca.tracer.Start(ctx, "code_analyzer.parse_code")
	defer span.End()

	span.SetAttributes(
		attribute.String("language", language),
		attribute.Int("code_length", len(code)),
	)

	// Simple parsing implementation
	// In a real implementation, this would use language-specific parsers
	parsedCode := &ParsedCode{
		Language:   language,
		AST:        ca.buildSimpleAST(code),
		Symbols:    ca.extractSymbols(code, language),
		Functions:  ca.extractFunctions(code, language),
		Classes:    ca.extractClasses(code, language),
		Interfaces: ca.extractInterfaces(code, language),
		Variables:  ca.extractVariables(code, language),
		Imports:    ca.extractImports(code, language),
		Comments:   ca.extractComments(code, language),
		Metadata:   make(map[string]interface{}),
	}

	ca.logger.WithFields(logrus.Fields{
		"language":  language,
		"functions": len(parsedCode.Functions),
		"classes":   len(parsedCode.Classes),
		"symbols":   len(parsedCode.Symbols),
	}).Debug("Code parsed successfully")

	return parsedCode, nil
}

// AnalyzeComplexity analyzes code complexity
func (ca *DefaultCodeAnalyzer) AnalyzeComplexity(ctx context.Context, code string, language string) (*ComplexityAnalysis, error) {
	ctx, span := ca.tracer.Start(ctx, "code_analyzer.analyze_complexity")
	defer span.End()

	// Parse code first
	parsedCode, err := ca.ParseCode(ctx, code, language)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	analysis := &ComplexityAnalysis{
		Overall:   ca.calculateOverallComplexity(parsedCode),
		Functions: make(map[string]*ComplexityMetrics),
		Classes:   make(map[string]*ComplexityMetrics),
		Modules:   make(map[string]*ComplexityMetrics),
		Hotspots:  []*ComplexityHotspot{},
		Metadata:  make(map[string]interface{}),
	}

	// Analyze function complexity
	for _, function := range parsedCode.Functions {
		metrics := ca.calculateFunctionComplexity(function)
		analysis.Functions[function.Name] = metrics

		// Check for complexity hotspots
		if metrics.Cyclomatic > ca.config.MaxComplexity {
			hotspot := &ComplexityHotspot{
				Name:       function.Name,
				Type:       "function",
				Position:   function.Position,
				Complexity: metrics,
				Severity:   ca.getComplexitySeverity(metrics.Cyclomatic),
				Metadata:   make(map[string]interface{}),
			}
			analysis.Hotspots = append(analysis.Hotspots, hotspot)
		}
	}

	// Analyze class complexity
	for _, class := range parsedCode.Classes {
		metrics := ca.calculateClassComplexity(class)
		analysis.Classes[class.Name] = metrics
	}

	return analysis, nil
}

// DetectPatterns detects code patterns and anti-patterns
func (ca *DefaultCodeAnalyzer) DetectPatterns(ctx context.Context, code string, language string) ([]*CodePattern, error) {
	ctx, span := ca.tracer.Start(ctx, "code_analyzer.detect_patterns")
	defer span.End()

	var patterns []*CodePattern

	// Detect common patterns based on language
	switch language {
	case "go":
		patterns = append(patterns, ca.detectGoPatterns(code)...)
	case "python":
		patterns = append(patterns, ca.detectPythonPatterns(code)...)
	case "javascript", "typescript":
		patterns = append(patterns, ca.detectJavaScriptPatterns(code)...)
	case "java":
		patterns = append(patterns, ca.detectJavaPatterns(code)...)
	default:
		patterns = append(patterns, ca.detectGenericPatterns(code)...)
	}

	ca.logger.WithFields(logrus.Fields{
		"language":       language,
		"patterns_found": len(patterns),
	}).Debug("Pattern detection completed")

	return patterns, nil
}

// AnalyzeSecurity analyzes code for security issues
func (ca *DefaultCodeAnalyzer) AnalyzeSecurity(ctx context.Context, code string, language string) ([]*SecurityIssue, error) {
	ctx, span := ca.tracer.Start(ctx, "code_analyzer.analyze_security")
	defer span.End()

	var issues []*SecurityIssue

	// Detect common security issues
	issues = append(issues, ca.detectSQLInjection(code)...)
	issues = append(issues, ca.detectXSS(code)...)
	issues = append(issues, ca.detectHardcodedSecrets(code)...)
	issues = append(issues, ca.detectInsecureCrypto(code)...)

	return issues, nil
}

// AnalyzePerformance analyzes code for performance issues
func (ca *DefaultCodeAnalyzer) AnalyzePerformance(ctx context.Context, code string, language string) ([]*PerformanceIssue, error) {
	ctx, span := ca.tracer.Start(ctx, "code_analyzer.analyze_performance")
	defer span.End()

	var issues []*PerformanceIssue

	// Detect common performance issues
	issues = append(issues, ca.detectInefficiientLoops(code)...)
	issues = append(issues, ca.detectMemoryLeaks(code, language)...)
	issues = append(issues, ca.detectBlockingOperations(code)...)

	return issues, nil
}

// ExtractMetrics extracts code metrics
func (ca *DefaultCodeAnalyzer) ExtractMetrics(ctx context.Context, code string, language string) (*CodeMetrics, error) {
	ctx, span := ca.tracer.Start(ctx, "code_analyzer.extract_metrics")
	defer span.End()

	lines := strings.Split(code, "\n")

	metrics := &CodeMetrics{
		LinesOfCode:     ca.countLinesOfCode(lines),
		LinesOfComments: ca.countLinesOfComments(lines, language),
		LinesBlank:      ca.countBlankLines(lines),
		Metadata:        make(map[string]interface{}),
	}

	// Parse code to get more detailed metrics
	parsedCode, err := ca.ParseCode(ctx, code, language)
	if err == nil {
		metrics.FunctionCount = len(parsedCode.Functions)
		metrics.ClassCount = len(parsedCode.Classes)
		metrics.VariableCount = len(parsedCode.Variables)

		// Calculate complexity
		complexity, err := ca.AnalyzeComplexity(ctx, code, language)
		if err == nil && complexity.Overall != nil {
			metrics.CyclomaticComplexity = complexity.Overall.Cyclomatic
			metrics.CognitiveComplexity = complexity.Overall.Cognitive
		}
	}

	// Calculate maintainability index (simplified)
	metrics.Maintainability = ca.calculateMaintainabilityIndex(metrics)

	return metrics, nil
}

// Helper methods for parsing and analysis

// buildSimpleAST builds a simple AST representation
func (ca *DefaultCodeAnalyzer) buildSimpleAST(code string) *ASTNode {
	// Simplified AST building
	return &ASTNode{
		Type:     "root",
		Value:    "",
		Position: &Position{Line: 1, Column: 1},
		Children: []*ASTNode{},
		Attributes: map[string]interface{}{
			"language": "unknown",
		},
	}
}

// extractSymbols extracts symbols from code
func (ca *DefaultCodeAnalyzer) extractSymbols(code string, language string) []*Symbol {
	var symbols []*Symbol

	// Simple symbol extraction based on common patterns
	// Function definitions
	funcRegex := regexp.MustCompile(`(?m)^func\s+(\w+)\s*\(`)
	matches := funcRegex.FindAllStringSubmatch(code, -1)
	for _, match := range matches {
		if len(match) > 1 {
			symbol := &Symbol{
				Name:       match[1],
				Type:       "function",
				Kind:       SymbolKindFunction,
				Position:   &Position{Line: 1, Column: 1}, // Simplified
				Scope:      "global",
				Visibility: "public",
				Metadata:   make(map[string]interface{}),
			}
			symbols = append(symbols, symbol)
		}
	}

	// Variable definitions
	varRegex := regexp.MustCompile(`(?m)^(?:var|let|const)\s+(\w+)`)
	matches = varRegex.FindAllStringSubmatch(code, -1)
	for _, match := range matches {
		if len(match) > 1 {
			symbol := &Symbol{
				Name:       match[1],
				Type:       "variable",
				Kind:       SymbolKindVariable,
				Position:   &Position{Line: 1, Column: 1}, // Simplified
				Scope:      "local",
				Visibility: "private",
				Metadata:   make(map[string]interface{}),
			}
			symbols = append(symbols, symbol)
		}
	}

	return symbols
}

// extractFunctions extracts function definitions
func (ca *DefaultCodeAnalyzer) extractFunctions(code string, language string) []*ParsedFunction {
	var functions []*ParsedFunction

	// Simple function extraction for Go
	if language == "go" {
		funcRegex := regexp.MustCompile(`(?m)^func\s+(\w+)\s*\([^)]*\)(?:\s*[^{]*)?{`)
		matches := funcRegex.FindAllStringSubmatch(code, -1)
		for _, match := range matches {
			if len(match) > 1 {
				function := &ParsedFunction{
					Name:       match[1],
					Signature:  match[0],
					Parameters: []*Parameter{},
					ReturnType: "",
					Body:       "",
					Position:   &Position{Line: 1, Column: 1},
					Visibility: "public",
					IsAsync:    false,
					IsStatic:   false,
					Decorators: []string{},
					Complexity: &ComplexityMetrics{Cyclomatic: 1},
					Metadata:   make(map[string]interface{}),
				}
				functions = append(functions, function)
			}
		}
	}

	return functions
}

// extractClasses extracts class definitions
func (ca *DefaultCodeAnalyzer) extractClasses(code string, language string) []*ParsedClass {
	var classes []*ParsedClass

	// Simple class extraction
	classRegex := regexp.MustCompile(`(?m)^(?:class|type)\s+(\w+)`)
	matches := classRegex.FindAllStringSubmatch(code, -1)
	for _, match := range matches {
		if len(match) > 1 {
			class := &ParsedClass{
				Name:        match[1],
				BaseClasses: []string{},
				Interfaces:  []string{},
				Methods:     []*ParsedFunction{},
				Properties:  []*ParsedProperty{},
				Position:    &Position{Line: 1, Column: 1},
				Visibility:  "public",
				IsAbstract:  false,
				IsInterface: false,
				Decorators:  []string{},
				Metadata:    make(map[string]interface{}),
			}
			classes = append(classes, class)
		}
	}

	return classes
}

// extractInterfaces extracts interface definitions
func (ca *DefaultCodeAnalyzer) extractInterfaces(code string, language string) []*ParsedInterface {
	var interfaces []*ParsedInterface

	// Simple interface extraction
	interfaceRegex := regexp.MustCompile(`(?m)^interface\s+(\w+)`)
	matches := interfaceRegex.FindAllStringSubmatch(code, -1)
	for _, match := range matches {
		if len(match) > 1 {
			iface := &ParsedInterface{
				Name:       match[1],
				Methods:    []*MethodSignature{},
				Properties: []*PropertySignature{},
				Extends:    []string{},
				Position:   &Position{Line: 1, Column: 1},
				Metadata:   make(map[string]interface{}),
			}
			interfaces = append(interfaces, iface)
		}
	}

	return interfaces
}

// extractVariables extracts variable definitions
func (ca *DefaultCodeAnalyzer) extractVariables(code string, language string) []*ParsedVariable {
	var variables []*ParsedVariable

	// Simple variable extraction
	varRegex := regexp.MustCompile(`(?m)^(?:var|let|const)\s+(\w+)`)
	matches := varRegex.FindAllStringSubmatch(code, -1)
	for _, match := range matches {
		if len(match) > 1 {
			variable := &ParsedVariable{
				Name:       match[1],
				Type:       "unknown",
				Value:      "",
				Position:   &Position{Line: 1, Column: 1},
				Scope:      "local",
				Visibility: "private",
				IsConstant: strings.Contains(match[0], "const"),
				IsStatic:   false,
				Metadata:   make(map[string]interface{}),
			}
			variables = append(variables, variable)
		}
	}

	return variables
}

// extractImports extracts import statements
func (ca *DefaultCodeAnalyzer) extractImports(code string, language string) []*Import {
	var imports []*Import

	// Simple import extraction
	importRegex := regexp.MustCompile(`(?m)^import\s+(?:"([^"]+)"|(\w+))`)
	matches := importRegex.FindAllStringSubmatch(code, -1)
	for _, match := range matches {
		module := ""
		if len(match) > 1 && match[1] != "" {
			module = match[1]
		} else if len(match) > 2 && match[2] != "" {
			module = match[2]
		}

		if module != "" {
			imp := &Import{
				Module:     module,
				Alias:      "",
				Items:      []string{},
				Position:   &Position{Line: 1, Column: 1},
				IsWildcard: false,
			}
			imports = append(imports, imp)
		}
	}

	return imports
}

// extractComments extracts comments from code
func (ca *DefaultCodeAnalyzer) extractComments(code string, language string) []*Comment {
	var comments []*Comment

	// Extract line comments
	lineCommentRegex := regexp.MustCompile(`//.*$`)
	matches := lineCommentRegex.FindAllString(code, -1)
	for _, match := range matches {
		comment := &Comment{
			Text:     strings.TrimPrefix(match, "//"),
			Type:     "line",
			Position: &Position{Line: 1, Column: 1},
		}
		comments = append(comments, comment)
	}

	// Extract block comments
	blockCommentRegex := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	matches = blockCommentRegex.FindAllString(code, -1)
	for _, match := range matches {
		comment := &Comment{
			Text:     strings.TrimSuffix(strings.TrimPrefix(match, "/*"), "*/"),
			Type:     "block",
			Position: &Position{Line: 1, Column: 1},
		}
		comments = append(comments, comment)
	}

	return comments
}

// Complexity calculation methods

// calculateOverallComplexity calculates overall code complexity
func (ca *DefaultCodeAnalyzer) calculateOverallComplexity(parsedCode *ParsedCode) *ComplexityMetrics {
	totalCyclomatic := 0
	totalCognitive := 0
	totalStatements := 0

	for _, function := range parsedCode.Functions {
		if function.Complexity != nil {
			totalCyclomatic += function.Complexity.Cyclomatic
			totalCognitive += function.Complexity.Cognitive
			totalStatements += function.Complexity.Statements
		}
	}

	return &ComplexityMetrics{
		Cyclomatic: totalCyclomatic,
		Cognitive:  totalCognitive,
		Halstead:   0.0, // Would be calculated with proper analysis
		Nesting:    0,   // Would be calculated with proper analysis
		Parameters: 0,   // Would be calculated with proper analysis
		Variables:  len(parsedCode.Variables),
		Branches:   0, // Would be calculated with proper analysis
		Statements: totalStatements,
	}
}

// calculateFunctionComplexity calculates complexity for a function
func (ca *DefaultCodeAnalyzer) calculateFunctionComplexity(function *ParsedFunction) *ComplexityMetrics {
	// Simple complexity calculation based on function body
	body := function.Body

	// Count decision points for cyclomatic complexity
	cyclomatic := 1 // Base complexity
	cyclomatic += strings.Count(body, "if")
	cyclomatic += strings.Count(body, "for")
	cyclomatic += strings.Count(body, "while")
	cyclomatic += strings.Count(body, "switch")
	cyclomatic += strings.Count(body, "case")
	cyclomatic += strings.Count(body, "&&")
	cyclomatic += strings.Count(body, "||")

	// Cognitive complexity (simplified)
	cognitive := cyclomatic

	return &ComplexityMetrics{
		Cyclomatic: cyclomatic,
		Cognitive:  cognitive,
		Halstead:   0.0,
		Nesting:    ca.calculateNestingDepth(body),
		Parameters: len(function.Parameters),
		Variables:  strings.Count(body, "var") + strings.Count(body, "let") + strings.Count(body, "const"),
		Branches:   strings.Count(body, "if") + strings.Count(body, "switch"),
		Statements: strings.Count(body, ";") + strings.Count(body, "\n"),
	}
}

// calculateClassComplexity calculates complexity for a class
func (ca *DefaultCodeAnalyzer) calculateClassComplexity(class *ParsedClass) *ComplexityMetrics {
	totalCyclomatic := 0
	totalCognitive := 0
	totalStatements := 0

	for _, method := range class.Methods {
		if method.Complexity != nil {
			totalCyclomatic += method.Complexity.Cyclomatic
			totalCognitive += method.Complexity.Cognitive
			totalStatements += method.Complexity.Statements
		}
	}

	return &ComplexityMetrics{
		Cyclomatic: totalCyclomatic,
		Cognitive:  totalCognitive,
		Halstead:   0.0,
		Nesting:    0,
		Parameters: 0,
		Variables:  len(class.Properties),
		Branches:   0,
		Statements: totalStatements,
	}
}

// calculateNestingDepth calculates nesting depth
func (ca *DefaultCodeAnalyzer) calculateNestingDepth(code string) int {
	maxDepth := 0
	currentDepth := 0

	for _, char := range code {
		switch char {
		case '{':
			currentDepth++
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		case '}':
			currentDepth--
		}
	}

	return maxDepth
}

// getComplexitySeverity determines severity based on complexity
func (ca *DefaultCodeAnalyzer) getComplexitySeverity(complexity int) string {
	switch {
	case complexity > 20:
		return "critical"
	case complexity > 15:
		return "high"
	case complexity > 10:
		return "medium"
	default:
		return "low"
	}
}

// Pattern detection methods

// detectGoPatterns detects Go-specific patterns
func (ca *DefaultCodeAnalyzer) detectGoPatterns(code string) []*CodePattern {
	var patterns []*CodePattern

	// Detect error handling pattern
	if strings.Contains(code, "if err != nil") {
		pattern := &CodePattern{
			Name:          "Error Handling",
			Type:          "idiom",
			Description:   "Go idiomatic error handling pattern",
			Position:      &Position{Line: 1, Column: 1},
			IsAntiPattern: false,
			Confidence:    0.9,
			Examples:      []string{"if err != nil { return err }"},
			Metadata:      map[string]interface{}{"language": "go"},
		}
		patterns = append(patterns, pattern)
	}

	// Detect defer pattern
	if strings.Contains(code, "defer") {
		pattern := &CodePattern{
			Name:          "Defer Pattern",
			Type:          "resource_management",
			Description:   "Go defer statement for cleanup",
			Position:      &Position{Line: 1, Column: 1},
			IsAntiPattern: false,
			Confidence:    0.8,
			Examples:      []string{"defer file.Close()"},
			Metadata:      map[string]interface{}{"language": "go"},
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

// detectPythonPatterns detects Python-specific patterns
func (ca *DefaultCodeAnalyzer) detectPythonPatterns(code string) []*CodePattern {
	var patterns []*CodePattern

	// Detect list comprehension
	listCompRegex := regexp.MustCompile(`\[.*for.*in.*\]`)
	if listCompRegex.MatchString(code) {
		pattern := &CodePattern{
			Name:          "List Comprehension",
			Type:          "idiom",
			Description:   "Python list comprehension pattern",
			Position:      &Position{Line: 1, Column: 1},
			IsAntiPattern: false,
			Confidence:    0.9,
			Examples:      []string{"[x for x in range(10)]"},
			Metadata:      map[string]interface{}{"language": "python"},
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

// detectJavaScriptPatterns detects JavaScript-specific patterns
func (ca *DefaultCodeAnalyzer) detectJavaScriptPatterns(code string) []*CodePattern {
	var patterns []*CodePattern

	// Detect callback pattern
	if strings.Contains(code, "function(") && strings.Contains(code, "callback") {
		pattern := &CodePattern{
			Name:          "Callback Pattern",
			Type:          "async",
			Description:   "JavaScript callback pattern",
			Position:      &Position{Line: 1, Column: 1},
			IsAntiPattern: false,
			Confidence:    0.7,
			Examples:      []string{"function(callback) { callback(result); }"},
			Metadata:      map[string]interface{}{"language": "javascript"},
		}
		patterns = append(patterns, pattern)
	}

	// Detect promise pattern
	if strings.Contains(code, "Promise") || strings.Contains(code, ".then(") {
		pattern := &CodePattern{
			Name:          "Promise Pattern",
			Type:          "async",
			Description:   "JavaScript Promise pattern",
			Position:      &Position{Line: 1, Column: 1},
			IsAntiPattern: false,
			Confidence:    0.8,
			Examples:      []string{"promise.then(result => { ... })"},
			Metadata:      map[string]interface{}{"language": "javascript"},
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

// detectJavaPatterns detects Java-specific patterns
func (ca *DefaultCodeAnalyzer) detectJavaPatterns(code string) []*CodePattern {
	var patterns []*CodePattern

	// Detect singleton pattern
	singletonRegex := regexp.MustCompile(`private static.*getInstance`)
	if singletonRegex.MatchString(code) {
		pattern := &CodePattern{
			Name:          "Singleton Pattern",
			Type:          "design_pattern",
			Description:   "Singleton design pattern implementation",
			Position:      &Position{Line: 1, Column: 1},
			IsAntiPattern: false,
			Confidence:    0.8,
			Examples:      []string{"private static Instance getInstance()"},
			Metadata:      map[string]interface{}{"language": "java"},
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

// detectGenericPatterns detects language-agnostic patterns
func (ca *DefaultCodeAnalyzer) detectGenericPatterns(code string) []*CodePattern {
	var patterns []*CodePattern

	// Detect long method anti-pattern
	lines := strings.Split(code, "\n")
	if len(lines) > 50 {
		pattern := &CodePattern{
			Name:          "Long Method",
			Type:          "anti_pattern",
			Description:   "Method is too long and should be refactored",
			Position:      &Position{Line: 1, Column: 1},
			IsAntiPattern: true,
			Confidence:    0.7,
			Examples:      []string{"Methods should be kept under 50 lines"},
			Metadata:      map[string]interface{}{"lines": len(lines)},
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

// Security detection methods

// detectSQLInjection detects potential SQL injection vulnerabilities
func (ca *DefaultCodeAnalyzer) detectSQLInjection(code string) []*SecurityIssue {
	var issues []*SecurityIssue

	// Simple SQL injection detection
	sqlPatterns := []string{
		`"SELECT.*\+.*"`,
		`'SELECT.*\+.*'`,
		`fmt\.Sprintf.*SELECT`,
		`string.*concatenation.*SQL`,
	}

	for _, pattern := range sqlPatterns {
		regex := regexp.MustCompile(pattern)
		if regex.MatchString(code) {
			issue := &SecurityIssue{
				ID:          uuid.New().String(),
				Type:        SecurityIssueTypeInjection,
				Severity:    IssueSeverityError,
				Title:       "Potential SQL Injection",
				Description: "SQL query construction using string concatenation may be vulnerable to injection",
				Position:    &Position{Line: 1, Column: 1},
				CWE:         "CWE-89",
				CVSS:        7.5,
				Remediation: "Use parameterized queries or prepared statements",
				References:  []string{"https://owasp.org/www-community/attacks/SQL_Injection"},
				Metadata:    map[string]interface{}{"pattern": pattern},
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

// detectXSS detects potential XSS vulnerabilities
func (ca *DefaultCodeAnalyzer) detectXSS(code string) []*SecurityIssue {
	var issues []*SecurityIssue

	// Simple XSS detection
	xssPatterns := []string{
		`innerHTML.*\+`,
		`document\.write.*\+`,
		`eval\(.*\+`,
		`setTimeout.*\+`,
	}

	for _, pattern := range xssPatterns {
		regex := regexp.MustCompile(pattern)
		if regex.MatchString(code) {
			issue := &SecurityIssue{
				ID:          uuid.New().String(),
				Type:        SecurityIssueTypeXSS,
				Severity:    IssueSeverityWarning,
				Title:       "Potential XSS Vulnerability",
				Description: "Dynamic content insertion may be vulnerable to XSS attacks",
				Position:    &Position{Line: 1, Column: 1},
				CWE:         "CWE-79",
				CVSS:        6.1,
				Remediation: "Sanitize user input and use safe DOM manipulation methods",
				References:  []string{"https://owasp.org/www-community/attacks/xss/"},
				Metadata:    map[string]interface{}{"pattern": pattern},
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

// detectHardcodedSecrets detects hardcoded secrets
func (ca *DefaultCodeAnalyzer) detectHardcodedSecrets(code string) []*SecurityIssue {
	var issues []*SecurityIssue

	// Simple secret detection patterns
	secretPatterns := map[string]string{
		`password\s*=\s*"[^"]+"`:    "Hardcoded Password",
		`api_key\s*=\s*"[^"]+"`:     "Hardcoded API Key",
		`secret\s*=\s*"[^"]+"`:      "Hardcoded Secret",
		`token\s*=\s*"[^"]+"`:       "Hardcoded Token",
		`private_key\s*=\s*"[^"]+"`: "Hardcoded Private Key",
	}

	for pattern, title := range secretPatterns {
		regex := regexp.MustCompile(pattern)
		if regex.MatchString(code) {
			issue := &SecurityIssue{
				ID:          uuid.New().String(),
				Type:        SecurityIssueTypeInsecureStorage,
				Severity:    IssueSeverityError,
				Title:       title,
				Description: "Hardcoded secrets should not be stored in source code",
				Position:    &Position{Line: 1, Column: 1},
				CWE:         "CWE-798",
				CVSS:        7.5,
				Remediation: "Use environment variables or secure secret management",
				References:  []string{"https://owasp.org/www-community/vulnerabilities/Use_of_hard-coded_password"},
				Metadata:    map[string]interface{}{"pattern": pattern},
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

// detectInsecureCrypto detects insecure cryptographic practices
func (ca *DefaultCodeAnalyzer) detectInsecureCrypto(code string) []*SecurityIssue {
	var issues []*SecurityIssue

	// Simple crypto detection patterns
	cryptoPatterns := map[string]string{
		`MD5`:  "Weak Hash Algorithm (MD5)",
		`SHA1`: "Weak Hash Algorithm (SHA1)",
		`DES`:  "Weak Encryption Algorithm (DES)",
		`RC4`:  "Weak Encryption Algorithm (RC4)",
	}

	for pattern, title := range cryptoPatterns {
		if strings.Contains(code, pattern) {
			issue := &SecurityIssue{
				ID:          uuid.New().String(),
				Type:        SecurityIssueTypeWeakCrypto,
				Severity:    IssueSeverityWarning,
				Title:       title,
				Description: "Weak cryptographic algorithm detected",
				Position:    &Position{Line: 1, Column: 1},
				CWE:         "CWE-327",
				CVSS:        5.3,
				Remediation: "Use strong cryptographic algorithms like SHA-256, AES",
				References:  []string{"https://owasp.org/www-community/vulnerabilities/Use_of_a_Broken_or_Risky_Cryptographic_Algorithm"},
				Metadata:    map[string]interface{}{"algorithm": pattern},
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

// Performance detection methods

// detectInefficiientLoops detects inefficient loop patterns
func (ca *DefaultCodeAnalyzer) detectInefficiientLoops(code string) []*PerformanceIssue {
	var issues []*PerformanceIssue

	// Detect nested loops
	nestedLoopRegex := regexp.MustCompile(`for.*{[^}]*for.*{`)
	if nestedLoopRegex.MatchString(code) {
		issue := &PerformanceIssue{
			ID:          uuid.New().String(),
			Type:        PerformanceIssueTypeInefficient,
			Severity:    IssueSeverityWarning,
			Title:       "Nested Loops Detected",
			Description: "Nested loops can cause performance issues with large datasets",
			Position:    &Position{Line: 1, Column: 1},
			Impact:      "O(n²) time complexity",
			Suggestion:  "Consider using more efficient algorithms or data structures",
			Metadata:    map[string]interface{}{"pattern": "nested_loops"},
		}
		issues = append(issues, issue)
	}

	return issues
}

// detectMemoryLeaks detects potential memory leaks
func (ca *DefaultCodeAnalyzer) detectMemoryLeaks(code string, language string) []*PerformanceIssue {
	var issues []*PerformanceIssue

	// Language-specific memory leak detection
	switch language {
	case "javascript":
		// Detect potential memory leaks in JavaScript
		if strings.Contains(code, "setInterval") && !strings.Contains(code, "clearInterval") {
			issue := &PerformanceIssue{
				ID:          uuid.New().String(),
				Type:        PerformanceIssueTypeMemoryLeak,
				Severity:    IssueSeverityWarning,
				Title:       "Potential Memory Leak",
				Description: "setInterval without clearInterval may cause memory leaks",
				Position:    &Position{Line: 1, Column: 1},
				Impact:      "Memory usage will grow over time",
				Suggestion:  "Always clear intervals when no longer needed",
				Metadata:    map[string]interface{}{"type": "interval_leak"},
			}
			issues = append(issues, issue)
		}
	case "go":
		// Detect goroutine leaks
		if strings.Contains(code, "go func") && !strings.Contains(code, "context") {
			issue := &PerformanceIssue{
				ID:          uuid.New().String(),
				Type:        PerformanceIssueTypeMemoryLeak,
				Severity:    IssueSeverityInfo,
				Title:       "Potential Goroutine Leak",
				Description: "Goroutines without proper cancellation may leak",
				Position:    &Position{Line: 1, Column: 1},
				Impact:      "Goroutines may not terminate properly",
				Suggestion:  "Use context for goroutine cancellation",
				Metadata:    map[string]interface{}{"type": "goroutine_leak"},
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

// detectBlockingOperations detects blocking operations
func (ca *DefaultCodeAnalyzer) detectBlockingOperations(code string) []*PerformanceIssue {
	var issues []*PerformanceIssue

	// Detect synchronous operations that might block
	blockingPatterns := []string{
		`time\.Sleep`,
		`http\.Get`,
		`io\.ReadAll`,
		`os\.ReadFile`,
	}

	for _, pattern := range blockingPatterns {
		if strings.Contains(code, pattern) {
			issue := &PerformanceIssue{
				ID:          uuid.New().String(),
				Type:        PerformanceIssueTypeBlocking,
				Severity:    IssueSeverityInfo,
				Title:       "Blocking Operation Detected",
				Description: fmt.Sprintf("Synchronous operation '%s' may block execution", pattern),
				Position:    &Position{Line: 1, Column: 1},
				Impact:      "May cause application to become unresponsive",
				Suggestion:  "Consider using asynchronous alternatives",
				Metadata:    map[string]interface{}{"operation": pattern},
			}
			issues = append(issues, issue)
		}
	}

	return issues
}

// Metrics calculation helper methods

// countLinesOfCode counts lines of code (excluding comments and blank lines)
func (ca *DefaultCodeAnalyzer) countLinesOfCode(lines []string) int {
	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") {
			count++
		}
	}
	return count
}

// countLinesOfComments counts lines of comments
func (ca *DefaultCodeAnalyzer) countLinesOfComments(lines []string, language string) int {
	count := 0
	inBlockComment := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle block comments
		if strings.Contains(trimmed, "/*") {
			inBlockComment = true
		}
		if inBlockComment {
			count++
		}
		if strings.Contains(trimmed, "*/") {
			inBlockComment = false
			continue
		}

		// Handle line comments
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			count++
		}
	}

	return count
}

// countBlankLines counts blank lines
func (ca *DefaultCodeAnalyzer) countBlankLines(lines []string) int {
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			count++
		}
	}
	return count
}

// calculateMaintainabilityIndex calculates maintainability index
func (ca *DefaultCodeAnalyzer) calculateMaintainabilityIndex(metrics *CodeMetrics) float32 {
	// Simplified maintainability index calculation
	// Real implementation would use Halstead metrics and other factors

	if metrics.LinesOfCode == 0 {
		return 100.0
	}

	// Base score
	score := 100.0

	// Penalize high complexity
	if metrics.CyclomaticComplexity > 0 {
		score -= float64(metrics.CyclomaticComplexity) * 2.0
	}

	// Penalize low comment ratio
	commentRatio := float64(metrics.LinesOfComments) / float64(metrics.LinesOfCode)
	if commentRatio < 0.1 {
		score -= 10.0
	}

	// Penalize large files
	if metrics.LinesOfCode > 500 {
		score -= float64(metrics.LinesOfCode-500) * 0.01
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return float32(score)
}
