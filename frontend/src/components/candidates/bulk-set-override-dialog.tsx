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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useBatchSetPairOverride } from "@/hooks/use-candidates";
import type { WorkspaceSummary } from "@/types";

interface BulkSetOverrideDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  candidateIds: string[];
  workspaces: WorkspaceSummary[];
}

export function BulkSetOverrideDialog({
  open,
  onOpenChange,
  candidateIds,
  workspaces,
}: BulkSetOverrideDialogProps) {
  const [partnerId, setPartnerId] = React.useState("");
  const [country, setCountry] = React.useState("");
  const [salary, setSalary] = React.useState("");
  const batchOverride = useBatchSetPairOverride();

  const handleConfirm = async () => {
    if (!partnerId) return;
    await batchOverride.mutateAsync({
      candidate_ids: candidateIds,
      pairing_id: partnerId,
      country_applied: country || undefined,
      salary_offered: salary || undefined,
    });
    onOpenChange(false);
    setPartnerId("");
    setCountry("");
    setSalary("");
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Set partner overrides for {candidateIds.length} candidates</DialogTitle>
          <DialogDescription>
            These values will override the default country and salary for these candidates with the selected partner.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="partner-select">Partner *</Label>
            <Select value={partnerId} onValueChange={setPartnerId}>
              <SelectTrigger id="partner-select">
                <SelectValue placeholder="Select a partner" />
              </SelectTrigger>
              <SelectContent>
                {workspaces.map((ws) => (
                  <SelectItem key={ws.id} value={ws.id}>
                    {ws.partner_agency?.company_name || ws.partner_agency?.full_name || ws.partner_agency?.email || "Partner"}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="override-country">Country (optional)</Label>
            <Input
              id="override-country"
              placeholder="e.g., Kuwait"
              value={country}
              onChange={(e) => setCountry(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="override-salary">Salary (optional)</Label>
            <Input
              id="override-salary"
              placeholder="e.g., 2500 KWD"
              value={salary}
              onChange={(e) => setSalary(e.target.value)}
            />
          </div>

          <p className="text-xs text-muted-foreground">
            These values override defaults for these candidates with this partner. Leave empty to keep existing values.
          </p>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={batchOverride.isPending}>
            Cancel
          </Button>
          <Button onClick={handleConfirm} disabled={batchOverride.isPending || !partnerId}>
            {batchOverride.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              "Confirm"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
