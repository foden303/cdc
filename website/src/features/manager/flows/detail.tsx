import { useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  ArrowLeft,
  Pause,
  Play,
  Trash2,
  Database,
  ArrowRight,
  TrendingUp,
  History,
  GitCommit,
  FolderSync,
  Terminal,
  Activity,
  Layers,
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
import { MetricCard } from '@/components/shared/MetricCard';
import { StatusBadge } from '@/components/shared/StatusBadge';
import { ROUTES } from '@/config/routes';
import { formatNumber, formatDuration } from '@/lib/format';

export default function FlowDetailPage() {
  const { t } = useTranslation();
  const { id: flowId } = useParams<{ id: string }>();
  const navigate = useNavigate();

  // Queries
  const { data: flowData, isLoading: flowLoading } = useFlow(flowId || '');
  const { data: statsData, isLoading: statsLoading } = useFlowStats(flowId || '');
  const { data: progressData, isLoading: progressLoading } = useFlowProgress(flowId || '');
  const { data: configData } = useConfig();

  // Mutations
  const pauseMutation = usePauseFlow();
  const resumeMutation = useResumeFlow();
  const deleteMutation = useDeleteFlow();

  const flow = flowData?.flow;

  // Map Source/Sink IDs to Config Objects
  const selectedSource = useMemo(() => {
    if (!flow || !configData?.config?.sources) return null;
    return configData.config.sources.find((s) => s.instance_id === flow.source_id);
  }, [flow, configData]);

  const selectedSink = useMemo(() => {
    if (!flow || !configData?.config?.sinks) return null;
    return configData.config.sinks.find((s) => s.instance_id === flow.sink_id);
  }, [flow, configData]);

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

  const mapStatus = (status: string) => {
    switch (status) {
      case 'FLOW_STATUS_RUNNING':
        return 'healthy';
      case 'FLOW_STATUS_PAUSED':
        return 'paused';
      case 'FLOW_STATUS_ERROR':
        return 'unhealthy';
      default:
        return 'idle';
    }
  };

  const mapTableState = (state: string) => {
    switch (state) {
      case 'TABLE_SYNC_STATE_COMPLETED':
        return 'healthy';
      case 'TABLE_SYNC_STATE_SYNCING':
        return 'idle';
      case 'TABLE_SYNC_STATE_ERROR':
        return 'unhealthy';
      default:
        return 'paused';
    }
  };

  const formatTableStateText = (state: string) => {
    switch (state) {
      case 'TABLE_SYNC_STATE_COMPLETED':
        return t('manager.flows.state.completed');
      case 'TABLE_SYNC_STATE_SYNCING':
        return t('manager.flows.state.syncing');
      case 'TABLE_SYNC_STATE_ERROR':
        return t('manager.flows.state.error');
      default:
        return t('manager.flows.state.pending');
    }
  };

  if (flowLoading || !flow) {
    return (
      <div className="space-y-6">
        <div className="h-8 w-48 bg-muted rounded animate-pulse" />
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-28 bg-muted rounded animate-pulse" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header bar */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between border-b border-border pb-5">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate(ROUTES.MANAGER_FLOWS)}
            className="p-1.5 rounded-lg border border-border hover:bg-muted text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
          >
            <ArrowLeft className="h-4 w-4" />
          </button>
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-xl font-bold tracking-tight text-foreground font-mono">
                {flow.name || flow.flow_id}
              </h1>
              <StatusBadge
                status={mapStatus(flow.status)}
                label={t('common.' + mapStatus(flow.status))}
              />
            </div>
            <p className="text-[10px] text-muted-foreground font-mono mt-1">{t('common.id')}: {flow.flow_id}</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {flow.status === 'FLOW_STATUS_RUNNING' ? (
            <Button
              variant="outline"
              size="sm"
              onClick={handlePause}
              disabled={pauseMutation.isPending}
              className="h-9 text-xs cursor-pointer"
            >
              <Pause className="h-3.5 w-3.5 mr-1" />
              {t('manager.flows.pause')}
            </Button>
          ) : (
            <Button
              variant="outline"
              size="sm"
              onClick={handleResume}
              disabled={resumeMutation.isPending}
              className="h-9 text-xs text-emerald-500 hover:text-emerald-600 dark:text-emerald-400 dark:hover:text-emerald-300 cursor-pointer"
            >
              <Play className="h-3.5 w-3.5 mr-1" />
              {t('manager.flows.resume')}
            </Button>
          )}

          <Button
            variant="destructive"
            size="sm"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
            className="h-9 text-xs font-semibold cursor-pointer"
          >
            <Trash2 className="h-3.5 w-3.5 mr-1" />
            {t('manager.flows.delete')}
          </Button>
        </div>
      </div>

      {/* Sync Metrics Row */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          title={t('manager.flows.metrics.syncRate')}
          value={`${formatNumber(statsData?.events_per_second || 0)}/s`}
          description={t('manager.flows.metrics.syncRateDesc')}
          icon={TrendingUp}
          iconClassName="bg-sky-500/10 text-sky-400"
          loading={statsLoading}
        />
        <MetricCard
          title={t('manager.flows.metrics.lag')}
          value={`${formatDuration((statsData?.replication_lag_ms || 0) / 1000)}`}
          description={t('manager.flows.metrics.lagDesc', { lag: (statsData?.replication_lag_ms || 0).toLocaleString() })}
          icon={Activity}
          iconClassName="bg-amber-500/10 text-amber-400"
          loading={statsLoading}
        />
        <MetricCard
          title={t('manager.flows.metrics.eventsSynced')}
          value={formatNumber(statsData?.total_events_processed || 0)}
          description={t('manager.flows.metrics.eventsSyncedDesc')}
          icon={GitCommit}
          iconClassName="bg-indigo-500/10 text-indigo-400"
          loading={statsLoading}
        />
        <MetricCard
          title={t('manager.flows.metrics.lastSynced')}
          value={
            statsData?.last_synced_at
              ? new Date(statsData.last_synced_at * 1000).toLocaleTimeString()
              : '-'
          }
          description={
            statsData?.last_synced_at
              ? new Date(statsData.last_synced_at * 1000).toLocaleDateString()
              : t('manager.flows.metrics.neverSynced')
          }
          icon={History}
          iconClassName="bg-emerald-500/10 text-emerald-400"
          loading={statsLoading}
        />
      </div>

      {/* Two Column Layout */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Left main: Sync progress & column mappings */}
        <div className="lg:col-span-2 space-y-6">
          {/* Table sync progress */}
          <div className="rounded-xl border border-border bg-card p-5">
            <h2 className="text-sm font-semibold text-foreground mb-4 flex items-center gap-2">
              <Database className="h-4 w-4 text-sky-400" />
              {t('manager.flows.detail.tableProgress')}
            </h2>

            {progressLoading ? (
              <div className="h-16 bg-muted rounded animate-pulse" />
            ) : !progressData?.tables || progressData.tables.length === 0 ? (
              <div className="text-xs text-muted-foreground py-4 text-center">
                {t('manager.flows.detail.noProgress')}
              </div>
            ) : (
              <div className="space-y-3">
                {progressData.tables.map((table) => {
                  const state = mapTableState(table.state);
                  const isErr = table.state === 'TABLE_SYNC_STATE_ERROR';

                  return (
                    <div
                      key={table.table_name}
                      className={`p-4 rounded-lg border bg-muted/20 flex flex-col gap-2.5 transition-colors ${
                        isErr ? 'border-destructive/20 bg-destructive/5' : 'border-border hover:border-border/80'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-mono text-xs font-semibold text-foreground">{table.table_name}</span>
                        <div className="flex items-center gap-2.5">
                          <span className="font-mono text-[10px] text-muted-foreground">
                            {t('manager.flows.detail.rows', { count: table.rows_synced })}
                          </span>
                          <StatusBadge status={state} label={formatTableStateText(table.state)} />
                        </div>
                      </div>

                      {table.last_offset && (
                        <div className="flex items-center gap-1.5 text-[9px] text-muted-foreground font-mono">
                          <Terminal className="h-3 w-3" />
                          <span>{t('manager.flows.detail.lastOffset', { offset: table.last_offset })}</span>
                        </div>
                      )}

                      {isErr && table.error_message && (
                        <div className="mt-1.5 rounded bg-red-500/5 border border-red-500/10 p-2 text-[10px] font-mono text-red-400 leading-normal break-all">
                          {table.error_message}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* Column mappings */}
          <div className="rounded-xl border border-border bg-card p-5">
            <h2 className="text-sm font-semibold text-foreground mb-4 flex items-center gap-2">
              <Layers className="h-4 w-4 text-indigo-400" />
              {t('manager.flows.detail.columnSchema')}
            </h2>

            <div className="rounded-lg border border-border bg-card overflow-hidden text-xs">
              <Table>
                <TableHeader className="bg-muted/50 border-b border-border">
                  <TableRow className="border-b border-border hover:bg-transparent">
                    <TableHead className="px-4 py-2 h-auto text-muted-foreground select-none font-semibold text-left">
                      {t('manager.flows.detail.sourceColumn')}
                    </TableHead>
                    <TableHead className="px-4 py-2 w-8 h-auto text-muted-foreground select-none font-semibold text-left"></TableHead>
                    <TableHead className="px-4 py-2 h-auto text-muted-foreground select-none font-semibold text-left">
                      {t('manager.flows.detail.targetColumn')}
                    </TableHead>
                    <TableHead className="px-4 py-2 w-24 h-auto text-muted-foreground select-none font-semibold text-right">
                      {t('manager.flows.detail.enabled')}
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody className="divide-y divide-border">
                  {flow.column_mappings?.map((m) => (
                    <TableRow key={m.source_column} className={`border-b border-border hover:bg-muted/50 ${!m.enabled ? 'opacity-40' : ''}`}>
                      <TableCell className="px-4 py-3 align-middle">
                        <span className="font-mono text-foreground font-medium block">{m.source_column}</span>
                        <span className="text-[9px] font-mono text-muted-foreground block mt-0.5">{m.source_type}</span>
                      </TableCell>
                      <TableCell className="px-4 py-3 align-middle">
                        <ArrowRight className="h-3.5 w-3.5 text-muted-foreground inline" />
                      </TableCell>
                      <TableCell className="px-4 py-3 align-middle">
                        <span className="font-mono text-foreground font-medium block">{m.sink_column}</span>
                        <span className="text-[9px] font-mono text-muted-foreground block mt-0.5">{m.sink_type}</span>
                      </TableCell>
                      <TableCell className="px-4 py-3 text-right align-middle">
                        <span className={`inline-block px-1.5 py-0.5 rounded text-[10px] font-bold ${
                          m.enabled ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/10' : 'bg-muted text-muted-foreground border border-border'
                        }`}>
                          {m.enabled ? t('common.yes') : t('common.no')}
                        </span>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
        </div>

        {/* Right side: Pipeline Meta & Configuration */}
        <div className="space-y-6">
          {/* Connector Meta details */}
          <div className="rounded-xl border border-border bg-card p-5 space-y-4">
            <h2 className="text-sm font-semibold text-foreground mb-2 flex items-center gap-2">
              <FolderSync className="h-4 w-4 text-sky-400" />
              {t('manager.flows.detail.pipelineOverview')}
            </h2>

            {/* Source Connector Card */}
            <div className="p-3 bg-muted/20 border border-border rounded-lg">
              <span className="text-[8px] uppercase tracking-wider text-muted-foreground block mb-0.5 font-semibold">
                {t('manager.flows.detail.pipelineOverviewSource')}
              </span>
              <span className="text-xs font-mono font-bold text-foreground truncate block">
                {selectedSource?.name || flow.source_id}
              </span>
              <span className="text-[10px] text-muted-foreground font-mono capitalize block mt-0.5">
                {t('manager.flows.detail.pipelineOverviewType', { type: selectedSource?.type || 'Postgres' })}
              </span>
            </div>

            {/* Sink Connector Card */}
            <div className="p-3 bg-muted/20 border border-border rounded-lg">
              <span className="text-[8px] uppercase tracking-wider text-muted-foreground block mb-0.5 font-semibold">
                {t('manager.flows.detail.pipelineOverviewSink')}
              </span>
              <span className="text-xs font-mono font-bold text-foreground truncate block">
                {selectedSink?.name || flow.sink_id}
              </span>
              <span className="text-[10px] text-muted-foreground font-mono capitalize block mt-0.5">
                {t('manager.flows.detail.pipelineOverviewType', { type: selectedSink?.type || 'Clickhouse' })}
              </span>
            </div>
          </div>

          {/* Sync Configuration Options */}
          <div className="rounded-xl border border-border bg-card p-5 space-y-4">
            <h2 className="text-sm font-semibold text-foreground mb-2 flex items-center gap-2">
              <History className="h-4 w-4 text-sky-400" />
              {t('manager.flows.detail.executionConfig')}
            </h2>

            <div className="space-y-3 font-mono text-[11px] text-muted-foreground">
              <div className="flex items-center justify-between border-b border-border pb-2">
                <span>{t('manager.flows.detail.batchSize')}</span>
                <span className="text-foreground font-semibold">{flow.options?.batch_size || 100}</span>
              </div>
              <div className="flex items-center justify-between border-b border-border pb-2">
                <span>{t('manager.flows.detail.flushInterval')}</span>
                <span className="text-foreground font-semibold">{flow.options?.flush_interval_ms || 1000}ms</span>
              </div>
              {flow.options?.filter_expression && (
                <div className="space-y-1.5 pt-1">
                  <span className="block text-muted-foreground">{t('manager.flows.detail.filterExpression')}</span>
                  <div className="p-2 bg-muted/30 border border-border rounded text-[10px] text-indigo-500 dark:text-indigo-400 break-all leading-normal">
                    {flow.options.filter_expression}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
