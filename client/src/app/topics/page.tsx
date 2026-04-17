"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { listTopicsAction } from "@/lib/actions";
import type { TopicSummary, PaginationResponse } from "@/lib/grpc";
import { useApp } from "@/lib/AppContext";
import { 
  Layers, 
  ChevronLeft, 
  ChevronRight, 
  MessageSquare, 
  Search,
  Plus,
  Box,
  Activity
} from "lucide-react";
import Link from "next/link";

export default function TopicsPage() {
  const { t } = useApp();
  const router = useRouter();
  const [allTopics, setAllTopics] = useState<TopicSummary[]>([]);
  const [showInternal, setShowInternal] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [pagination, setPagination] = useState<PaginationResponse | null>(null);
  const [page, setPage] = useState(1);
  const limit = 20;
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchTopics = async () => {
      setLoading(true);
      try {
        const res = await listTopicsAction(limit, page);
        setAllTopics(res.data || []);
        setPagination(res.pagination || null);
      } catch (err) {
        console.error("fetch topics:", err);
      } finally {
        setLoading(false);
      }
    };
    fetchTopics();
  }, [page]);

  const filteredTopics = allTopics.filter(t => {
    const isInternal = t.name.startsWith("__") || t.name.startsWith("dev.debezium");
    const matchesSearch = t.name.toLowerCase().includes(searchTerm.toLowerCase());
    return (showInternal || !isInternal) && matchesSearch;
  });

  const stats = {
    totalTopics: pagination?.total_rows ?? 0,
    totalPartitions: allTopics.reduce((s, t) => s + (t.partition_count || 0), 0),
    totalMessages: allTopics.reduce((s, t) => s + Number(t.message_count || 0), 0)
  };

  const totalPages = pagination ? Math.ceil(pagination.total_rows / limit) : 1;

  return (
    <div className="space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-500 pb-10">
      {/* Compact Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="space-y-0.5">
          <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-500 overline-label overline-label-solid">
            <Layers className="w-3 h-3" /> {t("orchestration")} Hub
          </div>
          <h1 className="text-xl md:text-2xl font-black text-foreground">{t("topics")}</h1>
          <p className="text-muted-foreground text-base-compact font-medium max-w-xl line-clamp-1">Change data streams monitoring across the infrastructure.</p>
        </div>
        <button className="flex items-center gap-2 px-3.5 py-1.5 rounded-lg bg-primary text-white text-base-compact font-black shadow-sm">
          <Plus className="w-3.5 h-3.5" /> {t("addSource")}
        </button>
      </div>

      {/* Stats Mini Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        <StatBox label={t("topics")} value={stats.totalTopics} icon={<Layers className="w-3.5 h-3.5" />} color="blue" />
        <StatBox label={t("partitions")} value={stats.totalPartitions} icon={<Box className="w-3.5 h-3.5" />} color="purple" />
        <StatBox label={t("processedEvents")} value={stats.totalMessages.toLocaleString()} icon={<Activity className="w-3.5 h-3.5" />} color="emerald" />
      </div>

      {/* Control Bar */}
      <div className="flex flex-col md:flex-row items-center gap-3 p-2 bg-card rounded-xl border border-border shadow-sm">
        <div className="relative flex-1 group w-full">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground group-focus-within:text-primary transition-colors" />
          <input
            type="text"
            placeholder={t("topics")}
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full bg-muted/30 border border-border rounded-lg py-1.5 pl-8 pr-3 text-base-compact font-bold text-foreground outline-none focus:border-primary/40 transition-all placeholder:text-muted-foreground/30"
          />
        </div>
        <div className="flex items-center gap-4 px-2">
          <label className="flex items-center gap-2 cursor-pointer group">
            <input type="checkbox" className="w-3 h-3 rounded border-border bg-muted text-primary" checked={showInternal} onChange={() => setShowInternal(!showInternal)} />
            <span className="overline-label overline-label-dim group-hover:opacity-100 transition-opacity">Show Internal</span>
          </label>
        </div>
      </div>

      {/* Topics List Table */}
      <div className="bg-card rounded-xl overflow-hidden border border-border shadow-sm">
        <div className="overflow-x-auto custom-scrollbar">
          <table className="w-full text-table-compact">
            <thead>
              <tr className="border-b border-border bg-muted/30">
                <th className="px-3 py-2 text-left overline-label overline-label-dim">Internal Name</th>
                <th className="px-3 py-2 text-center overline-label overline-label-dim">Partitions</th>
                <th className="px-3 py-2 text-right overline-label overline-label-dim">Messages</th>
                <th className="px-3 py-2 text-right overline-label overline-label-dim">Status</th>
                <th className="px-3 py-2 text-center overline-label overline-label-dim">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {filteredTopics.map((topic) => (
                <tr key={topic.name} className="group hover:bg-muted/30 transition-colors">
                  <td className="px-3 py-1.5">
                    <Link href={`/topics/${encodeURIComponent(topic.name)}`} className="flex items-center gap-2">
                      <div className={`w-1 h-1 rounded-full ${topic.message_count > 0 ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]" : "bg-muted-foreground/30"}`} />
                      <span className="font-mono font-bold text-foreground group-hover:text-primary transition-colors truncate max-w-[200px] lg:max-w-md">{topic.name}</span>
                    </Link>
                  </td>
                  <td className="px-3 py-1.5 text-center font-mono font-bold text-muted-foreground tabular-nums">{topic.partition_count}</td>
                  <td className="px-3 py-1.5 text-right font-mono font-bold text-foreground tabular-nums">{Number(topic.message_count).toLocaleString()}</td>
                  <td className="px-3 py-1.5 text-right">
                    <span className={`px-1.5 py-0.25 rounded-md text-micro-compact font-black uppercase ${topic.message_count > 0 ? "bg-emerald-500/10 text-emerald-500" : "bg-muted text-muted-foreground"}`}>
                      {topic.message_count > 0 ? "Streaming" : "Idle"}
                    </span>
                  </td>
                  <td className="px-3 py-1.5 text-center">
                    <Link href={`/topics/${encodeURIComponent(topic.name)}`} className="inline-flex px-2 py-0.5 rounded-md border border-border bg-muted/50 hover:bg-muted text-tiny font-black uppercase text-muted-foreground hover:text-foreground transition-all">
                      Inspect
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {pagination && totalPages > 1 && (
        <div className="flex items-center justify-center gap-3 pt-2">
          <button onClick={() => setPage(p => Math.max(1, p-1))} disabled={!pagination?.has_prev} className="px-3 py-1 rounded-lg bg-muted border border-border disabled:opacity-30 text-mini-compact font-black uppercase text-muted-foreground hover:bg-muted/80">
            Prev
          </button>
          <div className="text-mini-compact font-black text-muted-foreground uppercase tabular-nums">{page} / {totalPages}</div>
          <button onClick={() => setPage(p => p + 1)} disabled={!pagination?.has_next} className="px-3 py-1 rounded-lg bg-muted border border-border disabled:opacity-30 text-mini-compact font-black uppercase text-muted-foreground hover:bg-muted/80">
            Next
          </button>
        </div>
      )}
    </div>
  );
}

function StatBox({ label, value, icon, color }: any) {
  const colorMap: any = {
    blue: "text-primary bg-primary/10 border-primary/10",
    purple: "text-purple-500 bg-purple-500/10 border-purple-500/10",
    emerald: "text-emerald-500 bg-emerald-500/10 border-emerald-500/10"
  };
  return (
    <div className="bg-card p-2.5 rounded-xl flex items-center justify-between border border-border shadow-sm">
       <div className="space-y-0.5">
          <div className="overline-label overline-label-dim">{label}</div>
          <div className="text-lg font-black text-foreground tracking-tighter tabular-nums leading-none">{value}</div>
       </div>
       <div className={`p-1.5 rounded-lg border ${colorMap[color]}`}>{icon}</div>
    </div>
  );
}
