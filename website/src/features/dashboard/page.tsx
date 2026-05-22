import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Activity,
  Timer,
  AlertTriangle,
  MailWarning,
  Database,
  HardDrive,
  GitBranch,
  Zap,
} from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { MetricCard } from '@/components/shared/MetricCard';
import { StatusBadge } from '@/components/shared/StatusBadge';
import { ThroughputChart } from './components/ThroughputChart';
import { SystemHealthBar } from './components/SystemHealthBar';
import { ROUTES } from '@/config/routes';
import {
  useHealth,
  useStats,
  usePerformance,
  useFlows,
  useConfig,
  dashboardKeys,
} from '@/lib/query/dashboard';
import { api } from '@/lib/api/client';
import { ENDPOINTS } from '@/lib/api/endpoints';
import { formatNumber, formatDuration, formatPercent } from '@/lib/format';
import type { ReprocessDLQResponse } from '@/types/api';

/** Dashboard page — single overview of all system metrics (pure telemetries). */
export default function DashboardPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: health, isLoading: healthLoading } = useHealth();
  const { data: stats, isLoading: statsLoading } = useStats();
  const { data: perf, isLoading: perfLoading } = usePerformance();
  const { data: flows, isLoading: flowsLoading } = useFlows();
  const { data: config, isLoading: configLoading } = useConfig();

  // DLQ reprocess mutation
  const dlqMutation = useMutation({
    mutationFn: () => api.post<ReprocessDLQResponse>(ENDPOINTS.dlqReprocess),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: dashboardKeys.stats });
    },
  });

  // Compute total DLQ count from all sink failure_counts
  const dlqCount = useMemo(() => {
    if (!stats?.sink_stats) return 0;
    return Object.values(stats.sink_stats).reduce(
      (sum, s) => sum + (s.failure_count || 0),
      0,
    );
  }, [stats]);

  // Compute total data synced (success count from all sinks)
  const totalSynced = useMemo(() => {
    if (!stats?.sink_stats) return 0;
    return Object.values(stats.sink_stats).reduce(
      (sum, s) => sum + (s.success_count || 0),
      0,
    );
  }, [stats]);

  const sourcesCount = config?.config?.sources?.length ?? 0;
  const sinksCount = config?.config?.sinks?.length ?? 0;
  const flowsCount = flows?.flows?.length ?? 0;

  const isLoading = healthLoading || statsLoading || perfLoading || configLoading || flowsLoading;

  return (
    <div className="space-y-6">
      {/* Header: System Status */}
      <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:gap-4">
        <h1 className="text-2xl font-bold tracking-tight text-foreground">
          {t('dashboard.title')}
        </h1>
        {health ? (
          <div className="flex items-center gap-3 text-xs text-muted-foreground sm:mt-1">
            <StatusBadge
              status={health.status === 'ok' ? 'healthy' : 'unhealthy'}
            />
            <span className="font-mono bg-accent/30 text-accent-foreground px-1.5 py-0.5 rounded text-[10px] font-semibold">
              v{health.version}
            </span>
            <span className="flex items-center gap-1">
              {t('dashboard.uptime')}: <span className="font-semibold text-foreground">{formatDuration(health.uptime)}</span>
            </span>
          </div>
        ) : (
          healthLoading && <Skeleton className="h-5 w-48 sm:mt-1" />
        )}
      </div>

      {/* Row 1: System Scale / Inventory Cards */}
      <div className="space-y-2">
        <h2 className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground/80">
          {t('dashboard.systemInventory')}
        </h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {isLoading ? (
            Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-[108px]" />
            ))
          ) : (
            <>
              <MetricCard
                title={t('dashboard.activeSources')}
                value={sourcesCount}
                icon={Database}
                iconClassName="bg-blue-500/10 text-blue-500"
                onClick={() => navigate(ROUTES.MANAGER_SOURCES)}
              />
              <MetricCard
                title={t('dashboard.activeSinks')}
                value={sinksCount}
                icon={HardDrive}
                iconClassName="bg-indigo-500/10 text-indigo-500"
                onClick={() => navigate(ROUTES.MANAGER_SINKS)}
              />
              <MetricCard
                title={t('dashboard.activeFlows')}
                value={flowsCount}
                icon={GitBranch}
                iconClassName="bg-amber-500/10 text-amber-500"
                onClick={() => navigate(ROUTES.MANAGER_FLOWS)}
              />
              <MetricCard
                title={t('dashboard.totalSyncedEvents')}
                value={formatNumber(totalSynced)}
                icon={Zap}
                iconClassName="bg-emerald-500/10 text-emerald-500"
              />
            </>
          )}
        </div>
      </div>

      {/* Row 2: Live Traffic / Performance Metrics */}
      <div className="space-y-2">
        <h2 className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground/80">
          {t('dashboard.liveTelemetry')}
        </h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {isLoading ? (
            Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-[108px]" />
            ))
          ) : (
            <>
              <MetricCard
                title={t('dashboard.throughput')}
                value={formatNumber(perf?.throughput ?? 0)}
                unit={t('dashboard.eventsPerSec')}
                icon={Activity}
                iconClassName="bg-cyan-500/10 text-cyan-500"
              />
              <MetricCard
                title={t('dashboard.latency')}
                value={(perf?.latency_p99 ?? 0).toFixed(1)}
                unit={t('dashboard.ms')}
                icon={Timer}
                iconClassName="bg-violet-500/10 text-violet-500"
              />
              <MetricCard
                title={t('dashboard.errorRate')}
                value={formatPercent(perf?.error_rate ?? 0)}
                icon={AlertTriangle}
                iconClassName="bg-yellow-500/10 text-yellow-500"
              />
              <MetricCard
                title={t('dashboard.dlqCount')}
                value={formatNumber(dlqCount)}
                icon={MailWarning}
                iconClassName="bg-red-500/10 text-red-500"
              />
            </>
          )}
        </div>
      </div>

      {/* Throughput Chart */}
      <ThroughputChart currentValue={perf?.throughput} />

      {/* System Health */}
      <SystemHealthBar
        natsConnected={health?.status === 'ok'}
        channelUtilPercent={0}
        activeWorkers={perf?.active_workers ?? 0}
        dlqCount={dlqCount}
        onReprocessDlq={() => dlqMutation.mutate()}
        isReprocessing={dlqMutation.isPending}
      />
    </div>
  );
}
