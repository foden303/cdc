import { useTranslation } from 'react-i18next';
import { Card, CardContent } from '@/components/ui/card';
import { StatusBadge } from '@/components/shared/StatusBadge';
import type { Status } from '@/components/shared/StatusBadge';
import { Server, Radio, Users, AlertTriangle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface SystemHealthBarProps {
  natsConnected: boolean;
  channelUtilPercent: number;
  activeWorkers: number;
  dlqCount: number;
  onReprocessDlq: () => void;
  isReprocessing?: boolean;
}

/** System health row — NATS, channel utilization, workers, DLQ button. */
export function SystemHealthBar({
  natsConnected,
  channelUtilPercent,
  activeWorkers,
  dlqCount,
  onReprocessDlq,
  isReprocessing,
}: SystemHealthBarProps) {
  const { t } = useTranslation();

  const natsStatus: Status = natsConnected ? 'healthy' : 'unhealthy';

  const channelColor =
    channelUtilPercent > 80
      ? 'text-red-500'
      : channelUtilPercent > 50
        ? 'text-amber-500'
        : 'text-emerald-500';

  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex flex-wrap items-center gap-6">
          {/* NATS status */}
          <div className="flex items-center gap-2">
            <Server className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm text-muted-foreground">
              {t('dashboard.natsStatus')}
            </span>
            <StatusBadge status={natsStatus} />
          </div>

          {/* Channel utilization */}
          <div className="flex items-center gap-2">
            <Radio className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm text-muted-foreground">
              {t('dashboard.channelUtil')}
            </span>
            <span className={cn('text-sm font-semibold', channelColor)}>
              {channelUtilPercent.toFixed(0)}%
            </span>
          </div>

          {/* Active workers */}
          <div className="flex items-center gap-2">
            <Users className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm text-muted-foreground">
              {t('dashboard.activeWorkers')}
            </span>
            <span className="text-sm font-semibold text-foreground">
              {activeWorkers}
            </span>
          </div>

          {/* DLQ reprocess */}
          <div className="ml-auto flex items-center gap-2">
            {dlqCount > 0 && (
              <span className="flex items-center gap-1 text-sm text-amber-500">
                <AlertTriangle className="h-3.5 w-3.5" />
                {dlqCount} DLQ
              </span>
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={onReprocessDlq}
              disabled={isReprocessing || dlqCount === 0}
              className="cursor-pointer"
            >
              {t('dashboard.reprocessDlq')}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
