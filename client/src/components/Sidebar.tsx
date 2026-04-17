"use client";

import { usePathname, useRouter } from "next/navigation";
import { useApp } from "@/lib/AppContext";
import { TranslationKey } from "@/lib/i18n";
import { 
  LayoutDashboard, 
  Layers, 
  Grid2X2, 
  MessageSquare, 
  Settings, 
  Zap,
  Activity,
  Search,
  ShieldCheck,
  Clock
} from "lucide-react";
import { useEffect, useState } from "react";
import { HealthCheckResponse } from "@/lib/grpc";
import { healthCheckAction } from "@/lib/actions";

interface NavItem {
  label: TranslationKey;
  href: string;
  icon: React.ElementType;
}

const navItems: NavItem[] = [
  { label: "dashboard", href: "/", icon: LayoutDashboard },
  { label: "topics", href: "/topics", icon: Layers },
  { label: "performance", href: "/metrics", icon: Activity },
  { label: "manager", href: "/manager", icon: Settings },
  { label: "explorer", href: "/explorer", icon: Search },
  { label: "partitions", href: "/partitions", icon: Grid2X2 },
  { label: "messages", href: "/messages", icon: MessageSquare },
  { label: "sysConfig", href: "/config", icon: Settings },
];

export default function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { t } = useApp();
  const [health, setHealth] = useState<HealthCheckResponse | null>(null);

  useEffect(() => {
    const checkHealth = async () => {
      try {
        const h = await healthCheckAction();
        setHealth(h);
      } catch (err) {
        console.error("health check fail:", err);
        setHealth(null);
      }
    };
    checkHealth();
    const timer = setInterval(checkHealth, 30000);
    return () => clearInterval(timer);
  }, []);

  const formatUptime = (seconds: string | number) => {
    const s = typeof seconds === "string" ? parseInt(seconds, 10) : seconds;
    if (isNaN(s) || s <= 0) return "—";
    const h = Math.floor(s / 3600);
    const m = Math.floor((s % 3600) / 60);
    if (h > 0) return `${h}h ${m}m ${s % 60}s`;
    return `${m}m ${s % 60}s`;
  };

  return (
    <aside className="sidebar flex flex-col p-3 h-full border-r border-border bg-card">
      {/* Brand */}
      <div className="flex items-center gap-2 mb-6 px-1 cursor-pointer" onClick={() => router.push("/")}>
        <div className="p-1.5 rounded-lg bg-primary shadow-sm">
          <Zap className="w-3.5 h-3.5 text-white fill-white" />
        </div>
        <div>
          <h1 className="text-sm font-black tracking-tighter text-foreground leading-none">CDC Engine</h1>
          <p className="overline-label mt-0.5">v2.4 Core</p>
        </div>
      </div>

      {/* Primary Navigation */}
      <nav className="flex-1 space-y-0.5">
        <div className="overline-label overline-label-dim mb-2 px-2">{t("navigation")}</div>
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive =
            item.href === "/"
              ? pathname === "/"
              : pathname.startsWith(item.href);
          
          return (
            <button
              key={item.href}
              onClick={() => router.push(item.href)}
              className={`w-full flex items-center gap-2.5 px-2.5 py-1.5 rounded-lg text-base-compact font-bold transition-all group ${
                isActive 
                  ? "bg-primary/10 text-primary border border-primary/10 shadow-sm" 
                  : "text-muted-foreground hover:text-foreground hover:bg-muted"
              }`}
            >
              <Icon className={`w-3.5 h-3.5 transition-transform group-hover:scale-110 ${isActive ? "text-primary" : "text-muted-foreground/60"}`} />
              <span className="truncate">{t(item.label)}</span>
            </button>
          );
        })}
      </nav>

      {/* Footer Info */}
      <div className="mt-auto border-t border-border rounded-lg bg-muted/50 border border-border group hover:bg-muted transition-colors">
        <div>
          <div className="flex items-center gap-2 px-2 py-1.5">
            <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse shadow-[0_0_8px_rgba(16,185,129,0.3)]" />
            <div className="flex-1 min-w-0">
              <p className="text-mini-compact font-black text-foreground truncate">{t("operational")}</p>
              <p className="overline-label text-tiny !normal-case tracking-normal">{t("everythingNormal")}</p>
            </div>
            <Activity className="w-3 h-3 text-muted-foreground/40 group-hover:text-primary transition-colors" />
          </div>
        </div>
        <div className="hidden lg:flex border-t border-border items-center gap-4 px-2 py-1 text-mini-compact font-bold tabular-nums">
            {health ? (
              <>
                <div className="flex items-center gap-1.5 group transition-colors hover:text-emerald-500 cursor-default">
                  <ShieldCheck className="w-2.5 h-2.5 text-emerald-500" />
                  <span className="text-muted-foreground/80">v{health.version}</span>
                </div>
                
                <div className="flex items-center gap-1.5 group transition-colors hover:text-primary cursor-default">
                  <Clock className="w-2.5 h-2.5 text-primary" />
                  <span className="text-muted-foreground/80">{t("uptime")}: {formatUptime(health.uptime)}</span>
                </div>
              </>
            ) : (
              <div className="flex items-center gap-1.5 text-destructive animate-pulse">
                <Activity className="w-2.5 h-2.5" />
                <span>{t("offline")}</span>
              </div>
            )}
          </div>
      </div>
    </aside>
  );
}
