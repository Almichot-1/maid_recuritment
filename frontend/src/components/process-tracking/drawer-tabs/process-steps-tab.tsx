"use client";

import * as React from "react";
import { Stethoscope, FileText, Building2, Ticket, MapPin, UploadCloud, Loader2, Eye, Trash2, FileBadge2 } from "lucide-react";
import { StatusDropdown } from "../status-dropdown";
import { TICKET_STATUS_MAP, ARRIVAL_STATUS_MAP } from "../status-badge";
import { cn } from "@/lib/utils";
import type { StepItem } from "@/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useUploadProgressDocument, useDeleteProgressDocument, useSelectionProgress } from "@/hooks/use-selection-progress";
import { mapProgressToSteps } from "../process-tracking-page";

function useLiveSteps(selectionId: string, fallbackSteps: StepItem[]): StepItem[] {
  const { data: progress } = useSelectionProgress(selectionId);
  return React.useMemo(() => {
    if (!progress) return fallbackSteps;
    return mapProgressToSteps(progress);
  }, [progress, fallbackSteps]);
}

function StepDetailCard({
  selectionId,
  stepName,
  icon: Icon,
  steps: propSteps,
  canEdit,
  isSaving,
  onUpdate,
  accentClass,
}: {
  selectionId: string;
  stepName: string;
  icon: React.ElementType;
  steps: StepItem[];
  canEdit: boolean;
  isSaving: boolean;
  onUpdate: (stepName: string, status: string, extras?: { destination_country?: string; arrival_date?: string }) => void;
  accentClass: string;
}) {
  const steps = useLiveSteps(selectionId, propSteps);
  const step = steps.find((s) => s.step_name === stepName);
  const status = step?.step_status ?? "pending";
  
  // Ticket Extra Fields
  const [destCountry, setDestCountry] = React.useState("");
  const [arrivalDate, setArrivalDate] = React.useState("");

  // Document Upload / Delete
  const uploadDoc = useUploadProgressDocument(selectionId);
  const deleteDoc = useDeleteProgressDocument(selectionId);
  const fileInputRef = React.useRef<HTMLInputElement>(null);

  const getDocType = (): string => {
    if (stepName === "Medical") return "medical";
    if (stepName === "Visa") return "visa";
    if (stepName === "Ticket") return "ticket";
    if (stepName === "CoC") return "coc";
    if (stepName === "Arrival City") return "arrival";
    return "other";
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      await uploadDoc.mutateAsync({ documentType: getDocType(), file });
    } finally {
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  };

  const isTicket = stepName === "Ticket";

  return (
    <div className={cn("rounded-2xl border p-5 flex flex-col gap-4", accentClass)}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2.5">
          <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-white/60 shadow-sm">
            <Icon className="h-4 w-4" />
          </div>
          <h4 className="text-sm font-bold">{stepName}</h4>
        </div>
        <StatusDropdown
          value={status}
          disabled={!canEdit}
          isSaving={isSaving}
          onSelect={(v) => {
            if (isTicket) {
              onUpdate(stepName, v, { destination_country: destCountry, arrival_date: arrivalDate });
            } else {
              onUpdate(stepName, v);
            }
          }}
          options={
            stepName === "Ticket" ? TICKET_STATUS_MAP :
            stepName === "Arrival City" ? ARRIVAL_STATUS_MAP :
            undefined
          }
        />
      </div>

      {isTicket && (
        <div className="grid grid-cols-2 gap-3 mt-2">
          <div className="space-y-1">
            <Label className="text-[11px] font-semibold uppercase tracking-wider opacity-70">Destination Country</Label>
            <Input 
              value={destCountry} 
              onChange={(e) => setDestCountry(e.target.value)} 
              placeholder="e.g. UAE" 
              disabled={!canEdit}
              className="h-8 text-xs bg-white/50 border-white/40 shadow-sm"
              onBlur={() => onUpdate(stepName, status, { destination_country: destCountry, arrival_date: arrivalDate })}
            />
          </div>
          <div className="space-y-1">
            <Label className="text-[11px] font-semibold uppercase tracking-wider opacity-70">Arrival Date</Label>
            <Input 
              type="date"
              value={arrivalDate} 
              onChange={(e) => setArrivalDate(e.target.value)} 
              disabled={!canEdit}
              className="h-8 text-xs bg-white/50 border-white/40 shadow-sm"
              onBlur={() => onUpdate(stepName, status, { destination_country: destCountry, arrival_date: arrivalDate })}
            />
          </div>
        </div>
      )}

      {step?.document_url ? (
        <div className="flex items-center gap-2 rounded-xl bg-white/60 px-3 py-2 border border-border/40">
          <FileBadge2 className="h-4 w-4 text-muted-foreground shrink-0" />
          <span className="text-xs text-muted-foreground truncate flex-1">
            {step.document_name || "Document uploaded"}
          </span>
          <Button size="sm" variant="ghost" className="h-7 w-7 rounded-lg p-0" asChild>
            <a href={step.document_url} target="_blank" rel="noopener noreferrer" title="View document">
              <Eye className="h-3.5 w-3.5" />
            </a>
          </Button>
          {canEdit && (
            <Button
              size="sm"
              variant="ghost"
              className="h-7 w-7 rounded-lg p-0 text-destructive hover:text-destructive"
              onClick={() => deleteDoc.mutate(getDocType())}
              disabled={deleteDoc.isPending}
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          )}
        </div>
      ) : null}

      {step?.notes && (
        <div className="rounded-xl bg-white/50 px-3 py-2">
          <p className="text-[11px] text-muted-foreground font-medium uppercase tracking-widest mb-1">Notes</p>
          <p className="text-sm text-foreground">{step.notes}</p>
        </div>
      )}

      <div className="flex items-center justify-between mt-1">
        {step?.updated_by?.name ? (
          <p className="text-[11px] text-muted-foreground/80">
            Last updated by <span className="font-medium">{step.updated_by.name}</span>
          </p>
        ) : <div />}
        
        {canEdit && (
          <div>
            <input 
              type="file" 
              ref={fileInputRef} 
              className="hidden" 
              onChange={handleFileChange}
              accept=".pdf,.jpg,.jpeg,.png"
            />
            <Button 
              size="sm" 
              variant="outline" 
              className={cn("h-7 text-[11px] rounded-lg gap-1.5 bg-white/60 hover:bg-white", step?.document_url ? "text-muted-foreground" : "")}
              onClick={() => fileInputRef.current?.click()}
              disabled={uploadDoc.isPending}
            >
              {uploadDoc.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : <UploadCloud className="h-3 w-3" />}
              {step?.document_url ? "Change" : "Upload"} Doc
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}

interface ProcessStepsTabProps {
  selectionId: string;
  steps: StepItem[];
  canEdit: boolean;
  savingSteps: Set<string>;
  onUpdate: (stepName: string, status: string, extras?: { destination_country?: string; arrival_date?: string }) => void;
}

export function ProcessStepsTab({ selectionId, steps, canEdit, savingSteps, onUpdate }: ProcessStepsTabProps) {
  return (
    <div className="flex flex-col gap-4">
      <StepDetailCard
        selectionId={selectionId}
        stepName="Medical"
        icon={Stethoscope}
        steps={steps}
        canEdit={canEdit}
        isSaving={savingSteps.has("Medical")}
        onUpdate={onUpdate}
        accentClass="border-emerald-100 bg-emerald-50/60 text-emerald-900"
      />
      <StepDetailCard
        selectionId={selectionId}
        stepName="CoC"
        icon={FileText}
        steps={steps}
        canEdit={canEdit}
        isSaving={savingSteps.has("CoC")}
        onUpdate={onUpdate}
        accentClass="border-violet-100 bg-violet-50/60 text-violet-900"
      />
      <StepDetailCard
        selectionId={selectionId}
        stepName="Visa"
        icon={Building2}
        steps={steps}
        canEdit={canEdit}
        isSaving={savingSteps.has("Visa")}
        onUpdate={onUpdate}
        accentClass="border-sky-100 bg-sky-50/60 text-sky-900"
      />
      <StepDetailCard
        selectionId={selectionId}
        stepName="Ticket"
        icon={Ticket}
        steps={steps}
        canEdit={canEdit}
        isSaving={savingSteps.has("Ticket")}
        onUpdate={onUpdate}
        accentClass="border-amber-100 bg-amber-50/60 text-amber-900"
      />
      <StepDetailCard
        selectionId={selectionId}
        stepName="Arrival City"
        icon={MapPin}
        steps={steps}
        canEdit={canEdit}
        isSaving={savingSteps.has("Arrival City")}
        onUpdate={onUpdate}
        accentClass="border-teal-100 bg-teal-50/60 text-teal-900"
      />
    </div>
  );
}
