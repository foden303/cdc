/**
 * Frontend mirror of the backend's type compatibility matrix (pkg/flow/typecompat.go).
 */

export const TypeCompatibilityMatrix: Record<string, Record<string, string[]>> = {
  postgres: {
    integer: ['integer', 'bigint', 'numeric', 'text'],
    bigint: ['bigint', 'numeric', 'text'],
    smallint: ['smallint', 'integer', 'bigint', 'numeric', 'text'],
    numeric: ['numeric', 'text'],
    real: ['real', 'double precision', 'numeric', 'text'],
    'double precision': ['double precision', 'numeric', 'text'],
    boolean: ['boolean', 'integer', 'text'],
    text: ['text', 'character varying'],
    'character varying': ['text', 'character varying'],
    uuid: ['uuid', 'text', 'character varying'],
    'timestamp without time zone': ['timestamp without time zone', 'timestamp with time zone', 'text'],
    'timestamp with time zone': ['timestamp with time zone', 'text'],
    date: ['date', 'timestamp without time zone', 'text'],
    jsonb: ['jsonb', 'json', 'text'],
    json: ['json', 'jsonb', 'text'],
    bytea: ['bytea', 'text'],
  },
  clickhouse: {
    integer: ['Int32', 'Int64', 'UInt32', 'Float64', 'String'],
    bigint: ['Int64', 'UInt64', 'Float64', 'String'],
    smallint: ['Int16', 'Int32', 'Int64', 'String'],
    numeric: ['Decimal', 'Float64', 'String'],
    real: ['Float32', 'Float64', 'String'],
    'double precision': ['Float64', 'String'],
    boolean: ['Bool', 'UInt8', 'String'],
    text: ['String', 'FixedString'],
    'character varying': ['String', 'FixedString'],
    uuid: ['UUID', 'String'],
    'timestamp without time zone': ['DateTime', 'DateTime64', 'String'],
    'timestamp with time zone': ['DateTime64', 'String'],
    date: ['Date', 'Date32', 'DateTime', 'String'],
    jsonb: ['String'],
    json: ['String'],
  },
};

/**
 * Checks if a source type is compatible with a sink type for a given sink connector type.
 * If the sink type is pass-through (e.g. elasticsearch, stdout, webhook), it returns true.
 */
export function isTypeCompatible(
  sinkTypeConnector: string,
  sourceType: string,
  sinkType: string
): boolean {
  // Normalize connector names to lowercase
  const conn = sinkTypeConnector.toLowerCase();
  
  // Pass-through sinks accept all types (anything other than postgres & clickhouse is pass-through)
  if (conn !== 'postgres' && conn !== 'clickhouse') {
    return true;
  }

  const matrix = TypeCompatibilityMatrix[conn];
  if (!matrix) return false;

  // Let's normalize source type for comparison
  const normalizedSrc = sourceType.toLowerCase();
  const compatibleList = matrix[normalizedSrc];
  
  if (!compatibleList) {
    // If not explicitly declared in matrix, default to false for strict sinks,
    // or let's be graceful if exact match
    return normalizedSrc === sinkType.toLowerCase();
  }

  // Case-insensitive comparison with the list
  const lowerSink = sinkType.toLowerCase();
  return compatibleList.some(t => t.toLowerCase() === lowerSink);
}
