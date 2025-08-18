package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// FileSystemToolImpl implements FileSystemTool
type FileSystemToolImpl struct {
	name        string
	description string
	basePath    string
	allowedPaths []string
	logger      *logrus.Logger
	tracer      trace.Tracer
}

// NewFileSystemTool creates a new file system tool
func NewFileSystemTool(basePath string, allowedPaths []string, logger *logrus.Logger) FileSystemTool {
	return &FileSystemToolImpl{
		name:        "filesystem",
		description: "Provides file system operations for reading, writing, and managing files and directories",
		basePath:    basePath,
		allowedPaths: allowedPaths,
		logger:      logger,
		tracer:      otel.Tracer("mcp.tools.filesystem"),
	}
}

// GetName returns the tool name
func (t *FileSystemToolImpl) GetName() string {
	return t.name
}

// GetDescription returns the tool description
func (t *FileSystemToolImpl) GetDescription() string {
	return t.description
}

// GetInputSchema returns the input schema
func (t *FileSystemToolImpl) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type": "string",
				"enum": []string{"read_file", "write_file", "list_directory", "create_directory", "delete_file", "move_file", "get_file_info"},
				"description": "The file system operation to perform",
			},
			"path": map[string]interface{}{
				"type": "string",
				"description": "The file or directory path",
			},
			"content": map[string]interface{}{
				"type": "string",
				"description": "Content to write (for write_file operation)",
			},
			"destination": map[string]interface{}{
				"type": "string",
				"description": "Destination path (for move_file operation)",
			},
		},
		"required": []string{"operation", "path"},
	}
}

// GetOutputSchema returns the output schema
func (t *FileSystemToolImpl) GetOutputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type": "boolean",
				"description": "Whether the operation was successful",
			},
			"result": map[string]interface{}{
				"type": "object",
				"description": "The operation result",
			},
			"error": map[string]interface{}{
				"type": "string",
				"description": "Error message if operation failed",
			},
		},
	}
}

// Execute executes the tool with the given arguments
func (t *FileSystemToolImpl) Execute(ctx context.Context, arguments map[string]interface{}) (*protocol.CallToolResult, error) {
	ctx, span := t.tracer.Start(ctx, "filesystem_tool.execute")
	defer span.End()

	operation, ok := arguments["operation"].(string)
	if !ok {
		return t.createErrorResult("operation is required and must be a string"), nil
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return t.createErrorResult("path is required and must be a string"), nil
	}

	// Validate path
	if err := t.validatePath(path); err != nil {
		return t.createErrorResult(fmt.Sprintf("invalid path: %v", err)), nil
	}

	span.SetAttributes(
		attribute.String("filesystem.operation", operation),
		attribute.String("filesystem.path", path),
	)

	t.logger.WithFields(logrus.Fields{
		"operation": operation,
		"path":      path,
	}).Debug("Executing filesystem operation")

	switch operation {
	case "read_file":
		return t.executeReadFile(ctx, path)
	case "write_file":
		content, _ := arguments["content"].(string)
		return t.executeWriteFile(ctx, path, content)
	case "list_directory":
		return t.executeListDirectory(ctx, path)
	case "create_directory":
		return t.executeCreateDirectory(ctx, path)
	case "delete_file":
		return t.executeDeleteFile(ctx, path)
	case "move_file":
		destination, _ := arguments["destination"].(string)
		return t.executeMoveFile(ctx, path, destination)
	case "get_file_info":
		return t.executeGetFileInfo(ctx, path)
	default:
		return t.createErrorResult(fmt.Sprintf("unsupported operation: %s", operation)), nil
	}
}

// Validate validates the tool configuration
func (t *FileSystemToolImpl) Validate() error {
	if t.basePath == "" {
		return fmt.Errorf("base path cannot be empty")
	}

	// Check if base path exists
	if _, err := os.Stat(t.basePath); os.IsNotExist(err) {
		return fmt.Errorf("base path does not exist: %s", t.basePath)
	}

	return nil
}

// GetCategory returns the tool category
func (t *FileSystemToolImpl) GetCategory() string {
	return "filesystem"
}

// GetTags returns the tool tags
func (t *FileSystemToolImpl) GetTags() []string {
	return []string{"file", "directory", "io", "storage"}
}

// IsAsync returns whether the tool supports async execution
func (t *FileSystemToolImpl) IsAsync() bool {
	return false
}

// GetTimeout returns the tool execution timeout
func (t *FileSystemToolImpl) GetTimeout() time.Duration {
	return 30 * time.Second
}

// FileSystemTool interface methods

// ReadFile reads a file
func (t *FileSystemToolImpl) ReadFile(ctx context.Context, path string) (string, error) {
	if err := t.validatePath(path); err != nil {
		return "", err
	}

	fullPath := t.getFullPath(path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// WriteFile writes to a file
func (t *FileSystemToolImpl) WriteFile(ctx context.Context, path string, content string) error {
	if err := t.validatePath(path); err != nil {
		return err
	}

	fullPath := t.getFullPath(path)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ListDirectory lists directory contents
func (t *FileSystemToolImpl) ListDirectory(ctx context.Context, path string) ([]FileInfo, error) {
	if err := t.validatePath(path); err != nil {
		return nil, err
	}

	fullPath := t.getFullPath(path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var fileInfos []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Name:        entry.Name(),
			Path:        filepath.Join(path, entry.Name()),
			Size:        info.Size(),
			Mode:        info.Mode().String(),
			ModTime:     info.ModTime(),
			IsDirectory: entry.IsDir(),
			Permissions: info.Mode().Perm().String(),
		}

		// Check if it's a symlink
		if info.Mode()&fs.ModeSymlink != 0 {
			fileInfo.IsSymlink = true
			if target, err := os.Readlink(filepath.Join(fullPath, entry.Name())); err == nil {
				fileInfo.Target = target
			}
		}

		fileInfos = append(fileInfos, fileInfo)
	}

	return fileInfos, nil
}

// CreateDirectory creates a directory
func (t *FileSystemToolImpl) CreateDirectory(ctx context.Context, path string) error {
	if err := t.validatePath(path); err != nil {
		return err
	}

	fullPath := t.getFullPath(path)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// DeleteFile deletes a file
func (t *FileSystemToolImpl) DeleteFile(ctx context.Context, path string) error {
	if err := t.validatePath(path); err != nil {
		return err
	}

	fullPath := t.getFullPath(path)
	if err := os.RemoveAll(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// MoveFile moves/renames a file
func (t *FileSystemToolImpl) MoveFile(ctx context.Context, src, dst string) error {
	if err := t.validatePath(src); err != nil {
		return err
	}
	if err := t.validatePath(dst); err != nil {
		return err
	}

	srcPath := t.getFullPath(src)
	dstPath := t.getFullPath(dst)

	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// GetFileInfo gets file information
func (t *FileSystemToolImpl) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	if err := t.validatePath(path); err != nil {
		return nil, err
	}

	fullPath := t.getFullPath(path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	fileInfo := &FileInfo{
		Name:        info.Name(),
		Path:        path,
		Size:        info.Size(),
		Mode:        info.Mode().String(),
		ModTime:     info.ModTime(),
		IsDirectory: info.IsDir(),
		Permissions: info.Mode().Perm().String(),
	}

	// Check if it's a symlink
	if info.Mode()&fs.ModeSymlink != 0 {
		fileInfo.IsSymlink = true
		if target, err := os.Readlink(fullPath); err == nil {
			fileInfo.Target = target
		}
	}

	return fileInfo, nil
}

// Helper methods

func (t *FileSystemToolImpl) validatePath(path string) error {
	// Clean the path
	cleanPath := filepath.Clean(path)

	// Check for path traversal
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Check if path is within allowed paths
	if len(t.allowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range t.allowedPaths {
			if strings.HasPrefix(cleanPath, allowedPath) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("path not in allowed paths")
		}
	}

	return nil
}

func (t *FileSystemToolImpl) getFullPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(t.basePath, path)
}

func (t *FileSystemToolImpl) createErrorResult(message string) *protocol.CallToolResult {
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

func (t *FileSystemToolImpl) createSuccessResult(result interface{}) *protocol.CallToolResult {
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

// Execute operation methods

func (t *FileSystemToolImpl) executeReadFile(ctx context.Context, path string) (*protocol.CallToolResult, error) {
	content, err := t.ReadFile(ctx, path)
	if err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation": "read_file",
		"path":      path,
		"content":   content,
		"size":      len(content),
	}

	return t.createSuccessResult(result), nil
}

func (t *FileSystemToolImpl) executeWriteFile(ctx context.Context, path, content string) (*protocol.CallToolResult, error) {
	if err := t.WriteFile(ctx, path, content); err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation": "write_file",
		"path":      path,
		"size":      len(content),
	}

	return t.createSuccessResult(result), nil
}

func (t *FileSystemToolImpl) executeListDirectory(ctx context.Context, path string) (*protocol.CallToolResult, error) {
	files, err := t.ListDirectory(ctx, path)
	if err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation": "list_directory",
		"path":      path,
		"files":     files,
		"count":     len(files),
	}

	return t.createSuccessResult(result), nil
}

func (t *FileSystemToolImpl) executeCreateDirectory(ctx context.Context, path string) (*protocol.CallToolResult, error) {
	if err := t.CreateDirectory(ctx, path); err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation": "create_directory",
		"path":      path,
	}

	return t.createSuccessResult(result), nil
}

func (t *FileSystemToolImpl) executeDeleteFile(ctx context.Context, path string) (*protocol.CallToolResult, error) {
	if err := t.DeleteFile(ctx, path); err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation": "delete_file",
		"path":      path,
	}

	return t.createSuccessResult(result), nil
}

func (t *FileSystemToolImpl) executeMoveFile(ctx context.Context, src, dst string) (*protocol.CallToolResult, error) {
	if dst == "" {
		return t.createErrorResult("destination path is required for move operation"), nil
	}

	if err := t.MoveFile(ctx, src, dst); err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation":   "move_file",
		"source":      src,
		"destination": dst,
	}

	return t.createSuccessResult(result), nil
}

func (t *FileSystemToolImpl) executeGetFileInfo(ctx context.Context, path string) (*protocol.CallToolResult, error) {
	info, err := t.GetFileInfo(ctx, path)
	if err != nil {
		return t.createErrorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"operation": "get_file_info",
		"path":      path,
		"info":      info,
	}

	return t.createSuccessResult(result), nil
}
