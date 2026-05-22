/**
 * Utility functions for formatting display values.
 */
import i18n from './i18n';

/** Formats large numbers with K/M/B suffixes. */
export function formatNumber(value: number): string {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(1)}B`;
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(1)}K`;
  return value.toFixed(value % 1 === 0 ? 0 : 1);
}

/** Formats bytes into human-readable size. */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

/** Formats seconds into human-readable duration (e.g., "2d 5h 30m"). */
export function formatDuration(totalSeconds: number): string {
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = Math.floor(totalSeconds % 60);

  const parts: string[] = [];
  if (days > 0) parts.push(`${days}${i18n.t('common.days', { defaultValue: 'd' })}`);
  if (hours > 0) parts.push(`${hours}${i18n.t('common.hours', { defaultValue: 'h' })}`);
  if (minutes > 0) parts.push(`${minutes}${i18n.t('common.minutes', { defaultValue: 'm' })}`);
  if (parts.length === 0) parts.push(`${seconds}${i18n.t('common.seconds', { defaultValue: 's' })}`);

  return parts.join(' ');
}

/** Formats a percentage value with fixed decimals. */
export function formatPercent(value: number, decimals = 2): string {
  return `${value.toFixed(decimals)}%`;
}

/** Sums all partition lag values from a partition_lag map. */
export function sumPartitionLag(lag: Record<number, number> | undefined): number {
  if (!lag) return 0;
  return Object.values(lag).reduce((sum, v) => sum + v, 0);
}
