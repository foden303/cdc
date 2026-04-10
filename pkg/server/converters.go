package server

import (
	"errors"
	"fmt"

	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/foden/cdc/pkg/config"
	"github.com/foden/cdc/pkg/utils"
)

var (
	ErrSourceConfigRequired   = errors.New("source config is required")
	ErrSourceTypeRequired     = errors.New("source type is required")
	ErrSourceHostRequired     = errors.New("source host is required")
	ErrSourcePortRequired     = errors.New("source port must be positive")
	ErrSourceDatabaseRequired = errors.New("source database is required")
	ErrSinkConfigRequired     = errors.New("sink config is required")
	ErrSinkTypeRequired       = errors.New("sink type is required")
	ErrSinkURLRequired        = errors.New("sink URL is required")
)

// toSourceConfig converts proto SourceConfig to internal SourceConfig with validation and defaults.
func toSourceConfig(p *cdcpb.SourceConfig) (*config.SourceConfig, error) {
	if p == nil {
		return nil, ErrSourceConfigRequired
	}

	if p.Type == "" {
		return nil, ErrSourceTypeRequired
	}
	if p.Host == "" {
		return nil, ErrSourceHostRequired
	}
	if p.Port <= 0 {
		return nil, ErrSourcePortRequired
	}
	if p.Database == "" {
		return nil, ErrSourceDatabaseRequired
	}

	c := &config.SourceConfig{
		Type:            p.Type,
		Host:            p.Host,
		Port:            int(p.Port),
		Database:        p.Database,
		Tables:          p.Tables,
		Topic:           p.GetTopic(),
		Username:        p.GetUsername(),
		Password:        p.GetPassword(),
		SlotName:        p.GetSlotName(),
		PublicationName: p.GetPublicationName(),
		InstanceID:      p.InstanceId,
		Name:            p.GetName(),
	}

	// Apply Dynamic Defaults
	if c.InstanceID == "" {
		c.InstanceID = fmt.Sprintf("src-%s-%s", c.Type, utils.RandomString(4))
	}

	if c.Type == "postgres" {
		if c.SlotName == "" {
			c.SlotName = fmt.Sprintf("cdc_slot_%s", c.InstanceID)
		}
		if c.PublicationName == "" {
			c.PublicationName = fmt.Sprintf("cdc_pub_%s", c.InstanceID)
		}
	}

	return c, nil
}

// toSinkConfig converts proto SinkConfig to internal SinkConfig with validation and defaults.
func toSinkConfig(p *cdcpb.SinkConfig) (*config.SinkConfig, error) {
	if p == nil {
		return nil, ErrSinkConfigRequired
	}

	if p.Type == "" {
		return nil, ErrSinkTypeRequired
	}
	if len(p.Url) == 0 {
		return nil, ErrSinkURLRequired
	}

	c := &config.SinkConfig{
		Type:            p.Type,
		URL:             p.Url,
		Username:        p.GetUsername(),
		Password:        p.GetPassword(),
		APIKey:          p.GetApiKey(),
		Topic:           p.GetTopic(),
		Name:            p.GetName(),
		Index:           p.GetIndex(),
		IndexMapping:    p.IndexMapping,
		IndexPrefix:     p.GetIndexPrefix(),
		BatchSize:       p.GetBatchSize(),
		FlushIntervalMs: p.GetFlushIntervalMs(),
		MaxRetries:      p.GetMaxRetries(),
		RetryBaseMs:     p.GetRetryBaseMs(),
		InstanceID:      p.InstanceId,
		FieldMapping:    p.FieldMapping,
	}

	if p.Redis != nil {
		c.Redis = config.RedisSettings{
			Command:     p.Redis.Command,
			KeyTemplate: p.Redis.KeyTemplate,
			ValueFields: p.Redis.ValueFields,
			TTL:         int(p.Redis.Ttl),
		}
	}

	// Apply defaults
	if c.BatchSize <= 0 {
		c.BatchSize = 500
	}
	if c.FlushIntervalMs <= 0 {
		c.FlushIntervalMs = 1000
	}
	if c.IndexPrefix == "" {
		c.IndexPrefix = "cdc_"
	}
	if c.MaxRetries <= 0 {
		c.MaxRetries = 3
	}
	if c.RetryBaseMs <= 0 {
		c.RetryBaseMs = 100
	}

	return c, nil
}

// toSourceProto converts internal SourceConfig to proto SourceConfig.
func toSourceProto(c *config.SourceConfig) *cdcpb.SourceConfig {
	return &cdcpb.SourceConfig{
		Type:            c.Type,
		Host:            c.Host,
		Port:            int32(c.Port),
		Username:        &c.Username,
		Password:        &c.Password,
		Database:        c.Database,
		Tables:          c.Tables,
		Topic:           &c.Topic,
		SlotName:        &c.SlotName,
		PublicationName: &c.PublicationName,
		InstanceId:      c.InstanceID,
		Name:            &c.Name,
	}
}

// toSinkProto converts internal SinkConfig to proto SinkConfig.
func toSinkProto(c *config.SinkConfig) *cdcpb.SinkConfig {
	return &cdcpb.SinkConfig{
		Type:            c.Type,
		Url:             c.URL,
		Username:        &c.Username,
		Password:        &c.Password,
		IndexPrefix:     &c.IndexPrefix,
		Index:           &c.Index,
		IndexMapping:    c.IndexMapping,
		BatchSize:       &c.BatchSize,
		FlushIntervalMs: &c.FlushIntervalMs,
		MaxRetries:      &c.MaxRetries,
		RetryBaseMs:     &c.RetryBaseMs,
		ApiKey:          &c.APIKey,
		Topic:           &c.Topic,
		InstanceId:      c.InstanceID,
		Name:            &c.Name,
		FieldMapping:    c.FieldMapping,
		Redis: &cdcpb.RedisSettings{
			Command:     c.Redis.Command,
			KeyTemplate: c.Redis.KeyTemplate,
			ValueFields: c.Redis.ValueFields,
			Ttl:         int32(c.Redis.TTL),
		},
	}
}

func fromPipelineConfig(p *cdcpb.PipelineConfig) config.PipelineConfig {
	if p == nil {
		return config.PipelineConfig{}
	}
	return config.PipelineConfig{
		ChannelBufferSize: int(p.ChannelBufferSize),
		WorkerCount:       int(p.WorkerCount),
		BatchSize:         int(p.BatchSize),
		FlushIntervalMs:   int(p.FlushIntervalMs),
		SubjectFilter:     p.SubjectFilter,
	}
}

func toPipelineProto(c config.PipelineConfig) *cdcpb.PipelineConfig {
	return &cdcpb.PipelineConfig{
		ChannelBufferSize: int32(c.ChannelBufferSize),
		WorkerCount:       int32(c.WorkerCount),
		BatchSize:         int32(c.BatchSize),
		FlushIntervalMs:   int32(c.FlushIntervalMs),
		SubjectFilter:     c.SubjectFilter,
	}
}

func fromNATSConfig(p *cdcpb.NATSConfig) config.NATSConfig {
	if p == nil {
		return config.NATSConfig{}
	}
	return config.NATSConfig{
		Enabled:            p.Enabled,
		URL:                p.Url,
		StreamName:         p.StreamName,
		RetentionDays:      p.RetentionDays,
		MaxReconnects:      int(p.MaxReconnects),
		ReconnectWaitMs:    int(p.ReconnectWaitMs),
		ReconnectBufferSizeMb: int(p.ReconnectBufferSizeMb),
		MaxAckPending:      int(p.MaxAckPending),
		AckWaitMs:          int(p.AckWaitMs),
		MaxDeliver:         int(p.MaxDeliver),
	}
}

func toNATSProto(c config.NATSConfig) *cdcpb.NATSConfig {
	return &cdcpb.NATSConfig{
		Enabled:            c.Enabled,
		Url:                c.URL,
		StreamName:         c.StreamName,
		RetentionDays:      c.RetentionDays,
		MaxReconnects:      int32(c.MaxReconnects),
		ReconnectWaitMs:    int32(c.ReconnectWaitMs),
		ReconnectBufferSizeMb: int32(c.ReconnectBufferSizeMb),
		MaxAckPending:      int32(c.MaxAckPending),
		AckWaitMs:          int32(c.AckWaitMs),
		MaxDeliver:         int32(c.MaxDeliver),
	}
}

func toPartitionProto(p, topic string) *cdcpb.PartitionSummary {
	return &cdcpb.PartitionSummary{
		Id:           p,
		MessageCount: 0,
		Topic:        topic,
	}
}
