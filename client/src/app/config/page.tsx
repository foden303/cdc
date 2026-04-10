"use client";

import { useEffect, useState } from "react";
import { getConfigAction } from "@/lib/actions";
import { AppConfig } from "@/lib/grpc";
import { 
  Save, 
  Settings, 
  Zap, 
  Database, 
  Terminal, 
  Activity,
  CheckCircle2,
  AlertCircle,
  Loader2,
  Lock,
  Info
} from "lucide-react";

export default function ConfigPage() {
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    fetchConfig();
  }, []);

  const fetchConfig = async () => {
    setLoading(true);
    try {
      const res = await getConfigAction() as { config: AppConfig };
      if (res?.config) {
        setConfig(res.config);
      }
    } catch (err) {
      console.error(err);
    }
    setLoading(false);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="w-8 h-8 text-blue-500 animate-spin" />
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700 pb-20">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight flex items-center gap-3">
            <Settings className="w-8 h-8 text-blue-400" />
            System Settings
          </h1>
          <p className="text-slate-400 mt-1">Manage core engine configuration and runtime parameters.</p>
        </div>
      </div>

      {message && (
        <div className={`p-4 rounded-xl flex items-center gap-3 animate-in zoom-in-95 duration-300 ${
          message.type === "success" 
            ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20" 
            : "bg-red-500/10 text-red-400 border border-red-500/20"
        }`}>
          {message.type === "success" ? <CheckCircle2 className="w-5 h-5" /> : <AlertCircle className="w-5 h-5" />}
          <p className="text-sm font-medium">{message.text}</p>
        </div>
      )}

      <div className="space-y-6">
        {/* Basic Info */}
        <section className="bg-white/[0.03] border border-white/5 rounded-2xl overflow-hidden backdrop-blur-md">
          <div className="p-6 border-b border-white/5 bg-white/[0.02] flex items-center gap-3">
            <Activity className="w-5 h-5 text-blue-500" />
            <h2 className="text-lg font-bold text-white tracking-tight">App Settings</h2>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-6 opacity-60">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest">App Name</label>
                <Lock className="w-3 h-3 text-slate-500" />
              </div>
              <input 
                name="name"
                defaultValue={config?.name}
                readOnly
                className="w-full bg-white/[0.02] border border-white/5 rounded-xl px-4 py-2.5 text-slate-400 cursor-not-allowed font-medium"
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest">Log Mode</label>
                <Lock className="w-3 h-3 text-slate-500" />
              </div>
              <input 
                name="log_mode"
                defaultValue={config?.log_mode}
                readOnly
                className="w-full bg-white/[0.02] border border-white/5 rounded-xl px-4 py-2.5 text-slate-400 cursor-not-allowed font-medium"
              />
            </div>
          </div>
        </section>

        {/* Pipeline Controls */}
        <section className="bg-white/[0.03] border border-white/5 rounded-2xl overflow-hidden backdrop-blur-md">
          <div className="p-6 border-b border-white/5 bg-white/[0.02] flex items-center gap-3">
            <Zap className="w-5 h-5 text-emerald-500" />
            <h2 className="text-lg font-bold text-white tracking-tight">Pipeline Core</h2>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 opacity-60">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest text-[10px]">Channel Buffer Size</label>
                <Lock className="w-3 h-3 text-slate-500" />
              </div>
              <input 
                name="channel_buffer_size"
                defaultValue={config?.pipeline?.channel_buffer_size}
                readOnly
                className="w-full bg-white/[0.02] border border-white/5 rounded-xl px-4 py-2.5 text-slate-400 cursor-not-allowed font-medium"
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest text-[10px]">Worker Count</label>
                <Lock className="w-3 h-3 text-slate-500" />
              </div>
              <input 
                name="worker_count"
                defaultValue={config?.pipeline?.worker_count}
                readOnly
                className="w-full bg-white/[0.02] border border-white/5 rounded-xl px-4 py-2.5 text-slate-400 cursor-not-allowed font-medium"
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest text-[10px]">Batch Size</label>
                <Lock className="w-3 h-3 text-slate-500" />
              </div>
              <input 
                name="batch_size"
                defaultValue={config?.pipeline?.batch_size}
                readOnly
                className="w-full bg-white/[0.02] border border-white/5 rounded-xl px-4 py-2.5 text-slate-400 cursor-not-allowed font-medium"
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest text-[10px]">Flush Interval (ms)</label>
                <Lock className="w-3 h-3 text-slate-500" />
              </div>
              <input 
                name="flush_interval_ms"
                defaultValue={config?.pipeline?.flush_interval_ms}
                readOnly
                className="w-full bg-white/[0.02] border border-white/5 rounded-xl px-4 py-2.5 text-slate-400 cursor-not-allowed font-medium"
              />
            </div>
            <div className="space-y-2 lg:col-span-2">
              <div className="flex items-center gap-2">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest text-[10px]">Subject Filter</label>
                <Lock className="w-3 h-3 text-slate-500" />
              </div>
              <input 
                name="subject_filter"
                defaultValue={config?.pipeline?.subject_filter?.join(", ")}
                readOnly
                className="w-full bg-white/[0.02] border border-white/5 rounded-xl px-4 py-2.5 text-slate-400 cursor-not-allowed font-medium"
              />
            </div>
          </div>
        </section>

        {/* NATS Settings */}
        <section className="bg-white/[0.03] border border-white/5 rounded-2xl overflow-hidden backdrop-blur-md">
          <div className="p-6 border-b border-white/5 bg-white/[0.02] flex items-center gap-3">
            <Database className="w-5 h-5 text-indigo-500" />
            <h2 className="text-lg font-bold text-white tracking-tight">NATS Streaming</h2>
          </div>
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 opacity-60">
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest text-[10px]">Stream Name</label>
                <Lock className="w-3 h-3 text-slate-500" />
              </div>
              <input 
                name="stream_name"
                defaultValue={config?.nats?.stream_name}
                readOnly
                className="w-full bg-white/[0.02] border border-white/5 rounded-xl px-4 py-2.5 text-slate-400 cursor-not-allowed font-medium"
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest text-[10px]">Retention (Days)</label>
                <Lock className="w-3 h-3 text-slate-500" />
              </div>
              <input 
                name="retention_days"
                defaultValue={config?.nats?.retention_days}
                readOnly
                className="w-full bg-white/[0.02] border border-white/5 rounded-xl px-4 py-2.5 text-slate-400 cursor-not-allowed font-medium"
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest text-[10px]">Max Reconnects</label>
                <Lock className="w-3 h-3 text-slate-500" />
              </div>
              <input 
                name="max_reconnects"
                defaultValue={config?.nats?.max_reconnects}
                readOnly
                className="w-full bg-white/[0.02] border border-white/5 rounded-xl px-4 py-2.5 text-slate-400 cursor-not-allowed font-medium"
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest text-[10px]">Reconnect Wait (ms)</label>
                <Lock className="w-3 h-3 text-slate-500" />
              </div>
              <input 
                name="reconnect_wait_ms"
                defaultValue={config?.nats?.reconnect_wait_ms}
                readOnly
                className="w-full bg-white/[0.02] border border-white/5 rounded-xl px-4 py-2.5 text-slate-400 cursor-not-allowed font-medium"
              />
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <label className="text-sm font-bold text-slate-400 uppercase tracking-widest text-[10px]">Reconnect Buffer (MB)</label>
                <Lock className="w-3 h-3 text-slate-500" />
              </div>
              <input 
                name="reconnect_buffer_size_mb"
                defaultValue={config?.nats?.reconnect_buffer_size_mb}
                readOnly
                className="w-full bg-white/[0.02] border border-white/5 rounded-xl px-4 py-2.5 text-slate-400 cursor-not-allowed font-medium"
              />
            </div>
          </div>
        </section>
      </div>
    </div>
  );
}
