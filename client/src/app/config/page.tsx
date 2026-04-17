"use client";

import { useEffect, useState } from "react";
import { getConfigAction } from "@/lib/actions";
import { AppConfig } from "@/lib/grpc";
import { useApp } from "@/lib/AppContext";
import { 
  Settings, 
  Zap, 
  Database, 
  Activity, 
  Loader2,
  Lock
} from "lucide-react";

export default function ConfigPage() {
  const { t } = useApp();
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchConfig();
  }, []);

  const fetchConfig = async () => {
    setLoading(true);
    try {
      const res = await getConfigAction() as { config: AppConfig };
      if (res?.config) setConfig(res.config);
    } catch (err) {
      console.error(err);
    }
    setLoading(false);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="w-6 h-6 text-primary animate-spin" />
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-500 pb-10">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="space-y-0.5">
          <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-primary/10 border border-primary/10 text-primary overline-label overline-label-solid">
            <Settings className="w-3 h-3" /> Core Infrastructure
          </div>
          <h1 className="text-xl md:text-2xl font-black text-foreground">{t("settings")}</h1>
          <p className="text-muted-foreground body-compact">Read-only system core parameters determined by manifest.</p>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4">
        {/* App Info */}
        <ConfigSection icon={<Activity className="w-3.5 h-3.5 text-primary" />} title="App Engine">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <ReadOnlyInput label="App Name" value={config?.name} />
            <ReadOnlyInput label="Log Level" value={config?.log_mode} />
          </div>
        </ConfigSection>

        {/* Pipeline Info */}
        <ConfigSection icon={<Zap className="w-3.5 h-3.5 text-emerald-500" />} title="Pipeline Core">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <ReadOnlyInput label="Buffer" value={config?.pipeline?.channel_buffer_size} />
            <ReadOnlyInput label="Workers" value={config?.pipeline?.worker_count} />
            <ReadOnlyInput label="Batch" value={config?.pipeline?.batch_size} />
            <ReadOnlyInput label="Interval" value={config?.pipeline?.flush_interval_ms + "ms"} />
          </div>
        </ConfigSection>

        {/* NATS Info */}
        <ConfigSection icon={<Database className="w-3.5 h-3.5 text-indigo-500" />} title="NATS Cluster">
          <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
            <ReadOnlyInput label="Stream" value={config?.nats?.stream_name} />
            <ReadOnlyInput label="Retention" value={config?.nats?.retention_days + "d"} />
            <ReadOnlyInput label="Max Reconnect" value={config?.nats?.max_reconnects} />
          </div>
        </ConfigSection>
      </div>
    </div>
  );
}

function ConfigSection({ icon, title, children }: any) {
  return (
    <section className="bg-card rounded-xl overflow-hidden border border-border shadow-sm">
      <div className="px-3 py-1.5 border-b border-border bg-muted/30 flex items-center gap-2">
        {icon}
        <h2 className="text-sm-compact font-black text-foreground">{title}</h2>
      </div>
      <div className="p-3">{children}</div>
    </section>
  );
}

function ReadOnlyInput({ label, value }: any) {
  return (
    <div className="space-y-1 ui-opacity-soft">
      <div className="flex items-center gap-1 ml-0.5">
        <label className="overline-label-tiny">{label}</label>
        <Lock className="w-2 h-2 text-muted-foreground" />
      </div>
      <div className="w-full bg-muted/20 border border-border rounded-lg px-2.5 py-1 text-sm-compact text-muted-foreground font-bold tabular-nums">
        {value || "N/A"}
      </div>
    </div>
  );
}
