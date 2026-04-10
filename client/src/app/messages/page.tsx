"use client";

import { useSearchParams } from "next/navigation";
import MessageBrowser from "@/components/MessageBrowser";
import { 
  Terminal
} from "lucide-react";

export default function MessagesPage() {
  const searchParams = useSearchParams();
  const topicFilter = searchParams.get("topic") || "";
  const partitionFilter = searchParams.get("partition") || "";

  return (
    <div className="p-8 space-y-10 animate-in fade-in slide-in-from-bottom-4 duration-700">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
        <div>
          <h1 className="text-4xl font-bold tracking-tight text-white flex items-center gap-3">
             <Terminal className="w-10 h-10 text-blue-500" />
             Event Explorer
          </h1>
          <p className="text-slate-400 mt-2 text-sm font-medium">
            Real-time browse through all captured events across the pipeline.
          </p>
        </div>
      </div>

      {/* Main Browser */}
      <div className="bg-white/[0.01] border border-white/5 rounded-[2.5rem] p-8 shadow-2xl">
         <MessageBrowser 
           topic={topicFilter} 
           partition={partitionFilter} 
           limit={50} 
         />
      </div>
    </div>
  );
}
