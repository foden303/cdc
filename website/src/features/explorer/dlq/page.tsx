import { useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Inbox, RefreshCw, RotateCcw, X } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useDLQMessages, useReprocessDLQ } from '@/lib/query/explorer';
import type { DLQMessage } from '@/types/api';
import { MessageDetailSheet } from '../components/MessageDetailSheet';
import { formatBytes, formatTime, messageSize, StatusBadge } from '../shared';

export default function ExplorerDLQPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const topicFilter = searchParams.get('topic') || '';
  const [selectedMessage, setSelectedMessage] = useState<DLQMessage | null>(null);
  const { data, isLoading, isFetching, refetch } = useDLQMessages(1, 100);
  const reprocessMutation = useReprocessDLQ();
  const messages = useMemo(() => {
    const rows = data?.data ?? [];
    if (!topicFilter) return rows;
    return rows.filter((message) => {
      const originalSubject =
        message.original_subject || message.headers?.['X-DLQ-Original-Subject'] || message.subject;
      return originalSubject.startsWith(topicFilter);
    });
  }, [data, topicFilter]);

  const reprocessAll = async () => {
    try {
      const result = await reprocessMutation.mutateAsync();
      toast.success(`Reprocessed ${result.count || 0} DLQ messages`);
      refetch();
    } catch (error: any) {
      toast.error(error.message || 'Failed to reprocess DLQ');
    }
  };

  return (
    <div className="flex h-full flex-col gap-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">DLQ</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Global failed-message inbox with original subjects, reasons, and reprocessing controls.
          </p>
          {topicFilter ? (
            <button
              type="button"
              onClick={() => setSearchParams({})}
              className="mt-3 inline-flex max-w-full cursor-pointer items-center gap-2 rounded-full border border-rose-500/25 bg-rose-500/10 px-3 py-1 text-xs text-rose-300 transition-colors hover:bg-rose-500/15"
            >
              <span className="truncate font-mono">{topicFilter}</span>
              <X className="h-3.5 w-3.5" />
            </button>
          ) : null}
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            <RefreshCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          </Button>
          <Button size="sm" onClick={reprocessAll} disabled={reprocessMutation.isPending}>
            <RotateCcw className={`h-4 w-4 ${reprocessMutation.isPending ? 'animate-spin' : ''}`} />
            Reprocess all
          </Button>
        </div>
      </div>

      <div className="grid gap-3 md:grid-cols-3">
        <Metric label="Failed messages" value={messages.length.toLocaleString()} />
        <Metric label="Current page" value={String(data?.pagination?.page ?? 1)} />
        <Metric label="Page size" value={String(data?.pagination?.limit ?? 100)} />
      </div>

      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Failed At</TableHead>
              <TableHead>Original Subject</TableHead>
              <TableHead>Reason</TableHead>
              <TableHead className="text-right">Sequence</TableHead>
              <TableHead className="text-right">Size</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              Array.from({ length: 6 }).map((_, index) => (
                <TableRow key={index}>
                  <TableCell colSpan={6}>
                    <div className="h-6 animate-pulse rounded bg-muted" />
                  </TableCell>
                </TableRow>
              ))
            ) : messages.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="h-40 text-center text-sm text-muted-foreground">
                  <Inbox className="mx-auto mb-3 h-8 w-8 opacity-50" />
                  DLQ is clean.
                </TableCell>
              </TableRow>
            ) : (
              messages.map((message) => (
                <TableRow
                  key={`${message.subject}-${message.sequence}`}
                  className="cursor-pointer"
                  onClick={() => setSelectedMessage(message)}
                >
                  <TableCell className="whitespace-nowrap text-xs">{formatTime(message.timestamp)}</TableCell>
                  <TableCell className="max-w-[420px] truncate font-mono text-xs">
                    {message.original_subject || message.headers?.['X-DLQ-Original-Subject'] || '-'}
                  </TableCell>
                  <TableCell className="max-w-[320px] truncate text-xs">
                    {message.reason || message.headers?.['X-DLQ-Reason'] || '-'}
                  </TableCell>
                  <TableCell className="text-right font-mono text-xs">{message.sequence}</TableCell>
                  <TableCell className="text-right text-xs">{formatBytes(messageSize(message.data))}</TableCell>
                  <TableCell><StatusBadge status="dlq" /></TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <MessageDetailSheet message={selectedMessage} onOpenChange={(open) => !open && setSelectedMessage(null)} />
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{label}</div>
      <div className="mt-2 text-2xl font-semibold text-foreground">{value}</div>
    </div>
  );
}
