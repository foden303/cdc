"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { listTopicsAction } from "@/lib/actions";
import type { TopicSummary, PaginationResponse } from "@/lib/grpc";
import { 
  Layers, 
  ChevronLeft, 
  ChevronRight, 
  ExternalLink,
  MessageSquare,
  ArrowRight,
  Search,
  Plus,
  Box
} from "lucide-react";
import Link from "next/link";

export default function TopicsPage() {
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

  if (loading && allTopics.length === 0) {
    return (
      <div className="space-y-6 animate-pulse p-8">
        <div className="h-24 w-full rounded-2xl bg-white/5" />
        <div className="h-96 rounded-2xl bg-white/5" />
      </div>
    );
  }

  return (
    <div className="p-8 space-y-10 animate-in fade-in slide-in-from-bottom-4 duration-700">
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-4xl font-bold tracking-tight text-white flex items-center gap-3">
            <Layers className="w-10 h-10 text-blue-500" />
            Topics Hub
          </h1>
          <p className="text-slate-400 mt-2 text-sm font-medium">Manage and inspect change data capture streams across your infrastructure.</p>
        </div>
        <button className="flex items-center gap-2 px-6 py-2.5 rounded-2xl bg-blue-600 hover:bg-blue-500 text-white text-sm font-bold transition-all shadow-xl shadow-blue-600/20 active:scale-95">
          <Plus className="w-4 h-4" /> Create Topic
        </button>
      </div>

      {/* Stats Dashboard */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {[
          { label: "Total Topics", value: stats.totalTopics, icon: Layers, color: "text-blue-400" },
          { label: "Total Partitions", value: stats.totalPartitions, icon: Box, color: "text-purple-400" },
          { label: "Total Messages", value: stats.totalMessages.toLocaleString(), icon: MessageSquare, color: "text-emerald-400" }
        ].map((stat, i) => (
          <div key={i} className="bg-white/[0.02] border border-white/5 p-6 rounded-3xl relative overflow-hidden group hover:bg-white/[0.04] transition-all">
            <stat.icon className={`absolute -right-4 -bottom-4 w-24 h-24 opacity-[0.03] group-hover:scale-110 transition-transform ${stat.color}`} />
            <div className="text-[10px] font-bold text-slate-500 uppercase tracking-[0.2em] mb-1">{stat.label}</div>
            <div className="text-3xl font-mono font-bold text-white tracking-tighter">{stat.value}</div>
          </div>
        ))}
      </div>

      {/* Control Bar */}
      <div className="flex flex-col lg:flex-row items-center gap-4 bg-white/[0.02] border border-white/5 p-4 rounded-3xl backdrop-blur-sm">
        <div className="relative flex-1 group w-full">
          <Search className="absolute left-4 top-3.5 w-4 h-4 text-slate-500 group-focus-within:text-blue-400 transition-colors" />
          <input
            type="text"
            placeholder="Search topics by name..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full bg-white/[0.02] border border-white/5 rounded-2xl py-3 pl-12 pr-4 text-sm font-medium text-white placeholder:text-slate-600 focus:outline-none focus:border-blue-500/50 focus:bg-white/[0.04] transition-all"
          />
        </div>
        <div className="flex items-center gap-6 w-full lg:w-auto px-4">
          <label className="flex items-center gap-3 cursor-pointer group">
            <div className="relative">
              <input 
                type="checkbox" 
                className="sr-only" 
                checked={showInternal}
                onChange={() => setShowInternal(!showInternal)}
              />
              <div className={`w-10 h-5 rounded-full transition-colors ${showInternal ? "bg-blue-600" : "bg-slate-700"}`} />
              <div className={`absolute left-1 top-1 w-3 h-3 rounded-full bg-white transition-transform ${showInternal ? "translate-x-5" : ""}`} />
            </div>
            <span className="text-xs font-bold text-slate-500 group-hover:text-slate-300 transition-colors">Show Internal</span>
          </label>
        </div>
      </div>

      {/* Topics List */}
      <div className="bg-white/[0.01] border border-white/5 rounded-[2rem] overflow-hidden shadow-2xl">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-white/5 text-[10px] text-slate-500 uppercase font-bold tracking-[0.2em]">
                <th className="px-8 py-6 text-left">Internal Name</th>
                <th className="px-8 py-6 text-center">Partitions</th>
                <th className="px-8 py-6 text-right">Message Count</th>
                <th className="px-8 py-6 text-right">Status</th>
                <th className="px-8 py-6 text-center">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/[0.02]">
              {filteredTopics.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-8 py-32 text-center text-slate-600">
                    <div className="flex flex-col items-center gap-4">
                      <Layers className="w-16 h-16 opacity-10" />
                      <p className="text-sm font-bold tracking-tight">No topics found matching your criteria</p>
                    </div>
                  </td>
                </tr>
              ) : (
                filteredTopics.map((topic) => (
                  <tr key={topic.name} className="group hover:bg-blue-500/[0.02] transition-colors">
                    <td className="px-8 py-6">
                      <Link 
                        href={`/topics/${encodeURIComponent(topic.name)}`}
                        className="flex items-center gap-3"
                      >
                        <div className={`w-2 h-2 rounded-full ${topic.message_count > 0 ? "bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.4)]" : "bg-slate-700"} animate-pulse`} />
                        <span className="text-sm font-mono font-bold text-white group-hover:text-blue-400 transition-colors">{topic.name}</span>
                      </Link>
                    </td>
                    <td className="px-8 py-6 text-center font-mono font-bold text-slate-400">
                      {topic.partition_count}
                    </td>
                    <td className="px-8 py-6 text-right font-mono font-bold text-white">
                      {Number(topic.message_count).toLocaleString()}
                    </td>
                    <td className="px-8 py-6 text-right">
                      <span className={`px-2 py-0.5 rounded-full border text-[9px] font-bold uppercase tracking-widest ${
                        topic.message_count > 0 
                          ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20" 
                          : "bg-orange-500/10 text-orange-400 border-orange-500/20"
                      }`}>
                        {topic.message_count > 0 ? "Streaming" : "Idle"}
                      </span>
                    </td>
                    <td className="px-8 py-6 text-center">
                      <Link
                        href={`/topics/${encodeURIComponent(topic.name)}`}
                        className="px-4 py-1.5 rounded-xl border border-white/5 bg-white/[0.03] hover:bg-blue-600 hover:border-blue-500 hover:text-white text-[10px] font-bold uppercase tracking-widest transition-all"
                      >
                        Inspect
                      </Link>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Pagination */}
      {pagination && totalPages > 1 && (
        <div className="flex items-center justify-between px-8 py-6 rounded-3xl bg-white/[0.01] border border-white/5">
          <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
            Showing <span className="text-white">{(page-1)*limit+1} - {Math.min(page*limit, pagination.total_rows)}</span> of <span className="text-white">{pagination.total_rows}</span>
          </div>
          <div className="flex items-center gap-4">
            <button
              onClick={() => setPage(p => Math.max(1, p-1))}
              disabled={!pagination.has_prev}
              className="p-2.5 rounded-2xl bg-white/[0.03] border border-white/5 hover:bg-white/10 disabled:opacity-20 transition-all"
            >
              <ChevronLeft className="w-5 h-5 text-slate-300" />
            </button>
            <div className="flex items-center gap-2 font-mono font-bold text-xs">
              <span className="text-blue-400">{page}</span>
              <span className="text-slate-700">/</span>
              <span className="text-slate-400">{totalPages}</span>
            </div>
            <button
              onClick={() => setPage(p => p + 1)}
              disabled={!pagination.has_next}
              className="p-2.5 rounded-2xl bg-white/[0.03] border border-white/5 hover:bg-white/10 disabled:opacity-20 transition-all"
            >
              <ChevronRight className="w-5 h-5 text-slate-300" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
