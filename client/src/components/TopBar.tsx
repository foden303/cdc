"use client";

import { useEffect, useRef, useState } from "react";
import { useApp } from "@/lib/AppContext";
import { useNotifications } from "@/lib/NotificationContext";
import { 
  Bell, 
  Search, 
  Languages,
  Sun,
  Moon
} from "lucide-react";

export default function TopBar() {
  const { theme, toggleTheme, lang, setLang, t } = useApp();
  const { history, clearHistory } = useNotifications();
  const [langMenuOpen, setLangMenuOpen] = useState(false);
  const [notifOpen, setNotifOpen] = useState(false);
  const langMenuRef = useRef<HTMLDivElement>(null);
  const notifRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!langMenuOpen) return;
    const onPointerDown = (e: PointerEvent) => {
      if (langMenuRef.current && !langMenuRef.current.contains(e.target as Node)) {
        setLangMenuOpen(false);
      }
    };
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, [langMenuOpen]);

  useEffect(() => {
    if (!notifOpen) return;
    const onPointerDown = (e: PointerEvent) => {
      if (notifRef.current && !notifRef.current.contains(e.target as Node)) {
        setNotifOpen(false);
      }
    };
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, [notifOpen]);

  return (
    <header className="h-11 flex items-center justify-between px-4 border-b border-border bg-card/50 backdrop-blur-md sticky top-0 z-50">
      {/* Search Bar / Insight Area */}
      <div className="flex-1 max-w-sm group">
        <div className="relative flex items-center w-full">
          <Search className="absolute left-2.5 w-3 h-3 text-muted-foreground group-focus-within:text-primary transition-colors" />
          <input
            type="text"
            placeholder={t("explorer")}
            className="w-full bg-muted/30 border border-border rounded-lg py-1 pl-8 pr-3 text-base-compact font-medium text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:border-primary/40 focus:bg-muted/50 transition-all"
          />
        </div>
      </div>

      {/* Right Side Info Badges */}
      <div className="flex items-center gap-3 ml-4">
        {/* Action Icons */}
        <div className="flex items-center gap-1.5 border-border pl-3">
          {/* Theme Toggle */}
          <button 
            onClick={toggleTheme}
            className="flex items-center justify-center w-7 h-7 rounded-lg border border-border bg-muted/20 hover:bg-muted/50 transition-all text-muted-foreground hover:text-foreground"
          >
            {theme === 'dark' ? <Sun className="w-3.5 h-3.5" /> : <Moon className="w-3.5 h-3.5" />}
          </button>

          {/* Lang: click + hover bridge — pure hover breaks when cursor crosses mt-1 gap */}
          <div ref={langMenuRef} className="relative group/lang">
            <button
              type="button"
              aria-expanded={langMenuOpen}
              aria-haspopup="listbox"
              onClick={() => setLangMenuOpen((o) => !o)}
              className="flex items-center justify-center h-7 px-2 rounded-lg border border-border bg-muted/20 hover:bg-muted/50 transition-all text-muted-foreground hover:text-foreground gap-1.5 text-mini-compact font-black uppercase"
            >
              <Languages className="w-3.5 h-3.5" />
              {lang}
            </button>
            <div
              className={`absolute top-full right-0 z-[60] pt-1 w-28 transition-all ${
                langMenuOpen ? "opacity-100 visible" : "ui-fade-hidden group-hover/lang:opacity-100 group-hover/lang:visible"
              }`}
            >
              <div className="glass-panel p-1.5 space-y-0.5 shadow-xl">
                {(['en', 'vn', 'cn'] as const).map((l) => (
                  <button
                    key={l}
                    type="button"
                    onClick={() => {
                      setLang(l);
                      setLangMenuOpen(false);
                    }}
                    className={`w-full text-left px-2 py-1 rounded-md text-mini-compact font-black uppercase transition-colors ${lang === l ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-muted"}`}
                  >
                    {l === "en" ? "English" : l === "vn" ? "Tiếng Việt" : "中文"}
                  </button>
                ))}
              </div>
            </div>
          </div>
          
          <div ref={notifRef} className="relative">
            <button
              type="button"
              aria-expanded={notifOpen}
              onClick={() => setNotifOpen((o) => !o)}
              className="relative flex items-center justify-center w-7 h-7 rounded-lg border border-border bg-muted/20 hover:bg-muted/50 transition-all text-muted-foreground hover:text-foreground font-black text-mini-compact"
            >
              <Bell className="w-3.5 h-3.5" />
              {history.length > 0 ? (
                <span className="absolute -top-0.5 -right-0.5 min-w-[14px] h-3.5 px-0.5 rounded-full bg-primary text-[8px] font-black leading-none flex items-center justify-center text-primary-foreground">
                  {history.length > 9 ? "9+" : history.length}
                </span>
              ) : null}
            </button>
            {notifOpen && (
              <div className="absolute top-full right-0 mt-1 w-72 max-h-72 overflow-y-auto custom-scrollbar z-[70] glass-panel p-2 shadow-xl border border-border">
                <div className="flex items-center justify-between px-1 pb-2 border-b border-border/60">
                  <span className="overline-label overline-label-dim">{t("notificationsTitle")}</span>
                  {history.length > 0 ? (
                    <button
                      type="button"
                      onClick={() => clearHistory()}
                      className="text-mini-compact font-black uppercase text-muted-foreground hover:text-foreground"
                    >
                      {t("notificationsClear")}
                    </button>
                  ) : null}
                </div>
                {history.length === 0 ? (
                  <p className="text-mini-compact text-muted-foreground py-4 text-center">{t("notificationsEmpty")}</p>
                ) : (
                  <ul className="space-y-1 pt-1">
                    {history.map((n) => (
                      <li
                        key={n.id}
                        className={`rounded-md px-2 py-1.5 text-mini-compact font-medium ${
                          n.variant === "success" ? "bg-emerald-500/5 text-emerald-600 dark:text-emerald-400" : "bg-rose-500/5 text-rose-600 dark:text-rose-400"
                        }`}
                      >
                        <div className="leading-snug">{n.message}</div>
                        <div className="text-[8px] font-mono text-muted-foreground/80 mt-0.5 tabular-nums">
                          {new Date(n.createdAt).toLocaleString()}
                        </div>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
