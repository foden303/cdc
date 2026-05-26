import { useTranslation } from 'react-i18next';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';

export type Status = 'running' | 'paused' | 'error' | 'idle' | 'healthy' | 'unhealthy';

const STATUS_CONFIG: Record<Status, { label: string; className: string; dotClassName: string }> = {
  running: {
    label: 'Running',
    className: 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20',
    dotClassName: 'bg-emerald-500',
  },
  healthy: {
    label: 'Healthy',
    className: 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20',
    dotClassName: 'bg-emerald-500',
  },
  paused: {
    label: 'Paused',
    className: 'bg-amber-500/10 text-amber-500 border-amber-500/20',
    dotClassName: 'bg-amber-500',
  },
  error: {
    label: 'Error',
    className: 'bg-red-500/10 text-red-500 border-red-500/20',
    dotClassName: 'bg-red-500',
  },
  unhealthy: {
    label: 'Unhealthy',
    className: 'bg-red-500/10 text-red-500 border-red-500/20',
    dotClassName: 'bg-red-500',
  },
  idle: {
    label: 'Idle',
    className: 'bg-muted text-muted-foreground border-border',
    dotClassName: 'bg-muted-foreground',
  },
};

interface StatusBadgeProps {
  status: Status;
  label?: string;
  showDot?: boolean;
  className?: string;
}

/** Status badge — colored indicator for component state. */
export function StatusBadge({ status, label, showDot = true, className }: StatusBadgeProps) {
  const { t } = useTranslation();
  const config = STATUS_CONFIG[status] ?? STATUS_CONFIG.idle;

  return (
    <Badge
      variant="outline"
      className={cn('gap-1.5 font-medium', config.className, className)}
    >
      {showDot && (
        <span className={cn('h-1.5 w-1.5 rounded-full', config.dotClassName)} />
      )}
      {label || t('common.status.' + status, { defaultValue: config.label })}
    </Badge>
  );
}
