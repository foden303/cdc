import { useMemo } from 'react';
import type { ReactNode } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  AlertTriangle,
  ArrowLeft,
  ChevronRight,
  Copy,
  Database,
  GitBranch,
  RadioTower,
  RefreshCw,
} from 'lucide-react';
import { toast } from 'sonner';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ROUTES } from '@/config/routes';
import { usePartitions } from '@/lib/query/explorer';
import { parseSubject } from '../shared';

function partitionPath(topic: string, partition: string) {
  return ROUTES.EXPLORER_TOPIC_PARTITION.replace(':topic', encodeURIComponent(topic)).replace(
    ':partition',
    encodeURIComponent(partition),
  );
}

export default function ExplorerTopicDetailPage() {
  const navigate = useNavigate();
  const params = useParams();
  const topic = decodeURIComponent(params.topic ?? '');
  const parsed = useMemo(() => parseSubject(topic), [topic]);
  const { data, isLoading, isFetching, refetch } = usePartitions(topic, 1, 100);
  const partitions = data?.data ?? [];

  const copyTopic = async () => {
    await navigator.clipboard.writeText(topic);
    toast.success('Topic copied');
  };

  if (!topic) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        Missing topic.
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col gap-5">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <Button
            variant="ghost"
            size="sm"
            className="mb-3 -ml-2 text-muted-foreground"
            onClick={() => navigate(ROUTES.EXPLORER_TOPICS)}
          >
            <ArrowLeft className="h-4 w-4" />
            Topics
          </Button>
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-2xl font-semibold tracking-tight text-foreground">
              {parsed.shortName}
            </h1>
            <Badge variant="outline" className="border-sky-500/25 bg-sky-500/10 text-sky-300">
              Topic
            </Badge>
          </div>
          <p className="mt-1 max-w-4xl truncate font-mono text-xs text-muted-foreground">
            {topic}
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={copyTopic}>
            <Copy className="h-4 w-4" />
            Copy
          </Button>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            <RefreshCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>
      </div>

      <div className="grid gap-3 md:grid-cols-4">
        <Metric label="Source" value={parsed.sourceId || '-'} icon={<Database className="h-4 w-4" />} />
        <Metric label="Schema" value={parsed.schema || '-'} icon={<GitBranch className="h-4 w-4" />} />
        <Metric label="Table" value={parsed.table || '-'} icon={<Database className="h-4 w-4" />} />
        <Metric label="Partitions" value={String(partitions.length)} icon={<GitBranch className="h-4 w-4" />} />
      </div>

      <div className="grid gap-4 lg:grid-cols-[1fr_320px]">
        <Card className="overflow-hidden">
          <CardHeader>
            <CardTitle>Partitions</CardTitle>
            <CardDescription>Pick a partition to inspect ordered CDC messages.</CardDescription>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Partition</TableHead>
                  <TableHead className="text-right">Messages</TableHead>
                  <TableHead className="w-20 text-right">Open</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading ? (
                  Array.from({ length: 4 }).map((_, index) => (
                    <TableRow key={index}>
                      <TableCell colSpan={3}>
                        <div className="h-6 animate-pulse rounded bg-muted" />
                      </TableCell>
                    </TableRow>
                  ))
                ) : partitions.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={3} className="h-36 text-center text-sm text-muted-foreground">
                      No partitions found for this topic.
                    </TableCell>
                  </TableRow>
                ) : (
                  partitions.map((partition) => (
                    <TableRow
                      key={partition.id}
                      className="cursor-pointer"
                      onClick={() => navigate(partitionPath(topic, partition.id))}
                    >
                      <TableCell>
                        <div className="font-mono text-sm font-semibold text-foreground">
                          Partition {partition.id}
                        </div>
                        <div className="mt-1 font-mono text-[11px] text-muted-foreground">
                          {topic}.{partition.id}
                        </div>
                      </TableCell>
                      <TableCell className="text-right">{partition.message_count.toLocaleString()}</TableCell>
                      <TableCell className="text-right">
                        <ChevronRight className="ml-auto h-4 w-4 text-muted-foreground" />
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <div className="space-y-3">
          <ContextLink
            icon={<RadioTower className="h-4 w-4" />}
            title="Related consumers"
            description="Flow workers currently reading this topic."
            to={`${ROUTES.EXPLORER_CONSUMERS}?topic=${encodeURIComponent(topic)}`}
          />
          <ContextLink
            icon={<AlertTriangle className="h-4 w-4" />}
            title="Topic DLQ"
            description="Failed messages whose original subject belongs to this topic."
            to={`${ROUTES.EXPLORER_DLQ}?topic=${encodeURIComponent(topic)}`}
          />
        </div>
      </div>
    </div>
  );
}

function Metric({ label, value, icon }: { label: string; value: string; icon: ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
        {icon}
        {label}
      </div>
      <div className="mt-3 truncate font-mono text-sm font-semibold text-foreground">{value}</div>
    </div>
  );
}

function ContextLink({
  icon,
  title,
  description,
  to,
}: {
  icon: ReactNode;
  title: string;
  description: string;
  to: string;
}) {
  return (
    <Link
      to={to}
      className="block rounded-lg border border-border bg-card p-4 transition-colors hover:border-sky-500/40 hover:bg-muted/30"
    >
      <div className="flex items-start gap-3">
        <div className="rounded-md border border-sky-500/20 bg-sky-500/10 p-2 text-sky-300">
          {icon}
        </div>
        <div className="min-w-0">
          <div className="font-semibold text-foreground">{title}</div>
          <p className="mt-1 text-sm text-muted-foreground">{description}</p>
        </div>
      </div>
    </Link>
  );
}
