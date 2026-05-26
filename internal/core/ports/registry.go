package ports

// SourceFactory is a constructor function for creating Source instances.
type SourceFactory func(cfg *SourceConfig) (Source, error)

// SinkFactory is a constructor function for creating Sink instances.
type SinkFactory func(cfg *SinkConfig) (Sink, error)

// Registry provides factory pattern for creating connector instances.
type Registry interface {
	RegisterSource(typeName string, factory SourceFactory)
	RegisterSink(typeName string, factory SinkFactory)
	CreateSource(cfg *SourceConfig) (Source, error)
	CreateSink(cfg *SinkConfig) (Sink, error)
	SourceNames() []string
	SinkNames() []string
}
