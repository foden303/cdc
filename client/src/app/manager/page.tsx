"use client";

import { useEffect, useState } from "react";
import { Database, Box } from "lucide-react";
import {
  getConfigAction,
  addSourceAction,
  removeSourceAction,
  updateSourceAction,
  addSinkAction,
  removeSinkAction,
  updateSinkAction,
  getStatsAction,
} from "@/lib/actions";
import type { AppConfig, GetStatsResponse, SourceConfig, SinkConfig } from "@/lib/grpc";
import { useApp } from "@/lib/AppContext";
import { useNotifications } from "@/lib/NotificationContext";
import { DEFAULT_SOURCE_TYPE, DEFAULT_SINK_TYPE } from "./registry";
import { ManagerPageHeader } from "./components/ManagerPageHeader";
import { ManagerSectionHeader } from "./components/ManagerSectionHeader";
import { ManagerConfigCard } from "./components/ManagerConfigCard";
import { SourceSinkModal } from "./components/SourceSinkModal";
import { statusFromStat } from "./utils/stats";

export default function ManagerPage() {
  const { t } = useApp();
  const { notify } = useNotifications();
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [stats, setStats] = useState<GetStatsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [modalType, setModalType] = useState<"source" | "sink">("source");
  const [editingItem, setEditingItem] = useState<Record<string, unknown> | null>(null);
  const [selectedType, setSelectedType] = useState("");
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      const [cRes, sRes] = await Promise.all([getConfigAction(), getStatsAction()]);
      console.log("fetchData====>", cRes, sRes);
      if ((cRes as { config?: AppConfig })?.config) setConfig((cRes as { config: AppConfig }).config);
      setStats(sRes);
      setLoading(false);
    };
    fetchData();
    const timer = setInterval(fetchData, 10000);
    return () => clearInterval(timer);
  }, []);

  const refreshConfig = async () => {
    const cRes = await getConfigAction();
    console.log("refreshConfig====>", cRes);
    if ((cRes as { config?: AppConfig })?.config) setConfig((cRes as { config: AppConfig }).config);
  };

  const handleRemove = async (type: "source" | "sink", id: string) => {
    if (!confirm(t("attention"))) return;
    if (type === "source") await removeSourceAction(id);
    else await removeSinkAction(id);
    await refreshConfig();
  };

  const openAddModal = (type: "source" | "sink") => {
    setModalType(type);
    setEditingItem(null);
    setSelectedType(type === "source" ? DEFAULT_SOURCE_TYPE : DEFAULT_SINK_TYPE);
    setIsModalOpen(true);
  };

  const openEditModal = (type: "source" | "sink", item: Record<string, unknown>) => {
    setModalType(type);
    setEditingItem(item);
    setSelectedType(String(item.type ?? ""));
    setIsModalOpen(true);
  };

  const handleModalSave = async (payload: Record<string, unknown>) => {
    setIsSaving(true);
    try {
      if (modalType === "source") {
        if (editingItem) {
          await updateSourceAction({ ...(editingItem as unknown as SourceConfig), ...payload } as SourceConfig);
        } else {
          await addSourceAction(payload as unknown as SourceConfig);
        }
      } else {
        if (editingItem) {
          await updateSinkAction({ ...(editingItem as unknown as SinkConfig), ...payload } as SinkConfig);
        } else {
          await addSinkAction(payload as unknown as SinkConfig);
        }
      }
      setIsModalOpen(false);
      await refreshConfig();
      if (modalType === "source") {
        notify(
          editingItem ? t("managerToastSourceUpdated") : t("managerToastSourceAdded"),
          "success"
        );
      } else {
        notify(
          editingItem ? t("managerToastSinkUpdated") : t("managerToastSinkAdded"),
          "success"
        );
      }
    } catch (err) {
      console.error(err);
      const detail = err instanceof Error ? err.message : "";
      notify(detail ? `${t("managerToastSaveFailed")}: ${detail}` : t("managerToastSaveFailed"), "error");
    } finally {
      setIsSaving(false);
    }
  };

  if (loading && !config) return <div className="p-10 animate-pulse">{t("operational")}...</div>;

  return (
    <div className="space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-500">
      <ManagerPageHeader onAdd={openAddModal} t={t} />

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
        <div className="space-y-2">
          <ManagerSectionHeader icon={Database} title={t("ingestionSources")} count={config?.sources?.length || 0} color="blue" t={t} />
          <div className="space-y-1.5">
            {config?.sources?.map((s) => (
              <ManagerConfigCard
                key={s.instance_id}
                title={s.name || s.instance_id}
                details={`${s.database} @ ${s.host}`}
                status={statusFromStat(stats?.source_stats?.[s.instance_id])}
                onEdit={() => openEditModal("source", s as unknown as Record<string, unknown>)}
                onRemove={() => handleRemove("source", s.instance_id)}
                color="blue"
              />
            ))}
          </div>
        </div>

        <div className="space-y-2">
          <ManagerSectionHeader icon={Box} title={t("deliverySinks")} count={config?.sinks?.length || 0} color="emerald" t={t} />
          <div className="space-y-1.5">
            {config?.sinks?.map((s) => (
              <ManagerConfigCard
                key={s.instance_id}
                title={s.name || s.instance_id}
                details={s.url?.[0] || "N/A"}
                status={statusFromStat(stats?.sink_stats?.[s.instance_id])}
                onEdit={() => openEditModal("sink", s as unknown as Record<string, unknown>)}
                onRemove={() => handleRemove("sink", s.instance_id)}
                color="emerald"
              />
            ))}
          </div>
        </div>
      </div>

      {isModalOpen && (
        <SourceSinkModal
          kind={modalType}
          editingItem={editingItem}
          selectedType={selectedType}
          onSelectedTypeChange={setSelectedType}
          isSaving={isSaving}
          onClose={() => setIsModalOpen(false)}
          onSave={handleModalSave}
          t={t}
        />
      )}
    </div>
  );
}
