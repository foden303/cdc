import { useMemo, useState } from 'react';
import { Copy, Database, Eye, RefreshCw, Search } from 'lucide-react';
import { toast } from 'sonner';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ROUTES } from '@/config/routes';
import { useTopics } from '@/lib/query/explorer';
import { parseSubject } from '../shared';

export default function ExplorerTopicsPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [search, setSearch] = useState('');
  const { data, isLoading, isFetching, refetch } = useTopics(1, 100);

  const topics = useMemo(() => {
    const items = data?.data ?? [];
    const q = search.trim().toLowerCase();
    if (!q) return items;
    return items.filter((topic) => topic.name.toLowerCase().includes(q));
  }, [data, search]);

  const copy = async (value: string) => {
    await navigator.clipboard.writeText(value);
    toast.success(t('explorer.topicCopied'));
  };

  return (
    <div className="flex h-full flex-col gap-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">
            {t('explorer.topics')}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {t('explorer.topicsDesc')}
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={() => refetch()}>
          <RefreshCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
        </Button>
      </div>

      <div className="relative max-w-md">
        <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          className="pl-9"
          placeholder={t('explorer.searchTopicsDetailed')}
        />
      </div>

      <div className="overflow-hidden rounded-lg border border-border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('explorer.topic')}</TableHead>
              <TableHead>{t('explorer.source')}</TableHead>
              <TableHead>{t('explorer.schema')}</TableHead>
              <TableHead>{t('explorer.table')}</TableHead>
              <TableHead className="text-right">{t('explorer.partitions')}</TableHead>
              <TableHead className="text-right">{t('explorer.messages')}</TableHead>
              <TableHead className="w-24 text-right">{t('explorer.actions')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              Array.from({ length: 6 }).map((_, index) => (
                <TableRow key={index}>
                  <TableCell colSpan={7}>
                    <div className="h-6 animate-pulse rounded bg-muted" />
                  </TableCell>
                </TableRow>
              ))
            ) : topics.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="h-40 text-center text-sm text-muted-foreground">
                  <Database className="mx-auto mb-3 h-8 w-8 opacity-50" />
                  {t('explorer.noTopics')}
                </TableCell>
              </TableRow>
            ) : (
              topics.map((topic) => {
                const parsed = parseSubject(topic.name);
                return (
                  <TableRow key={topic.name}>
                    <TableCell>
                      <div className="font-mono text-xs font-semibold text-foreground">
                        {parsed.shortName}
                      </div>
                      <div className="mt-1 max-w-[360px] truncate font-mono text-[11px] text-muted-foreground">
                        {topic.name}
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-xs">{parsed.sourceId || '-'}</TableCell>
                    <TableCell>{parsed.schema || '-'}</TableCell>
                    <TableCell>{parsed.table || '-'}</TableCell>
                    <TableCell className="text-right">{topic.partition_count}</TableCell>
                    <TableCell className="text-right">{topic.message_count.toLocaleString()}</TableCell>
                    <TableCell>
                      <div className="flex justify-end gap-1">
                        <Button variant="ghost" size="icon-sm" onClick={() => copy(topic.name)}>
                          <Copy className="h-3.5 w-3.5" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() =>
                            navigate(
                              ROUTES.EXPLORER_TOPIC_DETAIL.replace(
                                ':topic',
                                encodeURIComponent(topic.name),
                              ),
                            )
                          }
                        >
                          <Eye className="h-3.5 w-3.5" />
                        </Button>
                      </div>
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
