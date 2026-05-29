"use client"

import * as React from "react"
import { Building2 } from "lucide-react"

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
  return companyName?.trim() || fullName?.trim() || "Partner"
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
        "flex min-w-0 items-center gap-2 rounded-lg border border-border bg-card px-2 py-1.5",
        compact ? "w-full max-w-full" : "max-w-[220px] lg:max-w-[260px]",
        className,
      )}
    >
      <Building2 className="h-4 w-4 shrink-0 text-muted-foreground" aria-hidden />
      <div className="min-w-0 flex-1">
        <Select value={activeWorkspace?.id} onValueChange={setActivePairingId}>
          <SelectTrigger className="h-9 border-0 bg-transparent px-1 text-left text-sm shadow-none focus:ring-0">
            <SelectValue placeholder="Partner" />
          </SelectTrigger>
          <SelectContent>
            {context.workspaces.map((workspace) => (
              <SelectItem key={workspace.id} value={workspace.id}>
                <span className="truncate">
                  {getWorkspaceLabel(
                    workspace.partner_agency.company_name,
                    workspace.partner_agency.full_name,
                  )}
                </span>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <Badge variant="outline" className="shrink-0 text-[10px]">
        {context.workspaces.length}
      </Badge>
    </div>
  )
}
