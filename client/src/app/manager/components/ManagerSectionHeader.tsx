import type { LucideIcon } from "lucide-react";
import type { TranslationKey } from "@/lib/i18n";

type TFn = (key: TranslationKey) => string;

type Props = {
  icon: LucideIcon;
  title: string;
  count: number;
  color: "blue" | "emerald";
  t: TFn;
};

export function ManagerSectionHeader({ icon: Icon, title, count, color, t }: Props) {
  return (
    <div className="flex items-center justify-between px-1">
      <h2 className="text-sm font-black flex items-center gap-2">
        <Icon className={`w-4 h-4 ${color === "blue" ? "text-primary" : "text-emerald-500"}`} />
        {title}
      </h2>
      <span className="overline-label text-tiny overline-label-dim">
        {count} {t("active")}
      </span>
    </div>
  );
}
