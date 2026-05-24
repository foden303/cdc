package server

import (
	cdcpb "github.com/foden/cdc/api/proto/v1"
	"github.com/foden/cdc/pkg/interfaces"
)

func protoToSourceConfig(pb *cdcpb.SourceConfig) *interfaces.SourceConfig {
	if pb == nil {
		return nil
	}
	cfg := &interfaces.SourceConfig{
		InstanceID: pb.InstanceId,
		Type:       pb.Type,
		Host:       pb.Host,
		Port:       int(pb.Port),
		Database:   pb.Database,
	}
	if pb.Username != nil {
		cfg.Username = *pb.Username
	}
	if pb.Password != nil {
		cfg.Password = *pb.Password
	}
	if pb.Name != nil {
		cfg.Name = *pb.Name
	}
	return cfg
}

func sourceConfigToProto(cfg *interfaces.SourceConfig) *cdcpb.SourceConfig {
	if cfg == nil {
		return nil
	}
	pb := &cdcpb.SourceConfig{
		InstanceId: cfg.InstanceID,
		Type:       cfg.Type,
		Host:       cfg.Host,
		Port:       int32(cfg.Port),
		Database:   cfg.Database,
	}
	if cfg.Username != "" {
		pb.Username = &cfg.Username
	}
	if cfg.Password != "" {
		pb.Password = &cfg.Password
	}
	if cfg.Name != "" {
		pb.Name = &cfg.Name
	}
	return pb
}

func protoToSinkConfig(pb *cdcpb.SinkConfig) *interfaces.SinkConfig {
	if pb == nil {
		return nil
	}
	cfg := &interfaces.SinkConfig{
		InstanceID: pb.InstanceId,
		Type:       pb.Type,
	}
	if pb.Host != nil {
		cfg.Host = *pb.Host
	}
	if pb.Port != nil {
		cfg.Port = int(*pb.Port)
	}
	if pb.Username != nil {
		cfg.Username = *pb.Username
	}
	if pb.Password != nil {
		cfg.Password = *pb.Password
	}
	if pb.Database != nil {
		cfg.Database = *pb.Database
	}
	if pb.Name != nil {
		cfg.Name = *pb.Name
	}
	if pb.ApiKey != nil {
		cfg.APIKey = *pb.ApiKey
	}
	if pb.IndexPrefix != nil {
		cfg.IndexPrefix = *pb.IndexPrefix
	}
	if pb.MaxRetries != nil {
		cfg.MaxRetries = *pb.MaxRetries
	}
	cfg.URL = pb.Url
	return cfg
}

func sinkConfigToProto(cfg *interfaces.SinkConfig) *cdcpb.SinkConfig {
	if cfg == nil {
		return nil
	}
	pb := &cdcpb.SinkConfig{
		InstanceId: cfg.InstanceID,
		Type:       cfg.Type,
		Url:        cfg.URL,
	}
	if cfg.Host != "" {
		pb.Host = &cfg.Host
	}
	if cfg.Port != 0 {
		port := int32(cfg.Port)
		pb.Port = &port
	}
	if cfg.Username != "" {
		pb.Username = &cfg.Username
	}
	if cfg.Password != "" {
		pb.Password = &cfg.Password
	}
	if cfg.Database != "" {
		pb.Database = &cfg.Database
	}
	if cfg.Name != "" {
		pb.Name = &cfg.Name
	}
	if cfg.APIKey != "" {
		pb.ApiKey = &cfg.APIKey
	}
	if cfg.IndexPrefix != "" {
		pb.IndexPrefix = &cfg.IndexPrefix
	}
	if cfg.MaxRetries != 0 {
		pb.MaxRetries = &cfg.MaxRetries
	}
	return pb
}

func flowConfigToProto(cfg *interfaces.FlowConfig) *cdcpb.FlowConfig {
	if cfg == nil {
		return nil
	}
	pb := &cdcpb.FlowConfig{
		FlowId:      cfg.FlowID,
		Name:        cfg.Name,
		SourceId:    cfg.SourceID,
		SinkId:      cfg.SinkID,
		SourceTable: cfg.SourceTable,
		SinkTable:   cfg.SinkTable,
		Status:      flowStatusToProto(cfg.Status),
		CreatedAt:   cfg.CreatedAt,
		UpdatedAt:   cfg.UpdatedAt,
	}

	if cfg.Options != nil {
		pb.Options = &cdcpb.FlowOptions{
			BatchSize:        cfg.Options.BatchSize,
			FlushIntervalMs:  cfg.Options.FlushIntervalMs,
			FilterExpression: cfg.Options.FilterExpression,
			PartitionCount:   int32(cfg.Options.PartitionCount),
		}
	}

	if len(cfg.ColumnMappings) > 0 {
		pb.ColumnMappings = make([]*cdcpb.ColumnMapping, len(cfg.ColumnMappings))
		for i, cm := range cfg.ColumnMappings {
			pb.ColumnMappings[i] = &cdcpb.ColumnMapping{
				SourceColumn: cm.SourceColumn,
				SinkColumn:   cm.SinkColumn,
				SourceType:   cm.SourceType,
				SinkType:     cm.SinkType,
				Enabled:      cm.Enabled,
			}
		}
	}

	return pb
}

func flowStatusToProto(s interfaces.FlowStatus) cdcpb.FlowStatus {
	switch s {
	case interfaces.FlowStatusRunning:
		return cdcpb.FlowStatus_FLOW_STATUS_RUNNING
	case interfaces.FlowStatusPaused:
		return cdcpb.FlowStatus_FLOW_STATUS_PAUSED
	case interfaces.FlowStatusError:
		return cdcpb.FlowStatus_FLOW_STATUS_ERROR
	default:
		return cdcpb.FlowStatus_FLOW_STATUS_UNSPECIFIED
	}
}
