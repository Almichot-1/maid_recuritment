"use client";

import * as React from "react";
import { Loader2, Download, RefreshCw } from "lucide-react";
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
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useBatchRegenerateCV, useBulkDownloadCVZip } from "@/hooks/use-candidates";
import type { WorkspaceSummary } from "@/types";

type Action = "regenerate" | "download";

interface BulkCvActionsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  candidateIds: string[];
  workspaces: WorkspaceSummary[];
  defaultAction?: Action;
}

export function BulkCvActionsDialog({
  open,
  onOpenChange,
  candidateIds,
  workspaces,
  defaultAction,
}: BulkCvActionsDialogProps) {
  const [action, setAction] = React.useState<Action>("regenerate");
  const [selectedPartner, setSelectedPartner] = React.useState<string>("");
  const [filenamePattern, setFilenamePattern] = React.useState("{name}");
  const batchRegen = useBatchRegenerateCV();
  const bulkZip = useBulkDownloadCVZip();
  const isPending = batchRegen.isPending || bulkZip.isPending;

  const handleConfirm = async () => {
    if (action === "regenerate") {
      await batchRegen.mutateAsync({
        candidateIds,
        pairingId: selectedPartner || undefined,
      });
    } else {
      const blob = await bulkZip.mutateAsync({
        candidateIds,
        pairingId: selectedPartner || undefined,
        filenamePattern: filenamePattern || "{name}",
      });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `candidates-${Date.now()}.zip`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
    }
    onOpenChange(false);
    setSelectedPartner("");
  };

  React.useEffect(() => {
    if (open) {
      setAction(defaultAction || "regenerate");
      setSelectedPartner("");
      setFilenamePattern("{name}");
    }
  }, [open, defaultAction]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {action === "regenerate"
              ? `Regenerate CVs for ${candidateIds.length} candidates`
              : `Download CVs for ${candidateIds.length} candidates`}
          </DialogTitle>
          <DialogDescription>
            {action === "regenerate"
              ? "Regenerate CVs using the latest candidate data and partner defaults."
              : "Download CVs as a ZIP archive with per-pairing overrides applied."}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="flex gap-2">
            <Button
              variant={action === "regenerate" ? "default" : "outline"}
              size="sm"
              onClick={() => setAction("regenerate")}
              className="flex-1"
            >
              <RefreshCw className="mr-2 h-4 w-4" />
              Regenerate
            </Button>
            <Button
              variant={action === "download" ? "default" : "outline"}
              size="sm"
              onClick={() => setAction("download")}
              className="flex-1"
            >
              <Download className="mr-2 h-4 w-4" />
              Download ZIP
            </Button>
          </div>

          <div className="space-y-2">
            <Label htmlFor="partner">Partner (optional)</Label>
            <Select value={selectedPartner} onValueChange={setSelectedPartner}>
              <SelectTrigger id="partner">
                <SelectValue placeholder="All partners" />
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

          {action === "download" && (
            <div className="space-y-2">
              <Label htmlFor="filename-pattern">Filename pattern</Label>
              <Input
                id="filename-pattern"
                value={filenamePattern}
                onChange={(e) => setFilenamePattern(e.target.value)}
                placeholder="{name}"
              />
              <p className="text-xs text-muted-foreground">
                Use {"{name}"}, {"{age}"}, {"{status}"}, {"{nationality}"}, {"{partner}"}
              </p>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isPending}>
            Cancel
          </Button>
          <Button onClick={handleConfirm} disabled={isPending}>
            {isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {action === "regenerate" ? "Regenerating..." : "Downloading..."}
              </>
            ) : (
              action === "regenerate" ? "Regenerate" : "Download ZIP"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
