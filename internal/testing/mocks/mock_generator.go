package mocks

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockGenerator provides utilities for generating and managing mocks
type MockGenerator struct {
	mocks map[string]*mock.Mock
	mu    sync.RWMutex
}

// NewMockGenerator creates a new mock generator
func NewMockGenerator() *MockGenerator {
	return &MockGenerator{
		mocks: make(map[string]*mock.Mock),
	}
}

// RegisterMock registers a mock with a name
func (g *MockGenerator) RegisterMock(name string, mockObj *mock.Mock) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.mocks[name] = mockObj
}

// GetMock retrieves a registered mock
func (g *MockGenerator) GetMock(name string) (*mock.Mock, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	mockObj, exists := g.mocks[name]
	return mockObj, exists
}

// AssertAllExpectations asserts expectations on all registered mocks
func (g *MockGenerator) AssertAllExpectations(t mock.TestingT) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for name, mockObj := range g.mocks {
		if !mockObj.AssertExpectations(t) {
			panic(fmt.Sprintf("Mock %s failed expectations", name))
		}
	}
}

// ClearAllMocks clears all registered mocks
func (g *MockGenerator) ClearAllMocks() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.mocks = make(map[string]*mock.Mock)
}

// MockBuilder provides a fluent interface for building mocks
type MockBuilder struct {
	mockObj *mock.Mock
	calls   []*mock.Call
}

// NewMockBuilder creates a new mock builder
func NewMockBuilder(mockObj *mock.Mock) *MockBuilder {
	return &MockBuilder{
		mockObj: mockObj,
		calls:   make([]*mock.Call, 0),
	}
}

// On adds an expectation
func (b *MockBuilder) On(methodName string, arguments ...interface{}) *MockBuilder {
	call := b.mockObj.On(methodName, arguments...)
	b.calls = append(b.calls, call)
	return b
}

// Return sets return values for the last expectation
func (b *MockBuilder) Return(returnArguments ...interface{}) *MockBuilder {
	if len(b.calls) > 0 {
		b.calls[len(b.calls)-1].Return(returnArguments...)
	}
	return b
}

// ReturnError sets an error return for the last expectation
func (b *MockBuilder) ReturnError(err error) *MockBuilder {
	if len(b.calls) > 0 {
		b.calls[len(b.calls)-1].Return(nil, err)
	}
	return b
}

// Times sets the number of times the method should be called
func (b *MockBuilder) Times(times int) *MockBuilder {
	if len(b.calls) > 0 {
		b.calls[len(b.calls)-1].Times(times)
	}
	return b
}

// Once sets the method to be called exactly once
func (b *MockBuilder) Once() *MockBuilder {
	return b.Times(1)
}

// Maybe sets the method to be called zero or more times
func (b *MockBuilder) Maybe() *MockBuilder {
	if len(b.calls) > 0 {
		b.calls[len(b.calls)-1].Maybe()
	}
	return b
}

// After sets a delay before the method returns
func (b *MockBuilder) After(duration time.Duration) *MockBuilder {
	if len(b.calls) > 0 {
		b.calls[len(b.calls)-1].After(duration)
	}
	return b
}

// Run sets a function to run when the method is called
func (b *MockBuilder) Run(fn func(args mock.Arguments)) *MockBuilder {
	if len(b.calls) > 0 {
		b.calls[len(b.calls)-1].Run(fn)
	}
	return b
}

// Build returns the configured mock
func (b *MockBuilder) Build() *mock.Mock {
	return b.mockObj
}

// ContextMock provides utilities for mocking context operations
type ContextMock struct {
	mock.Mock
	cancelled bool
	deadline  time.Time
	values    map[interface{}]interface{}
	mu        sync.RWMutex
}

// NewContextMock creates a new context mock
func NewContextMock() *ContextMock {
	return &ContextMock{
		values: make(map[interface{}]interface{}),
	}
}

// Deadline implements context.Context
func (c *ContextMock) Deadline() (deadline time.Time, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.deadline, !c.deadline.IsZero()
}

// Done implements context.Context
func (c *ContextMock) Done() <-chan struct{} {
	ch := make(chan struct{})
	if c.cancelled {
		close(ch)
	}
	return ch
}

// Err implements context.Context
func (c *ContextMock) Err() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cancelled {
		return context.Canceled
	}
	if !c.deadline.IsZero() && time.Now().After(c.deadline) {
		return context.DeadlineExceeded
	}
	return nil
}

// Value implements context.Context
func (c *ContextMock) Value(key interface{}) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[key]
}

// SetValue sets a value in the context
func (c *ContextMock) SetValue(key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

// Cancel cancels the context
func (c *ContextMock) Cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancelled = true
}

// SetDeadline sets a deadline for the context
func (c *ContextMock) SetDeadline(deadline time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deadline = deadline
}

// HTTPMock provides utilities for mocking HTTP operations
type HTTPMock struct {
	mock.Mock
	responses map[string]*HTTPResponse
	mu        sync.RWMutex
}

// HTTPResponse represents a mock HTTP response
type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	Delay      time.Duration
}

// NewHTTPMock creates a new HTTP mock
func NewHTTPMock() *HTTPMock {
	return &HTTPMock{
		responses: make(map[string]*HTTPResponse),
	}
}

// SetResponse sets a mock response for a URL pattern
func (h *HTTPMock) SetResponse(pattern string, response *HTTPResponse) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.responses[pattern] = response
}

// GetResponse gets a mock response for a URL
func (h *HTTPMock) GetResponse(url string) (*HTTPResponse, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Simple pattern matching - in real implementation, use regex
	for pattern, response := range h.responses {
		if pattern == url || pattern == "*" {
			return response, true
		}
	}

	return nil, false
}

// ClearResponses clears all mock responses
func (h *HTTPMock) ClearResponses() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.responses = make(map[string]*HTTPResponse)
}

// DatabaseMock provides utilities for mocking database operations
type DatabaseMock struct {
	mock.Mock
	transactions map[string]*TransactionMock
	queries      map[string]*QueryResult
	mu           sync.RWMutex
}

// TransactionMock represents a mock database transaction
type TransactionMock struct {
	ID         string
	Committed  bool
	RolledBack bool
	Operations []string
}

// QueryResult represents a mock query result
type QueryResult struct {
	Rows     []map[string]interface{}
	Error    error
	Affected int64
}

// NewDatabaseMock creates a new database mock
func NewDatabaseMock() *DatabaseMock {
	return &DatabaseMock{
		transactions: make(map[string]*TransactionMock),
		queries:      make(map[string]*QueryResult),
	}
}

// BeginTransaction starts a mock transaction
func (d *DatabaseMock) BeginTransaction(id string) *TransactionMock {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx := &TransactionMock{
		ID:         id,
		Operations: make([]string, 0),
	}
	d.transactions[id] = tx
	return tx
}

// CommitTransaction commits a mock transaction
func (d *DatabaseMock) CommitTransaction(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, exists := d.transactions[id]
	if !exists {
		return fmt.Errorf("transaction %s not found", id)
	}

	tx.Committed = true
	return nil
}

// RollbackTransaction rolls back a mock transaction
func (d *DatabaseMock) RollbackTransaction(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, exists := d.transactions[id]
	if !exists {
		return fmt.Errorf("transaction %s not found", id)
	}

	tx.RolledBack = true
	return nil
}

// SetQueryResult sets a mock result for a query
func (d *DatabaseMock) SetQueryResult(query string, result *QueryResult) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.queries[query] = result
}

// ExecuteQuery executes a mock query
func (d *DatabaseMock) ExecuteQuery(query string) (*QueryResult, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result, exists := d.queries[query]
	if !exists {
		return nil, fmt.Errorf("no mock result for query: %s", query)
	}

	return result, result.Error
}

// GetTransaction gets a transaction by ID
func (d *DatabaseMock) GetTransaction(id string) (*TransactionMock, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	tx, exists := d.transactions[id]
	return tx, exists
}

// ClearAll clears all mock data
func (d *DatabaseMock) ClearAll() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.transactions = make(map[string]*TransactionMock)
	d.queries = make(map[string]*QueryResult)
}

// ServiceMock provides utilities for mocking service dependencies
type ServiceMock struct {
	mock.Mock
	services map[string]interface{}
	mu       sync.RWMutex
}

// NewServiceMock creates a new service mock
func NewServiceMock() *ServiceMock {
	return &ServiceMock{
		services: make(map[string]interface{}),
	}
}

// RegisterService registers a mock service
func (s *ServiceMock) RegisterService(name string, service interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[name] = service
}

// GetService gets a mock service
func (s *ServiceMock) GetService(name string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	service, exists := s.services[name]
	return service, exists
}

// ClearServices clears all mock services
func (s *ServiceMock) ClearServices() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services = make(map[string]interface{})
}

// MockFactory provides a factory for creating common mocks
type MockFactory struct{}

// NewMockFactory creates a new mock factory
func NewMockFactory() *MockFactory {
	return &MockFactory{}
}

// CreateHTTPResponse creates a mock HTTP response
func (f *MockFactory) CreateHTTPResponse(statusCode int, body string) *HTTPResponse {
	return &HTTPResponse{
		StatusCode: statusCode,
		Headers:    make(map[string]string),
		Body:       []byte(body),
	}
}

// CreateQueryResult creates a mock query result
func (f *MockFactory) CreateQueryResult(rows []map[string]interface{}) *QueryResult {
	return &QueryResult{
		Rows:     rows,
		Affected: int64(len(rows)),
	}
}

// CreateErrorResult creates a mock error result
func (f *MockFactory) CreateErrorResult(err error) *QueryResult {
	return &QueryResult{
		Error: err,
	}
}

// ReflectionMock provides utilities for reflection-based mocking
type ReflectionMock struct {
	mock.Mock
	target interface{}
}

// NewReflectionMock creates a new reflection mock
func NewReflectionMock(target interface{}) *ReflectionMock {
	return &ReflectionMock{
		target: target,
	}
}

// MockMethod mocks a method using reflection
func (r *ReflectionMock) MockMethod(methodName string, args []interface{}, returns []interface{}) {
	r.On(methodName, args...).Return(returns...)
}

// GetMethodSignature gets the signature of a method
func (r *ReflectionMock) GetMethodSignature(methodName string) ([]reflect.Type, []reflect.Type, error) {
	targetType := reflect.TypeOf(r.target)
	method, exists := targetType.MethodByName(methodName)
	if !exists {
		return nil, nil, fmt.Errorf("method %s not found", methodName)
	}

	methodType := method.Type

	// Input types (excluding receiver)
	inputs := make([]reflect.Type, methodType.NumIn()-1)
	for i := 1; i < methodType.NumIn(); i++ {
		inputs[i-1] = methodType.In(i)
	}

	// Output types
	outputs := make([]reflect.Type, methodType.NumOut())
	for i := 0; i < methodType.NumOut(); i++ {
		outputs[i] = methodType.Out(i)
	}

	return inputs, outputs, nil
}
