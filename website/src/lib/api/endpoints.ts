/** All REST endpoints — mirrors the proto service definition. */
export const ENDPOINTS = {
  // System
  health: '/api/v1/health',
  stats: '/api/v1/stats',
  performance: '/api/v1/metrics/performance',
  dashboardInventory: '/api/v1/dashboard/system-inventory',
  dashboardLiveTelemetry: '/api/v1/dashboard/live-telemetry',
  dashboardThroughputOverTime: '/api/v1/dashboard/throughput-overtime',

  // Sources CRUD + Test
  sources: '/api/v1/sources',
  sourceById: (id: string) => `/api/v1/sources/${id}` as const,
  testSource: '/api/v1/test/source',
  discoverTables: (id: string) => `/api/v1/discover/tables/${id}` as const,

  // Sinks CRUD + Test
  sinks: '/api/v1/sinks',
  sinkById: (id: string) => `/api/v1/sinks/${id}` as const,
  testSink: '/api/v1/test/sink',
  discoverSinkTables: (id: string) => `/api/v1/discover/sink-tables/${id}` as const,

  // Flows CRUD + Lifecycle
  flows: '/api/v1/flows',
  flowById: (id: string) => `/api/v1/flows/${id}` as const,
  flowPause: (id: string) => `/api/v1/flows/${id}/pause` as const,
  flowResume: (id: string) => `/api/v1/flows/${id}/resume` as const,
  flowStats: (id: string) => `/api/v1/flows/${id}/stats` as const,
  flowProgress: (id: string) => `/api/v1/flows/${id}/progress` as const,

  // Explorer
  topics: '/api/v1/topics',
  consumers: '/api/v1/consumers',
  partitions: '/api/v1/partitions',
  messages: '/api/v1/messages',
  consumer: '/api/v1/consumer',

  // DLQ
  dlqMessages: '/api/v1/dlq/messages',
  dlqReprocess: '/api/v1/dlq/reprocess',
} as const;
