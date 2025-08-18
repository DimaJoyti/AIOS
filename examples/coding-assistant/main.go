package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aios/aios/pkg/coding"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	fmt.Println("🤖 AIOS AI Coding Assistant Demo")
	fmt.Println("=================================")

	// Run the comprehensive demo
	if err := runCodingAssistantDemo(logger); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n✅ AI Coding Assistant Demo completed successfully!")
}

func runCodingAssistantDemo(logger *logrus.Logger) error {
	ctx := context.Background()

	// Step 1: Create Coding Assistant
	fmt.Println("\n1. Creating AI Coding Assistant...")
	config := &coding.CodingAssistantConfig{
		DefaultLanguage:    "go",
		MaxConcurrentOps:   5,
		CacheEnabled:       true,
		CacheTTL:           1 * time.Hour,
		SecurityEnabled:    true,
		MetricsEnabled:     true,
		PluginsEnabled:     false, // Disable for demo
		AIModelProvider:    "openai",
		AIModelName:        "gpt-4",
		SupportedLanguages: []string{"go", "python", "javascript", "typescript", "java"},
	}

	assistant, err := coding.NewDefaultCodingAssistant(config, logger)
	if err != nil {
		return fmt.Errorf("failed to create coding assistant: %w", err)
	}
	fmt.Println("✓ AI Coding Assistant created successfully")

	// Step 2: Analyze Code
	fmt.Println("\n2. Analyzing Code...")
	sampleCode := `
package main

import (
	"fmt"
	"time"
)

func main() {
	for i := 0; i < 1000000; i++ {
		for j := 0; j < 1000; j++ {
			if i*j > 500000 {
				fmt.Printf("Result: %d\n", i*j)
				break
			}
		}
		time.Sleep(time.Millisecond)
	}
}

func calculateSum(a, b int) int {
	if a > 0 && b > 0 {
		if a > 100 {
			if b > 100 {
				if a+b > 1000 {
					return a + b + 100
				} else {
					return a + b + 50
				}
			} else {
				return a + b + 25
			}
		} else {
			return a + b
		}
	}
	return 0
}
`

	analysisRequest := &coding.CodeAnalysisRequest{
		Code:     sampleCode,
		Language: "go",
		FilePath: "main.go",
		Options: &coding.AnalysisOptions{
			IncludeSecurity:    true,
			IncludePerformance: true,
			IncludeComplexity:  true,
			IncludeStyle:       true,
		},
	}

	analysisResult, err := assistant.AnalyzeCode(ctx, analysisRequest)
	if err != nil {
		fmt.Printf("   ⚠️  Code analysis failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Code analysis completed:\n")
		fmt.Printf("     - Functions found: %d\n", len(analysisResult.ParsedCode.Functions))
		fmt.Printf("     - Issues detected: %d\n", len(analysisResult.Issues))
		fmt.Printf("     - Security issues: %d\n", len(analysisResult.SecurityIssues))
		fmt.Printf("     - Performance issues: %d\n", len(analysisResult.PerformanceIssues))
		fmt.Printf("     - Suggestions: %d\n", len(analysisResult.Suggestions))
		fmt.Printf("     - Cyclomatic complexity: %d\n", analysisResult.Metrics.CyclomaticComplexity)
		fmt.Printf("     - Lines of code: %d\n", analysisResult.Metrics.LinesOfCode)
		fmt.Printf("     - Processing time: %v\n", analysisResult.ProcessingTime)

		// Display some issues
		for i, issue := range analysisResult.Issues {
			if i >= 3 { // Show only first 3 issues
				break
			}
			fmt.Printf("     - Issue %d: %s (%s)\n", i+1, issue.Message, issue.Severity)
		}
	}

	// Step 3: Generate Code
	fmt.Println("\n3. Generating Code...")
	generationRequest := &coding.CodeGenerationRequest{
		Prompt:   "Create a function that calculates the factorial of a number using recursion",
		Language: "go",
		Style: &coding.CodingStyle{
			IndentSize:   4,
			IndentType:   "spaces",
			LineLength:   100,
			NamingStyle:  "camelCase",
			BraceStyle:   "K&R",
			CommentStyle: "line",
		},
	}

	generationResult, err := assistant.GenerateCode(ctx, generationRequest)
	if err != nil {
		fmt.Printf("   ⚠️  Code generation failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Code generation completed:\n")
		fmt.Printf("     - Generated code:\n%s\n", generationResult.GeneratedCode)
		fmt.Printf("     - Explanation: %s\n", generationResult.Explanation)
		fmt.Printf("     - Confidence: %.2f\n", generationResult.Confidence)
		fmt.Printf("     - Processing time: %v\n", generationResult.ProcessingTime)
	}

	// Step 4: Code Completion
	fmt.Println("\n4. Demonstrating Code Completion...")
	partialCode := `
package main

import "fmt"

func main() {
	numbers := []int{1, 2, 3, 4, 5}
	for _, num := range numbers {
		fmt.Print
`

	completionRequest := &coding.CodeCompletionRequest{
		Code:     partialCode,
		Position: &coding.Position{Line: 8, Column: 13},
		Language: "go",
		Context: &coding.CompletionContext{
			TriggerKind:     "invoked",
			IncludeSnippets: true,
			IncludeKeywords: true,
			IncludeSymbols:  true,
		},
		MaxResults: 5,
	}

	completionResult, err := assistant.CompleteCode(ctx, completionRequest)
	if err != nil {
		fmt.Printf("   ⚠️  Code completion failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Code completion completed:\n")
		fmt.Printf("     - Completions found: %d\n", len(completionResult.Completions))
		for i, completion := range completionResult.Completions {
			if i >= 3 { // Show only first 3 completions
				break
			}
			fmt.Printf("     - %s (%s)\n", completion.Label, completion.Kind)
		}
		fmt.Printf("     - Processing time: %v\n", completionResult.ProcessingTime)
	}

	// Step 5: Refactoring Suggestions
	fmt.Println("\n5. Getting Refactoring Suggestions...")
	complexCode := `
func processData(data []string) []string {
	var result []string
	for i := 0; i < len(data); i++ {
		if len(data[i]) > 0 {
			if data[i][0] == 'A' {
				if len(data[i]) > 5 {
					if data[i][len(data[i])-1] == 'Z' {
						result = append(result, data[i])
					}
				}
			}
		}
	}
	return result
}
`

	refactoringSuggestions, err := assistant.SuggestRefactoring(ctx, complexCode, "go")
	if err != nil {
		fmt.Printf("   ⚠️  Refactoring suggestions failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Refactoring suggestions generated:\n")
		fmt.Printf("     - Suggestions found: %d\n", len(refactoringSuggestions))
		for i, suggestion := range refactoringSuggestions {
			if i >= 2 { // Show only first 2 suggestions
				break
			}
			fmt.Printf("     - %s: %s (confidence: %.2f)\n",
				suggestion.Type, suggestion.Description, suggestion.Confidence)
		}
	}

	// Step 6: Apply Refactoring
	fmt.Println("\n6. Applying Refactoring...")
	refactoringRequest := &coding.RefactoringRequest{
		Code:     complexCode,
		Language: "go",
		Type:     coding.RefactoringTypeSimplify,
		Options: &coding.RefactoringOptions{
			PreserveComments: true,
			UpdateReferences: true,
			ValidateChanges:  true,
			DryRun:           false,
		},
	}

	refactoringResult, err := assistant.ApplyRefactoring(ctx, refactoringRequest)
	if err != nil {
		fmt.Printf("   ⚠️  Refactoring failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Refactoring completed:\n")
		fmt.Printf("     - Success: %t\n", refactoringResult.Success)
		fmt.Printf("     - Changes made: %d\n", len(refactoringResult.Changes))
		fmt.Printf("     - Processing time: %v\n", refactoringResult.ProcessingTime)
		if len(refactoringResult.Warnings) > 0 {
			fmt.Printf("     - Warnings: %v\n", refactoringResult.Warnings)
		}
	}

	// Step 7: Code Optimization
	fmt.Println("\n7. Optimizing Code...")
	inefficientCode := `
func findDuplicates(arr []int) []int {
	var duplicates []int
	for i := 0; i < len(arr); i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				duplicates = append(duplicates, arr[i])
			}
		}
	}
	return duplicates
}
`

	optimizationResult, err := assistant.OptimizeCode(ctx, inefficientCode, "go")
	if err != nil {
		fmt.Printf("   ⚠️  Code optimization failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Code optimization completed:\n")
		fmt.Printf("     - Improvements found: %d\n", len(optimizationResult.Improvements))
		fmt.Printf("     - Performance gain: %.2f%%\n", optimizationResult.PerformanceGain*100)
		fmt.Printf("     - Processing time: %v\n", optimizationResult.ProcessingTime)

		for i, improvement := range optimizationResult.Improvements {
			if i >= 2 { // Show only first 2 improvements
				break
			}
			fmt.Printf("     - %s: %s\n", improvement.Type, improvement.Description)
		}
	}

	// Step 8: Generate Documentation
	fmt.Println("\n8. Generating Documentation...")
	docRequest := &coding.DocumentationRequest{
		Code:     sampleCode,
		Language: "go",
		Type:     coding.DocumentationTypeFunction,
		Style: &coding.DocumentationStyle{
			Format:          "markdown",
			IncludeExamples: true,
			IncludeTypes:    true,
			IncludeParams:   true,
			IncludeReturns:  true,
		},
	}

	docResult, err := assistant.GenerateDocumentation(ctx, docRequest)
	if err != nil {
		fmt.Printf("   ⚠️  Documentation generation failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Documentation generation completed:\n")
		fmt.Printf("     - Format: %s\n", docResult.Format)
		fmt.Printf("     - Processing time: %v\n", docResult.ProcessingTime)
		fmt.Printf("     - Documentation preview:\n%s\n",
			truncateString(docResult.Documentation, 200))
	}

	// Step 9: Generate Tests
	fmt.Println("\n9. Generating Tests...")
	testRequest := &coding.TestGenerationRequest{
		Code:      "func add(a, b int) int { return a + b }",
		Language:  "go",
		TestType:  coding.TestTypeUnit,
		Framework: "testing",
		Options: &coding.TestGenerationOptions{
			IncludeEdgeCases:     true,
			IncludeNegativeCases: true,
			GenerateMocks:        false,
			GenerateSetup:        true,
			GenerateTeardown:     false,
			CoverageTarget:       90.0,
		},
	}

	testResult, err := assistant.GenerateTests(ctx, testRequest)
	if err != nil {
		fmt.Printf("   ⚠️  Test generation failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Test generation completed:\n")
		fmt.Printf("     - Tests generated: %d\n", len(testResult.Tests))
		fmt.Printf("     - Framework: %s\n", testResult.Framework)
		fmt.Printf("     - Expected coverage: %.1f%%\n", testResult.Coverage)
		fmt.Printf("     - Processing time: %v\n", testResult.ProcessingTime)

		if len(testResult.Tests) > 0 {
			fmt.Printf("     - Sample test: %s\n", testResult.Tests[0].Name)
		}
	}

	// Step 10: Analyze Project (if we had a real project path)
	fmt.Println("\n10. Project Analysis Demo...")
	projectOptions := &coding.ProjectAnalysisOptions{
		IncludeDependencies: true,
		IncludeMetrics:      true,
		IncludeIssues:       true,
		IncludeStructure:    true,
		MaxDepth:            3,
	}

	projectResult, err := assistant.AnalyzeProject(ctx, ".", projectOptions)
	if err != nil {
		fmt.Printf("   ⚠️  Project analysis failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Project analysis completed:\n")
		if projectResult.ProjectInfo != nil {
			fmt.Printf("     - Project: %s\n", projectResult.ProjectInfo.Name)
			fmt.Printf("     - Language: %s\n", projectResult.ProjectInfo.Language)
		}
		if projectResult.Structure != nil {
			fmt.Printf("     - Files: %d\n", projectResult.Structure.FileCount)
			fmt.Printf("     - Directories: %d\n", projectResult.Structure.DirCount)
		}
		if projectResult.Metrics != nil {
			fmt.Printf("     - Lines of code: %d\n", projectResult.Metrics.LinesOfCode)
			fmt.Printf("     - Functions: %d\n", projectResult.Metrics.FunctionCount)
		}
		fmt.Printf("     - Issues found: %d\n", len(projectResult.Issues))
		fmt.Printf("     - Processing time: %v\n", projectResult.ProcessingTime)
	}

	return nil
}

// Helper function to truncate strings for display
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
