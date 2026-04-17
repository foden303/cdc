/**
 * Single place to register source/sink types for the Manager UI.
 * Add a row here when the backend supports a new connector type.
 */

export interface ConnectorTypeOption {
  value: string;
  label: string;
}

export const SOURCE_TYPE_OPTIONS: ConnectorTypeOption[] = [
  { value: "postgres", label: "PostgreSQL" },
  { value: "mysql", label: "MySQL" },
];

export const SINK_TYPE_OPTIONS: ConnectorTypeOption[] = [
  { value: "clickhouse", label: "Clickhouse" },
  { value: "elasticsearch", label: "Elastic" },
  { value: "redis", label: "Redis" },
];

export const DEFAULT_SOURCE_TYPE = SOURCE_TYPE_OPTIONS[0]?.value ?? "postgres";
export const DEFAULT_SINK_TYPE = SINK_TYPE_OPTIONS[0]?.value ?? "clickhouse";

export function getSourcePortPlaceholder(type: string): string {
  return type === "mysql" ? "3306" : "5432";
}

export function sinkShowsElasticsearchFields(type: string): boolean {
  return type === "elasticsearch";
}

export function sinkShowsRedisFields(type: string): boolean {
  return type === "redis";
}
