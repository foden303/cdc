"use client";

import { useEffect, useState, useRef } from "react";
import { getPerformanceMetricsAction } from "@/lib/actions";
import { GetPerformanceMetricsResponse } from "@/lib/grpc";
import { 
  BarChart, 
  Bar, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer, 
  AreaChart, 
  Area,
  Cell,
  PieChart,
  Pie
} from "recharts";
import { 
  Activity, 
  Zap, 
  Clock, 
  Database, 
  RefreshCw, 
  AlertTriangle,
  ArrowUpRight,
  Monitor,
  Cpu,
  Server,
  ArrowRightLeft,
  ChevronsRight,
  Box
} from "lucide-react";

export default function MetricsPage() {
  const [metrics, setMetrics] = useState<GetPerformanceMetricsResponse | null>(null);
  const [history, setHistory] = useState<{ time: string; throughput: number }[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchMetrics = async () => {
      const data = await getPerformanceMetricsAction() as GetPerformanceMetricsResponse;
      if (data) {
        setMetrics(data);
        
        const now = new Date();
        const timeStr = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });

        setHistory(prev => {
          const newHistory = [...prev, { time: timeStr, throughput: data.throughput }].slice(-30);
          return newHistory;
        });
      }
      setLoading(false);
    };

    fetchMetrics();
    const timer = setInterval(fetchMetrics, 2000);
    return () => clearInterval(timer);
  }, []);

  if (loading && !metrics) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-6">
        <div className="relative">
          <div className="w-20 h-20 rounded-full border-4 border-blue-500/20 border-t-blue-500 animate-spin" />
          <div className="absolute inset-0 flex items-center justify-center">
             <Activity className="w-8 h-8 text-blue-400 animate-pulse" />
          </div>
        </div>
        <div className="text-center space-y-2">
          <h2 className="text-xl font-bold text-white">Connecting to Engine</h2>
          <p className="text-slate-500 animate-pulse">Synchronizing performance telemetry...</p>
        </div>
      </div>
    );
  }

  const sinkData = metrics?.sinks ? Object.values(metrics.sinks).map(s => ({
    name: s.sink_id,
    throughput: s.throughput,
    latency: s.avg_latency
  })) : [];

  const sourceData = metrics?.sources ? Object.values(metrics.sources).map(s => ({
    name: s.source_id,
    throughput: s.throughput,
    errorRate: s.error_rate
  })) : [];

  return (
    <div className="space-y-12 animate-in fade-in duration-700 pb-20">
      {/* Dynamic Header */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-8 py-4">
        <div className="space-y-3">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-[10px] font-black uppercase tracking-widest shadow-lg shadow-blue-500/5">
            <Monitor className="w-3 h-3" />
            Infrastructure Real-time
          </div>
          <h1 className="text-5xl font-black text-white tracking-tighter bg-clip-text text-transparent bg-gradient-to-r from-white via-white to-white/40">
            Engine Performance
          </h1>
          <p className="text-slate-400 text-lg max-w-xl font-medium leading-relaxed">
            End-to-end observability from data ingestion to delivery. Monitoring real-time pressure across all pipelines.
          </p>
        </div>
        
        <div className="flex flex-wrap gap-4">
           <div className="px-6 py-4 rounded-3xl bg-white/[0.03] border border-white/5 backdrop-blur-xl flex items-center gap-4 group hover:border-emerald-500/30 transition-all">
              <div className="w-10 h-10 rounded-2xl bg-emerald-500/10 flex items-center justify-center text-emerald-400">
                <Server className="w-5 h-5" />
              </div>
              <div>
                <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Node Status</div>
                <div className="text-sm font-bold text-emerald-400 flex items-center gap-2">
                  Operational
                  <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                </div>
              </div>
           </div>
           <div className="px-6 py-4 rounded-3xl bg-white/[0.03] border border-white/5 backdrop-blur-xl flex items-center gap-4 group hover:border-blue-500/30 transition-all">
              <div className="w-10 h-10 rounded-2xl bg-blue-500/10 flex items-center justify-center text-blue-400">
                <RefreshCw className="w-5 h-5 group-hover:rotate-180 transition-transform duration-700" />
              </div>
              <div>
                <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Update Rate</div>
                <div className="text-sm font-bold text-white">2,000ms</div>
              </div>
           </div>
        </div>
      </div>

      {/* Primary KPI Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
        <KPIWidget 
          label="Flow Input" 
          value={`${metrics?.throughput.toFixed(1)}`} 
          unit="evt/s"
          sub="Global ingestion flow"
          icon={<ArrowRightLeft />}
          color="emerald"
        />
        <KPIWidget 
          label="P99 Latency" 
          value={`${metrics?.latency_p99.toFixed(1)}`} 
          unit="ms"
          sub="End-to-end delay"
          icon={<Clock />}
          color="blue"
        />
        <KPIWidget 
          label="Execution Load" 
          value={`${metrics?.active_workers}`} 
          unit="threads"
          sub="Active parallel workers"
          icon={<Cpu />}
          color="indigo"
        />
        <KPIWidget 
          label="Health Matrix" 
          value={`${metrics?.error_rate.toFixed(2)}`} 
          unit="%"
          sub="Delivery failure"
          icon={<AlertTriangle />}
          color="red"
          alert={metrics && metrics.error_rate > 1}
        />
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        
        {/* Main Throughput Chart */}
        <div className="lg:col-span-2 p-10 rounded-[3rem] bg-slate-900/40 border border-white/5 shadow-2xl space-y-8 relative overflow-hidden group">
          <div className="absolute inset-0 bg-gradient-to-tr from-emerald-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-1000" />
          <div className="flex items-center justify-between relative z-10">
            <div className="space-y-1">
              <h3 className="text-2xl font-bold text-white tracking-tight">System Throughput</h3>
              <p className="text-slate-500 text-xs font-medium uppercase tracking-widest">Rolling flow telemetry</p>
            </div>
            <div className="flex gap-2">
               <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
               <span className="text-[10px] font-bold text-slate-400">LIVE SCANNING</span>
            </div>
          </div>
          <div className="h-[350px] w-full relative z-10">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={history}>
                <defs>
                  <linearGradient id="colorFlow" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="rgba(255,255,255,0.03)" />
                <XAxis dataKey="time" hide />
                <YAxis axisLine={false} tickLine={false} tick={{fill: '#475569', fontSize: 10}} />
                <Tooltip 
                  contentStyle={{ backgroundColor: 'rgba(15, 23, 42, 0.9)', borderColor: 'rgba(255,255,255,0.1)', borderRadius: '24px', backdropFilter: 'blur(10px)', border: '1px solid rgba(255,255,255,0.05)' }}
                  itemStyle={{color: '#10b981', fontSize: '14px', fontWeight: 'bold'}}
                  cursor={{ stroke: '#334155', strokeWidth: 2 }}
                />
                <Area 
                  type="monotone" 
                  dataKey="throughput" 
                  stroke="#10b981" 
                  strokeWidth={4} 
                  fillOpacity={1} 
                  fill="url(#colorFlow)" 
                  animationDuration={1500}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Source vs Sink Dist */}
        <div className="p-10 rounded-[3rem] bg-slate-900/40 border border-white/5 shadow-2xl space-y-8 flex flex-col relative overflow-hidden group">
           <div className="absolute inset-0 bg-gradient-to-tr from-blue-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-1000" />
           <div className="space-y-1 relative z-10">
              <h3 className="text-2xl font-bold text-white tracking-tight">Component Health</h3>
              <p className="text-slate-500 text-xs font-medium uppercase tracking-widest">Active pipeline distribution</p>
            </div>
            <div className="flex-1 flex flex-col justify-center relative z-10">
               <div className="h-[300px] w-full">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Tooltip 
                         contentStyle={{ backgroundColor: 'rgba(15, 23, 42, 0.9)', borderColor: 'rgba(255,255,255,0.1)', borderRadius: '20px', backdropFilter: 'blur(10px)' }}
                      />
                      <Pie 
                        data={[
                           { name: 'Sources', value: sourceData.length },
                           { name: 'Sinks', value: sinkData.length }
                        ]} 
                        cx="50%" 
                        cy="50%" 
                        innerRadius={80} 
                        outerRadius={100} 
                        paddingAngle={10} 
                        dataKey="value"
                        stroke="none"
                      >
                        <Cell fill="#3b82f6" />
                        <Cell fill="#10b981" />
                      </Pie>
                    </PieChart>
                  </ResponsiveContainer>
                  <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
                     <span className="text-4xl font-black text-white">{sourceData.length + sinkData.length}</span>
                     <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Total Nodes</span>
                  </div>
               </div>
            </div>
            <div className="space-y-4 relative z-10">
               <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                     <div className="w-2 h-2 rounded-full bg-blue-500" />
                     <span className="text-xs font-bold text-slate-400">Sources</span>
                  </div>
                  <span className="text-xs font-black text-white">{sourceData.length}</span>
               </div>
               <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                     <div className="w-2 h-2 rounded-full bg-emerald-500" />
                     <span className="text-xs font-bold text-slate-400">Sinks</span>
                  </div>
                  <span className="text-xs font-black text-white">{sinkData.length}</span>
               </div>
            </div>
        </div>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-12">
        {/* Source Activity Matrix */}
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-3xl font-black text-white tracking-tighter flex items-center gap-4">
              <div className="p-3 rounded-2xl bg-blue-500/10 border border-blue-500/20">
                <ChevronsRight className="w-6 h-6 text-blue-400" />
              </div>
              Source Ingestion
            </h2>
          </div>
          
          <div className="grid grid-cols-1 gap-4">
             {sourceData.length > 0 ? sourceData.map((source, idx) => (
               <div key={idx} className="p-8 rounded-[2.5rem] bg-white/[0.02] border border-white/5 hover:bg-white/[0.04] transition-all group relative overflow-hidden flex items-center justify-between">
                  <div className="flex items-center gap-6">
                    <div className="w-12 h-12 rounded-2xl bg-blue-500/10 flex items-center justify-center">
                       <Database className="w-6 h-6 text-blue-400" />
                    </div>
                    <div>
                      <div className="text-lg font-bold text-white tracking-tight">{source.name}</div>
                      <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Active Connector</div>
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-12 text-right">
                     <div className="space-y-1">
                        <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Throughput</div>
                        <div className="text-2xl font-mono font-bold text-emerald-400 tabular-nums">
                          {source.throughput.toFixed(1)} <span className="text-[10px] text-slate-500">evt/s</span>
                        </div>
                     </div>
                     <div className="space-y-1">
                        <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Error Count</div>
                        <div className={`text-2xl font-mono font-bold tabular-nums ${source.errorRate > 0 ? 'text-rose-400' : 'text-slate-600'}`}>
                          {source.errorRate}
                        </div>
                     </div>
                  </div>
               </div>
             )) : (
               <div className="p-12 rounded-[2.5rem] border border-dashed border-white/10 text-center space-y-3">
                  <div className="text-slate-500">No active ingestion sources detected</div>
               </div>
             )}
          </div>
        </div>

        {/* Sink Matrix */}
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-3xl font-black text-white tracking-tighter flex items-center gap-4">
              <div className="p-3 rounded-2xl bg-emerald-500/10 border border-emerald-500/20">
                <Box className="w-6 h-6 text-emerald-400" />
              </div>
              Sink Delivery
            </h2>
          </div>
          
          <div className="grid grid-cols-1 gap-4">
             {sinkData.length > 0 ? sinkData.map((sink, idx) => (
               <div key={idx} className="p-8 rounded-[2.5rem] bg-white/[0.02] border border-white/5 hover:bg-white/[0.04] transition-all group relative overflow-hidden flex items-center justify-between">
                  <div className="flex items-center gap-6">
                    <div className="w-12 h-12 rounded-2xl bg-emerald-500/10 flex items-center justify-center">
                       <Zap className="w-6 h-6 text-emerald-500" />
                    </div>
                    <div>
                      <div className="text-lg font-bold text-white tracking-tight">{sink.name}</div>
                      <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Active Delivery</div>
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-12 text-right">
                     <div className="space-y-1">
                        <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Throughput</div>
                        <div className="text-2xl font-mono font-bold text-emerald-400 tabular-nums">
                          {sink.throughput.toFixed(1)} <span className="text-[10px] text-slate-500">evt/s</span>
                        </div>
                     </div>
                     <div className="space-y-1">
                        <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Latency</div>
                        <div className={`text-2xl font-mono font-bold tabular-nums ${sink.latency > 100 ? 'text-rose-400' : 'text-blue-400'}`}>
                          {sink.latency.toFixed(1)} <span className="text-[10px] text-slate-500">ms</span>
                        </div>
                     </div>
                  </div>
               </div>
             )) : (
               <div className="p-12 rounded-[2.5rem] border border-dashed border-white/10 text-center space-y-3">
                  <div className="text-slate-500">No active delivery sinks detected</div>
               </div>
             )}
          </div>
        </div>
      </div>
    </div>
  );
}

function KPIWidget({ label, value, unit, sub, icon, color, alert }: any) {
  const colors: any = {
    emerald: "text-emerald-400 bg-emerald-500/10 border-emerald-500/20 shadow-emerald-500/5",
    blue: "text-blue-400 bg-blue-500/10 border-blue-500/20 shadow-blue-500/5",
    indigo: "text-indigo-400 bg-indigo-500/10 border-indigo-500/20 shadow-indigo-500/5",
    red: "text-rose-400 bg-rose-500/10 border-rose-500/20 shadow-rose-500/5"
  };

  return (
    <div className={`p-10 rounded-[3rem] border border-white/5 bg-slate-900/40 backdrop-blur-3xl hover:bg-slate-800/60 transition-all hover:-translate-y-2 shadow-2xl relative overflow-hidden group ${alert ? 'ring-2 ring-rose-500/50' : ''}`}>
      <div className="absolute top-0 right-0 p-8 opacity-10 group-hover:scale-150 transition-transform duration-1000">
        {icon}
      </div>
      
      <div className={`w-14 h-14 rounded-[1.25rem] flex items-center justify-center mb-8 shadow-2xl ${colors[color]}`}>
        {icon}
      </div>
      
      <div className="space-y-2">
        <div className="text-[11px] font-black text-slate-500 uppercase tracking-[0.2em]">{label}</div>
        <div className="flex items-baseline gap-2">
          <div className="text-5xl font-mono font-black text-white tracking-tighter tabular-nums">{value}</div>
          <div className="text-sm font-bold text-slate-500">{unit}</div>
        </div>
        <div className="text-xs text-slate-400 font-medium pt-2 flex items-center gap-2">
           <div className={`w-1 h-1 rounded-full ${alert ? 'bg-rose-500 animate-pulse' : 'bg-slate-700'}`} />
           {sub}
        </div>
      </div>
      
      {alert && (
        <div className="absolute inset-0 bg-rose-500/5 pointer-events-none animate-pulse" />
      )}
    </div>
  );
}
