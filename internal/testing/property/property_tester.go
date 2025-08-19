package property

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

// PropertyTester provides property-based testing capabilities
type PropertyTester struct {
	config PropertyConfig
	rand   *rand.Rand
}

// PropertyConfig defines configuration for property-based testing
type PropertyConfig struct {
	MaxTests     int           `json:"max_tests"`
	MaxShrinks   int           `json:"max_shrinks"`
	Timeout      time.Duration `json:"timeout"`
	Seed         int64         `json:"seed"`
	MinSuccess   int           `json:"min_success"`
	MaxDiscarded int           `json:"max_discarded"`
}

// Property represents a property to be tested
type Property struct {
	Name        string
	Description string
	Test        func(args ...interface{}) bool
	Generators  []Generator
}

// Generator generates test data
type Generator interface {
	Generate() interface{}
	Shrink(value interface{}) []interface{}
	String() string
}

// PropertyResult represents the result of property testing
type PropertyResult struct {
	Property       Property      `json:"property"`
	Passed         bool          `json:"passed"`
	TestsRun       int           `json:"tests_run"`
	Discarded      int           `json:"discarded"`
	Shrinks        int           `json:"shrinks"`
	CounterExample []interface{} `json:"counter_example,omitempty"`
	Error          string        `json:"error,omitempty"`
	Duration       time.Duration `json:"duration"`
}

// NewPropertyTester creates a new property tester
func NewPropertyTester(config PropertyConfig) *PropertyTester {
	if config.MaxTests == 0 {
		config.MaxTests = 100
	}
	if config.MaxShrinks == 0 {
		config.MaxShrinks = 100
	}
	if config.Timeout == 0 {
		config.Timeout = time.Minute
	}
	if config.MinSuccess == 0 {
		config.MinSuccess = config.MaxTests
	}
	if config.MaxDiscarded == 0 {
		config.MaxDiscarded = config.MaxTests * 5
	}

	seed := config.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	return &PropertyTester{
		config: config,
		rand:   rand.New(rand.NewSource(seed)),
	}
}

// TestProperty tests a single property
func (pt *PropertyTester) TestProperty(t *testing.T, property Property) *PropertyResult {
	start := time.Now()

	result := &PropertyResult{
		Property: property,
		Passed:   true,
		Duration: 0,
	}

	timeout := time.After(pt.config.Timeout)

	for result.TestsRun < pt.config.MaxTests && result.TestsRun-result.Discarded < pt.config.MinSuccess {
		select {
		case <-timeout:
			result.Passed = false
			result.Error = "Test timed out"
			result.Duration = time.Since(start)
			return result
		default:
		}

		// Generate test arguments
		args := make([]interface{}, len(property.Generators))
		for i, gen := range property.Generators {
			args[i] = gen.Generate()
		}

		// Test the property
		passed := pt.safeTestProperty(property.Test, args)
		result.TestsRun++

		if !passed {
			// Property failed, try to shrink
			counterExample := pt.shrink(property, args)
			result.Passed = false
			result.CounterExample = counterExample
			break
		}

		if result.Discarded > pt.config.MaxDiscarded {
			result.Passed = false
			result.Error = "Too many discarded tests"
			break
		}
	}

	result.Duration = time.Since(start)

	if t != nil {
		if !result.Passed {
			if result.CounterExample != nil {
				t.Errorf("Property %s failed with counter-example: %v", property.Name, result.CounterExample)
			} else {
				t.Errorf("Property %s failed: %s", property.Name, result.Error)
			}
		} else {
			t.Logf("Property %s passed (%d tests)", property.Name, result.TestsRun)
		}
	}

	return result
}

// safeTestProperty safely executes a property test, catching panics
func (pt *PropertyTester) safeTestProperty(test func(args ...interface{}) bool, args []interface{}) (passed bool) {
	defer func() {
		if r := recover(); r != nil {
			passed = false
		}
	}()

	return test(args...)
}

// shrink attempts to find a minimal counter-example
func (pt *PropertyTester) shrink(property Property, failingArgs []interface{}) []interface{} {
	current := make([]interface{}, len(failingArgs))
	copy(current, failingArgs)

	for shrinkAttempts := 0; shrinkAttempts < pt.config.MaxShrinks; shrinkAttempts++ {
		shrunk := false

		// Try to shrink each argument
		for i, gen := range property.Generators {
			candidates := gen.Shrink(current[i])

			for _, candidate := range candidates {
				testArgs := make([]interface{}, len(current))
				copy(testArgs, current)
				testArgs[i] = candidate

				// Test if this shrunk version still fails
				if !pt.safeTestProperty(property.Test, testArgs) {
					current = testArgs
					shrunk = true
					break
				}
			}

			if shrunk {
				break
			}
		}

		if !shrunk {
			break
		}
	}

	return current
}

// Built-in generators

// IntGenerator generates integers
type IntGenerator struct {
	Min int
	Max int
}

// NewIntGenerator creates a new integer generator
func NewIntGenerator(min, max int) *IntGenerator {
	return &IntGenerator{Min: min, Max: max}
}

// Generate generates a random integer
func (g *IntGenerator) Generate() interface{} {
	if g.Max <= g.Min {
		return g.Min
	}
	return g.Min + rand.Intn(g.Max-g.Min+1)
}

// Shrink shrinks an integer towards zero
func (g *IntGenerator) Shrink(value interface{}) []interface{} {
	v, ok := value.(int)
	if !ok {
		return nil
	}

	var candidates []interface{}

	// Shrink towards zero
	if v > 0 {
		candidates = append(candidates, 0)
		if v > 1 {
			candidates = append(candidates, v/2)
			candidates = append(candidates, v-1)
		}
	} else if v < 0 {
		candidates = append(candidates, 0)
		if v < -1 {
			candidates = append(candidates, v/2)
			candidates = append(candidates, v+1)
		}
	}

	return candidates
}

// String returns a string representation
func (g *IntGenerator) String() string {
	return fmt.Sprintf("Int(%d, %d)", g.Min, g.Max)
}

// StringGenerator generates strings
type StringGenerator struct {
	MinLength int
	MaxLength int
	Charset   string
}

// NewStringGenerator creates a new string generator
func NewStringGenerator(minLen, maxLen int) *StringGenerator {
	return &StringGenerator{
		MinLength: minLen,
		MaxLength: maxLen,
		Charset:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	}
}

// Generate generates a random string
func (g *StringGenerator) Generate() interface{} {
	length := g.MinLength
	if g.MaxLength > g.MinLength {
		length += rand.Intn(g.MaxLength - g.MinLength + 1)
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = g.Charset[rand.Intn(len(g.Charset))]
	}

	return string(result)
}

// Shrink shrinks a string by reducing its length
func (g *StringGenerator) Shrink(value interface{}) []interface{} {
	s, ok := value.(string)
	if !ok {
		return nil
	}

	var candidates []interface{}

	// Empty string
	if len(s) > 0 {
		candidates = append(candidates, "")
	}

	// Shorter strings
	if len(s) > 1 {
		candidates = append(candidates, s[:len(s)/2])
		candidates = append(candidates, s[:len(s)-1])
	}

	return candidates
}

// String returns a string representation
func (g *StringGenerator) String() string {
	return fmt.Sprintf("String(%d, %d)", g.MinLength, g.MaxLength)
}

// SliceGenerator generates slices
type SliceGenerator struct {
	ElementGenerator Generator
	MinLength        int
	MaxLength        int
}

// NewSliceGenerator creates a new slice generator
func NewSliceGenerator(elementGen Generator, minLen, maxLen int) *SliceGenerator {
	return &SliceGenerator{
		ElementGenerator: elementGen,
		MinLength:        minLen,
		MaxLength:        maxLen,
	}
}

// Generate generates a random slice
func (g *SliceGenerator) Generate() interface{} {
	length := g.MinLength
	if g.MaxLength > g.MinLength {
		length += rand.Intn(g.MaxLength - g.MinLength + 1)
	}

	result := make([]interface{}, length)
	for i := range result {
		result[i] = g.ElementGenerator.Generate()
	}

	return result
}

// Shrink shrinks a slice by reducing its length or shrinking elements
func (g *SliceGenerator) Shrink(value interface{}) []interface{} {
	slice, ok := value.([]interface{})
	if !ok {
		return nil
	}

	var candidates []interface{}

	// Empty slice
	if len(slice) > 0 {
		candidates = append(candidates, []interface{}{})
	}

	// Shorter slices
	if len(slice) > 1 {
		candidates = append(candidates, slice[:len(slice)/2])
		candidates = append(candidates, slice[:len(slice)-1])
	}

	// Shrink individual elements
	for i, elem := range slice {
		shrunkElements := g.ElementGenerator.Shrink(elem)
		for _, shrunkElem := range shrunkElements {
			newSlice := make([]interface{}, len(slice))
			copy(newSlice, slice)
			newSlice[i] = shrunkElem
			candidates = append(candidates, newSlice)
		}
	}

	return candidates
}

// String returns a string representation
func (g *SliceGenerator) String() string {
	return fmt.Sprintf("Slice[%s](%d, %d)", g.ElementGenerator.String(), g.MinLength, g.MaxLength)
}

// BoolGenerator generates booleans
type BoolGenerator struct{}

// NewBoolGenerator creates a new boolean generator
func NewBoolGenerator() *BoolGenerator {
	return &BoolGenerator{}
}

// Generate generates a random boolean
func (g *BoolGenerator) Generate() interface{} {
	return rand.Intn(2) == 1
}

// Shrink shrinks a boolean (false is smaller than true)
func (g *BoolGenerator) Shrink(value interface{}) []interface{} {
	v, ok := value.(bool)
	if !ok {
		return nil
	}

	if v {
		return []interface{}{false}
	}

	return nil
}

// String returns a string representation
func (g *BoolGenerator) String() string {
	return "Bool"
}

// Helper functions for common property patterns

// ForAll creates a property that should hold for all generated inputs
func ForAll(generators []Generator, test func(args ...interface{}) bool) Property {
	return Property{
		Name:       "ForAll",
		Test:       test,
		Generators: generators,
	}
}

// Exists creates a property that should hold for at least one generated input
func Exists(generators []Generator, test func(args ...interface{}) bool) Property {
	existsTest := func(args ...interface{}) bool {
		// This would require a different testing strategy
		// For now, just use the regular test
		return test(args...)
	}

	return Property{
		Name:       "Exists",
		Test:       existsTest,
		Generators: generators,
	}
}

// Implies creates an implication property (if condition then property)
func Implies(condition func(args ...interface{}) bool, property func(args ...interface{}) bool) func(args ...interface{}) bool {
	return func(args ...interface{}) bool {
		if !condition(args...) {
			return true // Vacuously true
		}
		return property(args...)
	}
}

// PropertyTestSuite manages multiple property tests
type PropertyTestSuite struct {
	properties []Property
	tester     *PropertyTester
}

// NewPropertyTestSuite creates a new property test suite
func NewPropertyTestSuite(config PropertyConfig) *PropertyTestSuite {
	return &PropertyTestSuite{
		properties: make([]Property, 0),
		tester:     NewPropertyTester(config),
	}
}

// AddProperty adds a property to the suite
func (pts *PropertyTestSuite) AddProperty(property Property) {
	pts.properties = append(pts.properties, property)
}

// RunAll runs all properties in the suite
func (pts *PropertyTestSuite) RunAll(t *testing.T) []*PropertyResult {
	results := make([]*PropertyResult, 0, len(pts.properties))

	for _, property := range pts.properties {
		result := pts.tester.TestProperty(t, property)
		results = append(results, result)
	}

	return results
}

// GenerateReport generates a report for property test results
func (pts *PropertyTestSuite) GenerateReport(results []*PropertyResult) string {
	var report strings.Builder

	report.WriteString("Property-Based Test Report\n")
	report.WriteString("=========================\n\n")

	passed := 0
	total := len(results)

	for _, result := range results {
		if result.Passed {
			passed++
			report.WriteString(fmt.Sprintf("✓ %s - PASSED (%d tests, %v)\n",
				result.Property.Name, result.TestsRun, result.Duration))
		} else {
			report.WriteString(fmt.Sprintf("✗ %s - FAILED (%d tests, %v)\n",
				result.Property.Name, result.TestsRun, result.Duration))
			if result.CounterExample != nil {
				report.WriteString(fmt.Sprintf("  Counter-example: %v\n", result.CounterExample))
			}
			if result.Error != "" {
				report.WriteString(fmt.Sprintf("  Error: %s\n", result.Error))
			}
		}
	}

	report.WriteString(fmt.Sprintf("\nSummary: %d/%d properties passed (%.1f%%)\n",
		passed, total, float64(passed)/float64(total)*100))

	return report.String()
}
