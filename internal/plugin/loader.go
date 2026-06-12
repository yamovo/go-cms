package plugin

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/vortexcms/go-cms/internal/models"
	"gorm.io/gorm"
)

// Plugin represents a loaded plugin instance.
type Plugin struct {
	Name        string
	Version     string
	Description string
	Author      string
	Enabled     bool
	Hooks       map[string][]HookFunc
	Config      map[string]interface{}
}

// HookFunc is a function called on a hook event.
type HookFunc func(args map[string]interface{}) (interface{}, error)

// Manager manages plugins lifecycle.
type Manager struct {
	db       *gorm.DB
	plugins  map[string]*Plugin
	pluginsDir string
}

// NewManager creates a new plugin manager.
func NewManager(db *gorm.DB, pluginsDir string) *Manager {
	return &Manager{
		db:         db,
		plugins:    make(map[string]*Plugin),
		pluginsDir: pluginsDir,
	}
}

// LoadAll scans the plugins directory and loads enabled plugins.
func (m *Manager) LoadAll() error {
	entries, err := os.ReadDir(m.pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("plugins directory not found, skipping", "dir", m.pluginsDir)
			return nil
		}
		return fmt.Errorf("failed to read plugins dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(m.pluginsDir, entry.Name(), "plugin.json")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue
		}

		plugin, err := m.loadPlugin(entry.Name(), manifestPath)
		if err != nil {
			slog.Warn("failed to load plugin", "name", entry.Name(), "error", err)
			continue
		}

		m.plugins[plugin.Name] = plugin
		slog.Info("plugin loaded", "name", plugin.Name, "version", plugin.Version, "enabled", plugin.Enabled)
	}

	return nil
}

// loadPlugin loads a single plugin from its manifest.
func (m *Manager) loadPlugin(dirName, manifestPath string) (*Plugin, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest struct {
		Name        string                 `json:"name"`
		Version     string                 `json:"version"`
		Description string                 `json:"description"`
		Author      string                 `json:"author"`
		Hooks       map[string][]string    `json:"hooks"`
		Config      map[string]interface{} `json:"config"`
	}

	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	// Check database for enabled state.
	var dbPlugin models.Plugin
	enabled := true
	if err := m.db.Where("slug = ?", manifest.Name).First(&dbPlugin).Error; err == nil {
		enabled = dbPlugin.IsEnabled
	}

	p := &Plugin{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		Enabled:     enabled,
		Hooks:       make(map[string][]HookFunc),
		Config:      manifest.Config,
	}

	// Register hooks from manifest (placeholder - actual hook loading would need Go plugin or scripting).
	for hookName, handlers := range manifest.Hooks {
		for _, handler := range handlers {
			_ = handler // In a real implementation, load the handler function
			p.Hooks[hookName] = append(p.Hooks[hookName], func(args map[string]interface{}) (interface{}, error) {
				slog.Debug("hook called", "plugin", manifest.Name, "hook", hookName)
				return nil, nil
			})
		}
	}

	return p, nil
}

// Get returns a loaded plugin by name.
func (m *Manager) Get(name string) (*Plugin, bool) {
	p, ok := m.plugins[name]
	return p, ok
}

// List returns all loaded plugins.
func (m *Manager) List() []*Plugin {
	var result []*Plugin
	for _, p := range m.plugins {
		result = append(result, p)
	}
	return result
}

// Enable activates a plugin.
func (m *Manager) Enable(name string) error {
	p, ok := m.plugins[name]
	if !ok {
		return fmt.Errorf("plugin %q not found", name)
	}

	p.Enabled = true
	m.db.Model(&models.Plugin{}).Where("slug = ?", name).Update("is_enabled", true)
	slog.Info("plugin enabled", "name", name)
	return nil
}

// Disable deactivates a plugin.
func (m *Manager) Disable(name string) error {
	p, ok := m.plugins[name]
	if !ok {
		return fmt.Errorf("plugin %q not found", name)
	}

	p.Enabled = false
	m.db.Model(&models.Plugin{}).Where("slug = ?", name).Update("is_enabled", false)
	slog.Info("plugin disabled", "name", name)
	return nil
}

// ExecuteHook runs all handlers for a given hook across enabled plugins.
func (m *Manager) ExecuteHook(hookName string, args map[string]interface{}) ([]interface{}, error) {
	var results []interface{}
	for _, p := range m.plugins {
		if !p.Enabled {
			continue
		}
		hooks, ok := p.Hooks[hookName]
		if !ok {
			continue
		}
		for _, hook := range hooks {
			result, err := hook(args)
			if err != nil {
				slog.Error("hook execution failed", "plugin", p.Name, "hook", hookName, "error", err)
				continue
			}
			if result != nil {
				results = append(results, result)
			}
		}
	}
	return results, nil
}

// IsEnabled checks if a plugin is enabled.
func (m *Manager) IsEnabled(name string) bool {
	p, ok := m.plugins[name]
	return ok && p.Enabled
}

// ListHookNames returns all registered hook names across plugins.
func (m *Manager) ListHookNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, p := range m.plugins {
		for hookName := range p.Hooks {
			if !seen[hookName] {
				seen[hookName] = true
				names = append(names, hookName)
			}
		}
	}
	return names
}

// InitDB ensures all discovered plugins have DB records.
func (m *Manager) InitDB() error {
	for _, p := range m.plugins {
		var count int64
		m.db.Model(&models.Plugin{}).Where("slug = ?", p.Name).Count(&count)
		if count == 0 {
			m.db.Create(&models.Plugin{
				Name:        p.Name,
				Slug:        strings.ToLower(p.Name),
				Description: p.Description,
				Version:     p.Version,
				IsEnabled:   p.Enabled,
			})
		}
	}
	return nil
}
