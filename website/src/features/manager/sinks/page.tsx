import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  HardDrive,
  Plus,
  Trash2,
  Edit3,
  RefreshCw,
  Server,
} from 'lucide-react';
import { toast } from 'sonner';

import { useConfig, useRemoveSink } from '@/lib/query/manager';
import { Button } from '@/components/ui/button';
import { SinkForm } from './SinkForm';
import type { SinkConfig } from '@/types/api';

export default function SinksPage() {
  const { t } = useTranslation();

  // Modal State
  const [formOpen, setFormOpen] = useState(false);
  const [editingSink, setEditingSink] = useState<SinkConfig | null>(null);

  // Queries & Mutations
  const { data, isLoading, refetch, isFetching } = useConfig();
  const removeMutation = useRemoveSink();

  const handleAddClick = () => {
    setEditingSink(null);
    setFormOpen(true);
  };

  const handleEditClick = (sink: SinkConfig) => {
    setEditingSink(sink);
    setFormOpen(true);
  };

  const handleDeleteClick = async (instanceId: string) => {
    if (
      confirm(
        t('manager.sinks.confirm.delete', { id: instanceId })
      )
    ) {
      try {
        await removeMutation.mutateAsync(instanceId);
        toast.success(t('manager.sinks.toast.deleted'));
      } catch (err: any) {
        toast.error(err.message || t('manager.sinks.toast.deleteFailed'));
      }
    }
  };

  const sinks = data?.config?.sinks || [];

  return (
    <div className="flex flex-col min-h-[calc(100vh-7.5rem)] space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
            <HardDrive className="h-6 w-6 text-sky-400" />
            {t('manager.sinks.title')}
          </h1>
          <p className="text-xs text-muted-foreground mt-1">
            {t('manager.sinks.desc')}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => refetch()}
            className="h-9 w-9 inline-flex items-center justify-center rounded-lg border border-border hover:bg-muted text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
            title={t('common.refresh')}
          >
            <RefreshCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          </button>
          <Button
            onClick={handleAddClick}
            className="h-9 text-xs bg-sky-500 text-slate-950 hover:bg-sky-400 font-semibold cursor-pointer"
          >
            <Plus className="h-4 w-4 mr-1" />
            {t('manager.sinks.add')}
          </Button>
        </div>
      </div>

      {/* Main Content */}
      {isLoading ? (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div
              key={i}
              className="h-40 rounded-xl bg-muted/40 animate-pulse border border-border"
            />
          ))}
        </div>
      ) : sinks.length === 0 ? (
        <div className="flex-1 flex flex-col items-center justify-center p-8 text-center rounded-xl border border-dashed border-border bg-card">
          <div className="p-3.5 rounded-full bg-muted border border-border mb-4">
            <HardDrive className="h-6 w-6 text-muted-foreground" />
          </div>
          <h3 className="text-sm font-semibold text-foreground mb-1">
            {t('manager.sinks.noSinks')}
          </h3>
          <p className="text-xs text-muted-foreground max-w-xs mb-4">
            {t('manager.sinks.noSinksDesc')}
          </p>
          <Button
            onClick={handleAddClick}
            className="h-8 text-xs bg-sky-500 text-slate-950 hover:bg-sky-400 font-semibold cursor-pointer"
          >
            <Plus className="h-3.5 w-3.5 mr-1" />
            {t('manager.sinks.add')}
          </Button>
        </div>
      ) : (
        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {sinks.map((sink) => (
            <div
              key={sink.instance_id}
              className="rounded-xl border border-border bg-card p-5 flex flex-col justify-between hover:border-border/80 hover:bg-card/90 transition-all duration-300 shadow-sm shadow-black/10"
            >
              <div className="space-y-4">
                {/* Top: Icon + Name + Type */}
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2.5">
                    <div className="p-2 bg-muted border border-border rounded-lg text-sky-500 dark:text-sky-400 shrink-0">
                      <Server className="h-4.5 w-4.5" />
                    </div>
                    <div className="truncate max-w-[170px]">
                      <span className="font-mono text-xs font-semibold text-foreground block truncate">
                        {sink.name || sink.database}
                      </span>
                      <span className="text-[10px] text-muted-foreground font-mono block mt-0.5 truncate">
                        {sink.host}:{sink.port}
                      </span>
                    </div>
                  </div>
                  <span className="inline-flex items-center rounded-md bg-sky-500/10 px-1.5 py-0.5 text-[9px] font-mono font-medium text-sky-400 ring-1 ring-inset ring-sky-500/20 select-none">
                    {sink.type === 'postgres' ? 'PostgreSQL' : sink.type.toUpperCase()}
                  </span>
                </div>

                {/* Connection info */}
                <div className="space-y-2 bg-muted/20 p-3 rounded-lg border border-border font-mono text-[10px] text-muted-foreground">
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground/80">Database</span>
                    <span className="text-foreground font-semibold">{sink.database}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground/80">Host</span>
                    <span className="text-foreground">{sink.host}:{sink.port}</span>
                  </div>
                  {sink.username && (
                    <div className="flex items-center justify-between">
                      <span className="text-muted-foreground/80">User</span>
                      <span className="text-foreground">{sink.username}</span>
                    </div>
                  )}
                </div>
              </div>

              {/* Bottom actions */}
              <div className="mt-5 pt-4 border-t border-border flex items-center justify-end gap-1.5">
                <button
                  onClick={() => handleEditClick(sink)}
                  className="p-1.5 rounded hover:bg-muted border border-border text-muted-foreground hover:text-foreground transition-all cursor-pointer"
                  title={t('manager.sinks.card.editTooltip')}
                >
                  <Edit3 className="h-3.5 w-3.5" />
                </button>
                <button
                  onClick={() => handleDeleteClick(sink.instance_id)}
                  disabled={removeMutation.isPending}
                  className="p-1.5 rounded hover:bg-destructive/10 border border-border hover:border-destructive/30 text-muted-foreground hover:text-destructive transition-all cursor-pointer"
                  title={t('manager.sinks.card.deleteTooltip')}
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* CREATE / EDIT DIALOG */}
      <SinkForm
        open={formOpen}
        onOpenChange={setFormOpen}
        sinkToEdit={editingSink}
      />
    </div>
  );
}
