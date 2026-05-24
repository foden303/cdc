import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api/client';
import { ENDPOINTS } from '@/lib/api/endpoints';
import { POLLING } from '@/config/constants';
import type {
  HealthCheckResponse,
  DashboardSystemInventoryResponse,
  DashboardLiveTelemetryResponse,
  DashboardThroughputOverTimeResponse,
} from '@/types/api';

/** Query key factory for dashboard queries. */
export const dashboardKeys = {
  health: ['health'] as const,
  inventory: ['dashboard', 'inventory'] as const,
  telemetry: ['dashboard', 'telemetry'] as const,
  throughputOverTime: ['dashboard', 'throughputOverTime'] as const,
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

/** Fetches inventory counts for the dashboard inventory row. */
export function useSystemInventory() {
  return useQuery({
    queryKey: dashboardKeys.inventory,
    queryFn: () =>
      api.get<DashboardSystemInventoryResponse>(ENDPOINTS.dashboardInventory),
    refetchInterval: POLLING.INVENTORY,
    placeholderData: keepPreviousData,
  });
}

/** Fetches live telemetry values for cards and throughput chart. */
export function useLiveTelemetry() {
  return useQuery({
    queryKey: dashboardKeys.telemetry,
    queryFn: () =>
      api.get<DashboardLiveTelemetryResponse>(ENDPOINTS.dashboardLiveTelemetry),
    refetchInterval: POLLING.PERFORMANCE,
    placeholderData: keepPreviousData,
  });
}

/** Fetches rolling throughput points for the dashboard chart. */
export function useThroughputOverTime() {
  return useQuery({
    queryKey: dashboardKeys.throughputOverTime,
    queryFn: () =>
      api.get<DashboardThroughputOverTimeResponse>(
        ENDPOINTS.dashboardThroughputOverTime,
      ),
    refetchInterval: POLLING.PERFORMANCE,
    placeholderData: keepPreviousData,
  });
}
