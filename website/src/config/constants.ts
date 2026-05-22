/** Polling intervals in milliseconds — used by TanStack Query refetchInterval. */
export const POLLING = {
  HEALTH: 30_000,
  STATS: 5_000,
  PERFORMANCE: 2_000,
  FLOWS: 10_000,
  FLOW_STATS: 3_000,
  TOPICS: 10_000,
  PARTITIONS: 5_000,
  MESSAGES: 0, // manual refresh only
} as const;

/** Default pagination sizes. */
export const PAGE_SIZES = [25, 50, 100] as const;
export const DEFAULT_PAGE_SIZE = 25;

/** Chart data points to keep in memory for rolling charts. */
export const CHART_MAX_POINTS = 60;

/** Debounce delay for search inputs (ms). */
export const SEARCH_DEBOUNCE_MS = 300;
