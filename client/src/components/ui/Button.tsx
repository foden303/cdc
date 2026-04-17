"use client";

import * as React from "react";

import { cn } from "@/lib/cn";

type ButtonVariant = "primary" | "muted" | "danger" | "ghost";
type ButtonSize = "sm" | "md";

export type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  size?: ButtonSize;
};

export function Button({
  className,
  variant = "primary",
  size = "md",
  type = "button",
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-lg font-black uppercase text-base-compact shadow-sm transition-colors ui-disabled-default",
        size === "sm" ? "px-3 py-1.5" : "px-5 py-2",
        variant === "primary" && "bg-foreground text-background hover:bg-foreground/90",
        variant === "muted" && "bg-muted text-muted-foreground hover:bg-muted/80",
        variant === "danger" && "bg-rose-600 text-white hover:bg-rose-600/90",
        variant === "ghost" && "bg-transparent text-foreground hover:bg-muted",
        className
      )}
      {...props}
    />
  );
}

