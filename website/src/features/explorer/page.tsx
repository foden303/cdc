import { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Search,
  RefreshCw,
  ChevronRight,
  Info,
  Layers,
  Database,
  ArrowRightLeft,
  CornerDownRight,
  Play,
  Copy,
  Check,
  AlertTriangle,
} from 'lucide-react';
import { toast } from 'sonner';

import { useTopics, usePartitions, useMessages, useReprocessDLQ } from '@/lib/query/explorer';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { JsonViewer } from '@/components/ui/json-viewer';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from '@/components/ui/sheet';
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table';
import type { MessageItem } from '@/types/api';

/** Decodes UTF-8 base64 safely */
function decodePayload(base64Str: string): string {
  if (!base64Str) return '';
  try {
    return decodeURIComponent(
      atob(base64Str)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );
  } catch {
    try {
      return atob(base64Str);
    } catch {
      return base64Str;
    }
  }
}

export default function ExplorerPage() {
  const { t } = useTranslation();
  const [topicSearch, setTopicSearch] = useState('');
  const [selectedTopic, setSelectedTopic] = useState<string | null>(null);
  const [selectedPartition, setSelectedPartition] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<number>(0); // 0 = all, 1 = DLQ/failed
  const [currentPage, setCurrentPage] = useState(1);
  const [selectedMessage, setSelectedMessage] = useState<MessageItem | null>(null);
  const [copiedText, setCopiedText] = useState(false);

  // Queries
  const { data: topicsData, isLoading: topicsLoading, refetch: refetchTopics } = useTopics();
  const { data: partitionsData, isLoading: partitionsLoading } = usePartitions(
    selectedTopic || '',
  );

  const messagesParams = useMemo(() => {
    return {
      topic: selectedTopic || undefined,
      partition: selectedPartition === 'all' ? undefined : selectedPartition,
      status: statusFilter === 1 ? 2 : undefined, // Assuming status 2 maps to failed/DLQ in backend
      page: currentPage,
      limit: 25,
    };
  }, [selectedTopic, selectedPartition, statusFilter, currentPage]);

  const {
    data: messagesData,
    isLoading: messagesLoading,
    refetch: refetchMessages,
    isFetching: messagesFetching,
  } = useMessages(messagesParams);

  const reprocessMutation = useReprocessDLQ();

  // Filter topics
  const filteredTopics = useMemo(() => {
    if (!topicsData?.data) return [];
    return topicsData.data.filter((topic) =>
      topic.name.toLowerCase().includes(topicSearch.toLowerCase()),
    );
  }, [topicsData, topicSearch]);

  const handleTopicSelect = (topicName: string) => {
    setSelectedTopic(topicName);
    setSelectedPartition('all');
    setCurrentPage(1);
  };

  const handleReprocessAll = async () => {
    try {
      const res = await reprocessMutation.mutateAsync();
      toast.success(t('explorer.reprocessSuccess') + ` (${res.count || 0} messages)`);
    } catch {
      toast.error(t('explorer.reprocessFailed'));
    }
  };

  const handleCopyText = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedText(true);
      setTimeout(() => setCopiedText(false), 2000);
    } catch (err) {
      console.error(err);
    }
  };

  // Process selected message info
  const messageDetail = useMemo(() => {
    if (!selectedMessage) return null;
    const raw = decodePayload(selectedMessage.data);
    let json: unknown = null;
    let isJson = false;
    try {
      json = JSON.parse(raw);
      isJson = true;
    } catch {
      // ignore
    }
    return { raw, json, isJson };
  }, [selectedMessage]);

  return (
    <div className="flex h-[calc(100vh-8rem)] gap-6 overflow-hidden">
      {/* LEFT COLUMN: Topics & Partitions */}
      <div className="flex w-80 flex-col rounded-xl border border-slate-800 bg-slate-950 p-4">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-slate-200 flex items-center gap-2">
            <Layers className="h-4 w-4 text-sky-400" />
            {t('explorer.topics')}
          </h2>
          <button
            onClick={() => refetchTopics()}
            className="rounded p-1 hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition-colors"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${topicsLoading ? 'animate-spin' : ''}`} />
          </button>
        </div>

        {/* Search */}
        <div className="relative mb-4">
          <Search className="absolute top-2.5 left-2.5 h-4 w-4 text-slate-500" />
          <Input
            value={topicSearch}
            onChange={(e) => setTopicSearch(e.target.value)}
            placeholder={t('explorer.searchPlaceholder')}
            className="pl-9 h-9 border-slate-800 bg-slate-900/50 text-xs"
          />
        </div>

        {/* Topics List */}
        <div className="flex-1 overflow-y-auto space-y-1.5 pr-1 select-none scrollbar-thin">
          {topicsLoading ? (
            Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="h-14 rounded-lg bg-slate-900/40 animate-pulse border border-slate-900" />
            ))
          ) : filteredTopics.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <Info className="h-5 w-5 text-slate-600 mb-2" />
              <span className="text-xs text-slate-500">{t('common.noData')}</span>
            </div>
          ) : (
            filteredTopics.map((topic) => {
              const isSelected = selectedTopic === topic.name;
              return (
                <button
                  key={topic.name}
                  onClick={() => handleTopicSelect(topic.name)}
                  className={`w-full text-left p-3 rounded-lg border transition-all duration-200 cursor-pointer ${
                    isSelected
                      ? 'border-sky-500/30 bg-sky-950/20 text-slate-100 shadow-sm shadow-sky-950/45'
                      : 'border-slate-900 bg-slate-900/10 hover:border-slate-800 hover:bg-slate-900/40 text-slate-400 hover:text-slate-200'
                  }`}
                >
                  <div className="flex items-center justify-between font-mono text-xs font-semibold truncate mb-1">
                    <span className="truncate">{topic.name}</span>
                    <ChevronRight className={`h-3 w-3 opacity-60 transition-transform ${isSelected ? 'rotate-90 text-sky-400' : ''}`} />
                  </div>
                  <div className="flex items-center gap-3 text-[10px] text-slate-500">
                    <span className="flex items-center gap-1">
                      <Database className="h-3 w-3" />
                      {topic.partition_count} {t('explorer.partitionsCount')}
                    </span>
                    <span>{topic.message_count} {t('explorer.msgsCount')}</span>
                  </div>
                </button>
              );
            })
          )}
        </div>

        {/* Partition Filter inside Selected Topic */}
        {selectedTopic && (
          <div className="mt-4 pt-4 border-t border-slate-900">
            <span className="text-[10px] font-semibold uppercase tracking-wider text-slate-500 block mb-2">
              {t('explorer.partitions')}
            </span>
            <div className="flex flex-wrap gap-1.5">
              <button
                onClick={() => setSelectedPartition('all')}
                className={`px-2.5 py-1 rounded text-[10px] font-semibold border transition-all cursor-pointer ${
                  selectedPartition === 'all'
                    ? 'bg-slate-800 border-slate-700 text-slate-100'
                    : 'bg-slate-900/20 border-slate-900 hover:border-slate-800 text-slate-400 hover:text-slate-200'
                }`}
              >
                {t('explorer.allPartition')}
              </button>
              {partitionsLoading ? (
                <div className="h-5 w-12 rounded bg-slate-900 animate-pulse" />
              ) : (
                partitionsData?.data?.map((p) => {
                  const isPSelected = selectedPartition === p.id;
                  return (
                    <button
                      key={p.id}
                      onClick={() => setSelectedPartition(p.id)}
                      className={`px-2.5 py-1 rounded text-[10px] font-semibold border transition-all cursor-pointer ${
                        isPSelected
                          ? 'bg-sky-500/20 border-sky-500/30 text-sky-400'
                          : 'bg-slate-900/20 border-slate-900 hover:border-slate-800 text-slate-400 hover:text-slate-200'
                      }`}
                    >
                      p{p.id}
                    </button>
                  );
                })
              )}
            </div>
          </div>
        )}
      </div>

      {/* RIGHT COLUMN: Messages Table & Actions */}
      <div className="flex-1 flex flex-col rounded-xl border border-slate-800 bg-slate-950 overflow-hidden">
        {/* Toolbar */}
        <div className="flex items-center justify-between border-b border-slate-900 bg-slate-900/20 p-4">
          <div className="flex items-center gap-3">
            <h3 className="text-sm font-semibold text-slate-200 font-mono">
              {selectedTopic ? selectedTopic : t('explorer.title')}
            </h3>
            {selectedTopic && (
              <span className="inline-flex items-center gap-1.5 rounded-full bg-slate-900 px-2 py-0.5 text-[10px] font-medium text-slate-400 border border-slate-800">
                <CornerDownRight className="h-3 w-3" />
                {selectedPartition === 'all' ? t('explorer.allPartition') : `p${selectedPartition}`}
              </span>
            )}
          </div>

          <div className="flex items-center gap-2">
            {/* Filter buttons */}
            <div className="flex items-center rounded-lg bg-slate-900 p-0.5 border border-slate-800">
              <button
                onClick={() => setStatusFilter(0)}
                className={`rounded px-2.5 py-1 text-xs transition-colors cursor-pointer ${
                  statusFilter === 0
                    ? 'bg-slate-800 text-slate-100 font-medium'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                {t('explorer.all')}
              </button>
              <button
                onClick={() => setStatusFilter(1)}
                className={`rounded px-2.5 py-1 text-xs transition-colors flex items-center gap-1 cursor-pointer ${
                  statusFilter === 1
                    ? 'bg-red-500/20 text-red-400 font-medium border border-red-500/10'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                <AlertTriangle className="h-3 w-3" />
                DLQ
              </button>
            </div>

            {/* Reprocess all DLQ */}
            {statusFilter === 1 && (
              <Button
                variant="destructive"
                onClick={handleReprocessAll}
                disabled={reprocessMutation.isPending}
                size="sm"
                className="h-8 text-xs font-semibold cursor-pointer"
              >
                <Play className="h-3.5 w-3.5 mr-1" />
                {t('dashboard.reprocessDlq')}
              </Button>
            )}

            <button
              onClick={() => refetchMessages()}
              disabled={!selectedTopic || messagesLoading}
              className="h-8 w-8 inline-flex items-center justify-center rounded-lg border border-slate-800 hover:bg-slate-900 text-slate-400 hover:text-slate-200 disabled:opacity-40 transition-colors cursor-pointer"
            >
              <RefreshCw className={`h-3.5 w-3.5 ${messagesFetching ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>

        {/* Messages Body */}
        <div className="flex-1 overflow-auto">
          {!selectedTopic ? (
            <div className="flex flex-col items-center justify-center h-full py-16 text-center">
              <div className="p-3 rounded-full bg-slate-900/60 border border-slate-800 mb-3 animate-pulse">
                <ArrowRightLeft className="h-6 w-6 text-slate-600" />
              </div>
              <h4 className="text-sm font-semibold text-slate-400 mb-1">{t('explorer.selectTopic')}</h4>
              <p className="text-xs text-slate-600 max-w-xs">{t('explorer.selectTopic')}</p>
            </div>
          ) : messagesLoading ? (
            <div className="p-6 space-y-4">
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className="flex gap-4 items-center">
                  <div className="h-4 w-12 bg-slate-900 rounded animate-pulse" />
                  <div className="h-4 w-32 bg-slate-900 rounded animate-pulse" />
                  <div className="h-4 flex-1 bg-slate-900 rounded animate-pulse" />
                </div>
              ))}
            </div>
          ) : !messagesData?.data || messagesData.data.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full py-16 text-center">
              <Info className="h-6 w-6 text-slate-700 mb-2" />
              <span className="text-xs text-slate-500">{t('explorer.noMessages')}</span>
            </div>
          ) : (
            <Table className="w-full text-xs">
              <TableHeader>
                <TableRow className="border-b border-slate-900 bg-slate-900/10 text-slate-500 select-none font-semibold hover:bg-transparent">
                  <TableHead className="px-4 py-3 font-mono w-24 text-slate-500">{t('explorer.offset')}</TableHead>
                  <TableHead className="px-4 py-3 w-40 text-slate-500">{t('explorer.timestamp')}</TableHead>
                  <TableHead className="px-4 py-3 w-28 text-slate-500">{t('manager.sources.card.subject')}</TableHead>
                  <TableHead className="px-4 py-3 text-slate-500">{t('explorer.payloadPreview')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {messagesData.data.map((msg) => {
                  const decoded = decodePayload(msg.data);
                  // Preview limits to 120 chars
                  const preview = decoded.length > 120 ? `${decoded.substring(0, 120)}...` : decoded;

                  return (
                    <TableRow
                      key={msg.sequence}
                      onClick={() => setSelectedMessage(msg)}
                      className="hover:bg-slate-900/30 border-b border-slate-900 transition-colors cursor-pointer group"
                    >
                      <TableCell className="px-4 py-3 font-mono text-slate-400">{msg.sequence}</TableCell>
                      <TableCell className="px-4 py-3 text-slate-500">
                        {new Date(msg.timestamp).toLocaleString()}
                      </TableCell>
                      <TableCell className="px-4 py-3">
                        <span className="inline-block px-1.5 py-0.5 rounded bg-slate-900 text-slate-400 font-mono text-[10px] border border-slate-800 truncate max-w-[200px]">
                          {msg.subject}
                        </span>
                      </TableCell>
                      <TableCell className="px-4 py-3 text-slate-400 font-mono truncate max-w-lg group-hover:text-slate-200">
                        {preview}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </div>

        {/* Pagination Footer */}
        {selectedTopic && messagesData?.pagination && (
          <div className="flex items-center justify-between border-t border-slate-900 bg-slate-900/10 px-4 py-3">
            <span className="text-[10px] text-slate-500">
              {t('explorer.totalMessages', { count: messagesData.total_count || messagesData.pagination.total_rows || 0 })}
            </span>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage === 1 || messagesLoading}
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                className="h-7 text-[10px] border-slate-800 bg-slate-950 text-slate-400 hover:text-slate-200 cursor-pointer"
              >
                {t('common.previous')}
              </Button>
              <span className="flex items-center px-3 text-[10px] font-mono text-slate-400">
                {t('common.page', { page: currentPage })}
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={!messagesData.pagination.has_next || messagesLoading}
                onClick={() => setCurrentPage((p) => p + 1)}
                className="h-7 text-[10px] border-slate-800 bg-slate-950 text-slate-400 hover:text-slate-200 cursor-pointer"
              >
                {t('common.next')}
              </Button>
            </div>
          </div>
        )}
      </div>

      {/* MESSAGE DETAILS SHEET */}
      <Sheet open={!!selectedMessage} onOpenChange={(open) => !open && setSelectedMessage(null)}>
        <SheetContent side="right" className="w-[500px] sm:max-w-[550px] border-slate-800 bg-slate-950 text-slate-300 p-0 flex flex-col h-full shadow-2xl shadow-black/80">
          <SheetHeader className="p-6 pb-4 border-b border-slate-900 bg-slate-950">
            <div className="flex items-center justify-between">
              <div>
                <SheetTitle className="text-base font-bold text-slate-100 font-mono flex items-center gap-2">
                  {t('explorer.messageDetails')}
                </SheetTitle>
                <SheetDescription className="text-xs text-slate-500 font-mono mt-1">
                  {t('explorer.sequenceOffset')} {selectedMessage?.sequence}
                </SheetDescription>
              </div>
            </div>
          </SheetHeader>

          {/* Details Scroll Area */}
          <div className="flex-1 overflow-y-auto p-6 space-y-6 scrollbar-thin">
            {/* Metadata Summary */}
            <div className="grid grid-cols-2 gap-4 rounded-xl border border-slate-900 bg-slate-900/20 p-4">
              <div>
                <span className="text-[10px] uppercase font-semibold text-slate-500 tracking-wider">{t('manager.sources.card.subject')}</span>
                <span className="block text-xs font-mono text-slate-300 font-semibold truncate mt-0.5">{selectedMessage?.subject}</span>
              </div>
              <div>
                <span className="text-[10px] uppercase font-semibold text-slate-500 tracking-wider">{t('explorer.timestamp')}</span>
                <span className="block text-xs text-slate-300 font-semibold mt-0.5">
                  {selectedMessage?.timestamp ? new Date(selectedMessage.timestamp).toLocaleString() : '-'}
                </span>
              </div>
            </div>

            {/* Error alerts if message carries error headers */}
            {selectedMessage?.headers?.['x-error'] && (
              <div className="rounded-xl border border-red-500/20 bg-red-500/5 p-4 flex gap-3 text-xs text-red-400">
                <AlertTriangle className="h-4 w-4 shrink-0 mt-0.5 text-red-400" />
                <div className="space-y-1">
                  <span className="font-semibold text-red-300">{t('explorer.reprocessError')}</span>
                  <p className="font-mono break-all text-[11px] leading-relaxed opacity-90">
                    {selectedMessage.headers['x-error']}
                  </p>
                </div>
              </div>
            )}

            {/* Decoded JSON / Plain Text Content */}
            <div className="space-y-2">
              <span className="text-[10px] uppercase font-semibold text-slate-500 tracking-wider block">{t('explorer.decodedPayload')}</span>
              {messageDetail?.isJson ? (
                <JsonViewer data={messageDetail.json} />
              ) : (
                <div className="relative group rounded-lg border border-slate-800 bg-slate-950 font-mono text-xs text-slate-300 overflow-hidden">
                  <div className="flex items-center justify-between px-4 py-2 border-b border-slate-900 bg-slate-900/60 select-none">
                    <span className="text-[10px] uppercase tracking-wider text-slate-500 font-semibold">{t('explorer.plainText')}</span>
                    <button
                      onClick={() => handleCopyText(messageDetail?.raw || '')}
                      className="p-1.5 rounded-md hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition-colors cursor-pointer"
                    >
                      {copiedText ? (
                        <Check className="h-3.5 w-3.5 text-emerald-500" />
                      ) : (
                        <Copy className="h-3.5 w-3.5" />
                      )}
                    </button>
                  </div>
                  <pre className="p-4 overflow-auto max-h-[300px] whitespace-pre-wrap break-all leading-relaxed">
                    {messageDetail?.raw}
                  </pre>
                </div>
              )}
            </div>

            {/* Headers Section */}
            {selectedMessage?.headers && Object.keys(selectedMessage.headers).length > 0 && (
              <div className="space-y-2">
                <span className="text-[10px] uppercase font-semibold text-slate-500 tracking-wider block">{t('explorer.messageHeaders')}</span>
                <Table className="w-full text-xs">
                  <TableHeader>
                    <TableRow className="bg-slate-900/50 text-slate-500 font-semibold border-b border-slate-900 hover:bg-transparent">
                      <TableHead className="px-4 py-2 font-mono text-[10px] h-auto text-slate-500">{t('explorer.headerKey')}</TableHead>
                      <TableHead className="px-4 py-2 font-mono text-[10px] h-auto text-slate-500">{t('explorer.headerValue')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {Object.entries(selectedMessage.headers).map(([key, val]) => (
                      <TableRow key={key} className="hover:bg-slate-900/10 border-b border-slate-900">
                        <TableCell className="px-4 py-2.5 text-sky-400/80 font-medium truncate max-w-[150px]">{key}</TableCell>
                        <TableCell className="px-4 py-2.5 text-slate-400 break-all">{val}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}

            {/* Raw Bytes */}
            <div className="space-y-2">
              <span className="text-[10px] uppercase font-semibold text-slate-500 tracking-wider block">{t('explorer.rawBase64')}</span>
              <div className="relative rounded-lg border border-slate-900 bg-slate-950/40 p-4 font-mono text-[10px] break-all leading-normal text-slate-500 select-all max-h-24 overflow-y-auto">
                {selectedMessage?.data}
              </div>
            </div>
          </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}
