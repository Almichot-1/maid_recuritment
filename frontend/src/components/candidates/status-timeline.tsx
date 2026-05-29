"use client";

import * as React from "react";
import {
  AlertTriangle,
  CheckCircle2,
  ChevronDown,
  ChevronUp,
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
  isUploadingMedicalDocument?: boolean;
}

type EditorMode = "note" | "failed";

const FOCUS_WINDOW = 3;

export function StatusTimeline({
  steps,
  canUpdate = false,
  onUpdateStep,
  isUpdating = false,
  onUploadMedicalDocument,
  isUploadingMedicalDocument = false,
}: StatusTimelineProps) {
  const [showAll, setShowAll] = React.useState(false);
  const [editorStepId, setEditorStepId] = React.useState<string | null>(null);
  const [editorMode, setEditorMode] = React.useState<EditorMode>("note");
  const [noteDraft, setNoteDraft] = React.useState("");
  const [editorError, setEditorError] = React.useState("");

  const focusIndex = React.useMemo(() => {
    const inProgress = steps.findIndex((s) => s.step_status === "in_progress");
    if (inProgress >= 0) return inProgress;
    const failed = steps.findIndex((s) => s.step_status === "failed");
    if (failed >= 0) return failed;
    const pending = steps.findIndex((s) => s.step_status === "pending");
    if (pending >= 0) return pending;
    return Math.max(0, steps.length - 1);
  }, [steps]);

  const completedCount = steps.filter((s) => s.step_status === "completed").length;
  const hiddenBefore = showAll ? 0 : Math.max(0, focusIndex);
  const visibleEnd = showAll ? steps.length : Math.min(steps.length, focusIndex + FOCUS_WINDOW);
  const visibleSteps = steps.slice(hiddenBefore, visibleEnd);
  const hiddenAfter = showAll ? 0 : steps.length - visibleEnd;

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
    if (!onUpdateStep) return;

    if (checked) {
      if (isMedicalStep(step) && !step.medical_document_url) {
        setEditorError("Upload the medical file before marking complete.");
        setEditorStepId(step.id);
        setEditorMode("note");
        setNoteDraft(step.notes || "");
        return;
      }
      onUpdateStep(step.step_name, "completed", noteDraftForStep(step, noteDraft));
      closeEditor();
      return;
    }

    onUpdateStep(step.step_name, "pending", noteDraftForStep(step, noteDraft));
    closeEditor();
  };

  const handleSaveFailed = (step: StatusStep) => {
    if (!onUpdateStep) return;
    const trimmed = noteDraft.trim();
    if (!trimmed) {
      setEditorError("Add a short reason.");
      return;
    }
    onUpdateStep(step.step_name, "failed", trimmed);
    closeEditor();
  };

  const handleSaveNote = (step: StatusStep) => {
    if (!onUpdateStep) return;
    onUpdateStep(step.step_name, step.step_status, noteDraftForStep(step, noteDraft));
    closeEditor();
  };

  const handleMarkInProgress = (step: StatusStep) => {
    if (!onUpdateStep) return;
    onUpdateStep(step.step_name, "in_progress", noteDraftForStep(step, noteDraft));
    closeEditor();
  };

  return (
    <div className="space-y-3">
      {!showAll && completedCount > 0 && focusIndex > 0 ? (
        <p className="text-sm text-muted-foreground">
          {completedCount} step{completedCount === 1 ? "" : "s"} completed
        </p>
      ) : null}

      {visibleSteps.map((step) => {
        const isFocus = step.id === steps[focusIndex]?.id;
        return (
          <StepCard
            key={step.id}
            step={step}
            isFocus={isFocus}
            canUpdate={canUpdate}
            isUpdating={isUpdating}
            isEditing={editorStepId === step.id}
            editorMode={editorMode}
            noteDraft={noteDraft}
            editorError={editorError}
            isUploadingMedicalDocument={isUploadingMedicalDocument}
            onOpenEditor={openEditor}
            onCloseEditor={closeEditor}
            onToggle={handleToggle}
            onMarkInProgress={handleMarkInProgress}
            onSaveFailed={handleSaveFailed}
            onSaveNote={handleSaveNote}
            onNoteDraftChange={(value) => {
              setEditorError("");
              setNoteDraft(value);
            }}
            onUploadMedicalDocument={onUploadMedicalDocument}
          />
        );
      })}

      {!showAll && (hiddenBefore > 0 || hiddenAfter > 0) ? (
        <Button
          type="button"
          variant="outline"
          className="w-full"
          onClick={() => setShowAll(true)}
        >
          <ChevronDown className="mr-2 h-4 w-4" />
          Show all {steps.length} steps
        </Button>
      ) : null}

      {showAll && steps.length > FOCUS_WINDOW ? (
        <Button
          type="button"
          variant="ghost"
          className="w-full"
          onClick={() => setShowAll(false)}
        >
          <ChevronUp className="mr-2 h-4 w-4" />
          Show current steps only
        </Button>
      ) : null}
    </div>
  );
}

function StepCard({
  step,
  isFocus,
  canUpdate,
  isUpdating,
  isEditing,
  editorMode,
  noteDraft,
  editorError,
  isUploadingMedicalDocument,
  onOpenEditor,
  onCloseEditor,
  onToggle,
  onMarkInProgress,
  onSaveFailed,
  onSaveNote,
  onNoteDraftChange,
  onUploadMedicalDocument,
}: {
  step: StatusStep;
  isFocus: boolean;
  canUpdate: boolean;
  isUpdating: boolean;
  isEditing: boolean;
  editorMode: EditorMode;
  noteDraft: string;
  editorError: string;
  isUploadingMedicalDocument: boolean;
  onOpenEditor: (step: StatusStep, mode: EditorMode) => void;
  onCloseEditor: () => void;
  onToggle: (step: StatusStep, checked: boolean) => void;
  onMarkInProgress: (step: StatusStep) => void;
  onSaveFailed: (step: StatusStep) => void;
  onSaveNote: (step: StatusStep) => void;
  onNoteDraftChange: (value: string) => void;
  onUploadMedicalDocument?: (file: File) => void | Promise<unknown>;
}) {
  const isMedical = isMedicalStep(step);
  const checked = step.step_status === "completed";
  const isFailed = step.step_status === "failed";
  const isInProgress = step.step_status === "in_progress";

  return (
    <div
      className={cn(
        "rounded-xl border px-4 py-4 transition-colors",
        checked && "border-primary/20 bg-primary/5",
        isFailed && "border-destructive/30 bg-destructive/5",
        isInProgress && "border-primary/40 bg-muted/50",
        !checked && !isFailed && !isInProgress && "border-border bg-card",
        isFocus && "ring-2 ring-primary/20",
      )}
    >
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex min-w-0 gap-3">
          <Checkbox
            checked={checked}
            disabled={isUpdating || (isMedical && !checked && !step.medical_document_url)}
            onCheckedChange={(value) => onToggle(step, Boolean(value))}
            aria-label={`Mark ${step.step_name} complete`}
            className="mt-1 h-5 w-5"
          />
          <div className="min-w-0 space-y-1">
            <div className="flex flex-wrap items-center gap-2">
              <h4 className="text-base font-semibold text-foreground">{step.step_name}</h4>
              <StatusBadge status={step.step_status} />
              {isFocus ? (
                <Badge variant="secondary" className="text-xs">
                  Current
                </Badge>
              ) : null}
            </div>
            {step.updated_at ? (
              <p className="text-xs text-muted-foreground">
                Updated {formatStepDate(step.updated_at, "")}
                {step.updated_by?.name ? ` · ${step.updated_by.name}` : ""}
              </p>
            ) : null}
          </div>
        </div>

        {canUpdate && !checked ? (
          <div className="flex flex-wrap gap-2 sm:shrink-0">
            <Button
              type="button"
              variant={isFailed ? "destructive" : "outline"}
              disabled={isUpdating}
              onClick={() => onOpenEditor(step, "failed")}
            >
              <AlertTriangle className="mr-2 h-4 w-4" />
              Blocked
            </Button>
            {!isInProgress ? (
              <Button
                type="button"
                disabled={isUpdating}
                onClick={() => onMarkInProgress(step)}
              >
                <PlayCircle className="mr-2 h-4 w-4" />
                Start
              </Button>
            ) : null}
            <Button
              type="button"
              variant="outline"
              disabled={isUpdating}
              onClick={() => onOpenEditor(step, "note")}
            >
              <FileBadge2 className="mr-2 h-4 w-4" />
              Note
            </Button>
          </div>
        ) : canUpdate && checked ? (
          <Button
            type="button"
            variant="outline"
            disabled={isUpdating}
            onClick={() => onOpenEditor(step, "note")}
          >
            <RotateCcw className="mr-2 h-4 w-4" />
            Edit note
          </Button>
        ) : null}
      </div>

      {isMedical ? (
        <div className="mt-4 rounded-lg border border-border bg-muted/30 p-4">
          <p className="text-sm font-medium text-foreground">Medical file</p>
          {step.medical_document_url ? (
            <Button variant="outline" className="mt-2" asChild>
              <a href={step.medical_document_url} target="_blank" rel="noreferrer">
                <FileBadge2 className="mr-2 h-4 w-4" />
                View file
              </a>
            </Button>
          ) : canUpdate ? (
            <div className="mt-3">
              <DocumentUpload
                documentType="medical"
                title="Upload medical document"
                description="PDF, JPG, or PNG"
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
          ) : (
            <p className="mt-1 text-sm text-muted-foreground">Waiting for upload.</p>
          )}
        </div>
      ) : null}

      {step.notes ? (
        <div
          className={cn(
            "mt-3 rounded-lg border px-3 py-2 text-sm",
            isFailed
              ? "border-destructive/30 bg-destructive/5 text-foreground"
              : "border-border bg-muted/20 text-muted-foreground",
          )}
        >
          {step.notes}
        </div>
      ) : null}

      {isEditing ? (
        <div className="mt-3 space-y-3 rounded-lg border border-border bg-background p-3">
          <Textarea
            className="min-h-[88px] resize-none"
            value={noteDraft}
            onChange={(event) => onNoteDraftChange(event.target.value)}
            placeholder={editorMode === "failed" ? "What blocked this step?" : "Optional note for your partner"}
          />
          {editorError ? (
            <p className="text-sm font-medium text-destructive">{editorError}</p>
          ) : null}
          <div className="flex flex-wrap gap-2">
            {editorMode === "failed" ? (
              <Button type="button" disabled={isUpdating} onClick={() => onSaveFailed(step)}>
                {isUpdating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                Save blocked
              </Button>
            ) : (
              <Button type="button" disabled={isUpdating} onClick={() => onSaveNote(step)}>
                {isUpdating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                Save note
              </Button>
            )}
            <Button type="button" variant="outline" disabled={isUpdating} onClick={onCloseEditor}>
              Cancel
            </Button>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function StatusBadge({ status }: { status: StatusStep["step_status"] }) {
  switch (status) {
    case "completed":
      return (
        <Badge className="bg-primary text-primary-foreground hover:bg-primary">
          <CheckCircle2 className="mr-1 h-3.5 w-3.5" />
          Done
        </Badge>
      );
    case "failed":
      return (
        <Badge variant="destructive">
          <AlertTriangle className="mr-1 h-3.5 w-3.5" />
          Blocked
        </Badge>
      );
    case "in_progress":
      return (
        <Badge variant="secondary">
          <Loader2 className="mr-1 h-3.5 w-3.5 animate-spin" />
          Active
        </Badge>
      );
    default:
      return (
        <Badge variant="outline">
          <Clock3 className="mr-1 h-3.5 w-3.5" />
          Waiting
        </Badge>
      );
  }
}

function formatStepDate(value: string | undefined, fallback: string) {
  if (!value) return fallback;
  const parsed = new Date(value);
  if (!isValid(parsed)) return fallback;
  return format(parsed, "MMM d, h:mm a");
}

function isMedicalStep(step: StatusStep) {
  return step.step_name.trim().toLowerCase() === "medical";
}

function noteDraftForStep(step: StatusStep, draft: string) {
  const trimmed = draft.trim();
  if (trimmed) return trimmed;
  if (step.notes?.trim()) return step.notes.trim();
  return undefined;
}
