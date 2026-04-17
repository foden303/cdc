package server

import (
	"errors"
	"fmt"
	"strings"

	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/foden/cdc/pkg/config"
	"github.com/foden/cdc/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrSourceConfigRequired     = errors.New("source config is required")
	ErrSourceTypeRequired       = errors.New("source type is required")
	ErrSourceHostRequired       = errors.New("source host is required")
	ErrSourcePortRequired       = errors.New("source port must be positive")
	ErrSourceDatabaseRequired   = errors.New("source database is required")
	ErrSourceURLRequired        = errors.New("source url is required")
	ErrSinkConfigRequired       = errors.New("sink config is required")
	ErrSinkTypeRequired         = errors.New("sink type is required")
	ErrSinkURLRequired          = errors.New("sink URL is required")
	ErrSourceInstanceIDRequired = errors.New("source instance_id is required")
	ErrSinkInstanceIDRequired   = errors.New("sink instance_id is required")
)

// toSourceConfig converts proto SourceConfig to internal SourceConfig with validation and defaults.
func toSourceConfig(p *cdcpb.SourceConfig) (*config.SourceConfig, error) {
	if p == nil {
		return nil, ErrSourceConfigRequired
	}

	if p.Type == "" {
		return nil, ErrSourceTypeRequired
	}

	c := &config.SourceConfig{
		Type:              p.Type,
		Host:              p.Host,
		Port:              int(p.Port),
		Database:          p.Database,
		Tables:            p.Tables,
		Topic:             p.GetTopic(),
		Username:          p.GetUsername(),
		Password:          p.GetPassword(),
		SlotName:          p.GetSlotName(),
		PublicationName:   p.GetPublicationName(),
		InstanceID:        p.GetInstanceId(),
		Name:              p.GetName(),
		URL:               p.GetUrl(),
		Headers:           p.Headers,
		PollingIntervalMs: int(p.GetPollingIntervalMs()),
		SnapshotMode:      p.GetSnapshotMode(),
	}

	switch c.Type {
	case "rest":
		if c.URL == "" {
			return nil, ErrSourceURLRequired
		}
		if c.PollingIntervalMs <= 0 {
			c.PollingIntervalMs = 5000
		}
	default:
		if c.Host == "" {
			return nil, ErrSourceHostRequired
		}
		if c.Port <= 0 {
			return nil, ErrSourcePortRequired
		}
		if c.Database == "" {
			return nil, ErrSourceDatabaseRequired
		}
	}

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

func requireSourceInstanceID(instanceID string) error {
	if instanceID == "" {
		return ErrSourceInstanceIDRequired
	}
	return nil
}

func requireSinkInstanceID(instanceID string) error {
	if instanceID == "" {
		return ErrSinkInstanceIDRequired
	}
	return nil
}

func isNotFoundError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "not found")
}

func isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "already exists") || strings.Contains(errText, "already connects")
}

func isInvalidConfigError(err error) bool {
	if err == nil {
		return false
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "invalid") || strings.Contains(errText, "unsupported")
}

func isStartupError(err error) bool {
	if err == nil {
		return false
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "start") || strings.Contains(errText, "consumer") || strings.Contains(errText, "connect")
}

func statusErrorForAction(err error, action string) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrSourceConfigRequired),
		errors.Is(err, ErrSourceTypeRequired),
		errors.Is(err, ErrSourceHostRequired),
		errors.Is(err, ErrSourcePortRequired),
		errors.Is(err, ErrSourceDatabaseRequired),
		errors.Is(err, ErrSourceURLRequired),
		errors.Is(err, ErrSinkConfigRequired),
		errors.Is(err, ErrSinkTypeRequired),
		errors.Is(err, ErrSinkURLRequired),
		errors.Is(err, ErrSourceInstanceIDRequired),
		errors.Is(err, ErrSinkInstanceIDRequired),
		isInvalidConfigError(err):
		return status.Errorf(codes.InvalidArgument, "%s: %v", action, err)
	case isAlreadyExistsError(err):
		return status.Errorf(codes.AlreadyExists, "%s: %v", action, err)
	case isNotFoundError(err):
		return status.Errorf(codes.NotFound, "%s: %v", action, err)
	case isStartupError(err):
		return status.Errorf(codes.FailedPrecondition, "%s: %v", action, err)
	default:
		return status.Errorf(codes.Internal, "%s: %v", action, err)
	}
}

func persistedConfigError(resource string, err error) error {
	return status.Errorf(codes.Internal, "persist %s config: %v", resource, err)
}

func successfulSourceProto(c *config.SourceConfig) *cdcpb.SourceConfig {
	return toSourceProto(c)
}

func successfulSinkProto(c *config.SinkConfig) *cdcpb.SinkConfig {
	return toSinkProto(c)
}

func normalizedConsumerName(consumerName string) string {
	return strings.TrimSpace(consumerName)
}

func defaultConsumerGroupPrefix() string {
	return "pipeline-"
}

func partitionConsumerName(instanceID string, partition int) string {
	return fmt.Sprintf("pipeline-%s-p%d", instanceID, partition)
}

func sinkTopicOrDefault(topic string) string {
	if topic != "" {
		return topic
	}
	return "cdc.>"
}

func sourceTopicOrDefault(topic string) string {
	if topic != "" {
		return topic
	}
	return ""
}

func sourceProtoPollingInterval(interval int) *int32 {
	if interval == 0 {
		return nil
	}
	v := int32(interval)
	return &v
}

func optionalStringPtr(v string) *string {
	return &v
}

func optionalInt32Ptr(v int32) *int32 {
	return &v
}

func sourceHeadersOrEmpty(headers map[string]string) map[string]string {
	if headers == nil {
		return map[string]string{}
	}
	return headers
}

func sinkFieldMappingOrEmpty(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}

func sinkIndexMappingOrEmpty(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}

func sourceTablesOrEmpty(tables []string) []string {
	if tables == nil {
		return []string{}
	}
	return tables
}

func sinkURLsOrEmpty(urls []string) []string {
	if urls == nil {
		return []string{}
	}
	return urls
}

func protoRedisSettings(c config.RedisSettings) *cdcpb.RedisSettings {
	return &cdcpb.RedisSettings{
		Command:     c.Command,
		KeyTemplate: c.KeyTemplate,
		ValueFields: c.ValueFields,
		Ttl:         int32(c.TTL),
	}
}

func configRedisSettings(p *cdcpb.RedisSettings) config.RedisSettings {
	if p == nil {
		return config.RedisSettings{}
	}
	return config.RedisSettings{
		Command:     p.Command,
		KeyTemplate: p.KeyTemplate,
		ValueFields: p.ValueFields,
		TTL:         int(p.Ttl),
	}
}

func sourceConfigFromProto(p *cdcpb.SourceConfig) *config.SourceConfig {
	return nil
}

func sinkConfigFromProto(p *cdcpb.SinkConfig) *config.SinkConfig {
	return nil
}

func sourceConfigToProto(c *config.SourceConfig) *cdcpb.SourceConfig {
	return nil
}

func sinkConfigToProto(c *config.SinkConfig) *cdcpb.SinkConfig {
	return nil
}

func sourceTypeNeedsDatabase(typeName string) bool {
	return typeName != "rest"
}

func sourceTypeNeedsHost(typeName string) bool {
	return typeName != "rest"
}

func sourceTypeNeedsPort(typeName string) bool {
	return typeName != "rest"
}

func sourceTypeNeedsURL(typeName string) bool {
	return typeName == "rest"
}

func sourceTypeDefaultPollingInterval(typeName string) int {
	if typeName == "rest" {
		return 5000
	}
	return 0
}

func sourceValidationError(typeName string) error {
	return status.Errorf(codes.InvalidArgument, "unsupported source config for type %s", typeName)
}

func sinkValidationError(typeName string) error {
	return status.Errorf(codes.InvalidArgument, "unsupported sink config for type %s", typeName)
}

func duplicateResourceError(resource, instanceID string) error {
	return status.Errorf(codes.AlreadyExists, "%s %s already exists", resource, instanceID)
}

func missingResourceError(resource, instanceID string) error {
	return status.Errorf(codes.NotFound, "%s %s not found", resource, instanceID)
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
		InstanceID:      p.GetInstanceId(),
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
		Type:              c.Type,
		Host:              c.Host,
		Port:              int32(c.Port),
		Username:          optionalStringPtr(c.Username),
		Password:          optionalStringPtr(c.Password),
		Database:          c.Database,
		Tables:            sourceTablesOrEmpty(c.Tables),
		SlotName:          optionalStringPtr(c.SlotName),
		PublicationName:   optionalStringPtr(c.PublicationName),
		InstanceId:        c.InstanceID,
		Name:              optionalStringPtr(c.Name),
		Topic:             optionalStringPtr(c.Topic),
		Url:               optionalStringPtr(c.URL),
		Headers:           sourceHeadersOrEmpty(c.Headers),
		PollingIntervalMs: sourceProtoPollingInterval(c.PollingIntervalMs),
		SnapshotMode:      optionalStringPtr(c.SnapshotMode),
	}
}

// toSinkProto converts internal SinkConfig to proto SinkConfig.
func toSinkProto(c *config.SinkConfig) *cdcpb.SinkConfig {
	return &cdcpb.SinkConfig{
		Type:            c.Type,
		Url:             sinkURLsOrEmpty(c.URL),
		Username:        optionalStringPtr(c.Username),
		Password:        optionalStringPtr(c.Password),
		IndexPrefix:     optionalStringPtr(c.IndexPrefix),
		Index:           optionalStringPtr(c.Index),
		IndexMapping:    sinkIndexMappingOrEmpty(c.IndexMapping),
		BatchSize:       optionalInt32Ptr(c.BatchSize),
		FlushIntervalMs: optionalInt32Ptr(c.FlushIntervalMs),
		MaxRetries:      optionalInt32Ptr(c.MaxRetries),
		RetryBaseMs:     optionalInt32Ptr(c.RetryBaseMs),
		ApiKey:          optionalStringPtr(c.APIKey),
		Topic:           optionalStringPtr(c.Topic),
		InstanceId:      c.InstanceID,
		Name:            optionalStringPtr(c.Name),
		FieldMapping:    sinkFieldMappingOrEmpty(c.FieldMapping),
		Redis:           protoRedisSettings(c.Redis),
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
