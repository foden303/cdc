import { useMemo, useState } from 'react';
import { Link, useParams, useSearchParams } from 'react-router-dom';
import { ArrowLeft, GitBranch, RefreshCw, Search } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ROUTES } from '@/config/routes';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useMessages, usePartitions, useTopics } from '@/lib/query/explorer';
import type { MessageItem } from '@/types/api';
import { MessageDetailSheet } from '../components/MessageDetailSheet';
import { formatBytes, formatTime, messageSize, parseSubject, StatusBadge } from '../shared';

export default function ExplorerMessagesPage() {
  const routeParams = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const routeTopic = routeParams.topic ? decodeURIComponent(routeParams.topic) : '';
  const routePartition = routeParams.partition ? decodeURIComponent(routeParams.partition) : '';
  const isNested = !!routeTopic && !!routePartition;
  const initialTopic = routeTopic || searchParams.get('topic') || '';
  const [topic, setTopic] = useState(initialTopic);
  const [partition, setPartition] = useState(routePartition || 'all');
  const [search, setSearch] = useState('');
  const [page] = useState(1);
  const [selectedMessage, setSelectedMessage] = useState<MessageItem | null>(null);

  const activeTopic = routeTopic || topic;
  const activePartition = routePartition || partition;
  const topicDetailPath = activeTopic
    ? ROUTES.EXPLORER_TOPIC_DETAIL.replace(':topic', encodeURIComponent(activeTopic))
    : ROUTES.EXPLORER_TOPICS;

  const { data: topicsData } = useTopics(1, 100);
  const { data: partitionsData } = usePartitions(activeTopic, 1, 100);
  const { data, isLoading, isFetching, refetch } = useMessages({
    topic: activeTopic || undefined,
    partition: activePartition === 'all' ? undefined : activePartition,
    page,
    limit: 50,
  });

  const messages = useMemo(() => {
    const rows = data?.data ?? [];
    const q = search.trim().toLowerCase();
    if (!q) return rows;
    return rows.filter((message) =>
      `${message.subject} ${message.sequence}`.toLowerCase().includes(q),
    );
  }, [data, search]);

  const selectTopic = (value: string) => {
    setTopic(value);
    setPartition('all');
    if (value) setSearchParams({ topic: value });
    else setSearchParams({});
  };

  return (
    <div className="flex h-full flex-col gap-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          {isNested ? (
            <Link
              to={topicDetailPath}
              className="mb-3 -ml-2 inline-flex h-7 items-center gap-1 rounded-lg px-2.5 text-[0.8rem] font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            >
              <ArrowLeft className="h-4 w-4" />
              Topic partitions
            </Link>
          ) : null}
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">
            {isNested ? `Partition ${activePartition}` : 'Messages'}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {isNested
              ? 'Ordered CDC messages for the selected topic partition.'
              : 'Inspect CDC payloads, headers, metadata, and stream sequence.'}
          </p>
          {activeTopic ? (
            <div className="mt-2 flex flex-wrap items-center gap-2 font-mono text-xs text-muted-foreground">
              <GitBranch className="h-3.5 w-3.5" />
              <span className="max-w-[720px] truncate">{activeTopic}</span>
            </div>
          ) : null}
        </div>
        <Button variant="outline" size="sm" onClick={() => refetch()}>
          <RefreshCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
        </Button>
      </div>

      <div className={`grid gap-3 ${isNested ? 'lg:grid-cols-[minmax(240px,1fr)]' : 'lg:grid-cols-[minmax(240px,1fr)_minmax(180px,240px)_minmax(240px,1fr)]'}`}>
        {!isNested ? (
          <>
            <select
              value={topic}
              onChange={(event) => selectTopic(event.target.value)}
              className="h-9 rounded-md border border-input bg-background px-3 text-sm text-foreground"
            >
              <option value="">All topics</option>
              {(topicsData?.data ?? []).map((item) => (
                <option key={item.name} value={item.name}>
                  {item.name}
                </option>
              ))}
            </select>
            <select
              value={partition}
              onChange={(event) => setPartition(event.target.value)}
              className="h-9 rounded-md border border-input bg-background px-3 text-sm text-foreground"
              disabled={!activeTopic}
            >
              <option value="all">All partitions</option>
              {(partitionsData?.data ?? []).map((item) => (
                <option key={item.id} value={item.id}>
                  {item.id}
                </option>
              ))}
            </select>
          </>
        ) : null}
        <div className="relative">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            className="pl-9"
            placeholder="Search subject or sequence"
          />
        </div>
      </div>

      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Time</TableHead>
              <TableHead>Topic</TableHead>
              <TableHead>Partition</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Sequence</TableHead>
              <TableHead className="text-right">Size</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              Array.from({ length: 8 }).map((_, index) => (
                <TableRow key={index}>
                  <TableCell colSpan={6}>
                    <div className="h-6 animate-pulse rounded bg-muted" />
                  </TableCell>
                </TableRow>
              ))
            ) : messages.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="h-40 text-center text-sm text-muted-foreground">
                  No messages match these filters.
                </TableCell>
              </TableRow>
            ) : (
              messages.map((message) => {
                const parsed = parseSubject(message.subject);
                return (
                  <TableRow
                    key={`${message.subject}-${message.sequence}`}
                    className="cursor-pointer"
                    onClick={() => setSelectedMessage(message)}
                  >
                    <TableCell className="whitespace-nowrap text-xs">{formatTime(message.timestamp)}</TableCell>
                    <TableCell>
                      <div className="font-mono text-xs">{parsed.shortName}</div>
                      <div className="max-w-[360px] truncate font-mono text-[11px] text-muted-foreground">
                        {message.subject}
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-xs">{parsed.partition || '-'}</TableCell>
                    <TableCell><StatusBadge status="sent" /></TableCell>
                    <TableCell className="text-right font-mono text-xs">{message.sequence}</TableCell>
                    <TableCell className="text-right text-xs">{formatBytes(messageSize(message.data))}</TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </div>

      <MessageDetailSheet message={selectedMessage} onOpenChange={(open) => !open && setSelectedMessage(null)} />
    </div>
  );
}
