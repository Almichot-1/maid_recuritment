"use client";

import * as React from "react";
import {
  AlertTriangle,
  CheckCircle2,
  Clock3,
  FileBadge2,
  Loader2,
} from "lucide-react";

import { StatusStep } from "@/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DocumentUpload } from "@/components/candidates/document-upload";
import { cn } from "@/lib/utils";

interface StatusTimelineProps {
  steps: StatusStep[];
  canUpdate?: boolean;
  onUpdateStep?: (
    stepName: string,
    status: string,
    notes?: string,
    cocStatus?: string,
    arrivalCity?: string,
  ) => void;
  isUpdating?: boolean;
  onUploadMedicalDocument?: (file: File) => void | Promise<unknown>;
  onRemoveMedicalDocument?: () => void | Promise<unknown>;
  isUploadingMedicalDocument?: boolean;
  isRemovingMedicalDocument?: boolean;
}

export function StatusTimeline({
  steps,
  canUpdate = false,
  onUpdateStep,
  isUpdating = false,
  onUploadMedicalDocument,
  onRemoveMedicalDocument,
  isUploadingMedicalDocument = false,
  isRemovingMedicalDocument = false,
}: StatusTimelineProps) {
  const [noteDrafts, setNoteDrafts] = React.useState<Record<string, string>>(
    {},
  );
  const [cocStatuses, setCoCStatuses] = React.useState<
    Record<string, string>
  >({});
  const [arrivalCities, setArrivalCities] = React.useState<
    Record<string, string>
  >({});

  const getNote = (step: StatusStep) =>
    noteDrafts[step.id] ?? step.notes ?? "";
  const getCoC = (step: StatusStep) =>
    cocStatuses[step.id] ?? step.coc_status ?? "";
  const getCity = (step: StatusStep) =>
    arrivalCities[step.id] ?? step.arrival_city ?? "";

  const handleStatusChange = (step: StatusStep, newStatus: string) => {
    if (!onUpdateStep || isUpdating) return;
    if (newStatus === step.step_status) return;

    if (
      isMedicalStep(step) &&
      newStatus === "completed" &&
      !step.medical_document_url
    ) {
      return;
    }

    onUpdateStep(
      step.step_name,
      newStatus,
      getNote(step),
      isCoCStep(step) ? getCoC(step) || undefined : undefined,
      isArrivalCityStep(step) ? getCity(step) || undefined : undefined,
    );
  };

  const handleSaveExtras = (step: StatusStep) => {
    if (!onUpdateStep || isUpdating) return;
    onUpdateStep(
      step.step_name,
      step.step_status,
      getNote(step),
      isCoCStep(step) ? getCoC(step) || undefined : undefined,
      isArrivalCityStep(step) ? getCity(step) || undefined : undefined,
    );
  };

  return (
    <div className="space-y-2">
      {steps.map((step, index) => {
        const isMedical = isMedicalStep(step);
        const isCoC = isCoCStep(step);
        const isArrival = isArrivalCityStep(step);
        const checked = step.step_status === "completed";
        const isFailed = step.step_status === "failed";
        const cocVal = getCoC(step);
        const cityVal = getCity(step);

        return (
          <div
            key={step.id}
            className={cn(
              "rounded-xl border px-4 py-3 transition-colors",
              checked
                ? "border-emerald-200/80 bg-emerald-50/60 dark:border-emerald-900/40 dark:bg-emerald-950/20"
                : isFailed
                  ? "border-rose-300/70 bg-rose-50/70 dark:border-rose-900/40 dark:bg-rose-950/20"
                  : "border-border/70 bg-card/95",
            )}
          >
            <div className="flex items-center gap-3">
              {/* Step indicator */}
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-border/70 bg-background/80 text-xs font-bold text-muted-foreground">
                {index + 1}
              </div>

              {/* Step name + status */}
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-semibold">
                    {step.step_name}
                  </span>
                  <StatusBadge status={step.step_status} />
                  {isCoC && cocVal ? (
                    <Badge
                      variant="outline"
                      className={cn(
                        "text-[10px] px-1.5 py-0 h-5",
                        cocVal === "online"
                          ? "border-sky-300 text-sky-700 bg-sky-50 dark:border-sky-800 dark:text-sky-300 dark:bg-sky-950/30"
                          : "border-amber-300 text-amber-700 bg-amber-50 dark:border-amber-800 dark:text-amber-300 dark:bg-amber-950/30",
                      )}
                    >
                      {cocVal === "online" ? "Online" : "Not Online"}
                    </Badge>
                  ) : null}
                  {isArrival && cityVal ? (
                    <Badge
                      variant="outline"
                      className="text-[10px] px-1.5 py-0 h-5"
                    >
                      {cityVal}
                    </Badge>
                  ) : null}
                </div>
              </div>

              {/* Status dropdown */}
              {canUpdate && !isMedical ? (
                <Select
                  value={step.step_status}
                  onValueChange={(v) => handleStatusChange(step, v)}
                  disabled={isUpdating}
                >
                  <SelectTrigger className="h-8 w-[130px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="pending">Pending</SelectItem>
                    <SelectItem value="in_progress">In Progress</SelectItem>
                    <SelectItem value="completed">Completed</SelectItem>
                    <SelectItem value="failed">Failed</SelectItem>
                  </SelectContent>
                </Select>
              ) : canUpdate && isMedical ? (
                <Select
                  value={step.step_status}
                  onValueChange={(v) => handleStatusChange(step, v)}
                  disabled={isUpdating}
                >
                  <SelectTrigger className="h-8 w-[130px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="pending">Pending</SelectItem>
                    <SelectItem value="in_progress">In Progress</SelectItem>
                    <SelectItem
                      value="completed"
                      disabled={!step.medical_document_url}
                    >
                      Completed
                    </SelectItem>
                    <SelectItem value="failed">Failed</SelectItem>
                  </SelectContent>
                </Select>
              ) : null}

              {/* Updated at */}
              <span className="hidden sm:block text-[11px] text-muted-foreground shrink-0">
                {step.updated_by?.name || "System"}
              </span>
            </div>

            {/* CoC sub-status */}
            {isCoC && canUpdate ? (
              <div className="mt-2 flex items-center gap-2 pl-11">
                <span className="text-xs text-muted-foreground shrink-0">
                  CoC status:
                </span>
                <Select
                  value={cocVal}
                  onValueChange={(v) => {
                    setCoCStatuses((prev) => ({ ...prev, [step.id]: v }));
                  }}
                >
                  <SelectTrigger className="h-7 w-[130px] text-xs">
                    <SelectValue placeholder="Select..." />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="not_online">Not Online</SelectItem>
                    <SelectItem value="online">Online</SelectItem>
                  </SelectContent>
                </Select>
                {cocVal !== (step.coc_status ?? "") ? (
                  <Button
                    size="sm"
                    variant="ghost"
                    className="h-7 text-xs"
                    onClick={() => handleSaveExtras(step)}
                    disabled={isUpdating}
                  >
                    Save
                  </Button>
                ) : null}
              </div>
            ) : null}

            {/* Arrival City */}
            {isArrival && canUpdate ? (
              <div className="mt-2 flex items-center gap-2 pl-11">
                <span className="text-xs text-muted-foreground shrink-0">
                  City:
                </span>
                <Input
                  className="h-7 w-[180px] text-xs"
                  placeholder="e.g. Dubai"
                  value={cityVal}
                  onChange={(e) =>
                    setArrivalCities((prev) => ({
                      ...prev,
                      [step.id]: e.target.value,
                    }))
                  }
                />
                {cityVal !== (step.arrival_city ?? "") ? (
                  <Button
                    size="sm"
                    variant="ghost"
                    className="h-7 text-xs"
                    onClick={() => handleSaveExtras(step)}
                    disabled={isUpdating}
                  >
                    Save
                  </Button>
                ) : null}
              </div>
            ) : null}

            {/* Notes */}
            {canUpdate ? (
              <div className="mt-2 pl-11">
                <Textarea
                  className="min-h-[32px] text-xs resize-none"
                  placeholder="Add a note..."
                  value={getNote(step)}
                  onChange={(e) =>
                    setNoteDrafts((prev) => ({
                      ...prev,
                      [step.id]: e.target.value,
                    }))
                  }
                  rows={1}
                />
                {getNote(step) !== (step.notes ?? "") ? (
                  <div className="mt-1 flex justify-end">
                    <Button
                      size="sm"
                      variant="ghost"
                      className="h-6 text-xs"
                      onClick={() => handleSaveExtras(step)}
                      disabled={isUpdating}
                    >
                      {isUpdating ? (
                        <Loader2 className="h-3 w-3 animate-spin mr-1" />
                      ) : null}
                      Save note
                    </Button>
                  </div>
                ) : null}
              </div>
            ) : step.notes ? (
              <div className="mt-2 pl-11">
                <p className="rounded-lg bg-muted/50 px-3 py-2 text-xs text-muted-foreground">
                  {step.notes}
                </p>
              </div>
            ) : null}

            {/* File upload per step */}
            {canUpdate && (
              <div className="mt-2 pl-11">
                <FileUploadSlot
                  isMedical={isMedical}
                  medicalDocumentUrl={step.medical_document_url}
                  onUploadMedicalDocument={onUploadMedicalDocument}
                  onRemoveMedicalDocument={onRemoveMedicalDocument}
                  isUploading={isUploadingMedicalDocument}
                  isRemoving={isRemovingMedicalDocument}
                />
              </div>
            )}

            {isFailed && step.notes ? (
              <div className="mt-2 pl-11">
                <p className="rounded-lg border border-rose-300/50 bg-rose-50/80 px-3 py-2 text-xs text-rose-900 dark:border-rose-900/50 dark:bg-rose-950/25 dark:text-rose-100">
                  <span className="font-semibold">Reason:</span> {step.notes}
                </p>
              </div>
            ) : null}
          </div>
        );
      })}
    </div>
  );
}

function FileUploadSlot({
  isMedical,
  medicalDocumentUrl,
  onUploadMedicalDocument,
  onRemoveMedicalDocument,
  isUploading,
  isRemoving,
}: {
  isMedical: boolean;
  medicalDocumentUrl?: string;
  onUploadMedicalDocument?: (file: File) => void | Promise<unknown>;
  onRemoveMedicalDocument?: () => void | Promise<unknown>;
  isUploading: boolean;
  isRemoving: boolean;
}) {
  if (isMedical && medicalDocumentUrl) {
    return (
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" className="h-7 text-xs" asChild>
          <a href={medicalDocumentUrl} target="_blank" rel="noreferrer">
            <FileBadge2 className="mr-1 h-3 w-3" />
            View file
          </a>
        </Button>
        {onRemoveMedicalDocument ? (
          <Button
            variant="ghost"
            size="sm"
            className="h-7 text-xs"
            disabled={isRemoving}
            onClick={() => onRemoveMedicalDocument()}
          >
            {isRemoving ? (
              <Loader2 className="h-3 w-3 animate-spin" />
            ) : (
              "Remove"
            )}
          </Button>
        ) : null}
      </div>
    );
  }

  if (isMedical && onUploadMedicalDocument) {
    return (
      <DocumentUpload
        documentType="medical"
        title=""
        description=""
        accept={{
          "application/pdf": [".pdf"],
          "image/jpeg": [".jpg", ".jpeg"],
          "image/png": [".png"],
        }}
        maxSize={10485760}
        mode="instant"
        disabled={isUploading}
        onUpload={(file) => onUploadMedicalDocument(file)}
        className="max-w-[200px]"
      />
    );
  }

  return null;
}

function StatusBadge({ status }: { status: StatusStep["step_status"] }) {
  switch (status) {
    case "completed":
      return (
        <Badge className="bg-emerald-600 text-white hover:bg-emerald-600 h-5 text-[10px] px-1.5">
          <CheckCircle2 className="mr-0.5 h-3 w-3" />
          Done
        </Badge>
      );
    case "failed":
      return (
        <Badge className="bg-rose-600 text-white hover:bg-rose-600 h-5 text-[10px] px-1.5">
          <AlertTriangle className="mr-0.5 h-3 w-3" />
          Failed
        </Badge>
      );
    case "in_progress":
      return (
        <Badge className="bg-sky-600 text-white hover:bg-sky-600 h-5 text-[10px] px-1.5">
          <Loader2 className="mr-0.5 h-3 w-3 animate-spin" />
          Active
        </Badge>
      );
    default:
      return (
        <Badge variant="outline" className="h-5 text-[10px] px-1.5">
          <Clock3 className="mr-0.5 h-3 w-3" />
          Pending
        </Badge>
      );
  }
}

function isMedicalStep(step: StatusStep) {
  return step.step_name.trim().toLowerCase() === "medical";
}

function isCoCStep(step: StatusStep) {
  return step.step_name.trim().toLowerCase() === "coc";
}

function isArrivalCityStep(step: StatusStep) {
  return step.step_name.trim().toLowerCase() === "arrival city";
}
