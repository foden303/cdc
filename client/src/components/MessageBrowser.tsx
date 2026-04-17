"use client";

import { useEffect, useState, useCallback, Fragment } from "react";
import { listMessagesAction } from "@/lib/actions";
import type { MessageItem, PaginationResponse } from "@/lib/grpc";
import { useApp } from "@/lib/AppContext";
import { 
  ChevronDown, 
  ChevronUp, 
  ChevronLeft, 
  ChevronRight, 
  RefreshCcw, 
  Copy
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
  const { t } = useApp();
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
    <div className="space-y-4">
      {/* Mini Control Bar */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 px-2 py-1 rounded-xl bg-white/[0.03] border border-white/5 text-mini-compact font-black text-slate-500 uppercase tracking-widest">
          <span className="text-primary">{pagination?.total_rows ?? 0}</span> {t("processedEvents")}
        </div>
        <button onClick={fetchMessages} className="p-1.5 rounded-lg border border-white/5 bg-white/[0.03] hover:bg-white/[0.08] transition-all">
          <RefreshCcw className={`w-3 h-3 ${loading ? "animate-spin text-primary" : "text-slate-500"}`} />
        </button>
      </div>

      {/* Messages Table */}
      <div className="rounded-xl border border-white/5 bg-white/[0.01] overflow-hidden shadow-sm">
        <div className="overflow-x-auto custom-scrollbar">
          <table className="w-full text-base-compact">
            <thead>
              <tr className="border-b border-white/5 text-mini-compact text-slate-500 uppercase font-black tracking-widest bg-white/[0.01]">
                <th className="px-3 py-2 text-center w-10">#</th>
                <th className="px-3 py-2 text-left">Seq</th>
                {!compact && <th className="px-3 py-2 text-left">Subject</th>}
                <th className="px-3 py-2 text-left">Time</th>
                <th className="px-3 py-2 text-center w-12">CMD</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/[0.01]">
              {messages.length === 0 ? (
                <tr>
                  <td colSpan={compact ? 4 : 5} className="px-4 py-10 text-center text-slate-500 text-compact font-medium">
                    No events discovered.
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
                      <tr onClick={() => toggleExpand(m.sequence)}
                        className={`group cursor-pointer transition-colors ${isExpanded ? "bg-primary/[0.06]" : "hover:bg-white/[0.01]"}`}>
                        <td className="px-3 py-1.5 text-center">
                          {isExpanded ? <ChevronUp className="w-3 h-3 text-primary m-auto" /> : <ChevronDown className="w-3 h-3 text-slate-700 m-auto" />}
                        </td>
                        <td className="px-3 py-1.5 font-mono font-bold text-foreground">{m.sequence}</td>
                        {!compact && <td className="px-3 py-1.5 font-mono text-slate-500 truncate max-w-[150px]">{m.subject}</td>}
                        <td className="px-3 py-1.5 text-slate-500 font-mono">
                          {new Date(parseInt(m.timestamp)).toLocaleString('en-US', { hour12: false, month: 'numeric', day: 'numeric', hour: 'numeric', minute: 'numeric', second: 'numeric' })}
                        </td>
                        <td className="px-3 py-1.5 text-center flex items-center justify-center gap-2">
                          <button onClick={(e) => { e.stopPropagation(); copyToClipboard(rawPayload); }}
                            className="p-1 rounded-md border border-white/5 bg-white/[0.02] hover:bg-white/[0.1] text-slate-500">
                            <Copy className="w-3 h-3" />
                          </button>
                        </td>
                      </tr>
                      {isExpanded && (
                        <tr className="bg-primary/[0.02]">
                          <td colSpan={compact ? 4 : 5} className="px-6 py-4">
                            <div className="space-y-2">
                              <pre className="p-3 rounded-xl bg-black/40 border border-white/5 font-mono text-compact leading-relaxed text-primary/90 overflow-x-auto max-h-[300px]">
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

      {pagination && totalPages > 1 && (
        <div className="flex items-center justify-center gap-3 pt-1">
          <button onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={!pagination.has_prev} className="p-1 px-3 rounded-lg border border-white/5 bg-white/[0.03] ui-disabled-faint text-compact font-black uppercase">
            <ChevronLeft className="w-3.5 h-3.5" />
          </button>
          <span className="text-compact font-black text-slate-500">{page} / {totalPages}</span>
          <button onClick={() => setPage((p) => p + 1)} disabled={!pagination.has_next} className="p-1 px-3 rounded-lg border border-white/5 bg-white/[0.03] ui-disabled-faint text-compact font-black uppercase">
            <ChevronRight className="w-3.5 h-3.5" />
          </button>
        </div>
      )}
    </div>
  );
}
