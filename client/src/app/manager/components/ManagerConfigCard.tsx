import { Database, Box, Settings, Trash2 } from "lucide-react";

type Props = {
  title: string;
  details: string;
  status: string;
  onEdit: () => void;
  onRemove: () => void;
  color: "blue" | "emerald";
};

export function ManagerConfigCard({ title, details, status, onEdit, onRemove, color }: Props) {
  const statusColors: Record<string, string> = {
    running: "text-emerald-500 bg-emerald-500/10",
    error: "text-rose-500 bg-rose-500/10",
    starting: "text-blue-500 bg-blue-500/10",
  };
  return (
    <div className="p-2 rounded-xl bg-card border border-border flex items-center justify-between group transition-all hover:border-primary/20 hover:shadow-sm">
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg bg-muted/50 border border-border flex items-center justify-center shadow-sm">
          {color === "blue" ? (
            <Database className="w-4 h-4 text-primary" />
          ) : (
            <Box className="w-4 h-4 text-emerald-500" />
          )}
        </div>
        <div className="space-y-0.5">
          <div className="flex items-center gap-1.5">
            <h4 className="text-base-compact font-black tracking-tight">{title}</h4>
            <div className={`px-1 py-0.25 rounded-md text-micro-compact font-black uppercase ${statusColors[status] ?? ""}`}>
              {status}
            </div>
          </div>
          <p className="text-mini-compact text-muted-foreground font-medium truncate max-w-[150px] lg:max-w-[250px]">
            {details}
          </p>
        </div>
      </div>
      <div className="flex gap-1">
        <button
          type="button"
          onClick={onEdit}
          className="p-1.5 rounded-md hover:bg-muted text-muted-foreground hover:text-primary transition-colors"
        >
          <Settings className="w-3.5 h-3.5" />
        </button>
        <button
          type="button"
          onClick={onRemove}
          className="p-1.5 rounded-md hover:bg-rose-500/10 text-muted-foreground hover:text-rose-500 transition-colors"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      </div>
    </div>
  );
}
