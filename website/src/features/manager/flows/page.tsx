import { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import {
  FolderSync,
  Plus,
  Play,
  Pause,
  Trash2,
  RefreshCw,
  GitFork,
  ArrowRight,
  Clock,
  AlertOctagon,
  ArrowUpRight,
} from 'lucide-react';
import { toast } from 'sonner';

import {
  useFlows,
  useConfig,
  usePauseFlow,
  useResumeFlow,
  useDeleteFlow,
} from '@/lib/query/manager';
import { Button } from '@/components/ui/button';
import { StatusBadge } from '@/components/shared/StatusBadge';
import { FlowWizard } from './FlowWizard';
import { ROUTES } from '@/config/routes';

export default function FlowsPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [wizardOpen, setWizardOpen] = useState(false);

  // Queries
  const { data: flowsData, isLoading: flowsLoading, refetch: refetchFlows, isFetching } = useFlows();
  const { data: configData } = useConfig();

  // Mutations
  const pauseMutation = usePauseFlow();
  const resumeMutation = useResumeFlow();
  const deleteMutation = useDeleteFlow();

  // Map Source/Sink IDs to Config Objects for rich UI names
  const sourcesMap = useMemo(() => {
    if (!configData?.config?.sources) return new Map();
    return new Map(configData.config.sources.map((s) => [s.instance_id, s]));
  }, [configData]);

  const sinksMap = useMemo(() => {
    if (!configData?.config?.sinks) return new Map();
    return new Map(configData.config.sinks.map((s) => [s.instance_id, s]));
  }, [configData]);

  const handlePause = async (flowId: string) => {
    try {
      await pauseMutation.mutateAsync(flowId);
      toast.success(t('manager.flows.toast.paused'));
    } catch {
      toast.error(t('manager.flows.toast.pauseFailed'));
    }
  };

  const handleResume = async (flowId: string) => {
    try {
      await resumeMutation.mutateAsync(flowId);
      toast.success(t('manager.flows.toast.resumed'));
    } catch {
      toast.error(t('manager.flows.toast.resumeFailed'));
    }
  };

  const handleDelete = async (flowId: string) => {
    if (confirm(t('manager.flows.confirm.delete'))) {
      try {
        await deleteMutation.mutateAsync(flowId);
        toast.success(t('manager.flows.toast.deleted'));
      } catch {
        toast.error(t('manager.flows.toast.deleteFailed'));
      }
    }
  };

  // Maps backend FlowStatus string to StatusBadge state
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

  return (
    <div className="flex flex-col min-h-[calc(100vh-7.5rem)] space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
            <FolderSync className="h-6 w-6 text-sky-400" />
            {t('manager.flows.title')}
          </h1>
          <p className="text-xs text-muted-foreground mt-1">
            {t('manager.flows.desc')}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => refetchFlows()}
            className="h-9 w-9 inline-flex items-center justify-center rounded-lg border border-border hover:bg-muted text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
            title={t('common.refresh')}
          >
            <RefreshCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          </button>
          <Button
            onClick={() => setWizardOpen(true)}
            className="h-9 text-xs bg-sky-500 text-slate-950 hover:bg-sky-400 font-semibold cursor-pointer"
          >
            <Plus className="h-4 w-4 mr-1" />
            {t('manager.flows.create')}
          </Button>
        </div>
      </div>

      {/* Main Content */}
      {flowsLoading ? (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-44 rounded-xl bg-muted/40 animate-pulse border border-border" />
          ))}
        </div>
      ) : !flowsData?.flows || flowsData.flows.length === 0 ? (
        <div className="flex-1 flex flex-col items-center justify-center p-8 text-center rounded-xl border border-dashed border-border bg-card">
          <div className="p-3.5 rounded-full bg-muted border border-border mb-4">
            <GitFork className="h-6 w-6 text-muted-foreground" />
          </div>
          <h3 className="text-sm font-semibold text-foreground mb-1">{t('manager.flows.noFlows')}</h3>
          <p className="text-xs text-muted-foreground max-w-xs mb-4">
            {t('manager.flows.noFlowsDesc')}
          </p>
          <Button
            onClick={() => setWizardOpen(true)}
            className="h-8 text-xs bg-sky-500 text-slate-950 hover:bg-sky-400 font-semibold cursor-pointer"
          >
            <Plus className="h-3.5 w-3.5 mr-1" />
            {t('manager.flows.create')}
          </Button>
        </div>
      ) : (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {flowsData.flows.map((flow) => {
            const src = sourcesMap.get(flow.source_id);
            const sink = sinksMap.get(flow.sink_id);

            const isError = flow.status === 'FLOW_STATUS_ERROR';

            return (
              <div
                key={flow.flow_id}
                className={`relative rounded-xl border bg-card p-5 flex flex-col justify-between transition-all duration-300 shadow-sm ${
                  isError 
                    ? 'border-destructive/25 hover:border-destructive/40 shadow-destructive/10' 
                    : 'border-border hover:border-border/80 hover:bg-card/90 shadow-black/10'
                }`}
              >
                {/* Error Banner indicator */}
                {isError && (
                  <div className="absolute top-2 right-2 p-1 rounded-full bg-destructive/10 border border-destructive/20 text-destructive">
                    <AlertOctagon className="h-3.5 w-3.5 animate-pulse" />
                  </div>
                )}

                {/* Top Section: Title & Details */}
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-xs font-semibold text-foreground block truncate max-w-[180px]">
                      {flow.name || flow.flow_id}
                    </span>
                    <StatusBadge status={mapStatus(flow.status)} />
                  </div>

                  {/* Flow Map Visualiser */}
                  <div className="flex items-center gap-2 bg-muted/30 p-2.5 rounded-lg border border-border font-mono text-[10px] text-muted-foreground">
                    <div className="truncate max-w-[100px]">
                      <span className="text-[8px] uppercase tracking-wider text-muted-foreground/80 block mb-0.5">{t('dashboard.source')}</span>
                      <span className="text-foreground font-semibold truncate block">
                        {src?.name || flow.source_id.substring(0, 6)}
                      </span>
                      <span className="text-muted-foreground block truncate">{flow.source_table}</span>
                    </div>
                    <ArrowRight className="h-3.5 w-3.5 text-muted-foreground/55 shrink-0 mt-2" />
                    <div className="truncate max-w-[100px]">
                      <span className="text-[8px] uppercase tracking-wider text-muted-foreground/80 block mb-0.5">{t('dashboard.sink')}</span>
                      <span className="text-foreground font-semibold truncate block">
                        {sink?.name || flow.sink_id.substring(0, 6)}
                      </span>
                      <span className="text-muted-foreground block truncate">{flow.sink_table || '-'}</span>
                    </div>
                  </div>
                </div>

                {/* Bottom Section: Info + Controls */}
                <div className="mt-4 pt-4 border-t border-border flex items-center justify-between">
                  <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground">
                    <Clock className="h-3 w-3" />
                    <span>
                      {new Date(flow.created_at * 1000).toLocaleDateString()}
                    </span>
                  </div>

                  <div className="flex items-center gap-1">
                    {/* Pause / Resume */}
                    {flow.status === 'FLOW_STATUS_RUNNING' ? (
                      <button
                        onClick={() => handlePause(flow.flow_id)}
                        disabled={pauseMutation.isPending}
                        className="p-1.5 rounded hover:bg-muted border border-border text-muted-foreground hover:text-foreground transition-all cursor-pointer"
                        title={t('manager.flows.tooltips.pause')}
                      >
                        <Pause className="h-3.5 w-3.5" />
                      </button>
                    ) : (
                      <button
                        onClick={() => handleResume(flow.flow_id)}
                        disabled={resumeMutation.isPending}
                        className="p-1.5 rounded hover:bg-muted border border-border text-emerald-500 dark:text-emerald-400 hover:text-emerald-600 dark:hover:text-emerald-300 transition-all cursor-pointer"
                        title={t('manager.flows.tooltips.resume')}
                      >
                        <Play className="h-3.5 w-3.5" />
                      </button>
                    )}

                    {/* View Details */}
                    <button
                      onClick={() => navigate(ROUTES.MANAGER_FLOW_DETAIL.replace(':id', flow.flow_id))}
                      className="p-1.5 rounded hover:bg-muted border border-border text-sky-500 dark:text-sky-400 hover:text-sky-600 dark:hover:text-sky-300 transition-all cursor-pointer"
                      title={t('manager.flows.tooltips.details')}
                    >
                      <ArrowUpRight className="h-3.5 w-3.5" />
                    </button>

                    {/* Delete */}
                    <button
                      onClick={() => handleDelete(flow.flow_id)}
                      disabled={deleteMutation.isPending}
                      className="p-1.5 rounded hover:bg-destructive/10 border border-border hover:border-destructive/30 text-muted-foreground hover:text-destructive transition-all cursor-pointer"
                      title={t('manager.flows.tooltips.delete')}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* CREATE FLOW WIZARD */}
      <FlowWizard open={wizardOpen} onOpenChange={setWizardOpen} />
    </div>
  );
}
