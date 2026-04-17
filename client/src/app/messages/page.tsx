"use client";

import { useSearchParams } from "next/navigation";
import MessageBrowser from "@/components/MessageBrowser";
import { useApp } from "@/lib/AppContext";
import { Terminal } from "lucide-react";

export default function MessagesPage() {
  const { t } = useApp();
  const searchParams = useSearchParams();
  const topicFilter = searchParams.get("topic") || "";
  const partitionFilter = searchParams.get("partition") || "";

  return (
    <div className="space-y-4 animate-in fade-in slide-in-from-bottom-2 duration-500 pb-10">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="space-y-0.5">
          <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-primary/10 border border-primary/10 text-primary overline-label overline-label-solid">
            <Terminal className="w-3 h-3" /> Raw Stream
          </div>
          <h1 className="text-xl md:text-2xl font-black text-foreground">{t("explorer")}</h1>
          <p className="text-muted-foreground text-base-compact font-medium">Real-time browse through captured events across the pipeline.</p>
        </div>
      </div>

      <div className="bg-card rounded-xl border border-border overflow-hidden shadow-sm">
         <MessageBrowser 
           topic={topicFilter} 
           partition={partitionFilter} 
           limit={50} 
         />
      </div>
    </div>
  );
}
