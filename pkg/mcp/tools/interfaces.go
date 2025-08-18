package tools

import (
	"context"
	"time"

	"github.com/aios/aios/pkg/mcp/protocol"
)

// ToolManager manages MCP tools
type ToolManager interface {
	// RegisterTool registers a tool
	RegisterTool(tool MCPTool) error

	// UnregisterTool unregisters a tool
	UnregisterTool(name string) error

	// GetTool retrieves a tool by name
	GetTool(name string) (MCPTool, error)

	// ListTools returns all registered tools
	ListTools() []protocol.Tool

	// CallTool calls a tool with the given arguments
	CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.CallToolResult, error)

	// ValidateTool validates a tool configuration
	ValidateTool(tool MCPTool) error

	// RegisterToolProvider registers all tools from a tool provider
	RegisterToolProvider(provider ToolProvider) error
}

// MCPTool represents an MCP tool implementation
type MCPTool interface {
	// GetName returns the tool name
	GetName() string

	// GetDescription returns the tool description
	GetDescription() string

	// GetInputSchema returns the input schema
	GetInputSchema() map[string]interface{}

	// GetOutputSchema returns the output schema
	GetOutputSchema() map[string]interface{}

	// Execute executes the tool with the given arguments
	Execute(ctx context.Context, arguments map[string]interface{}) (*protocol.CallToolResult, error)

	// Validate validates the tool configuration
	Validate() error

	// GetCategory returns the tool category
	GetCategory() string

	// GetTags returns the tool tags
	GetTags() []string

	// IsAsync returns whether the tool supports async execution
	IsAsync() bool

	// GetTimeout returns the tool execution timeout
	GetTimeout() time.Duration
}

// ToolProvider interface for providing multiple tools
type ToolProvider interface {
	GetTools() []MCPTool
}

// FileSystemTool provides file system operations
type FileSystemTool interface {
	MCPTool

	// ReadFile reads a file
	ReadFile(ctx context.Context, path string) (string, error)

	// WriteFile writes to a file
	WriteFile(ctx context.Context, path string, content string) error

	// ListDirectory lists directory contents
	ListDirectory(ctx context.Context, path string) ([]FileInfo, error)

	// CreateDirectory creates a directory
	CreateDirectory(ctx context.Context, path string) error

	// DeleteFile deletes a file
	DeleteFile(ctx context.Context, path string) error

	// MoveFile moves/renames a file
	MoveFile(ctx context.Context, src, dst string) error

	// GetFileInfo gets file information
	GetFileInfo(ctx context.Context, path string) (*FileInfo, error)
}

// GitTool provides Git operations
type GitTool interface {
	MCPTool

	// GetStatus gets Git status
	GetStatus(ctx context.Context, repoPath string) (*GitStatus, error)

	// GetLog gets Git log
	GetLog(ctx context.Context, repoPath string, options *GitLogOptions) ([]*GitCommit, error)

	// GetDiff gets Git diff
	GetDiff(ctx context.Context, repoPath string, options *GitDiffOptions) (string, error)

	// CreateBranch creates a new branch
	CreateBranch(ctx context.Context, repoPath, branchName string) error

	// SwitchBranch switches to a branch
	SwitchBranch(ctx context.Context, repoPath, branchName string) error

	// Commit creates a commit
	Commit(ctx context.Context, repoPath, message string) error

	// Push pushes changes
	Push(ctx context.Context, repoPath string, options *GitPushOptions) error

	// Pull pulls changes
	Pull(ctx context.Context, repoPath string, options *GitPullOptions) error
}

// BuildTool provides build system operations
type BuildTool interface {
	MCPTool

	// Build builds the project
	Build(ctx context.Context, projectPath string, options *BuildOptions) (*BuildResult, error)

	// Test runs tests
	Test(ctx context.Context, projectPath string, options *TestOptions) (*TestResult, error)

	// Clean cleans build artifacts
	Clean(ctx context.Context, projectPath string) error

	// Install installs dependencies
	Install(ctx context.Context, projectPath string, options *InstallOptions) error

	// GetBuildInfo gets build information
	GetBuildInfo(ctx context.Context, projectPath string) (*BuildInfo, error)
}

// ProcessTool provides process management operations
type ProcessTool interface {
	MCPTool

	// StartProcess starts a process
	StartProcess(ctx context.Context, command string, args []string, options *ProcessOptions) (*ProcessInfo, error)

	// StopProcess stops a process
	StopProcess(ctx context.Context, pid int) error

	// GetProcessInfo gets process information
	GetProcessInfo(ctx context.Context, pid int) (*ProcessInfo, error)

	// ListProcesses lists running processes
	ListProcesses(ctx context.Context) ([]*ProcessInfo, error)

	// ExecuteCommand executes a command and returns output
	ExecuteCommand(ctx context.Context, command string, args []string, options *ExecuteOptions) (*ExecuteResult, error)
}

// NetworkTool provides network operations
type NetworkTool interface {
	MCPTool

	// HTTPRequest makes an HTTP request
	HTTPRequest(ctx context.Context, options *HTTPRequestOptions) (*HTTPResponse, error)

	// Ping pings a host
	Ping(ctx context.Context, host string, options *PingOptions) (*PingResult, error)

	// PortScan scans ports on a host
	PortScan(ctx context.Context, host string, ports []int) (*PortScanResult, error)

	// DNSLookup performs DNS lookup
	DNSLookup(ctx context.Context, hostname string) (*DNSResult, error)
}

// Data structures

// FileInfo represents file information
type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	Mode        string    `json:"mode"`
	ModTime     time.Time `json:"mod_time"`
	IsDirectory bool      `json:"is_directory"`
	IsSymlink   bool      `json:"is_symlink"`
	Target      string    `json:"target,omitempty"`
	Permissions string    `json:"permissions"`
}

// GitStatus represents Git repository status
type GitStatus struct {
	Branch    string   `json:"branch"`
	Ahead     int      `json:"ahead"`
	Behind    int      `json:"behind"`
	Staged    []string `json:"staged"`
	Modified  []string `json:"modified"`
	Untracked []string `json:"untracked"`
	Deleted   []string `json:"deleted"`
	Renamed   []string `json:"renamed"`
	IsClean   bool     `json:"is_clean"`
	RemoteURL string   `json:"remote_url"`
}

// GitCommit represents a Git commit
type GitCommit struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"short_hash"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Date      time.Time `json:"date"`
	Message   string    `json:"message"`
	Files     []string  `json:"files"`
}

// GitLogOptions represents options for Git log
type GitLogOptions struct {
	MaxCount int    `json:"max_count,omitempty"`
	Since    string `json:"since,omitempty"`
	Until    string `json:"until,omitempty"`
	Author   string `json:"author,omitempty"`
	Grep     string `json:"grep,omitempty"`
}

// GitDiffOptions represents options for Git diff
type GitDiffOptions struct {
	Cached     bool   `json:"cached,omitempty"`
	Staged     bool   `json:"staged,omitempty"`
	FromCommit string `json:"from_commit,omitempty"`
	ToCommit   string `json:"to_commit,omitempty"`
	FilePath   string `json:"file_path,omitempty"`
}

// GitPushOptions represents options for Git push
type GitPushOptions struct {
	Remote      string `json:"remote,omitempty"`
	Branch      string `json:"branch,omitempty"`
	Force       bool   `json:"force,omitempty"`
	SetUpstream bool   `json:"set_upstream,omitempty"`
}

// GitPullOptions represents options for Git pull
type GitPullOptions struct {
	Remote string `json:"remote,omitempty"`
	Branch string `json:"branch,omitempty"`
	Rebase bool   `json:"rebase,omitempty"`
}

// BuildOptions represents build options
type BuildOptions struct {
	Target      string            `json:"target,omitempty"`
	Config      string            `json:"config,omitempty"`
	Parallel    bool              `json:"parallel,omitempty"`
	Verbose     bool              `json:"verbose,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Args        []string          `json:"args,omitempty"`
}

// BuildResult represents build result
type BuildResult struct {
	Success   bool          `json:"success"`
	ExitCode  int           `json:"exit_code"`
	Output    string        `json:"output"`
	Error     string        `json:"error"`
	Duration  time.Duration `json:"duration"`
	Artifacts []string      `json:"artifacts"`
}

// TestOptions represents test options
type TestOptions struct {
	Pattern     string            `json:"pattern,omitempty"`
	Verbose     bool              `json:"verbose,omitempty"`
	Coverage    bool              `json:"coverage,omitempty"`
	Parallel    bool              `json:"parallel,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Args        []string          `json:"args,omitempty"`
}

// TestResult represents test result
type TestResult struct {
	Success      bool          `json:"success"`
	ExitCode     int           `json:"exit_code"`
	Output       string        `json:"output"`
	Error        string        `json:"error"`
	Duration     time.Duration `json:"duration"`
	TestsPassed  int           `json:"tests_passed"`
	TestsFailed  int           `json:"tests_failed"`
	TestsSkipped int           `json:"tests_skipped"`
	Coverage     float64       `json:"coverage,omitempty"`
}

// InstallOptions represents install options
type InstallOptions struct {
	Production  bool              `json:"production,omitempty"`
	Development bool              `json:"development,omitempty"`
	Update      bool              `json:"update,omitempty"`
	Clean       bool              `json:"clean,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Args        []string          `json:"args,omitempty"`
}

// BuildInfo represents build information
type BuildInfo struct {
	BuildSystem  string            `json:"build_system"`
	Version      string            `json:"version"`
	Target       string            `json:"target"`
	Config       string            `json:"config"`
	Dependencies []string          `json:"dependencies"`
	Metadata     map[string]string `json:"metadata"`
}

// ProcessOptions represents process options
type ProcessOptions struct {
	WorkingDir  string            `json:"working_dir,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Stdin       string            `json:"stdin,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
}

// ProcessInfo represents process information
type ProcessInfo struct {
	PID         int               `json:"pid"`
	PPID        int               `json:"ppid"`
	Name        string            `json:"name"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Status      string            `json:"status"`
	CPU         float64           `json:"cpu"`
	Memory      int64             `json:"memory"`
	StartTime   time.Time         `json:"start_time"`
	Environment map[string]string `json:"environment,omitempty"`
}

// ExecuteOptions represents execute options
type ExecuteOptions struct {
	WorkingDir  string            `json:"working_dir,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Stdin       string            `json:"stdin,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty"`
	Shell       bool              `json:"shell,omitempty"`
}

// ExecuteResult represents execute result
type ExecuteResult struct {
	ExitCode int           `json:"exit_code"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	Duration time.Duration `json:"duration"`
	Success  bool          `json:"success"`
}

// HTTPRequestOptions represents HTTP request options
type HTTPRequestOptions struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Timeout time.Duration     `json:"timeout,omitempty"`
}

// HTTPResponse represents HTTP response
type HTTPResponse struct {
	StatusCode int               `json:"status_code"`
	Status     string            `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   time.Duration     `json:"duration"`
}

// PingOptions represents ping options
type PingOptions struct {
	Count   int           `json:"count,omitempty"`
	Timeout time.Duration `json:"timeout,omitempty"`
	Size    int           `json:"size,omitempty"`
}

// PingResult represents ping result
type PingResult struct {
	Host        string        `json:"host"`
	PacketsSent int           `json:"packets_sent"`
	PacketsRecv int           `json:"packets_recv"`
	PacketLoss  float64       `json:"packet_loss"`
	MinRTT      time.Duration `json:"min_rtt"`
	MaxRTT      time.Duration `json:"max_rtt"`
	AvgRTT      time.Duration `json:"avg_rtt"`
	Success     bool          `json:"success"`
}

// PortScanResult represents port scan result
type PortScanResult struct {
	Host      string        `json:"host"`
	OpenPorts []PortInfo    `json:"open_ports"`
	Duration  time.Duration `json:"duration"`
}

// PortInfo represents port information
type PortInfo struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Service  string `json:"service,omitempty"`
	State    string `json:"state"`
}

// DNSResult represents DNS lookup result
type DNSResult struct {
	Hostname string        `json:"hostname"`
	IPs      []string      `json:"ips"`
	CNAME    string        `json:"cname,omitempty"`
	MX       []string      `json:"mx,omitempty"`
	TXT      []string      `json:"txt,omitempty"`
	Duration time.Duration `json:"duration"`
}
