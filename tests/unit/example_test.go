package unit

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example unit tests demonstrating the testing framework

func TestStringOperations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "uppercase conversion",
			input:    "hello",
			expected: "HELLO",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "mixed case",
			input:    "Hello World",
			expected: "HELLO WORLD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toUpperCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMathOperations(t *testing.T) {
	t.Parallel()

	t.Run("addition", func(t *testing.T) {
		t.Parallel()
		result := add(2, 3)
		assert.Equal(t, 5, result)
	})

	t.Run("subtraction", func(t *testing.T) {
		t.Parallel()
		result := subtract(5, 3)
		assert.Equal(t, 2, result)
	})

	t.Run("multiplication", func(t *testing.T) {
		t.Parallel()
		result := multiply(4, 3)
		assert.Equal(t, 12, result)
	})

	t.Run("division", func(t *testing.T) {
		t.Parallel()
		result, err := divide(10, 2)
		require.NoError(t, err)
		assert.Equal(t, 5.0, result)
	})

	t.Run("division by zero", func(t *testing.T) {
		t.Parallel()
		_, err := divide(10, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "division by zero")
	})
}

func TestDataStructures(t *testing.T) {
	t.Run("slice operations", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}

		// Test length
		assert.Len(t, slice, 5)

		// Test contains
		assert.Contains(t, slice, 3)
		assert.NotContains(t, slice, 6)

		// Test first and last elements
		assert.Equal(t, 1, slice[0])
		assert.Equal(t, 5, slice[len(slice)-1])
	})

	t.Run("map operations", func(t *testing.T) {
		m := map[string]int{
			"apple":  5,
			"banana": 3,
			"orange": 8,
		}

		// Test key existence
		_, exists := m["apple"]
		assert.True(t, exists)

		_, exists = m["grape"]
		assert.False(t, exists)

		// Test values
		assert.Equal(t, 5, m["apple"])
		assert.Equal(t, 3, m["banana"])
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		result, err := processData("valid input")
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("invalid input", func(t *testing.T) {
		_, err := processData("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty input")
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := processDataPointer(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil input")
	})
}

func TestConcurrency(t *testing.T) {
	t.Run("goroutine safety", func(t *testing.T) {
		counter := &SafeCounter{}

		// Start multiple goroutines
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					counter.Increment()
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Check final count
		assert.Equal(t, 1000, counter.Value())
	})

	t.Run("channel operations", func(t *testing.T) {
		ch := make(chan int, 5)

		// Send values
		for i := 1; i <= 5; i++ {
			ch <- i
		}
		close(ch)

		// Receive values
		var received []int
		for val := range ch {
			received = append(received, val)
		}

		assert.Equal(t, []int{1, 2, 3, 4, 5}, received)
	})
}

func TestTimeOperations(t *testing.T) {
	t.Run("time formatting", func(t *testing.T) {
		testTime := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
		formatted := formatTime(testTime)
		assert.Equal(t, "2023-12-25 15:30:45", formatted)
	})

	t.Run("time parsing", func(t *testing.T) {
		timeStr := "2023-12-25 15:30:45"
		parsed, err := parseTime(timeStr)
		require.NoError(t, err)

		expected := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
		assert.True(t, expected.Equal(parsed))
	})

	t.Run("duration calculations", func(t *testing.T) {
		start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2023, 1, 1, 1, 30, 0, 0, time.UTC)

		duration := calculateDuration(start, end)
		assert.Equal(t, 90*time.Minute, duration)
	})
}

// Benchmark tests
func BenchmarkStringOperations(b *testing.B) {
	input := "hello world this is a test string"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = toUpperCase(input)
	}
}

func BenchmarkMathOperations(b *testing.B) {
	b.Run("addition", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = add(i, i+1)
		}
	})

	b.Run("multiplication", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = multiply(i, i+1)
		}
	})
}

func BenchmarkDataStructures(b *testing.B) {
	b.Run("slice append", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice := make([]int, 0, 1000)
			for j := 0; j < 1000; j++ {
				slice = append(slice, j)
			}
		}
	})

	b.Run("map operations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := make(map[int]string, 1000)
			for j := 0; j < 1000; j++ {
				m[j] = "value"
			}
		}
	})
}

// Helper functions for testing
func toUpperCase(s string) string {
	// Simple implementation for testing
	return strings.ToUpper(s)
}

func add(a, b int) int {
	return a + b
}

func subtract(a, b int) int {
	return a - b
}

func multiply(a, b int) int {
	return a * b
}

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func processData(input string) (string, error) {
	if input == "" {
		return "", errors.New("empty input not allowed")
	}
	return "processed: " + input, nil
}

func processDataPointer(input *string) (string, error) {
	if input == nil {
		return "", errors.New("nil input not allowed")
	}
	return processData(*input)
}

type SafeCounter struct {
	mu    sync.Mutex
	value int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func parseTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", s)
}

func calculateDuration(start, end time.Time) time.Duration {
	return end.Sub(start)
}
