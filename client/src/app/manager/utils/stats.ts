import type { ComponentStats } from "@/lib/grpc";

export function statusFromStat(s: ComponentStats | undefined): "running" | "error" | "starting" {
  if (!s) return "starting";
  return s.failure_count > 0 ? "error" : "running";
}
