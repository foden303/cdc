import type { LucideIcon } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { Skeleton } from '@/components/ui/skeleton';

interface MetricCardProps {
  title: string;
  value: string | number;
  unit?: string;
  icon: LucideIcon;
  trend?: 'up' | 'down' | 'neutral';
  trendValue?: string;
  className?: string;
  iconClassName?: string;
  onClick?: () => void;
  description?: string;
  loading?: boolean;
}

/** Dashboard KPI card — displays a single metric with icon and optional trend/description. Supports loading skeleton and onClick navigation. */
export function MetricCard({
  title,
  value,
  unit,
  icon: Icon,
  trend,
  trendValue,
  className,
  iconClassName,
  onClick,
  description,
  loading,
}: MetricCardProps) {
  if (loading) {
    return <Skeleton className={cn('h-[108px] w-full', className)} />;
  }

  return (
    <Card
      onClick={onClick}
      className={cn(
        'relative overflow-hidden transition-all duration-200 hover:shadow-lg',
        onClick && 'cursor-pointer hover:border-primary/50 hover:bg-accent/10 active:scale-[0.99]',
        className,
      )}
    >
      <CardContent className="p-5">
        <div className="flex items-start justify-between">
          <div className="space-y-2">
            <p className="text-sm font-medium text-muted-foreground">{title}</p>
            <div className="flex items-baseline gap-1.5">
              <span className="text-2xl font-bold tracking-tight text-foreground">
                {value}
              </span>
              {unit && (
                <span className="text-sm text-muted-foreground">{unit}</span>
              )}
            </div>
            {trend && trendValue && (
              <p
                className={cn(
                  'text-xs font-medium',
                  trend === 'up' && 'text-emerald-500',
                  trend === 'down' && 'text-red-500',
                  trend === 'neutral' && 'text-muted-foreground',
                )}
              >
                {trend === 'up' ? '↑' : trend === 'down' ? '↓' : '→'}{' '}
                {trendValue}
              </p>
            )}
            {description && (
              <p className="text-xs text-muted-foreground">{description}</p>
            )}
          </div>
          <div
            className={cn(
              'flex h-10 w-10 items-center justify-center rounded-lg',
              iconClassName || 'bg-primary/10 text-primary',
            )}
          >
            <Icon className="h-5 w-5" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
