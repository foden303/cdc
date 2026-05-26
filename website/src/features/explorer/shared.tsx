import { Badge } from '@/components/ui/badge';
import { useTranslation } from 'react-i18next';

export function decodePayload(base64Str: string): string {
  if (!base64Str) return '';
  try {
    return decodeURIComponent(
      atob(base64Str)
        .split('')
        .map((c) => `%${(`00${c.charCodeAt(0).toString(16)}`).slice(-2)}`)
        .join(''),
    );
  } catch {
    try {
      return atob(base64Str);
    } catch {
      return base64Str;
    }
  }
}

export function parseSubject(subject: string) {
  const parts = subject.split('.');
  return {
    stream: parts[0] || '',
    sourceId: parts[1] || '',
    schema: parts[2] || '',
    table: parts[3] || '',
    partition: parts[4] || '',
    topic: parts.length >= 4 ? parts.slice(0, 4).join('.') : subject,
    shortName: parts.length >= 4 ? `${parts[2]}.${parts[3]}` : subject,
  };
}

export function formatBytes(value: number) {
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${(value / 1024 / 1024).toFixed(1)} MB`;
}

export function formatCount(value: number | string | null | undefined) {
  const numeric = typeof value === 'string' ? Number(value) : (value ?? 0);
  if (!Number.isFinite(numeric)) return '0';
  return numeric.toLocaleString();
}

export function messageSize(data: string) {
  if (!data) return 0;
  return decodePayload(data).length;
}

export function formatTime(timestamp: string | number) {
  if (!timestamp) return '-';
  const raw = typeof timestamp === 'number' ? timestamp : Number(timestamp);
  const date = Number.isFinite(raw) ? new Date(raw) : new Date(timestamp);
  if (Number.isNaN(date.getTime())) return String(timestamp);
  return date.toLocaleString();
}

export function StatusBadge({ status }: { status: 'sent' | 'dlq' | 'pending' | 'active' | 'lagging' }) {
  const { t } = useTranslation();
  const className =
    status === 'sent' || status === 'active'
      ? 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400'
      : status === 'dlq'
        ? 'border-rose-500/25 bg-rose-500/10 text-rose-700 dark:text-rose-400'
        : 'border-amber-500/25 bg-amber-500/10 text-amber-700 dark:text-amber-400';

  return (
    <Badge variant="outline" className={className}>
      {t(`explorer.status.${status}`)}
    </Badge>
  );
}
