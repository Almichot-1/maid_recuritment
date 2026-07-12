"use client";

import * as React from "react";
import { Search, X, ChevronDown, SlidersHorizontal, RotateCcw, Filter } from "lucide-react";
import { Input } from "@/components/ui/input";
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
import { STEP_STATUS_MAP, APPLICANT_STATUS_MAP, ARRIVAL_STATUS_MAP } from "./status-badge";

export interface TrackingFilters {
  search: string;
  medical: string;
  coc: string;
  visa: string;
  ticket: string;
  arrival: string;
  status: string;
  foreignAgent: string;
  sort: string;
}

export const DEFAULT_FILTERS: TrackingFilters = {
  search: "",
  medical: "",
  coc: "",
  visa: "",
  ticket: "",
  arrival: "",
  status: "",
  foreignAgent: "",
  sort: "name_asc",
};

const SORT_OPTIONS = [
  { value: "name_asc",     label: "Name A → Z" },
  { value: "name_desc",    label: "Name Z → A" },
  { value: "updated_desc", label: "Recently Updated" },
  { value: "updated_asc",  label: "Oldest Updated" },
];

function FilterPill({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: string;
  options: Record<string, { label: string; className: string }>;
  onChange: (v: string) => void;
}) {
  const isActive = !!value;
  const current = options[value];

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          className={cn(
            "inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-medium transition-all focus:outline-none focus:ring-2 focus:ring-primary/30",
            isActive
              ? "border-primary/40 bg-primary/5 text-primary"
              : "border-border bg-background text-muted-foreground hover:bg-accent",
          )}
        >
          {isActive && current ? (
            <span className={cn("inline-flex rounded-full border px-1.5 py-0.5 text-[10px] font-semibold", current.className)}>
              {current.label}
            </span>
          ) : (
            label
          )}
          <ChevronDown className="h-3 w-3 opacity-50" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="min-w-[160px] p-1.5 rounded-xl shadow-lg border border-border/60">
        <DropdownMenuLabel className="text-[10px] uppercase tracking-widest text-muted-foreground px-2 py-1">
          {label}
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={() => onChange("")} className="text-xs rounded-lg cursor-pointer font-medium">
          All statuses
        </DropdownMenuItem>
        {Object.entries(options).map(([key, info]) => (
          <DropdownMenuItem
            key={key}
            onSelect={() => onChange(key)}
            className="flex items-center justify-between rounded-lg px-2 py-1.5 cursor-pointer"
          >
            <span className={cn("inline-flex rounded-full border px-2 py-0.5 text-[10px] font-semibold", info.className)}>
              {info.label}
            </span>
            {value === key && <span className="h-1.5 w-1.5 rounded-full bg-primary" />}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

interface SearchToolbarProps {
  filters: TrackingFilters;
  onChange: (filters: TrackingFilters) => void;
  totalCount: number;
  filteredCount: number;
  foreignAgentOptions?: string[];
}

export function SearchToolbar({ filters, onChange, totalCount, filteredCount, foreignAgentOptions }: SearchToolbarProps) {
  const [showMoreFilters, setShowMoreFilters] = React.useState(false);
  const hasFilters = filters.search || filters.medical || filters.coc || filters.visa || filters.ticket || filters.arrival || filters.status || filters.foreignAgent;
  const currentSort = SORT_OPTIONS.find((o) => o.value === filters.sort);
  
  const activeSecondaryFilterCount = [filters.medical, filters.coc, filters.visa, filters.ticket, filters.arrival, filters.foreignAgent].filter(Boolean).length;

  return (
    <div className="sticky top-0 z-30 bg-background/80 backdrop-blur-md border-b border-border/60 shadow-sm">
      <div className="flex flex-col gap-3 px-6 py-4">
        <div className="flex flex-wrap items-center gap-3">
          {/* Search */}
          <div className="relative flex-1 min-w-[280px]">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
            <Input
              value={filters.search}
              onChange={(e) => onChange({ ...filters, search: e.target.value })}
              placeholder="Search applicant name, passport number…"
              className="pl-9 pr-9 h-9 rounded-xl border-border/60 bg-background text-sm shadow-sm"
            />
            {filters.search && (
              <button
                onClick={() => onChange({ ...filters, search: "" })}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground p-1 rounded-md hover:bg-muted"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            )}
          </div>

          {/* Primary Status Filter */}
          <FilterPill label="Status" value={filters.status} options={APPLICANT_STATUS_MAP} onChange={(v) => onChange({ ...filters, status: v })} />

          {/* More Filters Toggle */}
          <button
            onClick={() => setShowMoreFilters(!showMoreFilters)}
            className={cn(
              "inline-flex items-center gap-2 rounded-xl border px-3 py-1.5 text-xs font-medium transition-all focus:outline-none focus:ring-2 focus:ring-primary/30 h-9",
              showMoreFilters || activeSecondaryFilterCount > 0
                ? "border-primary/40 bg-primary/5 text-primary"
                : "border-border bg-background text-muted-foreground hover:bg-accent"
            )}
          >
            <Filter className="h-3.5 w-3.5" />
            More Filters
            {activeSecondaryFilterCount > 0 && (
              <span className="ml-1 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-bold text-primary-foreground">
                {activeSecondaryFilterCount}
              </span>
            )}
          </button>

          {/* Sort */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button className="ml-auto inline-flex items-center gap-1.5 rounded-xl border border-border bg-background px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-accent transition-all focus:outline-none h-9 shadow-sm">
                <SlidersHorizontal className="h-3.5 w-3.5" />
                {currentSort?.label ?? "Sort"}
                <ChevronDown className="h-3 w-3 opacity-50" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="min-w-[180px] p-1.5 rounded-xl shadow-lg border border-border/60">
              <DropdownMenuLabel className="text-[10px] uppercase tracking-widest text-muted-foreground px-2 py-1">
                Sort by
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              {SORT_OPTIONS.map((opt) => (
                <DropdownMenuItem
                  key={opt.value}
                  onSelect={() => onChange({ ...filters, sort: opt.value })}
                  className="flex items-center justify-between rounded-lg px-2.5 py-1.5 text-xs cursor-pointer"
                >
                  {opt.label}
                  {filters.sort === opt.value && <span className="h-1.5 w-1.5 rounded-full bg-primary" />}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>

          {/* Reset */}
          {hasFilters && (
            <Button
              size="sm"
              variant="ghost"
              className="rounded-xl h-9 px-3 text-xs text-muted-foreground"
              onClick={() => onChange({ ...DEFAULT_FILTERS, sort: filters.sort })}
            >
              <RotateCcw className="h-3 w-3 mr-1" />
              Reset
            </Button>
          )}
        </div>

        {/* Expanded Filters */}
        {showMoreFilters && (
          <div className="flex flex-wrap items-center gap-2 pt-2 border-t border-border/40 mt-1">
            <span className="text-xs font-semibold text-muted-foreground mr-2">Stages:</span>
            <FilterPill label="Medical" value={filters.medical} options={STEP_STATUS_MAP} onChange={(v) => onChange({ ...filters, medical: v })} />
            <FilterPill label="CoC"     value={filters.coc}     options={STEP_STATUS_MAP}     onChange={(v) => onChange({ ...filters, coc: v })} />
            <FilterPill label="Visa"    value={filters.visa}    options={STEP_STATUS_MAP}    onChange={(v) => onChange({ ...filters, visa: v })} />
            <FilterPill label="Ticket"  value={filters.ticket}  options={STEP_STATUS_MAP}  onChange={(v) => onChange({ ...filters, ticket: v })} />
            <FilterPill label="Arrival" value={filters.arrival} options={ARRIVAL_STATUS_MAP} onChange={(v) => onChange({ ...filters, arrival: v })} />
            
            {foreignAgentOptions && foreignAgentOptions.length > 0 && (
              <>
                <span className="text-xs font-semibold text-muted-foreground ml-3 mr-2">Partner:</span>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button
                      className={cn(
                        "inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-medium transition-all focus:outline-none focus:ring-2 focus:ring-primary/30",
                        filters.foreignAgent
                          ? "border-primary/40 bg-primary/5 text-primary"
                          : "border-border bg-background text-muted-foreground hover:bg-accent",
                      )}
                    >
                      {filters.foreignAgent ? (
                        <span className="inline-flex rounded-full border border-primary/30 bg-primary/10 px-1.5 py-0.5 text-[10px] font-bold text-primary">
                          {filters.foreignAgent}
                        </span>
                      ) : (
                        "Foreign Agent"
                      )}
                      <ChevronDown className="h-3 w-3 opacity-50" />
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="start" className="max-h-[260px] overflow-y-auto min-w-[180px] p-1.5 rounded-xl shadow-lg border border-border/60">
                    <DropdownMenuLabel className="text-[10px] uppercase tracking-widest text-muted-foreground px-2 py-1">
                      Foreign Agent
                    </DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onSelect={() => onChange({ ...filters, foreignAgent: "" })} className="text-xs rounded-lg cursor-pointer font-medium">
                      All agents
                    </DropdownMenuItem>
                    {foreignAgentOptions.map((name) => (
                      <DropdownMenuItem
                        key={name}
                        onSelect={() => onChange({ ...filters, foreignAgent: name })}
                        className="flex items-center justify-between rounded-lg px-2 py-1.5 text-xs cursor-pointer"
                      >
                        {name}
                        {filters.foreignAgent === name && <span className="h-1.5 w-1.5 rounded-full bg-primary" />}
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
              </>
            )}
          </div>
        )}
      </div>

      {/* Results count */}
      <div className="px-6 pb-3 text-[11px] font-medium text-muted-foreground border-b border-border/30 bg-muted/20">
        {hasFilters
          ? `Showing ${filteredCount} of ${totalCount} applicants`
          : `${totalCount} applicant${totalCount !== 1 ? "s" : ""}`}
      </div>
    </div>
  );
}
