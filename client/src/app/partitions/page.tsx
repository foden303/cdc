"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { listPartitionsAction, getStatsAction } from "@/lib/actions";
import type { PartitionSummary, GetStatsResponse } from "@/lib/grpc";
import { useApp } from "@/lib/AppContext";
import { 
  Network, 
  ChevronLeft, 
  ChevronRight, 
  MessageSquare, 
  Activity,
  ArrowDownLeft,
  Search,
  Clock
} from "lucide-react";

export default function PartitionsPage() {
  const { t } = useApp();
  const [partitions, setPartitions] = useState<PartitionSummary[]>([]);
  const [stats, setStats] = useState<GetStatsResponse | null>(null);
  const [total, setTotal] = useState(0);
  const [hasNext, setHasNext] = useState(false);
  const [hasPrev, setHasPrev] = useState(false);
  const [page, setPage] = useState(1);
  const limit = 50;
  const [loading, setLoading] = useState(true);
  const [topicFilter, setTopicFilter] = useState("");

  const fetchData = useCallback(async () => {
    try {
      const [pRes, sRes] = await Promise.all([
        listPartitionsAction(topicFilter, limit, page),
        getStatsAction()
      ]);
      setPartitions(pRes.data || []);
      setTotal(pRes.pagination?.total_rows ?? 0);
      setHasNext(pRes.pagination?.has_next ?? false);
      setHasPrev(pRes.pagination?.has_prev ?? false);
      setStats(sRes);
    } catch (e) {
      console.error("fetch partitions/stats:", e);
    } finally {
      setLoading(false);
    }
  }, [topicFilter, page]);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, [fetchData]);

  const getPartitionLag = (partitionId: string) => {
    if (!stats?.sink_stats) return 0;
    const parts = partitionId.split('.');
    const partitionIdx = parseInt(parts[parts.length - 1]);
    if (isNaN(partitionIdx)) return 0;
    let totalLag = 0;
    Object.values(stats.sink_stats).forEach(sink => {
      if (sink.partition_lag && sink.partition_lag[partitionIdx] !== undefined) {
        totalLag += sink.partition_lag[partitionIdx];
      }
    });
    return totalLag;
  };

  const totalPages = Math.ceil(total / limit) || 1;

  return (
    <div className="space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-500 pb-10">
      {/* Compact Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="space-y-0.5">
          <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10 border border-emerald-500/10 text-emerald-500 overline-label overline-label-solid">
            <Activity className="w-3 h-3" /> {t("performance")} Monitoring
          </div>
          <h1 className="text-xl md:text-2xl font-black text-foreground">{t("partitions")}</h1>
          <p className="text-muted-foreground body-compact max-w-xl line-clamp-1">Monitoring lag pressure and shard health across the cluster.</p>
        </div>

        <div className="flex gap-4">
           <div className="relative group">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground group-focus-within:text-primary transition-colors" />
              <input
                type="text"
                placeholder={t("topics")}
                value={topicFilter}
                onChange={(e) => { setTopicFilter(e.target.value); setPage(1); }}
                className="pl-8 pr-3 py-1.5 rounded-lg bg-muted/30 border border-border text-sm-compact text-foreground w-64 focus:outline-none focus:border-primary/40 transition-all font-bold placeholder:text-muted-foreground/30"
              />
           </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
         <StatBox label={t("partitions")} value={String(total)} icon={<Network className="w-3.5 h-3.5" />} color="blue" />
         <StatBox label={t("operational")} value={t("ready")} icon={<Activity className="w-3.5 h-3.5" />} color="emerald" />
         <StatBox label="Avg Cluster Lag" value="0.4ms" icon={<Clock className="w-3.5 h-3.5" />} color="indigo" />
      </div>

      <div className="bg-card rounded-xl overflow-hidden border border-border shadow-sm">
        <div className="overflow-x-auto custom-scrollbar">
          <table className="w-full text-xs-compact">
            <thead>
              <tr className="border-b border-border bg-muted/30">
                <th className="px-3 py-2 text-left overline-label overline-label-dim">Identity</th>
                <th className="px-3 py-2 text-left overline-label overline-label-dim">Subject</th>
                <th className="px-3 py-2 text-right overline-label overline-label-dim">Messages</th>
                <th className="px-3 py-2 text-right overline-label overline-label-dim">Lag</th>
                <th className="px-3 py-2 text-center overline-label overline-label-dim">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {partitions.map((p) => {
                const lag = getPartitionLag(p.id);
                return (
                  <tr key={p.id} className="hover:bg-muted/30 transition-colors group">
                    <td className="px-3 py-1.5 font-mono text-primary font-bold">{p.id}</td>
                    <td className="px-3 py-1.5">
                      <Link href={`/topics/${encodeURIComponent(p.topic)}`} className="font-mono text-muted-foreground hover:text-foreground truncate block max-w-[200px]">
                        {p.topic}
                      </Link>
                    </td>
                    <td className="px-3 py-1.5 font-mono text-foreground text-right font-black tabular-nums">
                      {Number(p.message_count).toLocaleString()}
                    </td>
                    <td className="px-3 py-1.5 text-right">
                       <div className={`inline-flex items-center gap-1 font-mono font-black tabular-nums ${lag > 100 ? 'text-rose-500' : lag > 10 ? 'text-amber-500' : 'text-emerald-500'}`}>
                          {lag > 0 && <ArrowDownLeft className="w-2.5 h-2.5" />}
                          {lag.toLocaleString()}
                       </div>
                    </td>
                    <td className="px-3 py-1.5 text-center">
                      <Link href={`/messages?partition=${encodeURIComponent(p.id)}`} className="inline-flex p-1.5 rounded-lg border border-border bg-muted/50 hover:bg-muted text-muted-foreground hover:text-foreground transition-all">
                        <MessageSquare className="w-3.5 h-3.5" />
                      </Link>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-3 pt-2">
          <button onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={!hasPrev} className="px-3 py-1 rounded-lg bg-muted border border-border ui-disabled-soft overline-label mb-0 text-muted-foreground hover:bg-muted/80 transition-colors">
            Prev
          </button>
          <div className="overline-label mb-0 tabular-nums">{page} / {totalPages}</div>
          <button onClick={() => setPage((p) => p + 1)} disabled={!hasNext} className="px-3 py-1 rounded-lg bg-muted border border-border ui-disabled-soft overline-label mb-0 text-muted-foreground hover:bg-muted/80 transition-colors">
            Next
          </button>
        </div>
      )}
    </div>
  );
}

function StatBox({ label, value, icon, color }: any) {
  const colors: any = {
    blue: "text-primary bg-primary/10 border-primary/10",
    emerald: "text-emerald-500 bg-emerald-500/10 border-emerald-500/10",
    indigo: "text-indigo-500 bg-indigo-500/10 border-indigo-500/10"
  };
  return (
    <div className="bg-card p-2.5 rounded-xl flex items-center justify-between border border-border shadow-sm">
       <div className="space-y-0.5">
          <div className="overline-label overline-label-dim">{label}</div>
          <div className="text-lg font-black text-foreground tracking-tighter tabular-nums leading-none">{value}</div>
       </div>
       <div className={`p-1.5 rounded-lg border ${colors[color]}`}>
          {icon}
       </div>
    </div>
  );
}
