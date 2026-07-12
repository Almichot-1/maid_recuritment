"use client";

import * as React from "react";
import { Check, ChevronDown, Loader2 } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import { STEP_STATUS_MAP, type StepStatus } from "./status-badge";

interface StatusDropdownProps {
  value: string;
  disabled?: boolean;
  isSaving?: boolean;
  onSelect: (value: string) => void;
  className?: string;
  options?: Record<string, { label: string; className: string }>;
}

export function StatusDropdown({
  value,
  disabled = false,
  isSaving = false,
  onSelect,
  className,
  options,
}: StatusDropdownProps) {
  const statusMap = options ?? STEP_STATUS_MAP;
  const current = statusMap[value as StepStatus] ?? { label: value || "—", className: "bg-slate-100 text-slate-800 border-slate-300 dark:bg-slate-500/10 dark:text-slate-300 dark:border-slate-500/20" };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        disabled={disabled || isSaving}
        className={cn(
          "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] font-bold tracking-wide transition-all",
          "hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-primary/30",
          "disabled:opacity-50 disabled:cursor-not-allowed",
          current.className,
          className,
        )}
      >
        {isSaving && <Loader2 className="h-3 w-3 animate-spin" />}
        <span>{current.label}</span>
        {!disabled && <ChevronDown className="h-3 w-3 opacity-60" />}
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align="start"
        className="min-w-[160px] p-1.5 rounded-xl shadow-lg border border-border/60"
      >
        {Object.entries(statusMap).map(([key, info]) => (
          <DropdownMenuItem
            key={key}
            onSelect={() => onSelect(key)}
            className="flex items-center justify-between rounded-lg px-2.5 py-1.5 cursor-pointer"
          >
            <span
              className={cn(
                "inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-bold",
                info.className,
              )}
            >
              {info.label}
            </span>
            {value === key && <Check className="h-3.5 w-3.5 text-primary ml-2" />}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
