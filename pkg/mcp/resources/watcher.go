package resources

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// WatchCallback defines the callback interface for resource changes
type WatchCallback interface {
	OnResourceChanged(uri string, event WatchEvent) error
}

// WatchEvent represents a resource change event
type WatchEvent struct {
	Type      WatchEventType         `json:"type"`
	URI       string                 `json:"uri"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// WatchEventType represents the type of watch event
type WatchEventType string

const (
	WatchEventCreated  WatchEventType = "created"
	WatchEventModified WatchEventType = "modified"
	WatchEventDeleted  WatchEventType = "deleted"
	WatchEventMoved    WatchEventType = "moved"
)

// FileSystemResourceWatcher implements ResourceWatcher for file system resources
type FileSystemResourceWatcher struct {
	watcher     *fsnotify.Watcher
	watchedURIs map[string]string // URI -> file path
	callbacks   map[string]WatchCallback
	mu          sync.RWMutex
	logger      *logrus.Logger
	running     bool
	stopChan    chan struct{}
}

// NewResourceWatcher creates a new file system resource watcher
func NewResourceWatcher(logger *logrus.Logger) (ResourceWatcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileSystemResourceWatcher{
		watcher:     fsWatcher,
		watchedURIs: make(map[string]string),
		callbacks:   make(map[string]WatchCallback),
		logger:      logger,
		stopChan:    make(chan struct{}),
	}, nil
}

// Watch starts watching a resource URI
func (w *FileSystemResourceWatcher) Watch(uri string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Extract file path from URI
	filePath := w.uriToFilePath(uri)
	if filePath == "" {
		return nil // Not a file URI, skip
	}

	// Check if already watching
	if _, exists := w.watchedURIs[uri]; exists {
		return nil
	}

	// Add to file system watcher
	err := w.watcher.Add(filePath)
	if err != nil {
		return err
	}

	w.watchedURIs[uri] = filePath
	w.logger.WithFields(logrus.Fields{
		"uri":  uri,
		"path": filePath,
	}).Debug("Started watching resource")

	return nil
}

// Unwatch stops watching a resource URI
func (w *FileSystemResourceWatcher) Unwatch(uri string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	filePath, exists := w.watchedURIs[uri]
	if !exists {
		return nil
	}

	// Remove from file system watcher
	err := w.watcher.Remove(filePath)
	if err != nil {
		w.logger.WithError(err).WithField("path", filePath).Warn("Failed to remove file watcher")
	}

	delete(w.watchedURIs, uri)
	w.logger.WithField("uri", uri).Debug("Stopped watching resource")

	return nil
}

// AddCallback adds a watch callback
func (w *FileSystemResourceWatcher) AddCallback(callback WatchCallback) string {
	w.mu.Lock()
	defer w.mu.Unlock()

	callbackID := generateCallbackID()
	w.callbacks[callbackID] = callback
	return callbackID
}

// RemoveCallback removes a watch callback
func (w *FileSystemResourceWatcher) RemoveCallback(callbackID string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.callbacks, callbackID)
}

// Start starts the watcher
func (w *FileSystemResourceWatcher) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = true
	w.mu.Unlock()

	go w.watchLoop(ctx)
	return nil
}

// Stop stops the watcher
func (w *FileSystemResourceWatcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return nil
	}

	w.running = false
	close(w.stopChan)
	return w.watcher.Close()
}

// IsWatching checks if a URI is being watched
func (w *FileSystemResourceWatcher) IsWatching(uri string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	_, exists := w.watchedURIs[uri]
	return exists
}

// GetWatchedResources returns all watched resource URIs
func (w *FileSystemResourceWatcher) GetWatchedResources() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	uris := make([]string, 0, len(w.watchedURIs))
	for uri := range w.watchedURIs {
		uris = append(uris, uri)
	}
	return uris
}

// watchLoop processes file system events
func (w *FileSystemResourceWatcher) watchLoop(ctx context.Context) {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleFileEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.logger.WithError(err).Error("File watcher error")

		case <-w.stopChan:
			return

		case <-ctx.Done():
			return
		}
	}
}

// handleFileEvent processes a file system event
func (w *FileSystemResourceWatcher) handleFileEvent(event fsnotify.Event) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Find URI for this file path
	var uri string
	for u, path := range w.watchedURIs {
		if path == event.Name {
			uri = u
			break
		}
	}

	if uri == "" {
		return // Not watching this file
	}

	// Convert fsnotify event to watch event
	watchEvent := WatchEvent{
		URI:       uri,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"file_path": event.Name,
			"fs_op":     event.Op.String(),
		},
	}

	// Determine event type
	if event.Op&fsnotify.Create == fsnotify.Create {
		watchEvent.Type = WatchEventCreated
	} else if event.Op&fsnotify.Write == fsnotify.Write {
		watchEvent.Type = WatchEventModified
	} else if event.Op&fsnotify.Remove == fsnotify.Remove {
		watchEvent.Type = WatchEventDeleted
	} else if event.Op&fsnotify.Rename == fsnotify.Rename {
		watchEvent.Type = WatchEventMoved
	} else {
		watchEvent.Type = WatchEventModified // Default
	}

	// Notify callbacks
	for _, callback := range w.callbacks {
		go func(cb WatchCallback) {
			if err := cb.OnResourceChanged(uri, watchEvent); err != nil {
				w.logger.WithError(err).WithField("uri", uri).Error("Watch callback failed")
			}
		}(callback)
	}

	w.logger.WithFields(logrus.Fields{
		"uri":        uri,
		"event_type": watchEvent.Type,
		"file_path":  event.Name,
	}).Debug("Resource change detected")
}

// uriToFilePath converts a URI to a file path
func (w *FileSystemResourceWatcher) uriToFilePath(uri string) string {
	// Handle file:// URIs
	if len(uri) > 7 && uri[:7] == "file://" {
		return uri[7:]
	}

	// Handle relative paths
	if filepath.IsAbs(uri) {
		return uri
	}

	return ""
}

// generateCallbackID generates a unique callback ID
func generateCallbackID() string {
	return time.Now().Format("20060102150405") + "_" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// SimpleWatchCallback is a simple implementation of WatchCallback
type SimpleWatchCallback struct {
	OnChange func(uri string, event WatchEvent) error
}

// OnResourceChanged implements WatchCallback
func (c *SimpleWatchCallback) OnResourceChanged(uri string, event WatchEvent) error {
	if c.OnChange != nil {
		return c.OnChange(uri, event)
	}
	return nil
}
