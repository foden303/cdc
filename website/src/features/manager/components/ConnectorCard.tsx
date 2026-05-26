import type { LucideIcon } from "lucide-react";
import { Edit3, Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

type ConnectorTone = "source" | "sink";

interface ConnectorMetric {
  label: string;
  value: string;
  tone?: "default" | "danger";
}

interface ConnectorCardProps {
  tone: ConnectorTone;
  icon: LucideIcon;
  name: string;
  endpoint: string;
  typeLabel: string;
  instanceId: string;
  metrics: ConnectorMetric[];
  editLabel: string;
  deleteLabel: string;
  deleteDisabled?: boolean;
  onEdit: () => void;
  onDelete: () => void;
}

const toneStyles: Record<
  ConnectorTone,
  {
    icon: string;
    badge: string;
    hover: string;
    rail: string;
  }
> = {
  source: {
    icon: "border-sky-500/20 bg-sky-500/10 text-sky-600 dark:text-sky-400",
    badge: "border-sky-500/25 bg-sky-500/10 text-sky-700 dark:text-sky-300",
    hover: "hover:border-sky-500/40",
    rail: "from-sky-500/10 via-transparent to-transparent",
  },
  sink: {
    icon: "border-violet-500/20 bg-violet-500/10 text-violet-600 dark:text-violet-400",
    badge:
      "border-violet-500/25 bg-violet-500/10 text-violet-700 dark:text-violet-300",
    hover: "hover:border-violet-500/40",
    rail: "from-violet-500/10 via-transparent to-transparent",
  },
};

export function ConnectorCard({
  tone,
  icon: Icon,
  name,
  endpoint,
  typeLabel,
  instanceId,
  metrics,
  editLabel,
  deleteLabel,
  deleteDisabled,
  onEdit,
  onDelete,
}: ConnectorCardProps) {
  const styles = toneStyles[tone];

  return (
    <article
      className={cn(
        "relative overflow-hidden rounded-lg border border-border bg-card p-4 shadow-sm shadow-black/5 transition-colors duration-200 hover:bg-card/95",
        styles.hover,
      )}
    >
      <div
        className={cn(
          "pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r",
          styles.rail,
        )}
      />

      <div className="space-y-4">
        <div className="flex items-start justify-between gap-3">
          <div className="flex min-w-0 items-start gap-3">
            <div
              className={cn(
                "mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border",
                styles.icon,
              )}
            >
              <Icon className="h-4.5 w-4.5" />
            </div>
            <div className="min-w-0 space-y-0.5">
              <h2 className="truncate text-base font-semibold leading-5 text-foreground">
                {name}
              </h2>
              <p className="truncate font-mono text-[11px] font-medium leading-3 text-muted-foreground/75">
                {instanceId}
              </p>
              <p className="truncate font-mono text-[11px] leading-4 text-muted-foreground">
                {endpoint}
              </p>
            </div>
          </div>
          <div className="flex w-28 shrink-0 flex-col items-end gap-1">
            <div className="flex items-center gap-1.5">
              <button
                onClick={onEdit}
                className="cursor-pointer rounded-md border border-border bg-background/60 p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                title={editLabel}
              >
                <Edit3 className="h-3.5 w-3.5" />
              </button>
              <button
                onClick={onDelete}
                disabled={deleteDisabled}
                className="cursor-pointer rounded-md border border-border bg-background/60 p-1.5 text-muted-foreground transition-colors hover:border-destructive/30 hover:bg-destructive/10 hover:text-destructive disabled:cursor-not-allowed disabled:opacity-50"
                title={deleteLabel}
              >
                <Trash2 className="h-3.5 w-3.5" />
              </button>
            </div>
            <Badge
              variant="outline"
              className={cn(
                "max-w-28 px-2 py-0 text-[10px] font-semibold leading-5",
                styles.badge,
              )}
            >
              {typeLabel}
            </Badge>
          </div>
        </div>

        <div className="grid grid-cols-3 gap-2">
          {metrics.map((metric) => (
            <div
              key={metric.label}
              className="min-w-0 rounded-md border border-border bg-muted/20 px-3 py-2"
            >
              <p
                className={cn(
                  "truncate text-sm font-semibold leading-5",
                  metric.tone === "danger" ? "text-red-500" : "text-foreground",
                )}
              >
                {metric.value}
              </p>
              <p className="truncate text-[10px] uppercase tracking-wide text-muted-foreground">
                {metric.label}
              </p>
            </div>
          ))}
        </div>
      </div>
    </article>
  );
}
