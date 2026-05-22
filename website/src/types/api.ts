/**
 * TypeScript interfaces mirroring proto message definitions.
 * Single source of truth for all API response/request types.
 */

// ─── Health ──────────────────────────────────────────────────────────

export interface HealthCheckResponse {
  status: string;
  version: string;
  uptime: number; // seconds since start
}

// ─── Config ──────────────────────────────────────────────────────────

export interface SourceConfig {
  type: string;
  host: string;
  port: number;
  username?: string;
  password?: string;
  database: string;
  tables: string[];
  slot_name?: string;
  publication_name?: string;
  instance_id: string;
  name?: string;
  topic?: string;
  url?: string;
  headers?: Record<string, string>;
  polling_interval_ms?: number;
  snapshot_mode?: string;
}

export interface SinkConfig {
  type: string;
  url: string[];
  username?: string;
  password?: string;
  index_prefix?: string;
  index?: string;
  index_mapping?: Record<string, string>;
  batch_size?: number;
  flush_interval_ms?: number;
  max_retries?: number;
  retry_base_ms?: number;
  api_key?: string;
  instance_id: string;
  name?: string;
  topic?: string;
  field_mapping?: Record<string, string>;
  host?: string;
  port?: number;
  database?: string;
}

export interface PipelineConfig {
  channel_buffer_size: number;
  worker_count: number;
  batch_size: number;
  flush_interval_ms: number;
  subject_filter: string[];
}

export interface NATSConfig {
  enabled: boolean;
  url: string;
  stream_name: string;
  retention_days: number;
  max_reconnects: number;
  reconnect_wait_ms: number;
  reconnect_buffer_size_mb: number;
  max_ack_pending: number;
  ack_wait_ms: number;
  max_deliver: number;
}

export interface AppConfig {
  name: string;
  log_mode: string;
  sources: SourceConfig[];
  sinks: SinkConfig[];
  pipeline: PipelineConfig;
  nats: NATSConfig;
}

export interface GetConfigResponse {
  config: AppConfig;
  available_sources: string[];
  available_sinks: string[];
}

// ─── Stats ───────────────────────────────────────────────────────────

export interface ComponentStats {
  success_count: number;
  failure_count: number;
  last_error: string;
  partition_lag: Record<number, number>;
}

export interface GetStatsResponse {
  source_stats: Record<string, ComponentStats>;
  sink_stats: Record<string, ComponentStats>;
}

// ─── Performance ─────────────────────────────────────────────────────

export interface SourcePerformance {
  source_id: string;
  throughput: number;
  error_rate: number;
}

export interface SinkPerformance {
  sink_id: string;
  throughput: number;
  avg_latency: number;
}

export interface GetPerformanceMetricsResponse {
  throughput: number;
  latency_p99: number;
  active_workers: number;
  error_rate: number;
  sinks: Record<string, SinkPerformance>;
  sources: Record<string, SourcePerformance>;
}

// ─── Explorer ────────────────────────────────────────────────────────

export interface TopicSummary {
  name: string;
  message_count: number;
  partition_count: number;
}

export interface PartitionSummary {
  id: string;
  message_count: number;
  topic: string;
}

export interface MessageItem {
  sequence: number;
  timestamp: string;
  subject: string;
  data: string; // base64 encoded
  headers: Record<string, string>;
}

export interface Sort {
  field: string;
  order: 'SORT_ORDER_ASC' | 'SORT_ORDER_DESC' | 'SORT_ORDER_UNSPECIFIED';
}

export interface OffsetPaginationRequest {
  limit: number;
  page: number;
  sort?: Sort[];
}

export interface OffsetPaginationResponse {
  total_rows: number;
  limit: number;
  page: number;
  has_next: boolean;
  has_prev: boolean;
  sort?: Sort[];
}

export interface ListTopicsResponse {
  data: TopicSummary[];
  pagination: OffsetPaginationResponse;
}

export interface ListPartitionsResponse {
  data: PartitionSummary[];
  pagination: OffsetPaginationResponse;
}

export interface ListMessagesResponse {
  data: MessageItem[];
  total_count: number;
  pagination: OffsetPaginationResponse;
}

// ─── Flows ───────────────────────────────────────────────────────────

export type FlowStatus = 'FLOW_STATUS_RUNNING' | 'FLOW_STATUS_PAUSED' | 'FLOW_STATUS_ERROR' | 'FLOW_STATUS_UNSPECIFIED';

export interface FlowOptions {
  batch_size: number;
  flush_interval_ms: number;
  filter_expression: string;
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
  status: FlowStatus;
  created_at: number;
  updated_at: number;
  options: FlowOptions;
  sink_table: string;
  column_mappings: ColumnMapping[];
}

export interface ListFlowsResponse {
  flows: FlowConfig[];
}

export interface GetFlowStatsResponse {
  events_per_second: number;
  replication_lag_ms: number;
  last_synced_at: number;
  total_events_processed: number;
}

export type TableSyncState = 'TABLE_SYNC_STATE_SYNCING' | 'TABLE_SYNC_STATE_COMPLETED' | 'TABLE_SYNC_STATE_ERROR' | 'TABLE_SYNC_STATE_UNSPECIFIED';

export interface TableProgress {
  table_name: string;
  state: TableSyncState;
  rows_synced: number;
  last_offset: string;
  error_message: string;
}

export interface GetFlowTableProgressResponse {
  tables: TableProgress[];
}

// ─── Connection Testing ──────────────────────────────────────────────

export interface TestConnectionResponse {
  success: boolean;
  message: string;
  latency_ms: number;
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
