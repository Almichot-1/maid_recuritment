"use client";

import * as React from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Trash2, Download, Printer, UserCheck, Bell, X, Stethoscope, FileText, Building2, Ticket, MapPin } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import { STEP_STATUS_MAP, TICKET_STATUS_MAP, ARRIVAL_STATUS_MAP } from "./status-badge";

interface BulkToolbarProps {
  selectedCount: number;
  onClear: () => void;
  onDelete: () => void;
  onExport: () => void;
  onPrint: () => void;
  onAssignEmployee: () => void;
  onSendNotification: () => void;
  onChangeMedical: (status: string) => void;
  onChangeCoC: (status: string) => void;
  onChangeVisa: (status: string) => void;
  onChangeTicket: (status: string) => void;
  onChangeArrival: (status: string) => void;
}

export function BulkToolbar({
  selectedCount,
  onClear,
  onDelete,
  onExport,
  onPrint,
  onAssignEmployee,
  onSendNotification,
  onChangeMedical,
  onChangeCoC,
  onChangeVisa,
  onChangeTicket,
  onChangeArrival,
}: BulkToolbarProps) {
  return (
    <AnimatePresence>
      {selectedCount > 0 && (
        <motion.div
          initial={{ opacity: 0, y: 40, scale: 0.96 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0, y: 30, scale: 0.94 }}
          transition={{ type: "spring", damping: 28, stiffness: 320 }}
          className="fixed bottom-6 left-1/2 z-50 -translate-x-1/2"
        >
          <div className="flex items-center gap-3 rounded-2xl border border-border/70 bg-background/95 px-5 py-3 shadow-2xl backdrop-blur-xl ring-1 ring-black/5">
            {/* Count badge */}
            <div className="flex items-center gap-2 shrink-0">
              <span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary text-[11px] font-bold text-primary-foreground">
                {selectedCount}
              </span>
              <span className="text-sm font-semibold whitespace-nowrap">selected</span>
            </div>

            <div className="h-5 w-px bg-border mx-1 shrink-0" />

            {/* Actions */}
            <div className="flex items-center gap-1 flex-wrap">
              <Button size="sm" variant="ghost" className="h-8 rounded-xl text-xs gap-1.5" onClick={onAssignEmployee}>
                <UserCheck className="h-3.5 w-3.5" />
                Assign
              </Button>

              {/* Medical bulk */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button size="sm" variant="ghost" className="h-8 rounded-xl text-xs gap-1.5">
                    <Stethoscope className="h-3.5 w-3.5" />
                    Medical
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="min-w-[160px] p-1.5 rounded-xl shadow-lg">
                  <DropdownMenuLabel className="text-[10px] uppercase tracking-widest text-muted-foreground px-2">Set Medical</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {Object.entries(STEP_STATUS_MAP).map(([key, info]) => (
                    <DropdownMenuItem key={key} onSelect={() => onChangeMedical(key)} className="rounded-lg cursor-pointer">
                      <span className={cn("inline-flex rounded-full border px-2 py-0.5 text-[10px] font-semibold", info.className)}>{info.label}</span>
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

              {/* CoC bulk */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button size="sm" variant="ghost" className="h-8 rounded-xl text-xs gap-1.5">
                    <FileText className="h-3.5 w-3.5" />
                    CoC
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="min-w-[160px] p-1.5 rounded-xl shadow-lg">
                  <DropdownMenuLabel className="text-[10px] uppercase tracking-widest text-muted-foreground px-2">Set CoC</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {Object.entries(STEP_STATUS_MAP).map(([key, info]) => (
                    <DropdownMenuItem key={key} onSelect={() => onChangeCoC(key)} className="rounded-lg cursor-pointer">
                      <span className={cn("inline-flex rounded-full border px-2 py-0.5 text-[10px] font-semibold", info.className)}>{info.label}</span>
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

              {/* Visa bulk */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button size="sm" variant="ghost" className="h-8 rounded-xl text-xs gap-1.5">
                    <Building2 className="h-3.5 w-3.5" />
                    Visa
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="min-w-[160px] p-1.5 rounded-xl shadow-lg">
                  <DropdownMenuLabel className="text-[10px] uppercase tracking-widest text-muted-foreground px-2">Set Visa</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {Object.entries(STEP_STATUS_MAP).map(([key, info]) => (
                    <DropdownMenuItem key={key} onSelect={() => onChangeVisa(key)} className="rounded-lg cursor-pointer">
                      <span className={cn("inline-flex rounded-full border px-2 py-0.5 text-[10px] font-semibold", info.className)}>{info.label}</span>
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

              {/* Ticket bulk */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button size="sm" variant="ghost" className="h-8 rounded-xl text-xs gap-1.5">
                    <Ticket className="h-3.5 w-3.5" />
                    Ticket
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="min-w-[160px] p-1.5 rounded-xl shadow-lg">
                  <DropdownMenuLabel className="text-[10px] uppercase tracking-widest text-muted-foreground px-2">Set Ticket</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {Object.entries(TICKET_STATUS_MAP).map(([key, info]) => (
                    <DropdownMenuItem key={key} onSelect={() => onChangeTicket(key)} className="rounded-lg cursor-pointer">
                      <span className={cn("inline-flex rounded-full border px-2 py-0.5 text-[10px] font-semibold", info.className)}>{info.label}</span>
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

              {/* Arrival bulk */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button size="sm" variant="ghost" className="h-8 rounded-xl text-xs gap-1.5">
                    <MapPin className="h-3.5 w-3.5" />
                    Arrival
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="min-w-[160px] p-1.5 rounded-xl shadow-lg">
                  <DropdownMenuLabel className="text-[10px] uppercase tracking-widest text-muted-foreground px-2">Set Arrival</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {Object.entries(ARRIVAL_STATUS_MAP).map(([key, info]) => (
                    <DropdownMenuItem key={key} onSelect={() => onChangeArrival(key)} className="rounded-lg cursor-pointer">
                      <span className={cn("inline-flex rounded-full border px-2 py-0.5 text-[10px] font-semibold", info.className)}>{info.label}</span>
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

              <Button size="sm" variant="ghost" className="h-8 rounded-xl text-xs gap-1.5" onClick={onPrint}>
                <Printer className="h-3.5 w-3.5" />
                Print
              </Button>

              <Button size="sm" variant="ghost" className="h-8 rounded-xl text-xs gap-1.5" onClick={onExport}>
                <Download className="h-3.5 w-3.5" />
                Export
              </Button>

              <Button size="sm" variant="ghost" className="h-8 rounded-xl text-xs gap-1.5" onClick={onSendNotification}>
                <Bell className="h-3.5 w-3.5" />
                Notify
              </Button>

              <div className="h-5 w-px bg-border mx-1 shrink-0" />

              <Button
                size="sm"
                variant="ghost"
                className="h-8 rounded-xl text-xs gap-1.5 text-destructive hover:text-destructive hover:bg-destructive/10"
                onClick={onDelete}
              >
                <Trash2 className="h-3.5 w-3.5" />
                Delete
              </Button>
            </div>

            {/* Dismiss */}
            <button
              onClick={onClear}
              className="ml-1 shrink-0 rounded-full p-1 text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
