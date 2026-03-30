"use client";

import * as React from "react";
import {
  AlertTriangle,
  CheckCircle2,
  Clock3,
  FileBadge2,
  Loader2,
  PlayCircle,
  RotateCcw,
} from "lucide-react";
import { format, isValid } from "date-fns";

import { StatusStep } from "@/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Textarea } from "@/components/ui/textarea";
import { DocumentUpload } from "@/components/candidates/document-upload";
import { cn } from "@/lib/utils";

interface StatusTimelineProps {
  steps: StatusStep[];
  canUpdate?: boolean;
  onUpdateStep?: (stepName: string, status: string, notes?: string) => void;
  isUpdating?: boolean;
  onUploadMedicalDocument?: (file: File) => void | Promise<unknown>;
  onRemoveMedicalDocument?: () => void | Promise<unknown>;
  isUploadingMedicalDocument?: boolean;
  isRemovingMedicalDocument?: boolean;
}

type EditorMode = "note" | "failed";

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
  const [editorStepId, setEditorStepId] = React.useState<string | null>(null);
  const [editorMode, setEditorMode] = React.useState<EditorMode>("note");
  const [noteDraft, setNoteDraft] = React.useState("");
  const [editorError, setEditorError] = React.useState("");

  const openEditor = (step: StatusStep, mode: EditorMode) => {
    setEditorStepId(step.id);
    setEditorMode(mode);
    setNoteDraft(step.notes || "");
    setEditorError("");
  };

  const closeEditor = () => {
    setEditorStepId(null);
    setEditorMode("note");
    setNoteDraft("");
    setEditorError("");
  };

  const handleToggle = (step: StatusStep, checked: boolean) => {
    if (!onUpdateStep) {
      return;
    }

    if (checked) {
      if (isMedicalStep(step) && !step.medical_document_url) {
        setEditorError(
          "Upload the medical document first so this step can be checked off.",
        );
        setEditorStepId(step.id);
        setEditorMode("note");
        setNoteDraft(step.notes || "");
        return;
      }
      onUpdateStep(
        step.step_name,
        "completed",
        noteDraftForStep(step, noteDraft),
      );
      closeEditor();
      return;
    }

    onUpdateStep(step.step_name, "pending", noteDraftForStep(step, noteDraft));
    closeEditor();
  };

  const handleSaveFailed = (step: StatusStep) => {
    if (!onUpdateStep) {
      return;
    }

    const trimmed = noteDraft.trim();
    if (!trimmed) {
      setEditorError(
        "Add a short reason so both agencies understand what blocked this step.",
      );
      return;
    }
    onUpdateStep(step.step_name, "failed", trimmed);
    closeEditor();
  };

  const handleSaveNote = (step: StatusStep) => {
    if (!onUpdateStep) {
      return;
    }

    onUpdateStep(
      step.step_name,
      step.step_status,
      noteDraftForStep(step, noteDraft),
    );
    closeEditor();
  };

  const handleMarkInProgress = (step: StatusStep) => {
    if (!onUpdateStep) {
      return;
    }

    onUpdateStep(
      step.step_name,
      "in_progress",
      noteDraftForStep(step, noteDraft),
    );
    closeEditor();
  };

  return (
    <div className="space-y-3">
      {steps.map((step, index) => {
        const isMedical = isMedicalStep(step);
        const checked = step.step_status === "completed";
        const isFailed = step.step_status === "failed";
        const isInProgress = step.step_status === "in_progress";
        const isEditing = editorStepId === step.id;

        return (
          <div
            key={step.id}
            className={cn(
              "rounded-2xl border px-4 py-4 transition-colors duration-200",
              checked
                ? "border-emerald-200/80 bg-emerald-50/60 dark:border-emerald-900/40 dark:bg-emerald-950/20"
                : isFailed
                  ? "border-rose-300/70 bg-rose-50/70 dark:border-rose-900/40 dark:bg-rose-950/20"
                  : isInProgress
                    ? "border-sky-200/80 bg-sky-50/60 dark:border-sky-900/40 dark:bg-sky-950/20"
                    : "border-border/70 bg-card/95 hover:border-primary/20",
            )}
          >
            <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div className="flex min-w-0 gap-3">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-border/70 bg-background/80">
                  <Checkbox
                    checked={checked}
                    disabled={
                      isUpdating ||
                      (isMedical && !checked && !step.medical_document_url)
                    }
                    onCheckedChange={(value) =>
                      handleToggle(step, Boolean(value))
                    }
                    aria-label={`Toggle ${step.step_name}`}
                  />
                </div>
                <div className="min-w-0 space-y-1.5">
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="text-xs font-semibold uppercase tracking-[0.22em] text-muted-foreground">
                      Step {index + 1}
                    </p>
                    <StatusBadge status={step.step_status} />
                  </div>
                  <div>
                    <h4 className="text-base font-semibold text-foreground">
                      {step.step_name}
                    </h4>
                    <p className="mt-0.5 text-sm text-muted-foreground">
                      {checked
                        ? "Completed."
                        : isFailed
                          ? "Blocked until the issue is resolved."
                          : isInProgress
                            ? "Currently active."
                            : isMedical
                              ? "Upload the medical file, then complete the step."
                              : "Mark it in progress, complete it, or add a short failure note."}
                    </p>
                  </div>
                </div>
              </div>

              <div className="flex flex-wrap gap-2">
                {canUpdate && !checked ? (
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={isUpdating}
                    onClick={() => openEditor(step, "failed")}
                    className="h-8 rounded-lg border-rose-300 px-3 text-rose-700 hover:bg-rose-50 dark:border-rose-900/60 dark:text-rose-200 dark:hover:bg-rose-950/30"
                  >
                    <AlertTriangle className="mr-2 h-4 w-4" />
                    {isFailed ? "Failure note" : "Failed"}
                  </Button>
                ) : null}
                {canUpdate && !checked && step.step_status !== "in_progress" ? (
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={isUpdating}
                    onClick={() => handleMarkInProgress(step)}
                    className="h-8 rounded-lg border-sky-300 px-3 text-sky-700 hover:bg-sky-50 dark:border-sky-900/60 dark:text-sky-200 dark:hover:bg-sky-950/30"
                  >
                    <PlayCircle className="mr-2 h-4 w-4" />
                    In progress
                  </Button>
                ) : null}
                {canUpdate && (step.notes || !checked) ? (
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={isUpdating}
                    onClick={() => openEditor(step, "note")}
                    className="h-8 rounded-lg px-3"
                  >
                    {checked || step.step_status === "in_progress" ? (
                      <RotateCcw className="mr-2 h-4 w-4" />
                    ) : (
                      <FileBadge2 className="mr-2 h-4 w-4" />
                    )}
                    {checked
                      ? "Reopen note"
                      : step.step_status === "in_progress"
                        ? "Progress note"
                        : "Add note"}
                  </Button>
                ) : null}
              </div>
            </div>

            <div className="mt-3 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
              <span>
                <span className="font-semibold text-foreground">Updated:</span>{" "}
                {formatStepDate(step.updated_at, "Awaiting update")}
              </span>
              <span>
                <span className="font-semibold text-foreground">By:</span>{" "}
                {step.updated_by?.name || "System"}
              </span>
              <span>
                <span className="font-semibold text-foreground">
                  Completed:
                </span>{" "}
                {step.completed_at
                  ? formatStepDate(step.completed_at, "Not completed")
                  : "Not completed"}
              </span>
            </div>

            {isMedical ? (
              <div className="mt-3 rounded-xl border border-border/70 bg-background/70 p-3.5">
                <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                  <div className="space-y-1">
                    <p className="text-sm font-semibold text-foreground">
                      Medical document
                    </p>
                    <p className="text-sm text-muted-foreground">
                      Upload a PDF, JPG, or PNG medical report to unlock this step.
                    </p>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {step.medical_document_url ? (
                      <Button variant="outline" size="sm" asChild>
                        <a
                          href={step.medical_document_url}
                          target="_blank"
                          rel="noreferrer"
                        >
                          <FileBadge2 className="mr-2 h-4 w-4" />
                          View medical file
                        </a>
                      </Button>
                    ) : null}
                    {step.medical_document_url && canUpdate && onRemoveMedicalDocument ? (
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        disabled={isRemovingMedicalDocument}
                        onClick={() => {
                          void onRemoveMedicalDocument()
                        }}
                      >
                        {isRemovingMedicalDocument ? (
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        ) : (
                          <RotateCcw className="mr-2 h-4 w-4" />
                        )}
                        Remove medical file
                      </Button>
                    ) : null}
                  </div>
                </div>

                {!step.medical_document_url && canUpdate ? (
                  <div className="mt-4">
                    <DocumentUpload
                      documentType="medical"
                      title="Upload medical document"
                      description="Drag a medical PDF, JPG, or PNG here to unlock the Medical checklist item."
                      accept={{
                        "application/pdf": [".pdf"],
                        "image/jpeg": [".jpg", ".jpeg"],
                        "image/png": [".png"],
                      }}
                      maxSize={10485760}
                      mode="instant"
                      disabled={isUploadingMedicalDocument || isUpdating}
                      onUpload={(file) => onUploadMedicalDocument?.(file)}
                    />
                  </div>
                ) : null}
              </div>
            ) : null}

            {step.notes ? (
              <div
                className={cn(
                  "mt-3 rounded-xl border px-3.5 py-3 text-sm",
                  isFailed
                    ? "border-rose-300/50 bg-rose-50/80 text-rose-900 dark:border-rose-900/50 dark:bg-rose-950/25 dark:text-rose-100"
                    : "border-border/70 bg-background/80 text-muted-foreground",
                )}
              >
                <p className="mb-1 text-[11px] font-semibold uppercase tracking-[0.22em] text-muted-foreground">
                  {isFailed ? "Failure reason" : "Notes"}
                </p>
                {step.notes}
              </div>
            ) : null}

            {isEditing ? (
              <div className="mt-3 rounded-xl border border-primary/15 bg-background/90 p-3.5">
                <p className="text-sm font-semibold text-foreground">
                  {editorMode === "failed"
                    ? `Explain why ${step.step_name} failed`
                    : `Add a shared note for ${step.step_name}`}
                </p>
                <Textarea
                  className="mt-3 min-h-[88px] resize-none"
                  value={noteDraft}
                  onChange={(event) => {
                    setEditorError("");
                    setNoteDraft(event.target.value);
                  }}
                  placeholder={
                    editorMode === "failed"
                      ? "Explain the blocker clearly so the employer sees why this milestone stopped."
                      : "Add any note you want both agencies to see for this milestone."
                  }
                />
                {editorError ? (
                  <p className="mt-3 text-sm font-medium text-rose-600 dark:text-rose-300">
                    {editorError}
                  </p>
                ) : null}
                <div className="mt-3 flex flex-wrap gap-2">
                  {editorMode === "failed" ? (
                    <Button
                      type="button"
                      size="sm"
                      disabled={isUpdating}
                      onClick={() => handleSaveFailed(step)}
                    >
                      {isUpdating ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      ) : (
                        <AlertTriangle className="mr-2 h-4 w-4" />
                      )}
                      Save failure note
                    </Button>
                  ) : (
                    <Button
                      type="button"
                      size="sm"
                      disabled={isUpdating}
                      onClick={() => handleSaveNote(step)}
                    >
                      {isUpdating ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      ) : (
                        <FileBadge2 className="mr-2 h-4 w-4" />
                      )}
                      Save note
                    </Button>
                  )}
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={isUpdating}
                    onClick={closeEditor}
                  >
                    Cancel
                  </Button>
                </div>
              </div>
            ) : null}
          </div>
        );
      })}
    </div>
  );
}

function StatusBadge({ status }: { status: StatusStep["step_status"] }) {
  switch (status) {
    case "completed":
      return (
        <Badge className="bg-emerald-600 text-white hover:bg-emerald-600">
          <CheckCircle2 className="mr-1 h-3.5 w-3.5" />
          Completed
        </Badge>
      );
    case "failed":
      return (
        <Badge className="bg-rose-600 text-white hover:bg-rose-600">
          <AlertTriangle className="mr-1 h-3.5 w-3.5" />
          Failed
        </Badge>
      );
    case "in_progress":
      return (
        <Badge className="bg-sky-600 text-white hover:bg-sky-600">
          <Loader2 className="mr-1 h-3.5 w-3.5 animate-spin" />
          In progress
        </Badge>
      );
    default:
      return (
        <Badge variant="outline">
          <Clock3 className="mr-1 h-3.5 w-3.5" />
          Pending
        </Badge>
      );
  }
}

function formatStepDate(value: string | undefined, fallback: string) {
  if (!value) {
    return fallback;
  }

  const parsed = new Date(value);
  if (!isValid(parsed)) {
    return fallback;
  }

  return format(parsed, "MMM dd, yyyy 'at' h:mm a");
}

function isMedicalStep(step: StatusStep) {
  return step.step_name.trim().toLowerCase() === "medical";
}

function noteDraftForStep(step: StatusStep, draft: string) {
  const trimmed = draft.trim();
  if (trimmed) {
    return trimmed;
  }
  if (step.notes?.trim()) {
    return step.notes.trim();
  }
  return undefined;
}
