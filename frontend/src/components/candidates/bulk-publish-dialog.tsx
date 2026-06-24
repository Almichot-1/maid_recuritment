"use client";

import * as React from "react";
import { Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { useBulkPublish } from "@/hooks/use-candidates";
import type { WorkspaceSummary } from "@/types";

interface BulkPublishDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  candidateIds: string[];
  workspaces: WorkspaceSummary[];
}

export function BulkPublishDialog({
  open,
  onOpenChange,
  candidateIds,
  workspaces,
}: BulkPublishDialogProps) {
  const [selectedIds, setSelectedIds] = React.useState<string[]>([]);
  const bulkPublish = useBulkPublish();

  React.useEffect(() => {
    if (open) {
      setSelectedIds(workspaces.filter((ws) => ws.default_country && ws.default_currency).map((ws) => ws.id));
    }
  }, [open, workspaces]);

  const togglePartner = (id: string) => {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((p) => p !== id) : [...prev, id]
    );
  };

  const handlePublish = async () => {
    await bulkPublish.mutateAsync({
      candidate_ids: candidateIds,
      pairing_ids: selectedIds.length > 0 ? selectedIds : undefined,
    });
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Publish {candidateIds.length} candidates</DialogTitle>
          <DialogDescription>
            Select partners to publish these candidates to.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {workspaces.length === 0 && (
            <p className="text-sm text-muted-foreground">No partners available.</p>
          )}
          {workspaces.map((ws) => {
            const needsSetup = !ws.default_country || !ws.default_currency;
            return (
              <div key={ws.id} className="flex items-center gap-3">
                <Checkbox
                  id={ws.id}
                  checked={selectedIds.includes(ws.id)}
                  onCheckedChange={() => togglePartner(ws.id)}
                  disabled={needsSetup}
                />
                <Label
                  htmlFor={ws.id}
                  className={`flex items-center justify-between w-full cursor-pointer ${needsSetup ? "text-muted-foreground" : ""}`}
                >
                  <span>{ws.partner_agency?.company_name || ws.partner_agency?.full_name || ws.partner_agency?.email || "Partner"}</span>
                  <span className="text-xs">
                    {needsSetup
                      ? "⚠ Setup needed"
                      : `${ws.default_salary || ""} ${ws.default_currency || ""}`}
                  </span>
                </Label>
              </div>
            );
          })}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={bulkPublish.isPending}>
            Cancel
          </Button>
          <Button onClick={handlePublish} disabled={bulkPublish.isPending || selectedIds.length === 0}>
            {bulkPublish.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Publishing...
              </>
            ) : (
              `Publish to ${selectedIds.length} partner${selectedIds.length !== 1 ? "s" : ""}`
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
