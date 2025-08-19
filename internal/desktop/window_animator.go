package desktop

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// WindowAnimator handles smooth window animations and transitions
type WindowAnimator struct {
	logger *logrus.Logger
	tracer trace.Tracer
	config AnimatorConfig
	mu     sync.RWMutex

	// Animation state
	activeAnimations map[string]*Animation
	animationQueue   []*Animation
	running          bool
	stopCh           chan struct{}

	// Performance tracking
	frameRate     float64
	lastFrameTime time.Time
	droppedFrames int
	totalFrames   int
}

// AnimatorConfig defines animation configuration
type AnimatorConfig struct {
	DefaultDuration time.Duration `json:"default_duration"`
	TargetFPS       int           `json:"target_fps"`
	EasingFunction  string        `json:"easing_function"`
	EnableVSync     bool          `json:"enable_vsync"`
	ReduceMotion    bool          `json:"reduce_motion"`
	HardwareAccel   bool          `json:"hardware_acceleration"`
	MaxConcurrent   int           `json:"max_concurrent_animations"`
	QualityLevel    string        `json:"quality_level"` // "low", "medium", "high"
}

// Animation represents a window animation
type Animation struct {
	ID           string                 `json:"id"`
	WindowID     string                 `json:"window_id"`
	Type         AnimationType          `json:"type"`
	StartTime    time.Time              `json:"start_time"`
	Duration     time.Duration          `json:"duration"`
	Progress     float64                `json:"progress"`
	StartState   WindowState            `json:"start_state"`
	EndState     WindowState            `json:"end_state"`
	CurrentState WindowState            `json:"current_state"`
	EasingFunc   EasingFunction         `json:"-"`
	OnComplete   func(*Animation)       `json:"-"`
	OnUpdate     func(*Animation)       `json:"-"`
	Properties   map[string]interface{} `json:"properties"`
	Priority     int                    `json:"priority"`
	Completed    bool                   `json:"completed"`
}

// AnimationType defines the type of animation
type AnimationType string

const (
	AnimationMove   AnimationType = "move"
	AnimationResize AnimationType = "resize"
	AnimationFade   AnimationType = "fade"
	AnimationSlide  AnimationType = "slide"
	AnimationScale  AnimationType = "scale"
	AnimationRotate AnimationType = "rotate"
	AnimationFlip   AnimationType = "flip"
	AnimationBounce AnimationType = "bounce"
	AnimationShake  AnimationType = "shake"
	AnimationPulse  AnimationType = "pulse"
	AnimationCustom AnimationType = "custom"
)

// WindowState represents the state of a window during animation
type WindowState struct {
	Position    models.Position `json:"position"`
	Size        models.Size     `json:"size"`
	Opacity     float64         `json:"opacity"`
	Scale       float64         `json:"scale"`
	Rotation    float64         `json:"rotation"`
	ZIndex      int             `json:"z_index"`
	Visible     bool            `json:"visible"`
	BorderWidth int             `json:"border_width"`
	Shadow      ShadowState     `json:"shadow"`
}

// ShadowState represents window shadow properties
type ShadowState struct {
	Enabled bool    `json:"enabled"`
	OffsetX int     `json:"offset_x"`
	OffsetY int     `json:"offset_y"`
	Blur    int     `json:"blur"`
	Opacity float64 `json:"opacity"`
	Color   string  `json:"color"`
}

// EasingFunction defines animation easing
type EasingFunction func(t float64) float64

// NewWindowAnimator creates a new window animator
func NewWindowAnimator(logger *logrus.Logger, config AnimatorConfig) *WindowAnimator {
	tracer := otel.Tracer("window-animator")

	animator := &WindowAnimator{
		logger:           logger,
		tracer:           tracer,
		config:           config,
		activeAnimations: make(map[string]*Animation),
		animationQueue:   make([]*Animation, 0),
		stopCh:           make(chan struct{}),
		frameRate:        float64(config.TargetFPS),
		lastFrameTime:    time.Now(),
	}

	return animator
}

// Start starts the animation engine
func (wa *WindowAnimator) Start(ctx context.Context) error {
	ctx, span := wa.tracer.Start(ctx, "windowAnimator.Start")
	defer span.End()

	wa.mu.Lock()
	defer wa.mu.Unlock()

	if wa.running {
		return fmt.Errorf("animator already running")
	}

	wa.running = true

	// Start animation loop
	go wa.animationLoop()

	wa.logger.Info("Window animator started")
	return nil
}

// Stop stops the animation engine
func (wa *WindowAnimator) Stop(ctx context.Context) error {
	ctx, span := wa.tracer.Start(ctx, "windowAnimator.Stop")
	defer span.End()

	wa.mu.Lock()
	defer wa.mu.Unlock()

	if !wa.running {
		return nil
	}

	wa.running = false
	close(wa.stopCh)

	// Complete all active animations immediately
	for _, animation := range wa.activeAnimations {
		animation.Progress = 1.0
		animation.CurrentState = animation.EndState
		animation.Completed = true

		if animation.OnComplete != nil {
			animation.OnComplete(animation)
		}
	}

	wa.activeAnimations = make(map[string]*Animation)
	wa.animationQueue = make([]*Animation, 0)

	wa.logger.Info("Window animator stopped")
	return nil
}

// AnimateWindow animates a window with the specified parameters
func (wa *WindowAnimator) AnimateWindow(ctx context.Context, windowID string, animType AnimationType, startState, endState WindowState, duration time.Duration) (*Animation, error) {
	ctx, span := wa.tracer.Start(ctx, "windowAnimator.AnimateWindow")
	defer span.End()

	// Create animation
	animation := &Animation{
		ID:           fmt.Sprintf("anim_%s_%d", windowID, time.Now().UnixNano()),
		WindowID:     windowID,
		Type:         animType,
		StartTime:    time.Now(),
		Duration:     duration,
		Progress:     0.0,
		StartState:   startState,
		EndState:     endState,
		CurrentState: startState,
		EasingFunc:   wa.getEasingFunction(wa.config.EasingFunction),
		Properties:   make(map[string]interface{}),
		Priority:     1,
		Completed:    false,
	}

	// Apply motion reduction if enabled
	if wa.config.ReduceMotion {
		animation.Duration = animation.Duration / 4
		if animation.Duration < 50*time.Millisecond {
			animation.Duration = 50 * time.Millisecond
		}
	}

	// Add to queue or start immediately
	wa.mu.Lock()
	defer wa.mu.Unlock()

	if len(wa.activeAnimations) < wa.config.MaxConcurrent {
		wa.activeAnimations[animation.ID] = animation
	} else {
		wa.animationQueue = append(wa.animationQueue, animation)
	}

	wa.logger.WithFields(logrus.Fields{
		"animation_id": animation.ID,
		"window_id":    windowID,
		"type":         animType,
		"duration":     duration,
	}).Debug("Animation created")

	return animation, nil
}

// animationLoop runs the main animation loop
func (wa *WindowAnimator) animationLoop() {
	targetFrameTime := time.Duration(1000/wa.config.TargetFPS) * time.Millisecond
	ticker := time.NewTicker(targetFrameTime)
	defer ticker.Stop()

	for {
		select {
		case <-wa.stopCh:
			return
		case <-ticker.C:
			wa.updateAnimations()
		}
	}
}

// updateAnimations updates all active animations
func (wa *WindowAnimator) updateAnimations() {
	wa.mu.Lock()
	defer wa.mu.Unlock()

	now := time.Now()
	frameTime := now.Sub(wa.lastFrameTime)
	wa.lastFrameTime = now
	wa.totalFrames++

	// Update frame rate calculation
	if frameTime > 0 {
		currentFPS := 1.0 / frameTime.Seconds()
		wa.frameRate = wa.frameRate*0.9 + currentFPS*0.1 // Smooth average
	}

	// Check for dropped frames
	expectedFrameTime := time.Duration(1000/wa.config.TargetFPS) * time.Millisecond
	if frameTime > expectedFrameTime*2 {
		wa.droppedFrames++
	}

	// Update active animations
	completedAnimations := make([]string, 0)

	for id, animation := range wa.activeAnimations {
		if wa.updateAnimation(animation, now) {
			completedAnimations = append(completedAnimations, id)
		}
	}

	// Remove completed animations
	for _, id := range completedAnimations {
		delete(wa.activeAnimations, id)
	}

	// Start queued animations if slots available
	wa.startQueuedAnimations()
}

// updateAnimation updates a single animation
func (wa *WindowAnimator) updateAnimation(animation *Animation, now time.Time) bool {
	if animation.Completed {
		return true
	}

	// Calculate progress
	elapsed := now.Sub(animation.StartTime)
	if elapsed >= animation.Duration {
		animation.Progress = 1.0
		animation.Completed = true
	} else {
		animation.Progress = float64(elapsed) / float64(animation.Duration)
	}

	// Apply easing function
	easedProgress := animation.EasingFunc(animation.Progress)

	// Interpolate current state
	animation.CurrentState = wa.interpolateState(animation.StartState, animation.EndState, easedProgress)

	// Call update callback
	if animation.OnUpdate != nil {
		animation.OnUpdate(animation)
	}

	// Call completion callback if finished
	if animation.Completed && animation.OnComplete != nil {
		animation.OnComplete(animation)
	}

	return animation.Completed
}

// interpolateState interpolates between two window states
func (wa *WindowAnimator) interpolateState(start, end WindowState, progress float64) WindowState {
	return WindowState{
		Position: models.Position{
			X: int(wa.lerp(float64(start.Position.X), float64(end.Position.X), progress)),
			Y: int(wa.lerp(float64(start.Position.Y), float64(end.Position.Y), progress)),
		},
		Size: models.Size{
			Width:  int(wa.lerp(float64(start.Size.Width), float64(end.Size.Width), progress)),
			Height: int(wa.lerp(float64(start.Size.Height), float64(end.Size.Height), progress)),
		},
		Opacity:     wa.lerp(start.Opacity, end.Opacity, progress),
		Scale:       wa.lerp(start.Scale, end.Scale, progress),
		Rotation:    wa.lerp(start.Rotation, end.Rotation, progress),
		ZIndex:      int(wa.lerp(float64(start.ZIndex), float64(end.ZIndex), progress)),
		Visible:     progress > 0.5, // Simple visibility logic
		BorderWidth: int(wa.lerp(float64(start.BorderWidth), float64(end.BorderWidth), progress)),
		Shadow: ShadowState{
			Enabled: end.Shadow.Enabled,
			OffsetX: int(wa.lerp(float64(start.Shadow.OffsetX), float64(end.Shadow.OffsetX), progress)),
			OffsetY: int(wa.lerp(float64(start.Shadow.OffsetY), float64(end.Shadow.OffsetY), progress)),
			Blur:    int(wa.lerp(float64(start.Shadow.Blur), float64(end.Shadow.Blur), progress)),
			Opacity: wa.lerp(start.Shadow.Opacity, end.Shadow.Opacity, progress),
			Color:   end.Shadow.Color, // Color interpolation would be more complex
		},
	}
}

// lerp performs linear interpolation
func (wa *WindowAnimator) lerp(start, end, progress float64) float64 {
	return start + (end-start)*progress
}

// startQueuedAnimations starts animations from the queue
func (wa *WindowAnimator) startQueuedAnimations() {
	availableSlots := wa.config.MaxConcurrent - len(wa.activeAnimations)

	if availableSlots > 0 && len(wa.animationQueue) > 0 {
		// Sort queue by priority
		wa.sortAnimationQueue()

		// Start highest priority animations
		toStart := min(availableSlots, len(wa.animationQueue))
		for i := 0; i < toStart; i++ {
			animation := wa.animationQueue[i]
			animation.StartTime = time.Now() // Reset start time
			wa.activeAnimations[animation.ID] = animation
		}

		// Remove started animations from queue
		wa.animationQueue = wa.animationQueue[toStart:]
	}
}

// sortAnimationQueue sorts animations by priority
func (wa *WindowAnimator) sortAnimationQueue() {
	// Sort by priority (higher first), then by creation time
	for i := 0; i < len(wa.animationQueue)-1; i++ {
		for j := i + 1; j < len(wa.animationQueue); j++ {
			if wa.animationQueue[i].Priority < wa.animationQueue[j].Priority {
				wa.animationQueue[i], wa.animationQueue[j] = wa.animationQueue[j], wa.animationQueue[i]
			}
		}
	}
}

// getEasingFunction returns the specified easing function
func (wa *WindowAnimator) getEasingFunction(name string) EasingFunction {
	switch name {
	case "linear":
		return EaseLinear
	case "ease_in":
		return EaseInQuad
	case "ease_out":
		return EaseOutQuad
	case "ease_in_out":
		return EaseInOutQuad
	case "ease_in_cubic":
		return EaseInCubic
	case "ease_out_cubic":
		return EaseOutCubic
	case "ease_in_out_cubic":
		return EaseInOutCubic
	case "bounce":
		return EaseOutBounce
	case "elastic":
		return EaseOutElastic
	default:
		return EaseInOutQuad
	}
}

// Predefined animations

// FadeIn creates a fade-in animation
func (wa *WindowAnimator) FadeIn(ctx context.Context, windowID string, duration time.Duration) (*Animation, error) {
	startState := WindowState{Opacity: 0.0, Visible: false}
	endState := WindowState{Opacity: 1.0, Visible: true}
	return wa.AnimateWindow(ctx, windowID, AnimationFade, startState, endState, duration)
}

// FadeOut creates a fade-out animation
func (wa *WindowAnimator) FadeOut(ctx context.Context, windowID string, duration time.Duration) (*Animation, error) {
	startState := WindowState{Opacity: 1.0, Visible: true}
	endState := WindowState{Opacity: 0.0, Visible: false}
	return wa.AnimateWindow(ctx, windowID, AnimationFade, startState, endState, duration)
}

// SlideIn creates a slide-in animation
func (wa *WindowAnimator) SlideIn(ctx context.Context, windowID string, fromPosition, toPosition models.Position, duration time.Duration) (*Animation, error) {
	startState := WindowState{Position: fromPosition, Opacity: 0.8}
	endState := WindowState{Position: toPosition, Opacity: 1.0}
	return wa.AnimateWindow(ctx, windowID, AnimationSlide, startState, endState, duration)
}

// ScaleIn creates a scale-in animation
func (wa *WindowAnimator) ScaleIn(ctx context.Context, windowID string, duration time.Duration) (*Animation, error) {
	startState := WindowState{Scale: 0.0, Opacity: 0.0}
	endState := WindowState{Scale: 1.0, Opacity: 1.0}
	return wa.AnimateWindow(ctx, windowID, AnimationScale, startState, endState, duration)
}

// GetPerformanceMetrics returns animation performance metrics
func (wa *WindowAnimator) GetPerformanceMetrics() map[string]interface{} {
	wa.mu.RLock()
	defer wa.mu.RUnlock()

	dropRate := 0.0
	if wa.totalFrames > 0 {
		dropRate = float64(wa.droppedFrames) / float64(wa.totalFrames) * 100
	}

	return map[string]interface{}{
		"frame_rate":        wa.frameRate,
		"dropped_frames":    wa.droppedFrames,
		"total_frames":      wa.totalFrames,
		"drop_rate":         dropRate,
		"active_animations": len(wa.activeAnimations),
		"queued_animations": len(wa.animationQueue),
	}
}

// Easing functions

func EaseLinear(t float64) float64 {
	return t
}

func EaseInQuad(t float64) float64 {
	return t * t
}

func EaseOutQuad(t float64) float64 {
	return t * (2 - t)
}

func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

func EaseInCubic(t float64) float64 {
	return t * t * t
}

func EaseOutCubic(t float64) float64 {
	t--
	return t*t*t + 1
}

func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	t = 2*t - 2
	return 1 + t*t*t/2
}

func EaseOutBounce(t float64) float64 {
	if t < 1/2.75 {
		return 7.5625 * t * t
	} else if t < 2/2.75 {
		t -= 1.5 / 2.75
		return 7.5625*t*t + 0.75
	} else if t < 2.5/2.75 {
		t -= 2.25 / 2.75
		return 7.5625*t*t + 0.9375
	} else {
		t -= 2.625 / 2.75
		return 7.5625*t*t + 0.984375
	}
}

func EaseOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return math.Pow(2, -10*t)*math.Sin((t-0.1)*2*math.Pi/0.4) + 1
}
