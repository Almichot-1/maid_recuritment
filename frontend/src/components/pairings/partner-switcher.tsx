"use client"

import * as React from "react"
import { Building2, Link2 } from "lucide-react"

import { usePairingContext } from "@/hooks/use-pairings"
import { useCurrentUser } from "@/hooks/use-auth"
import { Badge } from "@/components/ui/badge"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { cn } from "@/lib/utils"

interface PartnerSwitcherProps {
  compact?: boolean
  className?: string
}

function getWorkspaceLabel(companyName?: string, fullName?: string) {
  return companyName?.trim() || fullName?.trim() || "Partner workspace"
}

export function PartnerSwitcher({ compact = false, className }: PartnerSwitcherProps) {
  const { user } = useCurrentUser()
  const { context, activePairingId, setActivePairingId, hasActivePairs, isLoading } = usePairingContext()

  if (!user || isLoading || !context || !hasActivePairs) {
    return null
  }

  const activeWorkspace =
    context.workspaces.find((workspace) => workspace.id === activePairingId) || context.workspaces[0] || null

  return (
    <div
      className={cn(
        "flex items-center gap-3 rounded-2xl border border-border/70 bg-card/80 px-3 py-2 shadow-sm backdrop-blur",
        compact ? "w-full" : "min-w-[250px]",
        className
      )}
    >
      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-primary/10 text-primary">
        <Link2 className="h-4 w-4" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="mb-1 flex items-center gap-2">
          <p className="text-[11px] font-semibold uppercase tracking-[0.24em] text-muted-foreground">
            Partner Workspace
          </p>
          <Badge variant="outline" className="rounded-full px-2 py-0 text-[10px] uppercase tracking-[0.18em]">
            {context.workspaces.length}
          </Badge>
        </div>
        <Select value={activeWorkspace?.id} onValueChange={setActivePairingId}>
          <SelectTrigger className="h-auto min-h-[44px] border-0 bg-transparent px-0 py-0 text-left shadow-none focus:ring-0">
            <SelectValue placeholder="Select a partner workspace" />
          </SelectTrigger>
          <SelectContent>
            {context.workspaces.map((workspace) => (
              <SelectItem key={workspace.id} value={workspace.id}>
                <div className="flex min-w-0 items-center gap-2">
                  <Building2 className="h-4 w-4 shrink-0 text-muted-foreground" />
                  <span className="truncate">
                    {getWorkspaceLabel(workspace.partner_agency.company_name, workspace.partner_agency.full_name)}
                  </span>
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {activeWorkspace ? (
          <p className="mt-1 truncate text-xs text-muted-foreground">
            {activeWorkspace.partner_agency.email}
          </p>
        ) : null}
      </div>
    </div>
  )
}
