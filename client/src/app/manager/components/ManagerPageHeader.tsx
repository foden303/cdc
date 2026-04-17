import { Plus, Layers } from "lucide-react";
import type { TranslationKey } from "@/lib/i18n";

type TFn = (key: TranslationKey) => string;

type Props = {
  onAdd: (kind: "source" | "sink") => void;
  t: TFn;
};

export function ManagerPageHeader({ onAdd, t }: Props) {
  return (
    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div className="space-y-0.5">
        <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-primary/10 border border-primary/10 text-primary overline-label overline-label-solid">
          <Layers className="w-3 h-3" /> {t("orchestration")}
        </div>
        <h1 className="text-xl md:text-2xl font-black text-foreground">{t("pipelineManager")}</h1>
        <p className="text-muted-foreground text-base-compact font-medium max-w-lg">{t("controlPlane")}</p>
      </div>
      <div className="flex gap-2">
        <button
          type="button"
          onClick={() => onAdd("source")}
          className="px-3.5 py-1.5 rounded-lg bg-primary text-white font-black text-base-compact flex items-center gap-2 transition-all shadow-sm"
        >
          <Plus className="w-3.5 h-3.5" /> {t("addSource")}
        </button>
        <button
          type="button"
          onClick={() => onAdd("sink")}
          className="px-3.5 py-1.5 rounded-lg bg-emerald-600 text-white font-black text-base-compact flex items-center gap-2 transition-all shadow-sm"
        >
          <Plus className="w-3.5 h-3.5" /> {t("addSink")}
        </button>
      </div>
    </div>
  );
}
