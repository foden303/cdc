package registry

import (
	"fmt"
	"sort"

	"github.com/foden/cdc/internal/core/ports"
)

// Compile-time assertion that Registry implements ports.Registry.
var _ ports.Registry = (*Registry)(nil)

// Registry implements ports.Registry using the factory pattern
// for creating source and sink connector instances.
type Registry struct {
	sources map[string]ports.SourceFactory
	sinks   map[string]ports.SinkFactory
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		sources: make(map[string]ports.SourceFactory),
		sinks:   make(map[string]ports.SinkFactory),
	}
}

// RegisterSource registers a source factory under a given type name.
// Panics if the type name is already registered.
func (r *Registry) RegisterSource(typeName string, factory ports.SourceFactory) {
	if _, exists := r.sources[typeName]; exists {
		panic(fmt.Sprintf("source type %q already registered", typeName))
	}
	r.sources[typeName] = factory
}

// RegisterSink registers a sink factory under a given type name.
// Panics if the type name is already registered.
func (r *Registry) RegisterSink(typeName string, factory ports.SinkFactory) {
	if _, exists := r.sinks[typeName]; exists {
		panic(fmt.Sprintf("sink type %q already registered", typeName))
	}
	r.sinks[typeName] = factory
}

// CreateSource looks up a registered source factory and creates a Source.
func (r *Registry) CreateSource(cfg *ports.SourceConfig) (ports.Source, error) {
	factory, ok := r.sources[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported source type: %q (registered: %v)", cfg.Type, r.SourceNames())
	}
	return factory(cfg)
}

// CreateSink looks up a registered sink factory and creates a Sink.
func (r *Registry) CreateSink(cfg *ports.SinkConfig) (ports.Sink, error) {
	factory, ok := r.sinks[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported sink type: %q (registered: %v)", cfg.Type, r.SinkNames())
	}
	return factory(cfg)
}

// SourceNames returns a sorted list of registered source type names.
func (r *Registry) SourceNames() []string {
	names := make([]string, 0, len(r.sources))
	for k := range r.sources {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// SinkNames returns a sorted list of registered sink type names.
func (r *Registry) SinkNames() []string {
	names := make([]string, 0, len(r.sinks))
	for k := range r.sinks {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// --- Package-level functions for backward compatibility with init() registrations ---

// defaultRegistry is the global registry used by package-level functions.
var defaultRegistry = NewRegistry()

// Default returns the package-level default registry instance.
func Default() *Registry {
	return defaultRegistry
}

// RegisterSource registers a source factory in the default registry.
// Typically called from init() in each source package.
func RegisterSource(typeName string, factory ports.SourceFactory) {
	defaultRegistry.RegisterSource(typeName, factory)
}

// RegisterSink registers a sink factory in the default registry.
// Typically called from init() in each sink package.
func RegisterSink(typeName string, factory ports.SinkFactory) {
	defaultRegistry.RegisterSink(typeName, factory)
}

// CreateSource creates a source using the default registry.
func CreateSource(cfg *ports.SourceConfig) (ports.Source, error) {
	return defaultRegistry.CreateSource(cfg)
}

// CreateSink creates a sink using the default registry.
func CreateSink(cfg *ports.SinkConfig) (ports.Sink, error) {
	return defaultRegistry.CreateSink(cfg)
}

// SourceNames returns registered source type names from the default registry.
func SourceNames() []string {
	return defaultRegistry.SourceNames()
}

// SinkNames returns registered sink type names from the default registry.
func SinkNames() []string {
	return defaultRegistry.SinkNames()
}
