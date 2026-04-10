"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { listPartitionsAction } from "@/lib/actions";
import type { PartitionSummary, PaginationResponse } from "@/lib/grpc";
import MessageBrowser from "@/components/MessageBrowser";
import { 
  Layers, 
  ArrowLeft, 
  MessageSquare,
  Box,
  Database,
  Terminal,
  Settings,
  Users
} from "lucide-react";

type Tab = "messages" | "partitions" | "consumers" | "config";

export default function TopicDetailPage() {
  const { name } = useParams();
  const router = useRouter();
  const topicName = decodeURIComponent(name as string);

  const [activeTab, setActiveTab] = useState<Tab>("messages");
  const [partitions, setPartitions] = useState<PartitionSummary[]>([]);
  const [pagination, setPagination] = useState<PaginationResponse | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchPartitions = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listPartitionsAction(topicName, 100, 1);
      setPartitions(res.data || []);
      setPagination(res.pagination || null);
    } catch (err) {
      console.error("fetch partitions:", err);
    } finally {
      setLoading(false);
    }
  }, [topicName]);

  useEffect(() => {
    fetchPartitions();
  }, [fetchPartitions]);

  const totalMessages = partitions.reduce((s, p) => s + Number(p.message_count), 0);

  const tabs = [
    { id: "messages", label: "Messages", icon: MessageSquare },
    { id: "partitions", label: "Partitions", icon: Box },
    { id: "consumers", label: "Consumers", icon: Users },
    { id: "config", label: "Configuration", icon: Settings },
  ];

  return (
    <div className="p-8 space-y-8 animate-in fade-in duration-700">
      {/* Detail Header */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6 bg-white/[0.02] border border-white/5 p-8 rounded-[2.5rem] relative overflow-hidden">
        <div className="absolute top-0 right-0 p-8 opacity-[0.03]">
           <Layers className="w-32 h-32" />
        </div>
        
        <div className="flex items-center gap-6 relative z-10">
          <button
            onClick={() => router.push("/topics")}
            className="p-4 rounded-2xl bg-white/[0.03] border border-white/5 hover:bg-white/10 transition-all group"
          >
            <ArrowLeft className="w-5 h-5 text-slate-400 group-hover:text-white transition-colors" />
          </button>
          <div>
            <div className="flex items-center gap-2 text-blue-400 font-mono text-[10px] uppercase font-bold tracking-[0.2em] mb-2">
              <Database className="w-3 h-3" /> Topic Instance
            </div>
            <h1 className="text-4xl font-mono font-bold text-white tracking-tight leading-none">
              {topicName}
            </h1>
            <div className="flex items-center gap-4 mt-4">
               <div className="flex items-center gap-1.5 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-[10px] font-bold text-emerald-400 uppercase tracking-wider">
                  Live Stream
               </div>
               <div className="text-slate-500 text-xs font-medium">Capture source: <span className="text-slate-300">Native CDC v1</span></div>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4 relative z-10">
           <div className="px-6 py-4 rounded-2xl bg-white/[0.02] border border-white/5">
              <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1">Partitions</div>
              <div className="text-xl font-mono font-bold text-white">{partitions.length}</div>
           </div>
           <div className="px-6 py-4 rounded-2xl bg-white/[0.02] border border-white/5">
              <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1">Total Payload</div>
              <div className="text-xl font-mono font-bold text-white">{totalMessages.toLocaleString()}</div>
           </div>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex items-center gap-1 p-1 bg-white/[0.02] border border-white/5 rounded-2xl w-fit">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id as Tab)}
            className={`flex items-center gap-2 px-6 py-2.5 rounded-xl transition-all text-xs font-bold uppercase tracking-widest ${
              activeTab === tab.id 
                ? "bg-blue-600 text-white shadow-lg shadow-blue-600/20" 
                : "text-slate-500 hover:text-slate-300 hover:bg-white/[0.03]"
            }`}
          >
            <tab.icon className="w-4 h-4" />
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div className="bg-white/[0.01] border border-white/5 rounded-[2.5rem] p-8 shadow-2xl min-h-[500px]">
        {activeTab === "messages" && (
           <div className="space-y-6 animate-in fade-in slide-in-from-top-2 duration-500">
              <div className="flex items-center gap-3 mb-2">
                 <Terminal className="w-5 h-5 text-blue-400" />
                 <h2 className="text-lg font-bold text-white tracking-tight">Stream Browser</h2>
              </div>
              <MessageBrowser topic={topicName} />
           </div>
        )}

        {activeTab === "partitions" && (
          <div className="animate-in fade-in slide-in-from-top-2 duration-500">
            <h2 className="text-lg font-bold text-white tracking-tight mb-6 flex items-center gap-3">
               <Box className="w-5 h-5 text-purple-400" /> Topic Topology
            </h2>
            <div className="rounded-2xl border border-white/5 overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-white/5 text-[10px] text-slate-500 uppercase font-bold tracking-widest">
                    <th className="px-6 py-4 text-left">Partition Path</th>
                    <th className="px-6 py-4 text-center">Status</th>
                    <th className="px-6 py-4 text-right">Sequence Gap</th>
                    <th className="px-6 py-4 text-right">Message Count</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/[0.02]">
                  {partitions.map((p) => (
                    <tr key={p.id} className="hover:bg-white/[0.01] transition-colors">
                      <td className="px-6 py-5 font-mono text-xs font-bold text-white">{p.id}</td>
                      <td className="px-6 py-5 text-center">
                        <span className="px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-[9px] font-bold uppercase">Healthy</span>
                      </td>
                      <td className="px-6 py-5 text-right font-mono text-xs text-slate-500">None</td>
                      <td className="px-6 py-5 text-right font-mono font-bold text-white">{Number(p.message_count).toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {activeTab === "consumers" && (
           <div className="flex flex-col items-center justify-center py-32 opacity-20">
              <Users className="w-16 h-16 mb-4" />
              <p className="font-bold tracking-widest uppercase text-xs">No active consumers detected</p>
           </div>
        )}

        {activeTab === "config" && (
           <div className="flex flex-col items-center justify-center py-32 opacity-20">
              <Settings className="w-16 h-16 mb-4" />
              <p className="font-bold tracking-widest uppercase text-xs">Runtime configuration is read-only</p>
           </div>
        )}
      </div>
    </div>
  );
}
