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
    <div className="space-y-4 animate-in fade-in duration-700 pb-10">
      {/* Dynamic Header */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
        <div className="space-y-0.5">
          <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-blue-500/10 border border-blue-500/10 text-blue-500 overline-label overline-label-solid">
            <Monitor className="w-3 h-3" />
            Infrastructure Real-time
          </div>
          <h1 className="text-xl md:text-2xl font-black text-foreground tracking-tight">
            Engine Performance
          </h1>
          <p className="text-muted-foreground body-compact max-w-xl line-clamp-1">
            End-to-end observability from ingestion to delivery. Monitoring real-time pressure across all pipelines.
          </p>
        </div>
        
        <div className="flex gap-2">
           <div className="px-3 py-1.5 rounded-xl bg-card border border-border flex items-center gap-2 group hover:border-emerald-500/30 transition-all shadow-sm">
              <div className="w-6 h-6 rounded-lg bg-emerald-500/10 flex items-center justify-center text-emerald-500">
                <Server className="w-3.5 h-3.5" />
              </div>
              <div className="flex flex-col">
                <div className="overline-label-micro">Node Status</div>
                <div className="text-xs-compact font-bold text-emerald-500 flex items-center gap-1">
                  Operational
                  <div className="w-1 h-1 rounded-full bg-emerald-500 animate-pulse" />
                </div>
              </div>
           </div>
           <div className="px-3 py-1.5 rounded-xl bg-card border border-border flex items-center gap-2 group hover:border-blue-500/30 transition-all shadow-sm">
              <div className="w-6 h-6 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-500">
                <RefreshCw className="w-3.5 h-3.5 group-hover:rotate-180 transition-transform duration-700" />
              </div>
              <div className="flex flex-col">
                <div className="overline-label-micro">Update Rate</div>
                <div className="text-xs-compact font-bold text-foreground">2,000ms</div>
              </div>
           </div>
        </div>
      </div>

      {/* Primary KPI Grid */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
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
          label="Load" 
          value={`${metrics?.active_workers}`} 
          unit="threads"
          sub="Active workers"
          icon={<Cpu />}
          color="indigo"
        />
        <KPIWidget 
          label="Health" 
          value={`${metrics?.error_rate.toFixed(2)}`} 
          unit="%"
          sub="Delivery failure"
          icon={<AlertTriangle />}
          color="red"
          alert={metrics && metrics.error_rate > 1}
        />
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Main Throughput Chart */}
        <div className="lg:col-span-2 p-4 rounded-xl bg-card border border-border shadow-sm space-y-4 relative overflow-hidden group">
          <div className="flex items-center justify-between relative z-10">
            <div className="space-y-0.5">
              <h3 className="text-sm font-bold text-foreground">System Throughput</h3>
              <p className="overline-label overline-label-dim">Rolling flow telemetry</p>
            </div>
            <div className="flex items-center gap-1.5">
               <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
               <span className="overline-label overline-label-solid">Live</span>
            </div>
          </div>
          <div className="h-[200px] w-full relative z-10">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={history}>
                <defs>
                  <linearGradient id="colorFlow" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.1}/>
                    <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--border))" opacity={0.5} />
                <XAxis dataKey="time" hide />
                <YAxis axisLine={false} tickLine={false} tick={{fill: 'currentColor', fontSize: 10, opacity: 0.5}} />
                <Tooltip 
                  contentStyle={{ backgroundColor: 'hsl(var(--card))', borderColor: 'hsl(var(--border))', borderRadius: '8px', fontSize: '12px' }}
                />
                <Area 
                  type="monotone" 
                  dataKey="throughput" 
                  stroke="hsl(var(--primary))" 
                  strokeWidth={2} 
                  fillOpacity={1} 
                  fill="url(#colorFlow)" 
                  animationDuration={1500}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Source vs Sink Dist */}
        <div className="p-4 rounded-xl bg-card border border-border shadow-sm space-y-4 flex flex-col relative overflow-hidden group">
           <div className="space-y-0.5 relative z-10">
              <h3 className="text-sm font-bold text-foreground">Component Health</h3>
              <p className="overline-label overline-label-dim">Node distribution</p>
            </div>
            <div className="flex-1 flex flex-col justify-center relative z-10 py-2">
               <div className="h-[140px] w-full relative">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Tooltip 
                         contentStyle={{ backgroundColor: 'hsl(var(--card))', borderColor: 'hsl(var(--border))', borderRadius: '8px', fontSize: '11px' }}
                      />
                      <Pie 
                        data={[
                           { name: 'Sources', value: sourceData.length },
                           { name: 'Sinks', value: sinkData.length }
                        ]} 
                        cx="50%" 
                        cy="50%" 
                        innerRadius={45} 
                        outerRadius={60} 
                        paddingAngle={8} 
                        dataKey="value"
                        stroke="none"
                      >
                        <Cell fill="hsl(var(--primary))" />
                        <Cell fill="#10b981" />
                      </Pie>
                    </PieChart>
                  </ResponsiveContainer>
                  <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
                     <span className="text-xl font-black text-foreground leading-none">{sourceData.length + sinkData.length}</span>
                     <span className="overline-label-micro">Nodes</span>
                  </div>
               </div>
            </div>
            <div className="space-y-1.5 relative z-10">
               <div className="flex items-center justify-between px-1">
                  <div className="flex items-center gap-1.5">
                     <div className="w-1.5 h-1.5 rounded-full bg-primary" />
                     <span className="text-xs-compact font-bold text-muted-foreground">Sources</span>
                  </div>
                  <span className="text-xs-compact font-black text-foreground">{sourceData.length}</span>
               </div>
               <div className="flex items-center justify-between px-1">
                  <div className="flex items-center gap-1.5">
                     <div className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
                     <span className="text-xs-compact font-bold text-muted-foreground">Sinks</span>
                  </div>
                  <span className="text-xs-compact font-black text-foreground">{sinkData.length}</span>
               </div>
            </div>
        </div>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
        {/* Source Activity Matrix */}
        <div className="space-y-3">
          <h2 className="text-sm font-black text-foreground flex items-center gap-2 px-1">
            <ChevronsRight className="w-4 h-4 text-primary" />
            Source Ingestion
          </h2>
          
          <div className="grid grid-cols-1 gap-2">
             {sourceData.length > 0 ? sourceData.map((source, idx) => (
               <div key={idx} className="p-3 rounded-xl bg-card border border-border hover:bg-muted/30 transition-all group flex items-center justify-between shadow-sm">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center">
                       <Database className="w-4 h-4 text-primary" />
                    </div>
                    <div>
                      <div className="text-xs font-bold text-foreground leading-none mb-1">{source.name}</div>
                      <div className="overline-label-micro">Active Connector</div>
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-6 text-right">
                     <div className="space-y-0.5">
                        <div className="overline-label-micro">Throughput</div>
                        <div className="text-xs font-black text-emerald-500 tabular-nums">
                          {source.throughput.toFixed(1)} <span className="text-xs-compact opacity-40">E/s</span>
                        </div>
                     </div>
                     <div className="space-y-0.5">
                        <div className="overline-label-micro">Errors</div>
                        <div className={`text-xs font-black tabular-nums ${source.errorRate > 0 ? 'text-rose-500' : 'text-muted-foreground'}`}>
                          {source.errorRate}
                        </div>
                     </div>
                  </div>
               </div>
             )) : (
               <div className="p-8 rounded-xl border-2 border-dashed border-border text-center">
                  <p className="text-xs text-muted-foreground font-medium">No active ingestion sources</p>
               </div>
             )}
          </div>
        </div>

        {/* Sink Matrix */}
        <div className="space-y-3">
          <h2 className="text-sm font-black text-foreground flex items-center gap-2 px-1">
            <Box className="w-4 h-4 text-emerald-500" />
            Sink Delivery
          </h2>
          
          <div className="grid grid-cols-1 gap-2">
             {sinkData.length > 0 ? sinkData.map((sink, idx) => (
               <div key={idx} className="p-3 rounded-xl bg-card border border-border hover:bg-muted/30 transition-all group flex items-center justify-between shadow-sm">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-lg bg-emerald-500/10 flex items-center justify-center">
                       <Zap className="w-4 h-4 text-emerald-500" />
                    </div>
                    <div>
                      <div className="text-xs font-bold text-foreground leading-none mb-1">{sink.name}</div>
                      <div className="overline-label-micro">Active Delivery</div>
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-6 text-right">
                     <div className="space-y-0.5">
                        <div className="overline-label-micro">Throughput</div>
                        <div className="text-xs font-black text-emerald-500 tabular-nums">
                          {sink.throughput.toFixed(1)} <span className="text-xs-compact opacity-40">E/s</span>
                        </div>
                     </div>
                     <div className="space-y-0.5">
                        <div className="overline-label-micro">Latency</div>
                        <div className={`text-xs font-black tabular-nums ${sink.latency > 100 ? 'text-rose-500' : 'text-primary'}`}>
                          {sink.latency.toFixed(1)} <span className="text-xs-compact opacity-40">ms</span>
                        </div>
                     </div>
                  </div>
               </div>
             )) : (
               <div className="p-8 rounded-xl border-2 border-dashed border-border text-center">
                  <p className="text-xs text-muted-foreground font-medium">No active delivery sinks</p>
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
    emerald: "text-emerald-500 bg-emerald-500/10 border-emerald-500/10",
    blue: "text-primary bg-primary/10 border-primary/10",
    indigo: "text-indigo-500 bg-indigo-500/10 border-indigo-500/10",
    red: "text-rose-500 bg-rose-500/10 border-rose-500/10"
  };

  return (
    <div className={`p-4 rounded-xl border border-border bg-card shadow-sm hover:translate-y-[-2px] transition-all relative overflow-hidden group ${alert ? 'ring-1 ring-rose-500/50' : ''}`}>
      <div className="flex items-center justify-between mb-4">
        <div className={`w-8 h-8 rounded-lg flex items-center justify-center border ${colors[color]}`}>
          {icon && (typeof icon === 'object' ? { ...icon, props: { ...icon.props, className: "w-4 h-4" } } : icon)}
        </div>
        <div className="overline-label overline-label-faint">{unit}</div>
      </div>
      
      <div className="space-y-1">
        <div className="overline-label-micro">{label}</div>
        <div className="text-2xl font-black text-foreground tracking-tighter tabular-nums leading-none">{value}</div>
        <div className="text-2xs text-muted-foreground font-medium flex items-center gap-1.5 pt-1">
           <div className={`w-1 h-1 rounded-full ${alert ? 'bg-rose-500 animate-pulse' : 'bg-muted-foreground/30'}`} />
           {sub}
        </div>
      </div>
      
      {alert && (
        <div className="absolute inset-0 bg-rose-500/5 pointer-events-none animate-pulse" />
      )}
    </div>
  );
}
