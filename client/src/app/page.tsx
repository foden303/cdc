"use client";

import { useEffect, useState } from "react";
import { getPerformanceMetricsAction, getStatsAction, reprocessDLQAction, healthCheckAction } from "@/lib/actions";
import type { GetStatsResponse, HealthCheckResponse, GetPerformanceMetricsResponse } from "@/lib/grpc";
import { useApp } from "@/lib/AppContext";
import { 
  Zap, 
  Activity, 
  ArrowUpRight, 
  ShieldCheck, 
  Clock, 
  Database,
  Search,
  AlertCircle,
  RotateCcw,
  CheckCircle2,
  AlertTriangle,
  Cpu,
  Network,
  Boxes,
  ArrowRight,
  Server
} from "lucide-react";
import Link from "next/link";

export default function DashboardPage() {
  const { t } = useApp();
  const [stats, setStats] = useState<GetStatsResponse | null>(null);
  const [health, setHealth] = useState<HealthCheckResponse | null>(null);
  const [performance, setPerformance] = useState<GetPerformanceMetricsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [reprocessing, setReprocessing] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [s, h, p] = await Promise.all([
          getStatsAction(), 
          healthCheckAction(),
          getPerformanceMetricsAction()
        ]);
        setStats(s);
        setHealth(h);
        setPerformance(p);
      } catch (err) {
        console.error("fetch dashboard data:", err);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
    const statsTimer = setInterval(fetchData, 10000);
    const fastTimer = setInterval(async () => {
      const p = await getPerformanceMetricsAction();
      if (p) setPerformance(p);
    }, 2000);

    return () => {
      clearInterval(statsTimer);
      clearInterval(fastTimer);
    };
  }, []);

  const handleReprocessDLQ = async () => {
    setReprocessing(true);
    try {
      const res = await reprocessDLQAction();
      alert(`Success: ${res?.processed_count || 0} events re-queued.`);
    } catch (err) {
      console.error(err);
      alert("Failed to reprocess DLQ messages.");
    } finally {
      setReprocessing(false);
    }
  };

  if (loading && !stats) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 animate-pulse">
        {[...Array(4)].map((_, i) => (
          <div key={i} className="h-28 rounded-2xl bg-white/5 border border-white/5" />
        ))}
      </div>
    );
  }

  const totalFailure = Object.values(stats?.source_stats || {}).reduce((s, c) => s + c.failure_count, 0) +
                       Object.values(stats?.sink_stats || {}).reduce((s, c) => s + c.failure_count, 0);

  return (
    <div className="space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-700">
      {/* Compact Header */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
        <div className="space-y-0.5">
          <div className="flex items-center gap-2 mb-0.5">
            <div className={`w-1.5 h-1.5 rounded-full ${health?.status === 'healthy' ? 'bg-emerald-500' : 'bg-rose-500'} animate-pulse shadow-[0_0_8px_rgba(var(--primary),0.2)]`} />
            <span className="overline-label overline-label-soft">{t("operational")}</span>
          </div>
          <h1 className="text-xl md:text-2xl font-black tracking-tight text-foreground">{t("engineCore")}</h1>
          <p className="text-muted-foreground body-compact max-w-sm">{t("syncEngine")}</p>
        </div>

        <div className="flex items-center gap-2">
          <MetricBadge icon={Cpu} label={t("activeThreads")} value={String(performance?.active_workers || 0)} color="text-blue-500" />
          <MetricBadge icon={Database} label={t("globalNodes")} value={String(Object.keys(stats?.source_stats || {}).length + Object.keys(stats?.sink_stats || {}).length)} color="text-emerald-500" />
        </div>
      </div>

      {/* KPI Strip */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <KPICard 
          title={t("throughput")}
          value={`${performance?.throughput?.toLocaleString() || "0"}`}
          unit="EPS"
          subValue={t("processedEvents")}
          icon={Zap}
          trend="+12.5%"
          color="blue"
        />
        <KPICard 
          title={t("p99Latency")}
          value={`${performance?.latency_p99?.toFixed(1) || "0"}`}
          unit="ms"
          subValue={t("avgP99")}
          icon={Clock}
          trend="-2.4ms"
          color="emerald"
        />
        <KPICard 
          title={t("dlq")}
          value={totalFailure.toString()}
          unit={t("attention")}
          subValue="Unresolved failures"
          icon={AlertCircle}
          trend="+0"
          color="rose"
        />
        <KPICard 
          title={t("systemStatus")}
          value={health?.status === 'healthy' ? "99.9" : "85.2"}
          unit="%"
          subValue={t("uptime")}
          icon={ShieldCheck}
          trend={health?.status === 'healthy' ? "STABLE" : "DEGRADED"}
          color={health?.status === 'healthy' ? 'indigo' : 'rose'}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-4">
        {/* Phase 1 & 2 Section */}
        <section className="lg:col-span-8 space-y-3">
          <div className="flex items-center gap-2 px-1">
            <div className="p-1.5 rounded-lg bg-primary/10 border border-primary/10">
              <Network className="w-3.5 h-3.5 text-primary" />
            </div>
            <div>
              <h2 className="text-sm font-black text-foreground leading-none">{t("phase1_2")}</h2>
              <p className="overline-label mt-1 overline-label-dim">{t("ingestionDelivery")}</p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
            {Object.entries(stats?.source_stats || {}).map(([name, s]) => (
              <DetailCard key={name} name={name} success={s.success_count} failure={s.failure_count} />
            ))}
          </div>
        </section>

        {/* Phase 3 & Reliability */}
        <section className="lg:col-span-4 space-y-3">
          <div className="flex items-center gap-2 px-1">
            <div className="p-1.5 rounded-lg bg-rose-500/10 border border-rose-500/10">
              <ShieldCheck className="w-3.5 h-3.5 text-rose-500" />
            </div>
            <div>
              <h2 className="text-sm font-black text-foreground leading-none">{t("phase3")}</h2>
              <p className="overline-label mt-1 overline-label-dim">{t("reliabilityControl")}</p>
            </div>
          </div>

          <div className="glass-panel p-4 space-y-4">
            <div className="flex items-center justify-between gap-4">
              <div className="space-y-1">
                <p className="text-sm-compact font-black text-foreground">{t("requeueFailures")}</p>
                <p className="caption-compact text-muted-foreground">
                  Manual trigger to replay events currently stored in the Dead Letter Queue.
                </p>
              </div>
              <div className="shrink-0 p-2 rounded-full bg-rose-500/10 border border-rose-500/10">
                <AlertTriangle className="w-4 h-4 text-rose-500" />
              </div>
            </div>

            <button 
              onClick={handleReprocessDLQ}
              disabled={reprocessing || totalFailure === 0}
              className="w-full py-2 rounded-lg bg-foreground text-background text-sm-compact font-black hover:opacity-90 disabled:opacity-30 transition-all flex items-center justify-center gap-2 shadow-sm"
            >
              {reprocessing ? <Activity className="w-3 h-3 animate-spin" /> : <RotateCcw className="w-3 h-3" />}
              {t("requeueFailures")}
            </button>
          </div>
        </section>
      </div>
    </div>
  );
}

function MetricBadge({ icon: Icon, label, value, color }: { icon: any, label: string, value: string, color: string }) {
  return (
    <div className="flex items-center gap-2 px-4 py-2 rounded-lg border border-border bg-muted/20">
      <Icon className={`w-3 h-3 ${color}`} />
      <div className="flex flex-col">
        <span className="overline-label-tiny leading-none mb-0.5">{label}</span>
        <span className="text-sm-compact font-black text-foreground leading-none tabular-nums">{value}</span>
      </div>
    </div>
  );
}

function KPICard({ title, value, unit, subValue, icon: Icon, trend, color }: any) {
  const colorMap: any = {
    blue: "bg-blue-500/10 text-blue-500 border-blue-500/10",
    emerald: "bg-emerald-500/10 text-emerald-500 border-emerald-500/10",
    rose: "bg-rose-500/10 text-rose-500 border-rose-500/10",
    indigo: "bg-indigo-500/10 text-indigo-500 border-indigo-500/10"
  };

  return (
    <div className="glass-panel p-3 flex flex-col justify-between group">
      <div className="flex items-center justify-between mb-2">
        <div className={`p-1.5 rounded-lg ${colorMap[color]} group-hover:scale-105 transition-transform`}>
          <Icon className="w-3.5 h-3.5" />
        </div>
        <div className="flex items-center gap-1 px-1.5 py-0.5 rounded-full bg-muted/30 border border-border">
          <span className={`text-tiny font-black tabular-nums ${trend.startsWith('+') ? 'text-emerald-500' : 'text-muted-foreground'}`}>{trend}</span>
        </div>
      </div>
      <div>
        <p className="overline-label mb-1">{title}</p>
        <div className="flex items-baseline gap-1">
          <span className="text-lg font-black text-foreground tracking-tighter tabular-nums leading-none">{value}</span>
          <span className="overline-label text-muted-foreground">{unit}</span>
        </div>
        <p className="text-tiny font-bold text-muted-foreground/60 mt-1 uppercase tracking-tighter">{subValue}</p>
      </div>
    </div>
  );
}

function DetailCard({ name, success, failure }: { name: string, success: number, failure: number }) {
  return (
    <div className="glass-panel p-2.5 group hover:bg-muted/30 transition-all border-l-2 border-l-primary/30">
      <div className="flex items-center justify-between mb-1.5">
        <div className="flex items-center gap-1.5">
          <div className="w-1 h-1 rounded-full bg-emerald-500" />
          <span className="text-xs-compact font-black text-foreground uppercase truncate w-full">{name}</span>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-2 border-t border-border/50 pt-1.5">
        <div>
          <p className="overline-label-tiny overline-label-dim mb-0.5">Success</p>
          <p className="text-sm-compact font-black text-emerald-500 tabular-nums">{success.toLocaleString()}</p>
        </div>
        <div>
          <p className="overline-label-tiny overline-label-dim mb-0.5">Failures</p>
          <p className="text-sm-compact font-black text-rose-500 tabular-nums">{failure.toLocaleString()}</p>
        </div>
      </div>
    </div>
  );
}
