import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api/client';
import { ENDPOINTS } from '@/lib/api/endpoints';
import { POLLING } from '@/config/constants';
import type {
  HealthCheckResponse,
  GetStatsResponse,
  GetPerformanceMetricsResponse,
  ListFlowsResponse,
  GetConfigResponse,
} from '@/types/api';

/** Query key factory for dashboard queries. */
export const dashboardKeys = {
  health: ['health'] as const,
  stats: ['stats'] as const,
  performance: ['performance'] as const,
  flows: ['flows'] as const,
  config: ['config'] as const,
};

/** Fetches health check data with 30s polling. */
export function useHealth() {
  return useQuery({
    queryKey: dashboardKeys.health,
    queryFn: () => api.get<HealthCheckResponse>(ENDPOINTS.health),
    refetchInterval: POLLING.HEALTH,
  });
}

/** Fetches source/sink statistics with 5s polling. */
export function useStats() {
  return useQuery({
    queryKey: dashboardKeys.stats,
    queryFn: () => api.get<GetStatsResponse>(ENDPOINTS.stats),
    refetchInterval: POLLING.STATS,
  });
}

/** Fetches performance metrics (throughput, latency, workers) with 2s polling. */
export function usePerformance() {
  return useQuery({
    queryKey: dashboardKeys.performance,
    queryFn: () => api.get<GetPerformanceMetricsResponse>(ENDPOINTS.performance),
    refetchInterval: POLLING.PERFORMANCE,
  });
}

/** Fetches flows list with 10s polling. */
export function useFlows() {
  return useQuery({
    queryKey: dashboardKeys.flows,
    queryFn: () => api.get<ListFlowsResponse>(ENDPOINTS.flows),
    refetchInterval: POLLING.FLOWS,
  });
}

/** Fetches full system config. */
export function useConfig() {
  return useQuery({
    queryKey: dashboardKeys.config,
    queryFn: () => api.get<GetConfigResponse>(ENDPOINTS.config),
    refetchInterval: POLLING.FLOWS,
  });
}
