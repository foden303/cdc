// Package storage provides NATS KV-backed persistence for CDC system state.
package storage

const (
	// BucketName is the NATS KV bucket used for all CDC state persistence.
	BucketName = "CDC_STATE"

	// PrefixSources is the key prefix for source configurations.
	PrefixSources = "sources."

	// PrefixSinks is the key prefix for sink configurations.
	PrefixSinks = "sinks."

	// PrefixFlows is the key prefix for flow configurations.
	PrefixFlows = "flows."

	// PrefixOffsets is the key prefix for flow consumer offsets.
	PrefixOffsets = "offsets."
)
