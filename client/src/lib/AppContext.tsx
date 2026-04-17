"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { translations, Locale, TranslationKey } from "./i18n";

interface AppContextType {
  theme: "dark" | "light";
  lang: Locale;
  toggleTheme: () => void;
  setLang: (lang: Locale) => void;
  t: (key: TranslationKey) => string;
}

const AppContext = createContext<AppContextType | undefined>(undefined);

export function AppProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<"dark" | "light">("dark");
  const [lang, setLangState] = useState<Locale>("en");
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    // Load persisted settings
    const savedTheme = localStorage.getItem("cdc-theme") as "dark" | "light";
    const savedLang = localStorage.getItem("cdc-lang") as Locale;
    
    if (savedTheme) setTheme(savedTheme);
    if (savedLang) setLangState(savedLang);
    setMounted(true);
  }, []);

  useEffect(() => {
    if (mounted) {
      localStorage.setItem("cdc-theme", theme);
      document.documentElement.setAttribute("data-theme", theme);
      if (theme === "light") {
        document.documentElement.classList.add("light");
      } else {
        document.documentElement.classList.remove("light");
      }
    }
  }, [theme, mounted]);

  useEffect(() => {
    if (mounted) {
      localStorage.setItem("cdc-lang", lang);
      document.documentElement.setAttribute("lang", lang);
    }
  }, [lang, mounted]);

  const toggleTheme = () => setTheme(prev => (prev === "dark" ? "light" : "dark"));
  
  const setLang = (newLang: Locale) => setLangState(newLang);

  const t = (key: TranslationKey): string => {
    return translations[lang][key] || translations["en"][key] || key;
  };

  return (
    <AppContext.Provider value={{ theme, lang, toggleTheme, setLang, t }}>
      {mounted ? children : <div className="invisible">{children}</div>}
    </AppContext.Provider>
  );
}

export function useApp() {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error("useApp must be used within an AppProvider");
  }
  return context;
}
