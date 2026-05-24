import { useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import { RefreshCw, RadioTower, X } from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useConsumers } from '@/lib/query/explorer';
import { StatusBadge } from '../shared';

export default function ExplorerConsumersPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const topicFilter = searchParams.get('topic') || '';
  const { data, isLoading, isFetching, refetch } = useConsumers(1, 100);
  const consumers = useMemo(() => {
    const rows = data?.data ?? [];
    if (!topicFilter) return rows;
    return rows.filter((consumer) =>
      consumer.filter_subjects.some((subject) => subject.startsWith(topicFilter)),
    );
  }, [data, topicFilter]);

  return (
    <div className="flex h-full flex-col gap-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">Consumers</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Global flow consumers, filter subjects, pending messages, and ack lag.
          </p>
          {topicFilter ? (
            <button
              type="button"
              onClick={() => setSearchParams({})}
              className="mt-3 inline-flex max-w-full cursor-pointer items-center gap-2 rounded-full border border-sky-500/25 bg-sky-500/10 px-3 py-1 text-xs text-sky-300 transition-colors hover:bg-sky-500/15"
            >
              <span className="truncate font-mono">{topicFilter}</span>
              <X className="h-3.5 w-3.5" />
            </button>
          ) : null}
        </div>
        <Button variant="outline" size="sm" onClick={() => refetch()}>
          <RefreshCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
        </Button>
      </div>

      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Consumer</TableHead>
              <TableHead>Filter Subjects</TableHead>
              <TableHead className="text-right">Pending</TableHead>
              <TableHead className="text-right">Ack Pending</TableHead>
              <TableHead className="text-right">Delivered Seq</TableHead>
              <TableHead className="text-right">Ack Floor</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              Array.from({ length: 5 }).map((_, index) => (
                <TableRow key={index}>
                  <TableCell colSpan={7}>
                    <div className="h-6 animate-pulse rounded bg-muted" />
                  </TableCell>
                </TableRow>
              ))
            ) : consumers.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="h-40 text-center text-sm text-muted-foreground">
                  <RadioTower className="mx-auto mb-3 h-8 w-8 opacity-50" />
                  No active flow consumers.
                </TableCell>
              </TableRow>
            ) : (
              consumers.map((consumer) => {
                const lagging = consumer.num_pending > 0 || consumer.num_ack_pending > 0;
                return (
                  <TableRow key={consumer.name}>
                    <TableCell className="font-mono text-xs font-semibold">{consumer.name}</TableCell>
                    <TableCell>
                      <div className="space-y-1">
                        {consumer.filter_subjects.map((subject) => (
                          <div key={subject} className="font-mono text-[11px] text-muted-foreground">
                            {subject}
                          </div>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell className="text-right">{consumer.num_pending.toLocaleString()}</TableCell>
                    <TableCell className="text-right">{consumer.num_ack_pending.toLocaleString()}</TableCell>
                    <TableCell className="text-right font-mono text-xs">{consumer.delivered_stream_seq}</TableCell>
                    <TableCell className="text-right font-mono text-xs">{consumer.ack_floor_stream_seq}</TableCell>
                    <TableCell>
                      <StatusBadge status={lagging ? 'lagging' : 'active'} />
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
