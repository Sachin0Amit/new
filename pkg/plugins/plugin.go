package plugins

import (
	"sync"
	"time"

	"github.com/papi-ai/sovereign-core/internal/models"
	"github.com/papi-ai/sovereign-core/pkg/logger"
)

// Plugin defines the standard contract for all Sovereign system extensions.
type Plugin interface {
	Metadata() models.PluginManifest
	OnHook(hook models.HookPoint, task *models.Task) error
}

// Registry manages the lifecycle and execution of active plugins.
type Registry struct {
	plugins map[string]Plugin
	hooks   map[models.HookPoint][]Plugin
	logger  logger.Logger
	mu      sync.RWMutex
}

// NewRegistry initializes a thread-safe plugin management layer.
func NewRegistry(l logger.Logger) *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
		hooks:   make(map[models.HookPoint][]Plugin),
		logger:  l,
	}
}

// Register adds a plugin to the active registry and maps its capabilities to hook points.
func (r *Registry) Register(p Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()

	manifest := p.Metadata()
	r.plugins[manifest.ID] = p

	// In a real-world scenario, we'd inspect the plugin capabilities to map hooks properly
	r.hooks[models.HookPreProcessing] = append(r.hooks[models.HookPreProcessing], p)
	r.hooks[models.HookPostInference] = append(r.hooks[models.HookPostInference], p)

	r.logger.Info("Plugin registered", logger.String("plugin_id", manifest.ID))
}

// InvokeHook executes all plugins registered for a specific cognitive stage.
func (r *Registry) InvokeHook(hook models.HookPoint, task *models.Task) {
	r.mu.RLock()
	activePlugins := r.hooks[hook]
	r.mu.RUnlock()

	for _, p := range activePlugins {
		start := time.Now()
		if err := p.OnHook(hook, task); err != nil {
			r.logger.Error("Plugin hook execution failed", 
				logger.String("plugin_id", p.Metadata().ID),
				logger.String("hook", string(hook)),
				logger.ErrorF(err))
			continue
		}
		
		r.logger.Debug("Plugin hook executed", 
			logger.String("plugin_id", p.Metadata().ID),
			logger.Duration("duration", time.Since(start)))
	}
}

// ListActivePlugins returns telemetry for all currently loaded extensions.
func (r *Registry) ListActivePlugins() []models.PluginManifest {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var manifests []models.PluginManifest
	for _, p := range r.plugins {
		manifests = append(manifests, p.Metadata())
	}
	return manifests
}
