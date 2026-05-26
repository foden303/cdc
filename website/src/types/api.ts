/**
 * TypeScript interfaces mirroring proto message definitions.
 * Single source of truth for all API response/request types.
 */

// ─── Health ──────────────────────────────────────────────────────────

export interface HealthCheckResponse {
  status: string;
  uptime?: number | string; // seconds since start; int64 may be encoded as string by gRPC gateway
  version?: string;
}

// ─── Source ──────────────────────────────────────────────────────────

export interface SourceConfig {
  type: 'postgres' | 'mysql';
  host: string;
  port: number;
  username?: string;
  password?: string;
  database: string;
  instance_id: string;
  name?: string;
}

export interface ListSourcesResponse {
  sources: SourceConfig[];
}

// ─── Sink ────────────────────────────────────────────────────────────

export interface SinkConfig {
  type: 'postgres' | 'mysql' | 'elasticsearch' | 'clickhouse';
  host?: string;
  port?: number;
  username?: string;
  password?: string;
  database?: string;
  instance_id: string;
  name?: string;
  url?: string[];       // elasticsearch
  api_key?: string;     // elasticsearch
  index_prefix?: string; // elasticsearch
}

export interface ListSinksResponse {
  sinks: SinkConfig[];
}

// ─── Flow ────────────────────────────────────────────────────────────

export type FlowStatus = 'FLOW_STATUS_RUNNING' | 'FLOW_STATUS_PAUSED' | 'FLOW_STATUS_ERROR' | 'FLOW_STATUS_UNSPECIFIED';

export interface FlowOptions {
  batch_size: number;
  flush_interval_ms: number;
  filter_expression: string;
  partition_count: number;
}

export interface ColumnMapping {
  source_column: string;
  sink_column: string;
  source_type: string;
  sink_type: string;
  enabled: boolean;
}

export interface FlowConfig {
  flow_id: string;
  name: string;
  source_id: string;
  sink_id: string;
  source_table: string;
  sink_table: string;
  status: FlowStatus;
  created_at: number;
  updated_at: number;
  options?: FlowOptions;
  column_mappings?: ColumnMapping[];
}

export interface ListFlowsResponse {
  flows: FlowConfig[];
}

export interface CreateFlowRequest {
  name?: string;
  source_id: string;
  sink_id: string;
  source_table: string;
  sink_table: string;
  options?: Partial<FlowOptions>;
  column_mappings?: ColumnMapping[];
}

export interface CreateFlowResponse {
  flow_id: string;
  status: FlowStatus;
}

export interface GetFlowStatsResponse {
  events_per_second: number;
  replication_lag_ms: number;
  total_events_processed: number;
  running_workers: number;
  pool_capacity: number;
  worker_utilization: number;
  failure_count: number;
  dlq_count: number;
  filtered_count: number;
  last_error: string;
}

// ─── Dashboard Aggregates ────────────────────────────────────────────

export interface DashboardSystemInventoryResponse {
  sources_count: number;
  sinks_count: number;
  flows_count: number;
}

export interface DashboardLiveTelemetryResponse {
  throughput: number;
  latency_p99: number;
  active_workers: number;
  channel_utilization: number;
  nats_healthy: boolean;
  error_rate: number;
  total_synced_events: number;
  failure_count: number;
}

export interface DashboardSummaryResponse {
  inventory?: DashboardSystemInventoryResponse;
  telemetry?: DashboardLiveTelemetryResponse;
}

// ─── Component Stats ─────────────────────────────────────────────────

/** Stats for a single source or sink component. */
export interface ComponentStats {
  success_count: number;
  failure_count: number;
  last_error: string;
  partition_lag: Record<number, number>;
  last_event_at: number;
  active_flows: number;
  throughput: number;
  error_rate: number;
  avg_latency_ms: number;
}

/** Aggregated stats response keyed by component instance ID. */
export interface GetStatsResponse {
  source_stats: Record<string, ComponentStats>;
  sink_stats: Record<string, ComponentStats>;
}

// ─── Connection Testing ──────────────────────────────────────────────

export interface TestConnectionResponse {
  success: boolean;
  message: string;
  latency_ms?: number;
  latencyMs?: number;
}

// ─── Table Discovery ─────────────────────────────────────────────────

export interface ColumnInfo {
  name: string;
  type: string;
  is_primary_key: boolean;
  is_nullable: boolean;
}

export interface TableInfo {
  schema: string;
  name: string;
  columns: ColumnInfo[];
}

export interface DiscoverTablesResponse {
  tables: TableInfo[];
}

// ─── DLQ ─────────────────────────────────────────────────────────────

export interface ReprocessDLQResponse {
  count: number;
}

// ─── Explorer ────────────────────────────────────────────────────────

export interface TopicSummary {
  name: string;
  message_count?: number;
  partition_count?: number;
}

export interface PartitionSummary {
  id: string;
  message_count?: number;
  topic: string;
}

export interface MessageItem {
  sequence: number;
  timestamp: string | number;
  subject: string;
  data: string;
  headers: Record<string, string>;
}

export interface DLQMessage extends MessageItem {
  reason?: string;
  original_subject?: string;
}

export interface ConsumerSummary {
  name: string;
  filter_subjects?: string[];
  num_pending?: number;
  num_ack_pending?: number;
  delivered_stream_seq?: number;
  ack_floor_stream_seq?: number;
}

export interface OffsetPaginationResponse {
  total_rows: number;
  page: number;
  limit: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface ListTopicsResponse {
  data: TopicSummary[];
  pagination?: OffsetPaginationResponse;
}

export interface ListPartitionsResponse {
  data: PartitionSummary[];
  pagination?: OffsetPaginationResponse;
}

export interface ListMessagesResponse {
  data: MessageItem[];
  total_count: number;
  pagination?: OffsetPaginationResponse;
}

export interface ListDLQMessagesResponse {
  data: DLQMessage[];
  pagination?: OffsetPaginationResponse;
}

export interface ListConsumersResponse {
  data: ConsumerSummary[];
  pagination?: OffsetPaginationResponse;
}
