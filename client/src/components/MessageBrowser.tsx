"use client";

import { useEffect, useState, useCallback, Fragment } from "react";
import { listMessagesAction } from "@/lib/actions";
import type { MessageItem, PaginationResponse } from "@/lib/grpc";
import { 
  ChevronDown, 
  ChevronUp, 
  ChevronLeft, 
  ChevronRight, 
  RefreshCcw, 
  FileJson, 
  Copy, 
  Terminal,
  Fingerprint,
  Search,
  Box
} from "lucide-react";

interface MessageBrowserProps {
  topic?: string;
  partition?: string;
  limit?: number;
  compact?: boolean;
}

export default function MessageBrowser({ 
  topic = "", 
  partition = "", 
  limit: initialLimit = 25,
  compact = false
}: MessageBrowserProps) {
  const [messages, setMessages] = useState<MessageItem[]>([]);
  const [pagination, setPagination] = useState<PaginationResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [expandedSeq, setExpandedSeq] = useState<number | null>(null);

  const fetchMessages = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listMessagesAction(topic, partition, initialLimit, page);
      setMessages(res.data || []);
      setPagination(res.pagination || null);
    } catch (err) {
      console.error("fetch messages:", err);
    } finally {
      setLoading(false);
    }
  }, [topic, partition, initialLimit, page]);

  useEffect(() => {
    fetchMessages();
  }, [fetchMessages]);

  const totalPages = pagination ? Math.ceil(pagination.total_rows / initialLimit) : 1;

  const toggleExpand = (seq: number) => {
    setExpandedSeq(expandedSeq === seq ? null : seq);
  };

  const decodePayload = (data: any) => {
    if (!data) return "";
    try {
      if (typeof data === "string") {
        const binString = atob(data);
        const bytes = Uint8Array.from(binString, (m) => m.codePointAt(0)!);
        return new TextDecoder().decode(bytes);
      }
      return new TextDecoder().decode(data as Uint8Array);
    } catch (e) {
      return typeof data === "string" ? data : "Unable to decode binary payload.";
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  return (
    <div className="space-y-6">
      {/* Mini Control Bar */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-white/[0.03] border border-white/5 text-[10px] font-bold text-slate-500 uppercase tracking-widest">
            <span className="text-blue-400">{pagination?.total_rows ?? 0}</span> Events Captured
          </div>
        </div>
        <button 
          onClick={fetchMessages}
          className="p-2 rounded-xl border border-white/5 bg-white/[0.03] hover:bg-white/[0.08] transition-all"
        >
          <RefreshCcw className={`w-3.5 h-3.5 ${loading ? "animate-spin text-blue-400" : "text-slate-400"}`} />
        </button>
      </div>

      {/* Messages Table */}
      <div className="rounded-2xl border border-white/5 bg-white/[0.01] overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-white/5 text-[10px] text-slate-500 uppercase font-bold tracking-wider">
                <th className="px-6 py-4 text-center w-12">#</th>
                <th className="px-6 py-4 text-left">Sequence</th>
                {!compact && <th className="px-6 py-4 text-left">Subject</th>}
                <th className="px-6 py-4 text-left">Timestamp</th>
                <th className="px-6 py-4 text-center w-16">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/[0.01]">
              {loading && messages.length === 0 ? (
                [...Array(5)].map((_, i) => (
                  <tr key={i} className="animate-pulse">
                    {[...Array(compact ? 4 : 5)].map((__, j) => (
                      <td key={j} className="px-6 py-5">
                        <div className="h-4 rounded bg-white/[0.05]" />
                      </td>
                    ))}
                  </tr>
                ))
              ) : messages.length === 0 ? (
                <tr>
                  <td colSpan={compact ? 4 : 5} className="px-6 py-16 text-center text-slate-600">
                    <p className="text-xs font-medium">No messages captured under these filters.</p>
                  </td>
                </tr>
              ) : (
                messages.map((m) => {
                  const isExpanded = expandedSeq === m.sequence;
                  const rawPayload = decodePayload(m.data);
                  let formattedPayload = rawPayload;
                  try {
                    formattedPayload = JSON.stringify(JSON.parse(rawPayload), null, 2);
                  } catch {}

                  return (
                    <Fragment key={m.sequence}>
                      <tr 
                        onClick={() => toggleExpand(m.sequence)}
                        className={`group cursor-pointer transition-colors ${isExpanded ? "bg-blue-500/[0.06]" : "hover:bg-white/[0.01]"}`}
                      >
                        <td className="px-6 py-4 text-center">
                          {isExpanded ? <ChevronUp className="w-3.5 h-3.5 text-blue-400 m-auto" /> : <ChevronDown className="w-3.5 h-3.5 text-slate-700 m-auto" />}
                        </td>
                        <td className="px-6 py-4 font-mono text-xs font-bold text-white">{m.sequence}</td>
                        {!compact && <td className="px-6 py-4 font-mono text-xs text-slate-400">{m.subject}</td>}
                        <td className="px-6 py-4 text-slate-500 font-mono text-xs">
                          {new Date(parseInt(m.timestamp)).toLocaleString('en-US', { hour12: false })}
                        </td>
                        <td className="px-6 py-4 text-center">
                          <button 
                            onClick={(e) => { e.stopPropagation(); copyToClipboard(rawPayload); }}
                            className="p-1.5 rounded-lg border border-white/5 bg-white/[0.02] hover:bg-white/[0.1] text-slate-500 transition-all"
                          >
                            <Copy className="w-3.5 h-3.5" />
                          </button>
                        </td>
                      </tr>
                      {isExpanded && (
                        <tr className="bg-blue-500/[0.02]">
                          <td colSpan={compact ? 4 : 5} className="px-10 py-6">
                            <div className="space-y-4">
                              <div className="flex items-center justify-between">
                                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Payload Preview</span>
                                <button onClick={() => copyToClipboard(rawPayload)} className="text-[10px] font-bold text-blue-400 hover:text-blue-300">Copy JSON</button>
                              </div>
                              <pre className="p-4 rounded-xl bg-black/40 border border-white/5 font-mono text-[11px] leading-relaxed text-blue-100/90 overflow-x-auto max-h-[400px]">
                                {formattedPayload}
                              </pre>
                            </div>
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Pagination */}
      {pagination && totalPages > 1 && (
        <div className="flex items-center justify-center gap-4 py-2">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={!pagination.has_prev}
            className="p-2 rounded-xl border border-white/5 bg-white/[0.03] disabled:opacity-20 transition-all"
          >
            <ChevronLeft className="w-4 h-4 text-slate-300" />
          </button>
          <span className="text-xs font-mono font-bold text-slate-500">
            <span className="text-blue-400">{page}</span> / {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => p + 1)}
            disabled={!pagination.has_next}
            className="p-2 rounded-xl border border-white/5 bg-white/[0.03] disabled:opacity-20 transition-all"
          >
            <ChevronRight className="w-4 h-4 text-slate-300" />
          </button>
        </div>
      )}
    </div>
  );
}
