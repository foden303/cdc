/** All REST endpoints — mirrors the proto service definition. */
export const ENDPOINTS = {
  health: '/api/v1/health',
  config: '/api/v1/config',
  stats: '/api/v1/stats',
  performance: '/api/v1/metrics/performance',

  // Sources
  sources: '/api/v1/sources',
  sourcesTest: '/api/v1/sources/test',
  sourceById: (id: string) => `/api/v1/sources/${id}` as const,
  sourceTables: (id: string) => `/api/v1/sources/${id}/tables` as const,

  // Sinks
  sinks: '/api/v1/sinks',
  sinksTest: '/api/v1/sinks/test',
  sinkById: (id: string) => `/api/v1/sinks/${id}` as const,
  sinkTables: (id: string) => `/api/v1/sinks/${id}/tables` as const,

  // Explorer
  topics: '/api/v1/topics',
  partitions: '/api/v1/partitions',
  messages: '/api/v1/messages',
  consumer: '/api/v1/consumer',

  // Flows
  flows: '/api/v1/flows',
  flowById: (id: string) => `/api/v1/flows/${id}` as const,
  flowPause: (id: string) => `/api/v1/flows/${id}/pause` as const,
  flowResume: (id: string) => `/api/v1/flows/${id}/resume` as const,
  flowStats: (id: string) => `/api/v1/flows/${id}/stats` as const,
  flowTables: (id: string) => `/api/v1/flows/${id}/tables` as const,

  // DLQ
  dlqReprocess: '/api/v1/dlq/reprocess',
} as const;
