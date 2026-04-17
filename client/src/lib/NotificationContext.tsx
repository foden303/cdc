"use client";

import React, { createContext, useCallback, useContext, useMemo, useState } from "react";

export type NotificationVariant = "success" | "error";

export type NotificationItem = {
  id: string;
  message: string;
  variant: NotificationVariant;
  createdAt: number;
};

type NotificationContextValue = {
  history: NotificationItem[];
  notify: (message: string, variant: NotificationVariant) => void;
  clearHistory: () => void;
};

const NotificationContext = createContext<NotificationContextValue | undefined>(undefined);

const MAX_HISTORY = 50;

export function NotificationProvider({ children }: { children: React.ReactNode }) {
  const [activeToasts, setActiveToasts] = useState<NotificationItem[]>([]);
  const [history, setHistory] = useState<NotificationItem[]>([]);

  const notify = useCallback((message: string, variant: NotificationVariant) => {
    const id = `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
    const createdAt = Date.now();
    const item: NotificationItem = { id, message, variant, createdAt };
    setHistory((h) => [item, ...h].slice(0, MAX_HISTORY));
    setActiveToasts((t) => [...t, item]);
    window.setTimeout(() => {
      setActiveToasts((t) => t.filter((x) => x.id !== id));
    }, 4200);
  }, []);

  const clearHistory = useCallback(() => setHistory([]), []);

  const value = useMemo(
    () => ({ history, notify, clearHistory }),
    [history, notify, clearHistory]
  );

  return (
    <NotificationContext.Provider value={value}>
      {children}
      <ToastStack toasts={activeToasts} onDismiss={(id) => setActiveToasts((t) => t.filter((x) => x.id !== id))} />
    </NotificationContext.Provider>
  );
}

function ToastStack({
  toasts,
  onDismiss,
}: {
  toasts: NotificationItem[];
  onDismiss: (id: string) => void;
}) {
  if (toasts.length === 0) return null;
  return (
    <div className="fixed top-14 right-4 z-[100] flex max-w-[min(100vw-2rem,20rem)] flex-col gap-2 pointer-events-none">
      {toasts.map((n) => (
        <div
          key={n.id}
          className={`pointer-events-auto animate-in slide-in-from-right-2 fade-in duration-200 flex items-start gap-2 rounded-lg border px-3 py-2 text-base-compact font-bold shadow-lg backdrop-blur-sm ${
            n.variant === "success"
              ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
              : "border-rose-500/30 bg-rose-500/10 text-rose-600 dark:text-rose-400"
          }`}
          role="status"
        >
          <span className="flex-1 leading-snug">{n.message}</span>
          <button
            type="button"
            onClick={() => onDismiss(n.id)}
            className="shrink-0 opacity-60 hover:opacity-100 text-[10px] uppercase"
            aria-label="Dismiss"
          >
            ×
          </button>
        </div>
      ))}
    </div>
  );
}

export function useNotifications() {
  const ctx = useContext(NotificationContext);
  if (!ctx) throw new Error("useNotifications must be used within NotificationProvider");
  return ctx;
}
