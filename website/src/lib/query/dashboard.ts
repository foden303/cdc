import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api/client';
import { ENDPOINTS } from '@/lib/api/endpoints';
import { POLLING } from '@/config/constants';
import type {
  HealthCheckResponse,
  DashboardSummaryResponse,
} from '@/types/api';

/** Query key factory for dashboard queries. */
export const dashboardKeys = {
  health: ['health'] as const,
  summary: ['dashboard', 'summary'] as const,
};

/** Fetches health check data with 30s polling. */
export function useHealth() {
  return useQuery({
    queryKey: dashboardKeys.health,
    queryFn: () => api.get<HealthCheckResponse>(ENDPOINTS.health),
    refetchInterval: POLLING.HEALTH,
    placeholderData: keepPreviousData,
  });
}

/** Fetches inventory and live telemetry values for the dashboard. */
export function useDashboardSummary() {
  return useQuery({
    queryKey: dashboardKeys.summary,
    queryFn: () => api.get<DashboardSummaryResponse>(ENDPOINTS.dashboard),
    refetchInterval: POLLING.PERFORMANCE,
    placeholderData: keepPreviousData,
  });
}
