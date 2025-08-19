package devtools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Profiler handles performance profiling
type Profiler struct {
	logger   *logrus.Logger
	tracer   trace.Tracer
	config   ProfilingConfig
	profiles map[string]*models.Profile
	mu       sync.RWMutex
	running  bool
	stopCh   chan struct{}
}

// NewProfiler creates a new profiler
func NewProfiler(logger *logrus.Logger, config ProfilingConfig) (*Profiler, error) {
	tracer := otel.Tracer("profiler")

	// Create output directory if it doesn't exist
	if config.OutputDir != "" {
		if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create profile output directory: %w", err)
		}
	}

	return &Profiler{
		logger:   logger,
		tracer:   tracer,
		config:   config,
		profiles: make(map[string]*models.Profile),
		stopCh:   make(chan struct{}),
	}, nil
}

// Start initializes the profiler
func (p *Profiler) Start(ctx context.Context) error {
	ctx, span := p.tracer.Start(ctx, "profiler.Start")
	defer span.End()

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("profiler is already running")
	}

	if !p.config.Enabled {
		p.logger.Info("Profiler is disabled")
		return nil
	}

	p.logger.Info("Starting profiler")

	// Enable profiling types
	if p.config.BlockProfiling {
		runtime.SetBlockProfileRate(1)
	}

	if p.config.MutexProfiling {
		runtime.SetMutexProfileFraction(1)
	}

	// Start continuous profiling if enabled
	go p.continuousProfiling()

	p.running = true
	p.logger.Info("Profiler started successfully")

	return nil
}

// Stop shuts down the profiler
func (p *Profiler) Stop(ctx context.Context) error {
	ctx, span := p.tracer.Start(ctx, "profiler.Stop")
	defer span.End()

	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	p.logger.Info("Stopping profiler")

	// Disable profiling
	runtime.SetBlockProfileRate(0)
	runtime.SetMutexProfileFraction(0)

	close(p.stopCh)
	p.running = false
	p.logger.Info("Profiler stopped")

	return nil
}

// GetStatus returns the current profiler status
func (p *Profiler) GetStatus(ctx context.Context) (*models.ProfilerStatus, error) {
	ctx, span := p.tracer.Start(ctx, "profiler.GetStatus")
	defer span.End()

	p.mu.RLock()
	defer p.mu.RUnlock()

	profiles := make([]*models.Profile, 0, len(p.profiles))
	for _, profile := range p.profiles {
		profiles = append(profiles, profile)
	}

	return &models.ProfilerStatus{
		Enabled:            p.config.Enabled,
		Running:            p.running,
		CPUProfiling:       p.config.CPUProfiling,
		MemoryProfiling:    p.config.MemoryProfiling,
		GoroutineProfiling: p.config.GoroutineProfiling,
		BlockProfiling:     p.config.BlockProfiling,
		MutexProfiling:     p.config.MutexProfiling,
		Profiles:           profiles,
		Timestamp:          time.Now(),
	}, nil
}

// StartCPUProfile starts CPU profiling
func (p *Profiler) StartCPUProfile(ctx context.Context) (*models.Profile, error) {
	ctx, span := p.tracer.Start(ctx, "profiler.StartCPUProfile")
	defer span.End()

	if !p.config.CPUProfiling {
		return nil, fmt.Errorf("CPU profiling is disabled")
	}

	profileID := fmt.Sprintf("cpu-%d", time.Now().Unix())
	filename := filepath.Join(p.config.OutputDir, fmt.Sprintf("%s.prof", profileID))

	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create CPU profile file: %w", err)
	}

	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to start CPU profiling: %w", err)
	}

	profile := &models.Profile{
		ID:        profileID,
		Type:      "cpu",
		Filename:  filename,
		StartTime: time.Now(),
		Active:    true,
	}

	p.mu.Lock()
	p.profiles[profileID] = profile
	p.mu.Unlock()

	p.logger.WithFields(logrus.Fields{
		"profile_id": profileID,
		"filename":   filename,
	}).Info("CPU profiling started")

	// Auto-stop after configured duration
	go func() {
		time.Sleep(p.config.ProfileDuration)
		p.StopCPUProfile(context.Background(), profileID)
		file.Close()
	}()

	return profile, nil
}

// StopCPUProfile stops CPU profiling
func (p *Profiler) StopCPUProfile(ctx context.Context, profileID string) error {
	ctx, span := p.tracer.Start(ctx, "profiler.StopCPUProfile")
	defer span.End()

	p.mu.Lock()
	defer p.mu.Unlock()

	profile, exists := p.profiles[profileID]
	if !exists {
		return fmt.Errorf("CPU profile %s not found", profileID)
	}

	if profile.Type != "cpu" {
		return fmt.Errorf("profile %s is not a CPU profile", profileID)
	}

	pprof.StopCPUProfile()

	profile.Active = false
	profile.EndTime = time.Now()
	profile.Duration = profile.EndTime.Sub(profile.StartTime)

	p.logger.WithFields(logrus.Fields{
		"profile_id": profileID,
		"duration":   profile.Duration,
	}).Info("CPU profiling stopped")

	return nil
}

// CreateMemoryProfile creates a memory profile
func (p *Profiler) CreateMemoryProfile(ctx context.Context) (*models.Profile, error) {
	ctx, span := p.tracer.Start(ctx, "profiler.CreateMemoryProfile")
	defer span.End()

	if !p.config.MemoryProfiling {
		return nil, fmt.Errorf("memory profiling is disabled")
	}

	profileID := fmt.Sprintf("memory-%d", time.Now().Unix())
	filename := filepath.Join(p.config.OutputDir, fmt.Sprintf("%s.prof", profileID))

	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory profile file: %w", err)
	}
	defer file.Close()

	// Force garbage collection before profiling
	runtime.GC()

	if err := pprof.WriteHeapProfile(file); err != nil {
		return nil, fmt.Errorf("failed to write memory profile: %w", err)
	}

	profile := &models.Profile{
		ID:        profileID,
		Type:      "memory",
		Filename:  filename,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Duration:  0,
		Active:    false,
	}

	p.mu.Lock()
	p.profiles[profileID] = profile
	p.mu.Unlock()

	p.logger.WithFields(logrus.Fields{
		"profile_id": profileID,
		"filename":   filename,
	}).Info("Memory profile created")

	return profile, nil
}

// CreateGoroutineProfile creates a goroutine profile
func (p *Profiler) CreateGoroutineProfile(ctx context.Context) (*models.Profile, error) {
	ctx, span := p.tracer.Start(ctx, "profiler.CreateGoroutineProfile")
	defer span.End()

	if !p.config.GoroutineProfiling {
		return nil, fmt.Errorf("goroutine profiling is disabled")
	}

	profileID := fmt.Sprintf("goroutine-%d", time.Now().Unix())
	filename := filepath.Join(p.config.OutputDir, fmt.Sprintf("%s.prof", profileID))

	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create goroutine profile file: %w", err)
	}
	defer file.Close()

	goroutineProfile := pprof.Lookup("goroutine")
	if err := goroutineProfile.WriteTo(file, 0); err != nil {
		return nil, fmt.Errorf("failed to write goroutine profile: %w", err)
	}

	profile := &models.Profile{
		ID:        profileID,
		Type:      "goroutine",
		Filename:  filename,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Duration:  0,
		Active:    false,
	}

	p.mu.Lock()
	p.profiles[profileID] = profile
	p.mu.Unlock()

	p.logger.WithFields(logrus.Fields{
		"profile_id": profileID,
		"filename":   filename,
	}).Info("Goroutine profile created")

	return profile, nil
}

// ListProfiles returns all profiles
func (p *Profiler) ListProfiles(ctx context.Context) ([]*models.Profile, error) {
	ctx, span := p.tracer.Start(ctx, "profiler.ListProfiles")
	defer span.End()

	p.mu.RLock()
	defer p.mu.RUnlock()

	profiles := make([]*models.Profile, 0, len(p.profiles))
	for _, profile := range p.profiles {
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// GetProfile returns a specific profile
func (p *Profiler) GetProfile(ctx context.Context, profileID string) (*models.Profile, error) {
	ctx, span := p.tracer.Start(ctx, "profiler.GetProfile")
	defer span.End()

	p.mu.RLock()
	defer p.mu.RUnlock()

	profile, exists := p.profiles[profileID]
	if !exists {
		return nil, fmt.Errorf("profile %s not found", profileID)
	}

	return profile, nil
}

// DeleteProfile deletes a profile
func (p *Profiler) DeleteProfile(ctx context.Context, profileID string) error {
	ctx, span := p.tracer.Start(ctx, "profiler.DeleteProfile")
	defer span.End()

	p.mu.Lock()
	defer p.mu.Unlock()

	profile, exists := p.profiles[profileID]
	if !exists {
		return fmt.Errorf("profile %s not found", profileID)
	}

	// Delete the profile file
	if err := os.Remove(profile.Filename); err != nil {
		p.logger.WithError(err).Warn("Failed to delete profile file")
	}

	delete(p.profiles, profileID)

	p.logger.WithField("profile_id", profileID).Info("Profile deleted")

	return nil
}

// GetRuntimeStats returns current runtime statistics
func (p *Profiler) GetRuntimeStats(ctx context.Context) (*models.RuntimeStats, error) {
	ctx, span := p.tracer.Start(ctx, "profiler.GetRuntimeStats")
	defer span.End()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &models.RuntimeStats{
		Goroutines:    runtime.NumGoroutine(),
		CGoCalls:      runtime.NumCgoCall(),
		HeapAlloc:     memStats.HeapAlloc,
		HeapSys:       memStats.HeapSys,
		HeapIdle:      memStats.HeapIdle,
		HeapInuse:     memStats.HeapInuse,
		HeapReleased:  memStats.HeapReleased,
		HeapObjects:   memStats.HeapObjects,
		StackInuse:    memStats.StackInuse,
		StackSys:      memStats.StackSys,
		MSpanInuse:    memStats.MSpanInuse,
		MSpanSys:      memStats.MSpanSys,
		MCacheInuse:   memStats.MCacheInuse,
		MCacheSys:     memStats.MCacheSys,
		GCSys:         memStats.GCSys,
		OtherSys:      memStats.OtherSys,
		NextGC:        memStats.NextGC,
		LastGC:        time.Unix(0, int64(memStats.LastGC)),
		PauseTotalNs:  memStats.PauseTotalNs,
		NumGC:         memStats.NumGC,
		NumForcedGC:   memStats.NumForcedGC,
		GCCPUFraction: memStats.GCCPUFraction,
		Timestamp:     time.Now(),
	}, nil
}

// continuousProfiling runs continuous profiling in the background
func (p *Profiler) continuousProfiling() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Create periodic memory profiles
			if p.config.MemoryProfiling {
				_, err := p.CreateMemoryProfile(context.Background())
				if err != nil {
					p.logger.WithError(err).Error("Failed to create periodic memory profile")
				}
			}

			// Create periodic goroutine profiles
			if p.config.GoroutineProfiling {
				_, err := p.CreateGoroutineProfile(context.Background())
				if err != nil {
					p.logger.WithError(err).Error("Failed to create periodic goroutine profile")
				}
			}

			// Cleanup old profiles
			p.cleanupOldProfiles()

		case <-p.stopCh:
			p.logger.Debug("Continuous profiling stopped")
			return
		}
	}
}

// cleanupOldProfiles removes profiles older than 24 hours
func (p *Profiler) cleanupOldProfiles() {
	p.mu.Lock()
	defer p.mu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	for id, profile := range p.profiles {
		if profile.StartTime.Before(cutoff) {
			// Delete the profile file
			if err := os.Remove(profile.Filename); err != nil {
				p.logger.WithError(err).Warn("Failed to delete old profile file")
			}

			delete(p.profiles, id)
			p.logger.WithField("profile_id", id).Debug("Cleaned up old profile")
		}
	}
}
