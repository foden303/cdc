/** API configuration — single source of truth for all connection settings. */
export const API_CONFIG = {
  baseURL: import.meta.env.VITE_API_URL || '',
  timeout: 10_000,
} as const;
