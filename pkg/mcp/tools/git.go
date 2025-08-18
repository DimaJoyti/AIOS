package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GitToolImpl implements GitTool
type GitToolImpl struct {
	name        string
	description string
	gitPath     string
	logger      *logrus.Logger
	tracer      trace.Tracer
}

// NewGitTool creates a new Git tool
func NewGitTool(logger *logrus.Logger) GitTool {
	gitPath, _ := exec.LookPath("git")

	return &GitToolImpl{
		name:        "git",
		description: "Provides Git version control operations for managing repositories",
		gitPath:     gitPath,
		logger:      logger,
		tracer:      otel.Tracer("mcp.tools.git"),
	}
}

// GetName returns the tool name
func (t *GitToolImpl) GetName() string {
	return t.name
}

// GetDescription returns the tool description
func (t *GitToolImpl) GetDescription() string {
	return t.description
}

// GetInputSchema returns the input schema
func (t *GitToolImpl) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"status", "log", "diff", "create_branch", "switch_branch", "commit", "push", "pull"},
				"description": "The Git operation to perform",
			},
			"repo_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the Git repository",
			},
			"branch_name": map[string]interface{}{
				"type":        "string",
				"description": "Branch name (for branch operations)",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Commit message (for commit operation)",
			},
			"options": map[string]interface{}{
				"type":        "object",
				"description": "Additional options for the operation",
			},
		},
		"required": []string{"operation", "repo_path"},
	}
}

// GetOutputSchema returns the output schema
func (t *GitToolImpl) GetOutputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the operation was successful",
			},
			"result": map[string]interface{}{
				"type":        "object",
				"description": "The operation result",
			},
			"error": map[string]interface{}{
				"type":        "string",
				"description": "Error message if operation failed",
			},
		},
	}
}

// Execute executes the tool with the given arguments
func (t *GitToolImpl) Execute(ctx context.Context, arguments map[string]interface{}) (*protocol.CallToolResult, error) {
	ctx, span := t.tracer.Start(ctx, "git_tool.execute")
	defer span.End()

	operation, ok := arguments["operation"].(string)
	if !ok {
		return t.createErrorResult("operation is required and must be a string"), nil
	}

	repoPath, ok := arguments["repo_path"].(string)
	if !ok {
		return t.createErrorResult("repo_path is required and must be a string"), nil
	}

	span.SetAttributes(
		attribute.String("git.operation", operation),
		attribute.String("git.repo_path", repoPath),
	)

	t.logger.WithFields(logrus.Fields{
		"operation": operation,
		"repo_path": repoPath,
	}).Debug("Executing Git operation")

	switch operation {
	case "status":
		return t.executeGetStatus(ctx, repoPath)
	case "log":
		options := t.parseLogOptions(arguments["options"])
		return t.executeGetLog(ctx, repoPath, options)
	case "diff":
		options := t.parseDiffOptions(arguments["options"])
		return t.executeGetDiff(ctx, repoPath, options)
	case "create_branch":
		branchName, _ := arguments["branch_name"].(string)
		return t.executeCreateBranch(ctx, repoPath, branchName)
	case "switch_branch":
		branchName, _ := arguments["branch_name"].(string)
		return t.executeSwitchBranch(ctx, repoPath, branchName)
	case "commit":
		message, _ := arguments["message"].(string)
		return t.executeCommit(ctx, repoPath, message)
	case "push":
		options := t.parsePushOptions(arguments["options"])
		return t.executePush(ctx, repoPath, options)
	case "pull":
		options := t.parsePullOptions(arguments["options"])
		return t.executePull(ctx, repoPath, options)
	default:
		return t.createErrorResult(fmt.Sprintf("unsupported operation: %s", operation)), nil
	}
}

// Validate validates the tool configuration
func (t *GitToolImpl) Validate() error {
	if t.gitPath == "" {
		return fmt.Errorf("git command not found in PATH")
	}
	return nil
}

// GetCategory returns the tool category
func (t *GitToolImpl) GetCategory() string {
	return "version_control"
}

// GetTags returns the tool tags
func (t *GitToolImpl) GetTags() []string {
	return []string{"git", "vcs", "version-control", "repository"}
}

// IsAsync returns whether the tool supports async execution
func (t *GitToolImpl) IsAsync() bool {
	return false
}

// GetTimeout returns the tool execution timeout
func (t *GitToolImpl) GetTimeout() time.Duration {
	return 60 * time.Second
}

// GitTool interface methods

// GetStatus gets Git status
func (t *GitToolImpl) GetStatus(ctx context.Context, repoPath string) (*GitStatus, error) {
	// Get current branch
	branch, err := t.runGitCommand(ctx, repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Get status
	statusOutput, err := t.runGitCommand(ctx, repoPath, "status", "--porcelain=v1")
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Get remote URL
	remoteURL, _ := t.runGitCommand(ctx, repoPath, "remote", "get-url", "origin")

	// Parse status
	status := &GitStatus{
		Branch:    strings.TrimSpace(branch),
		RemoteURL: strings.TrimSpace(remoteURL),
		Staged:    make([]string, 0),
		Modified:  make([]string, 0),
		Untracked: make([]string, 0),
		Deleted:   make([]string, 0),
		Renamed:   make([]string, 0),
	}

	lines := strings.Split(strings.TrimSpace(statusOutput), "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		statusCode := line[:2]
		filename := line[3:]

		switch statusCode[0] {
		case 'A', 'M', 'D', 'R', 'C':
			status.Staged = append(status.Staged, filename)
		}

		switch statusCode[1] {
		case 'M':
			status.Modified = append(status.Modified, filename)
		case 'D':
			status.Deleted = append(status.Deleted, filename)
		case '?':
			status.Untracked = append(status.Untracked, filename)
		}

		if statusCode[0] == 'R' {
			status.Renamed = append(status.Renamed, filename)
		}
	}

	status.IsClean = len(status.Staged) == 0 && len(status.Modified) == 0 &&
		len(status.Untracked) == 0 && len(status.Deleted) == 0

	return status, nil
}

// GetLog gets Git log
func (t *GitToolImpl) GetLog(ctx context.Context, repoPath string, options *GitLogOptions) ([]*GitCommit, error) {
	args := []string{"log", "--pretty=format:%H|%h|%an|%ae|%ad|%s", "--date=iso"}

	if options != nil {
		if options.MaxCount > 0 {
			args = append(args, "-n", strconv.Itoa(options.MaxCount))
		}
		if options.Since != "" {
			args = append(args, "--since", options.Since)
		}
		if options.Until != "" {
			args = append(args, "--until", options.Until)
		}
		if options.Author != "" {
			args = append(args, "--author", options.Author)
		}
		if options.Grep != "" {
			args = append(args, "--grep", options.Grep)
		}
	}

	output, err := t.runGitCommand(ctx, repoPath, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	var commits []*GitCommit
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 6 {
			continue
		}

		date, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[4])

		commit := &GitCommit{
			Hash:      parts[0],
			ShortHash: parts[1],
			Author:    parts[2],
			Email:     parts[3],
			Date:      date,
			Message:   parts[5],
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

// GetDiff gets Git diff
func (t *GitToolImpl) GetDiff(ctx context.Context, repoPath string, options *GitDiffOptions) (string, error) {
	args := []string{"diff"}

	if options != nil {
		if options.Cached {
			args = append(args, "--cached")
		}
		if options.Staged {
			args = append(args, "--staged")
		}
		if options.FromCommit != "" && options.ToCommit != "" {
			args = append(args, options.FromCommit+".."+options.ToCommit)
		}
		if options.FilePath != "" {
			args = append(args, "--", options.FilePath)
		}
	}

	output, err := t.runGitCommand(ctx, repoPath, args...)
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}

	return output, nil
}

// CreateBranch creates a new branch
func (t *GitToolImpl) CreateBranch(ctx context.Context, repoPath, branchName string) error {
	_, err := t.runGitCommand(ctx, repoPath, "checkout", "-b", branchName)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	return nil
}

// SwitchBranch switches to a branch
func (t *GitToolImpl) SwitchBranch(ctx context.Context, repoPath, branchName string) error {
	_, err := t.runGitCommand(ctx, repoPath, "checkout", branchName)
	if err != nil {
		return fmt.Errorf("failed to switch branch: %w", err)
	}
	return nil
}

// Commit creates a commit
func (t *GitToolImpl) Commit(ctx context.Context, repoPath, message string) error {
	if message == "" {
		return fmt.Errorf("commit message is required")
	}

	_, err := t.runGitCommand(ctx, repoPath, "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

// Push pushes changes
func (t *GitToolImpl) Push(ctx context.Context, repoPath string, options *GitPushOptions) error {
	args := []string{"push"}

	if options != nil {
		if options.Remote != "" {
			args = append(args, options.Remote)
		}
		if options.Branch != "" {
			args = append(args, options.Branch)
		}
		if options.Force {
			args = append(args, "--force")
		}
		if options.SetUpstream {
			args = append(args, "--set-upstream")
		}
	}

	_, err := t.runGitCommand(ctx, repoPath, args...)
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	return nil
}

// Pull pulls changes
func (t *GitToolImpl) Pull(ctx context.Context, repoPath string, options *GitPullOptions) error {
	args := []string{"pull"}

	if options != nil {
		if options.Remote != "" {
			args = append(args, options.Remote)
		}
		if options.Branch != "" {
			args = append(args, options.Branch)
		}
		if options.Rebase {
			args = append(args, "--rebase")
		}
	}

	_, err := t.runGitCommand(ctx, repoPath, args...)
	if err != nil {
		return fmt.Errorf("failed to pull: %w", err)
	}
	return nil
}

// Helper methods

func (t *GitToolImpl) runGitCommand(ctx context.Context, repoPath string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, t.gitPath, args...)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git command failed: %s", string(exitErr.Stderr))
		}
		return "", err
	}

	return string(output), nil
}

func (t *GitToolImpl) createErrorResult(message string) *protocol.CallToolResult {
	return &protocol.CallToolResult{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: message,
			},
		},
		IsError: true,
	}
}

func (t *GitToolImpl) createSuccessResult(result interface{}) *protocol.CallToolResult {
	data, _ := json.Marshal(result)
	return &protocol.CallToolResult{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Data: result,
				Metadata: map[string]interface{}{
					"json": string(data),
				},
			},
		},
		IsError: false,
	}
}

// Parse options methods

func (t *GitToolImpl) parseLogOptions(options interface{}) *GitLogOptions {
	if options == nil {
		return nil
	}

	optMap, ok := options.(map[string]interface{})
	if !ok {
		return nil
	}

	logOptions := &GitLogOptions{}

	if maxCount, ok := optMap["max_count"].(float64); ok {
		logOptions.MaxCount = int(maxCount)
	}
	if since, ok := optMap["since"].(string); ok {
		logOptions.Since = since
	}
	if until, ok := optMap["until"].(string); ok {
		logOptions.Until = until
	}
	if author, ok := optMap["author"].(string); ok {
		logOptions.Author = author
	}
	if grep, ok := optMap["grep"].(string); ok {
		logOptions.Grep = grep
	}

	return logOptions
}

func (t *GitToolImpl) parseDiffOptions(options interface{}) *GitDiffOptions {
	if options == nil {
		return nil
	}

	optMap, ok := options.(map[string]interface{})
	if !ok {
		return nil
	}

	diffOptions := &GitDiffOptions{}

	if cached, ok := optMap["cached"].(bool); ok {
		diffOptions.Cached = cached
	}
	if staged, ok := optMap["staged"].(bool); ok {
		diffOptions.Staged = staged
	}
	if fromCommit, ok := optMap["from_commit"].(string); ok {
		diffOptions.FromCommit = fromCommit
	}
	if toCommit, ok := optMap["to_commit"].(string); ok {
		diffOptions.ToCommit = toCommit
	}
	if filePath, ok := optMap["file_path"].(string); ok {
		diffOptions.FilePath = filePath
	}

	return diffOptions
}

func (t *GitToolImpl) parsePushOptions(options interface{}) *GitPushOptions {
	if options == nil {
		return nil
	}

	optMap, ok := options.(map[string]interface{})
	if !ok {
		return nil
	}

	pushOptions := &GitPushOptions{}

	if remote, ok := optMap["remote"].(string); ok {
		pushOptions.Remote = remote
	}
	if branch, ok := optMap["branch"].(string); ok {
		pushOptions.Branch = branch
	}
	if force, ok := optMap["force"].(bool); ok {
		pushOptions.Force = force
	}
	if setUpstream, ok := optMap["set_upstream"].(bool); ok {
		pushOptions.SetUpstream = setUpstream
	}

	return pushOptions
}

func (t *GitToolImpl) parsePullOptions(options interface{}) *GitPullOptions {
	if options == nil {
		return nil
	}

	optMap, ok := options.(map[string]interface{})
	if !ok {
		return nil
	}

	pullOptions := &GitPullOptions{}

	if remote, ok := optMap["remote"].(string); ok {
		pullOptions.Remote = remote
	}
	if branch, ok := optMap["branch"].(string); ok {
		pullOptions.Branch = branch
	}
	if rebase, ok := optMap["rebase"].(bool); ok {
		pullOptions.Rebase = rebase
	}

	return pullOptions
}

// Execute operation methods

func (t *GitToolImpl) executeGetStatus(ctx context.Context, repoPath string) (*protocol.CallToolResult, error) {
	status, err := t.GetStatus(ctx, repoPath)
	if err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation": "status",
		"repo_path": repoPath,
		"status":    status,
	}

	return t.createSuccessResult(result), nil
}

func (t *GitToolImpl) executeGetLog(ctx context.Context, repoPath string, options *GitLogOptions) (*protocol.CallToolResult, error) {
	commits, err := t.GetLog(ctx, repoPath, options)
	if err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation": "log",
		"repo_path": repoPath,
		"commits":   commits,
		"count":     len(commits),
	}

	return t.createSuccessResult(result), nil
}

func (t *GitToolImpl) executeGetDiff(ctx context.Context, repoPath string, options *GitDiffOptions) (*protocol.CallToolResult, error) {
	diff, err := t.GetDiff(ctx, repoPath, options)
	if err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation": "diff",
		"repo_path": repoPath,
		"diff":      diff,
	}

	return t.createSuccessResult(result), nil
}

func (t *GitToolImpl) executeCreateBranch(ctx context.Context, repoPath, branchName string) (*protocol.CallToolResult, error) {
	if branchName == "" {
		return t.createErrorResult("branch_name is required for create_branch operation"), nil
	}

	if err := t.CreateBranch(ctx, repoPath, branchName); err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation":   "create_branch",
		"repo_path":   repoPath,
		"branch_name": branchName,
	}

	return t.createSuccessResult(result), nil
}

func (t *GitToolImpl) executeSwitchBranch(ctx context.Context, repoPath, branchName string) (*protocol.CallToolResult, error) {
	if branchName == "" {
		return t.createErrorResult("branch_name is required for switch_branch operation"), nil
	}

	if err := t.SwitchBranch(ctx, repoPath, branchName); err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation":   "switch_branch",
		"repo_path":   repoPath,
		"branch_name": branchName,
	}

	return t.createSuccessResult(result), nil
}

func (t *GitToolImpl) executeCommit(ctx context.Context, repoPath, message string) (*protocol.CallToolResult, error) {
	if message == "" {
		return t.createErrorResult("message is required for commit operation"), nil
	}

	if err := t.Commit(ctx, repoPath, message); err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation": "commit",
		"repo_path": repoPath,
		"message":   message,
	}

	return t.createSuccessResult(result), nil
}

func (t *GitToolImpl) executePush(ctx context.Context, repoPath string, options *GitPushOptions) (*protocol.CallToolResult, error) {
	if err := t.Push(ctx, repoPath, options); err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation": "push",
		"repo_path": repoPath,
		"options":   options,
	}

	return t.createSuccessResult(result), nil
}

func (t *GitToolImpl) executePull(ctx context.Context, repoPath string, options *GitPullOptions) (*protocol.CallToolResult, error) {
	if err := t.Pull(ctx, repoPath, options); err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation": "pull",
		"repo_path": repoPath,
		"options":   options,
	}

	return t.createSuccessResult(result), nil
}
