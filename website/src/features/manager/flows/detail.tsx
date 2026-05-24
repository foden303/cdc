import { type ComponentType, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  ArrowLeft,
  ArrowRight,
  Activity,
  Clock3,
  Columns3,
  Filter,
  FolderSync,
  GitCommit,
  History,
  Pause,
  Play,
  Route,
  SlidersHorizontal,
  Terminal,
  Trash2,
  TrendingUp,
} from 'lucide-react';
import { toast } from 'sonner';

import {
  useFlow,
  useFlowStats,
  useFlowProgress,
  useConfig,
  usePauseFlow,
  useResumeFlow,
  useDeleteFlow,
} from '@/lib/query/manager';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { StatusBadge, type Status } from '@/components/shared/StatusBadge';
import { ROUTES } from '@/config/routes';
import { formatNumber, formatDuration } from '@/lib/format';
import type { FlowStatus } from '@/types/api';

function timestampToDate(value?: number) {
  if (!value || value <= 0) return null;
  const date = new Date(value > 1_000_000_000_000 ? value : value * 1000);
  return Number.isNaN(date.getTime()) ? null : date;
}

function statusToBadge(status: FlowStatus | string): Status {
  if (status === 'FLOW_STATUS_RUNNING') return 'healthy';
  if (status === 'FLOW_STATUS_PAUSED') return 'paused';
  if (status === 'FLOW_STATUS_ERROR') return 'unhealthy';
  return 'idle';
}

function statusLabel(status: FlowStatus | string) {
  if (status === 'FLOW_STATUS_RUNNING') return 'Running';
  if (status === 'FLOW_STATUS_PAUSED') return 'Paused';
  if (status === 'FLOW_STATUS_ERROR') return 'Error';
  return 'Idle';
}

function progressStateToBadge(state?: string): Status {
  const normalized = state?.toLowerCase();
  if (normalized?.includes('completed') || normalized?.includes('synced')) return 'healthy';
  if (normalized?.includes('error') || normalized?.includes('failed')) return 'unhealthy';
  if (normalized?.includes('syncing') || normalized?.includes('running')) return 'healthy';
  if (normalized?.includes('pending')) return 'paused';
  return 'idle';
}

function progressStateLabel(state?: string) {
  const normalized = state?.toLowerCase();
  if (normalized?.includes('completed') || normalized?.includes('synced')) return 'completed';
  if (normalized?.includes('error') || normalized?.includes('failed')) return 'error';
  if (normalized?.includes('syncing') || normalized?.includes('running')) return 'syncing';
  return 'pending';
}

function RuntimeStatCard({
  title,
  value,
  description,
  icon: Icon,
  tone,
  loading,
}: {
  title: string;
  value: string;
  description: string;
  icon: ComponentType<{ className?: string }>;
  tone: 'sky' | 'amber' | 'indigo' | 'emerald';
  loading?: boolean;
}) {
  const toneClass = {
    sky: 'bg-sky-500/10 text-sky-500 dark:text-sky-400',
    amber: 'bg-amber-500/10 text-amber-500 dark:text-amber-400',
    indigo: 'bg-indigo-500/10 text-indigo-500 dark:text-indigo-400',
    emerald: 'bg-emerald-500/10 text-emerald-500 dark:text-emerald-400',
  }[tone];

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      {loading ? (
        <div className="space-y-3">
          <div className="h-4 w-24 animate-pulse rounded bg-muted" />
          <div className="h-7 w-28 animate-pulse rounded bg-muted" />
          <div className="h-3 w-32 animate-pulse rounded bg-muted" />
        </div>
      ) : (
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <p className="truncate text-xs font-medium text-muted-foreground">{title}</p>
            <p className="mt-2 truncate text-2xl font-semibold tracking-tight text-foreground">
              {value}
            </p>
            <p className="mt-1 truncate text-xs text-muted-foreground">{description}</p>
          </div>
          <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg ${toneClass}`}>
            <Icon className="h-5 w-5" />
          </div>
        </div>
      )}
    </div>
  );
}

export default function FlowDetailPage() {
  const { t } = useTranslation();
  const { id: flowId } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: flowData, isLoading: flowLoading } = useFlow(flowId || '');
  const { data: statsData, isLoading: statsLoading } = useFlowStats(flowId || '');
  const { data: progressData, isLoading: progressLoading } = useFlowProgress(flowId || '');
  const { data: configData } = useConfig();

  const pauseMutation = usePauseFlow();
  const resumeMutation = useResumeFlow();
  const deleteMutation = useDeleteFlow();

  const flow = flowData?.flow;

  const selectedSource = useMemo(() => {
    if (!flow || !configData?.config?.sources) return null;
    return configData.config.sources.find((s) => s.instance_id === flow.source_id) ?? null;
  }, [flow, configData]);

  const selectedSink = useMemo(() => {
    if (!flow || !configData?.config?.sinks) return null;
    return configData.config.sinks.find((s) => s.instance_id === flow.sink_id) ?? null;
  }, [flow, configData]);

  const lastSyncedAt = timestampToDate(statsData?.last_synced_at);
  const mappingCount = flow?.column_mappings?.filter((mapping) => mapping.enabled).length ?? 0;
  const totalMappings = flow?.column_mappings?.length ?? 0;
  const hasFilter = Boolean(flow?.options?.filter_expression);

  const handlePause = async () => {
    if (!flowId) return;
    try {
      await pauseMutation.mutateAsync(flowId);
      toast.success(t('manager.flows.toast.paused'));
    } catch {
      toast.error(t('manager.flows.toast.pauseFailed'));
    }
  };

  const handleResume = async () => {
    if (!flowId) return;
    try {
      await resumeMutation.mutateAsync(flowId);
      toast.success(t('manager.flows.toast.resumed'));
    } catch {
      toast.error(t('manager.flows.toast.resumeFailed'));
    }
  };

  const handleDelete = async () => {
    if (!flowId) return;
    if (confirm(t('manager.flows.confirm.delete'))) {
      try {
        await deleteMutation.mutateAsync(flowId);
        toast.success(t('manager.flows.toast.deleted'));
        navigate(ROUTES.MANAGER_FLOWS);
      } catch {
        toast.error(t('manager.flows.toast.deleteFailed'));
      }
    }
  };

  if (flowLoading || !flow) {
    return (
      <div className="space-y-6">
        <div className="h-16 animate-pulse rounded-lg bg-muted" />
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-32 animate-pulse rounded-lg bg-muted" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 border-b border-border pb-5 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex min-w-0 items-start gap-3">
          <button
            onClick={() => navigate(ROUTES.MANAGER_FLOWS)}
            className="mt-1 cursor-pointer rounded-lg border border-border p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            title={t('common.back')}
          >
            <ArrowLeft className="h-4 w-4" />
          </button>
          <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg border border-amber-500/20 bg-amber-500/10 text-amber-600 dark:text-amber-400">
            <FolderSync className="h-5 w-5" />
          </div>
          <div className="min-w-0">
            <p className="inline-flex max-w-full truncate rounded border border-border bg-muted/30 px-1.5 py-0.5 font-mono text-[10px] font-medium leading-3 text-muted-foreground">
              {flow.flow_id}
            </p>
            <div className="mt-1 flex min-w-0 flex-wrap items-center gap-3">
              <h1 className="truncate text-2xl font-bold tracking-tight text-foreground">
                {flow.name || flow.flow_id}
              </h1>
              <StatusBadge
                status={statusToBadge(flow.status)}
                label={t(`common.status.${statusToBadge(flow.status)}`, {
                  defaultValue: statusLabel(flow.status),
                })}
              />
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {flow.status === 'FLOW_STATUS_RUNNING' ? (
            <Button
              variant="outline"
              size="sm"
              onClick={handlePause}
              disabled={pauseMutation.isPending}
              className="h-9 cursor-pointer text-xs"
            >
              <Pause className="mr-1 h-3.5 w-3.5" />
              {t('manager.flows.pause')}
            </Button>
          ) : (
            <Button
              variant="outline"
              size="sm"
              onClick={handleResume}
              disabled={resumeMutation.isPending}
              className="h-9 cursor-pointer text-xs text-emerald-600 hover:text-emerald-500"
            >
              <Play className="mr-1 h-3.5 w-3.5" />
              {t('manager.flows.resume')}
            </Button>
          )}

          <Button
            variant="destructive"
            size="sm"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
            className="h-9 cursor-pointer text-xs font-semibold"
          >
            <Trash2 className="mr-1 h-3.5 w-3.5" />
            {t('manager.flows.delete')}
          </Button>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <RuntimeStatCard
          title={t('manager.flows.metrics.syncRate')}
          value={`${formatNumber(statsData?.events_per_second || 0)}/s`}
          description={t('manager.flows.metrics.syncRateDesc')}
          icon={TrendingUp}
          tone="sky"
          loading={statsLoading}
        />
        <RuntimeStatCard
          title={t('manager.flows.metrics.lag')}
          value={formatDuration((statsData?.replication_lag_ms || 0) / 1000)}
          description={t('manager.flows.metrics.lagDesc', {
            lag: (statsData?.replication_lag_ms || 0).toLocaleString(),
          })}
          icon={Activity}
          tone="amber"
          loading={statsLoading}
        />
        <RuntimeStatCard
          title={t('manager.flows.metrics.eventsSynced')}
          value={formatNumber(statsData?.total_events_processed || 0)}
          description={t('manager.flows.metrics.eventsSyncedDesc')}
          icon={GitCommit}
          tone="indigo"
          loading={statsLoading}
        />
        <RuntimeStatCard
          title={t('manager.flows.metrics.lastSynced')}
          value={lastSyncedAt ? lastSyncedAt.toLocaleTimeString() : '-'}
          description={lastSyncedAt
            ? lastSyncedAt.toLocaleDateString()
            : t('manager.flows.metrics.neverSynced')}
          icon={Clock3}
          tone="emerald"
          loading={statsLoading}
        />
      </div>

      <div className="rounded-lg border border-border bg-card p-4">
        <div className="mb-4 flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Route className="h-4 w-4 text-amber-500 dark:text-amber-400" />
            <h2 className="text-sm font-semibold text-foreground">
              {t('manager.flows.detail.pipelineOverview')}
            </h2>
          </div>
          <Badge variant="outline" className="border-border bg-muted/30 text-[10px] uppercase tracking-wide text-muted-foreground">
            {selectedSource?.type || 'source'} -&gt; {selectedSink?.type || 'sink'}
          </Badge>
        </div>

        <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] md:items-center">
          <div className="min-w-0 rounded-md border border-sky-500/15 bg-sky-500/[0.04] px-3 py-2.5">
            <p className="text-[10px] font-semibold uppercase tracking-wide text-sky-600 dark:text-sky-400">
              {t('manager.flows.detail.pipelineOverviewSource')}
            </p>
            <p className="mt-1 truncate text-sm font-semibold text-foreground">
              {selectedSource?.name || selectedSource?.database || t('nav.sources')}
            </p>
            <p className="mt-0.5 truncate font-mono text-xs text-muted-foreground">
              {flow.source_table || '-'}
            </p>
          </div>

          <div className="flex h-8 w-8 items-center justify-center justify-self-center rounded-full border border-amber-500/20 bg-amber-500/10 text-amber-600 dark:text-amber-400">
            <ArrowRight className="h-4 w-4" />
          </div>

          <div className="min-w-0 rounded-md border border-violet-500/15 bg-violet-500/[0.04] px-3 py-2.5 md:text-right">
            <p className="text-[10px] font-semibold uppercase tracking-wide text-violet-600 dark:text-violet-400">
              {t('manager.flows.detail.pipelineOverviewSink')}
            </p>
            <p className="mt-1 truncate text-sm font-semibold text-foreground">
              {selectedSink?.name || selectedSink?.database || t('nav.sinks')}
            </p>
            <p className="mt-0.5 truncate font-mono text-xs text-muted-foreground">
              {flow.sink_table || '-'}
            </p>
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <div className="rounded-lg border border-border bg-card p-4">
            <div className="mb-4 flex items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <Columns3 className="h-4 w-4 text-indigo-500 dark:text-indigo-400" />
                <h2 className="text-sm font-semibold text-foreground">
                  {t('manager.flows.detail.columnSchema')}
                </h2>
              </div>
              <Badge variant="outline" className="border-border bg-muted/30 text-xs text-muted-foreground">
                {mappingCount}/{totalMappings}
              </Badge>
            </div>

            {!flow.column_mappings?.length ? (
              <div className="rounded-lg border border-dashed border-border py-8 text-center text-xs text-muted-foreground">
                {t('manager.flows.detail.noMappings', { defaultValue: 'No column mappings configured.' })}
              </div>
            ) : (
              <div className="overflow-hidden rounded-lg border border-border">
                <Table>
                  <TableHeader className="bg-muted/40">
                    <TableRow className="border-b border-border hover:bg-transparent">
                      <TableHead className="h-auto px-4 py-2 text-left text-xs font-semibold text-muted-foreground">
                        {t('manager.flows.detail.sourceColumn')}
                      </TableHead>
                      <TableHead className="h-auto w-8 px-4 py-2" />
                      <TableHead className="h-auto px-4 py-2 text-left text-xs font-semibold text-muted-foreground">
                        {t('manager.flows.detail.targetColumn')}
                      </TableHead>
                      <TableHead className="h-auto w-24 px-4 py-2 text-right text-xs font-semibold text-muted-foreground">
                        {t('manager.flows.detail.enabled')}
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody className="divide-y divide-border">
                    {flow.column_mappings.map((mapping) => (
                      <TableRow
                        key={mapping.source_column}
                        className={`border-b border-border hover:bg-muted/40 ${!mapping.enabled ? 'opacity-45' : ''}`}
                      >
                        <TableCell className="px-4 py-3 align-middle">
                          <span className="block font-mono text-xs font-medium text-foreground">
                            {mapping.source_column}
                          </span>
                          <span className="mt-0.5 block font-mono text-[10px] text-muted-foreground">
                            {mapping.source_type}
                          </span>
                        </TableCell>
                        <TableCell className="px-4 py-3 align-middle">
                          <ArrowRight className="inline h-3.5 w-3.5 text-muted-foreground" />
                        </TableCell>
                        <TableCell className="px-4 py-3 align-middle">
                          <span className="block font-mono text-xs font-medium text-foreground">
                            {mapping.sink_column}
                          </span>
                          <span className="mt-0.5 block font-mono text-[10px] text-muted-foreground">
                            {mapping.sink_type}
                          </span>
                        </TableCell>
                        <TableCell className="px-4 py-3 text-right align-middle">
                          <Badge
                            variant="outline"
                            className={mapping.enabled
                              ? 'border-emerald-500/20 bg-emerald-500/10 text-[10px] text-emerald-600 dark:text-emerald-400'
                              : 'border-border bg-muted text-[10px] text-muted-foreground'}
                          >
                            {mapping.enabled ? t('common.yes') : t('common.no')}
                          </Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </div>

          <div className="rounded-lg border border-border bg-card p-4">
            <div className="mb-4 flex items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <History className="h-4 w-4 text-sky-500 dark:text-sky-400" />
                <h2 className="text-sm font-semibold text-foreground">
                  {t('manager.flows.detail.tableProgress')}
                </h2>
              </div>
              <Badge variant="outline" className="border-border bg-muted/30 text-xs text-muted-foreground">
                {progressData?.tables?.length ?? 0}
              </Badge>
            </div>

            {progressLoading ? (
              <div className="h-16 animate-pulse rounded-lg bg-muted" />
            ) : !progressData?.tables?.length ? (
              <div className="rounded-lg border border-dashed border-border py-8 text-center text-xs text-muted-foreground">
                {t('manager.flows.detail.noProgress')}
              </div>
            ) : (
              <div className="space-y-2">
                {progressData.tables.map((table) => (
                  <div
                    key={table.table_name}
                    className="rounded-lg border border-border bg-muted/20 p-3"
                  >
                    <div className="flex items-center justify-between gap-3">
                      <div className="min-w-0">
                        <span className="block truncate font-mono text-xs font-semibold text-foreground">
                          {table.table_name}
                        </span>
                        <span className="mt-0.5 block font-mono text-[10px] text-muted-foreground">
                          {t('manager.flows.detail.rows', { count: table.rows_synced })}
                        </span>
                      </div>
                      <StatusBadge
                        status={progressStateToBadge(table.state)}
                        label={t(`manager.flows.state.${progressStateLabel(table.state)}`)}
                      />
                    </div>
                    {table.last_offset && (
                      <div className="mt-2 flex items-center gap-1.5 text-[10px] text-muted-foreground">
                        <Terminal className="h-3 w-3" />
                        <span className="truncate font-mono">
                          {t('manager.flows.detail.lastOffset', { offset: table.last_offset })}
                        </span>
                      </div>
                    )}
                    {table.error_message && (
                      <div className="mt-2 rounded border border-red-500/10 bg-red-500/5 p-2 font-mono text-[10px] text-red-500">
                        {table.error_message}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="space-y-6">
          <div className="rounded-lg border border-border bg-card p-4">
            <div className="mb-4 flex items-center gap-2">
              <SlidersHorizontal className="h-4 w-4 text-sky-500 dark:text-sky-400" />
              <h2 className="text-sm font-semibold text-foreground">
                {t('manager.flows.detail.executionConfig')}
              </h2>
            </div>

            <div className="grid gap-2">
              <div className="rounded-md border border-border bg-muted/20 px-3 py-2">
                <p className="text-[10px] uppercase tracking-wide text-muted-foreground">
                  {t('manager.flows.detail.batchSize')}
                </p>
                <p className="mt-1 font-mono text-sm font-semibold text-foreground">
                  {flow.options?.batch_size || 100}
                </p>
              </div>
              <div className="rounded-md border border-border bg-muted/20 px-3 py-2">
                <p className="text-[10px] uppercase tracking-wide text-muted-foreground">
                  {t('manager.flows.detail.flushInterval')}
                </p>
                <p className="mt-1 font-mono text-sm font-semibold text-foreground">
                  {flow.options?.flush_interval_ms || 1000}ms
                </p>
              </div>
              <div className="rounded-md border border-border bg-muted/20 px-3 py-2">
                <p className="text-[10px] uppercase tracking-wide text-muted-foreground">
                  {t('manager.flows.detail.partitionCount')}
                </p>
                <p className="mt-1 font-mono text-sm font-semibold text-foreground">
                  {flow.options?.partition_count || 4}
                </p>
              </div>
              <div className="rounded-md border border-border bg-muted/20 px-3 py-2">
                <p className="text-[10px] uppercase tracking-wide text-muted-foreground">
                  {t('manager.flows.columns')}
                </p>
                <p className="mt-1 font-mono text-sm font-semibold text-foreground">
                  {mappingCount}/{totalMappings}
                </p>
              </div>
            </div>
          </div>

          <div className="rounded-lg border border-border bg-card p-4">
            <div className="mb-4 flex items-center gap-2">
              <Filter className="h-4 w-4 text-amber-500 dark:text-amber-400" />
              <h2 className="text-sm font-semibold text-foreground">
                {t('manager.flows.detail.filterExpression')}
              </h2>
            </div>

            {hasFilter ? (
              <pre className="max-h-56 overflow-auto rounded-md border border-border bg-muted/30 p-3 font-mono text-xs leading-relaxed text-foreground">
                {flow.options?.filter_expression}
              </pre>
            ) : (
              <div className="rounded-md border border-dashed border-border py-6 text-center text-xs text-muted-foreground">
                {t('manager.cards.filterOff', { defaultValue: 'No filter' })}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
