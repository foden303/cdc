"use client"

import { useThemeStore } from "@/stores/theme"
import { Toaster as Sonner, type ToasterProps } from "sonner"
import { CircleCheckIcon, InfoIcon, TriangleAlertIcon, OctagonXIcon, Loader2Icon } from "lucide-react"

const Toaster = ({ ...props }: ToasterProps) => {
  const theme = useThemeStore((s) => s.theme)

  return (
    <Sonner
      theme={theme as ToasterProps["theme"]}
      className="toaster group"
      expand
      closeButton
      richColors
      visibleToasts={5}
      duration={4_000}
      icons={{
        success: (
          <CircleCheckIcon className="size-4" />
        ),
        info: (
          <InfoIcon className="size-4" />
        ),
        warning: (
          <TriangleAlertIcon className="size-4" />
        ),
        error: (
          <OctagonXIcon className="size-4" />
        ),
        loading: (
          <Loader2Icon className="size-4 animate-spin" />
        ),
      }}
      style={
        {
          "--normal-bg": "var(--popover)",
          "--normal-text": "var(--popover-foreground)",
          "--normal-border": "var(--border)",
          "--border-radius": "var(--radius)",
        } as React.CSSProperties
      }
      toastOptions={{
        classNames: {
          toast:
            "cn-toast border shadow-2xl backdrop-blur-md",
          title: "text-sm font-semibold tracking-tight",
          description: "text-xs leading-relaxed opacity-90",
          actionButton:
            "rounded-md bg-primary px-3 py-1.5 text-xs font-semibold text-primary-foreground",
          cancelButton:
            "rounded-md bg-muted px-3 py-1.5 text-xs font-semibold text-muted-foreground",
          success:
            "border-emerald-500/35 bg-emerald-950/90 text-emerald-50 shadow-emerald-950/25 [&_[data-icon]]:text-emerald-300",
          error:
            "border-rose-500/35 bg-rose-950/90 text-rose-50 shadow-rose-950/25 [&_[data-icon]]:text-rose-300",
          warning:
            "border-amber-500/35 bg-amber-950/90 text-amber-50 shadow-amber-950/25 [&_[data-icon]]:text-amber-300",
          info:
            "border-sky-500/35 bg-sky-950/90 text-sky-50 shadow-sky-950/25 [&_[data-icon]]:text-sky-300",
          loading:
            "border-violet-500/35 bg-violet-950/90 text-violet-50 shadow-violet-950/25 [&_[data-icon]]:text-violet-300",
        },
      }}
      {...props}
    />
  )
}

export { Toaster }
