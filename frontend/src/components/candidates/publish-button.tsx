"use client";

import * as React from "react";
import { ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { WorkspaceSummary } from "@/types";

interface PublishButtonProps {
  workspaces: WorkspaceSummary[];
  isPublishing: boolean;
  onPublish: (pairingId?: string) => void;
}

export function PublishButton({ workspaces, isPublishing, onPublish }: PublishButtonProps) {
  const partnerName = (ws: WorkspaceSummary) =>
    ws.partner_agency?.company_name || ws.partner_agency?.full_name || ws.partner_agency?.email || "Partner";

  const needsSetup = (ws: WorkspaceSummary) =>
    !ws.default_country || !ws.default_currency;

  const readyWorkspaces = workspaces.filter((ws) => !needsSetup(ws));

  if (readyWorkspaces.length === 0) {
    return (
      <Button disabled title="Set up partner defaults on the Partners page first" className="w-full">
        Publish
      </Button>
    );
  }

  if (readyWorkspaces.length === 1) {
    const ws = readyWorkspaces[0];
    return (
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button disabled={isPublishing} className="w-full gap-1">
            {isPublishing ? "Publishing..." : `Publish to ${partnerName(ws)}`}
            <ChevronDown className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={() => onPublish(ws.id)}>
            {partnerName(ws)}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    );
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button disabled={isPublishing} className="w-full gap-1">
          {isPublishing ? "Publishing..." : "Publish"}
          <ChevronDown className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {workspaces.map((ws) => (
          <DropdownMenuItem
            key={ws.id}
            disabled={needsSetup(ws)}
            onClick={() => !needsSetup(ws) && onPublish(ws.id)}
          >
            <div className="flex items-center justify-between w-full gap-4">
              <span>{partnerName(ws)}</span>
              <span className="text-xs text-muted-foreground">
                {needsSetup(ws)
                  ? "⚠ Setup needed"
                  : `${ws.default_salary || ""} ${ws.default_currency || ""}`}
              </span>
            </div>
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => onPublish()}>
          All Partners
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
