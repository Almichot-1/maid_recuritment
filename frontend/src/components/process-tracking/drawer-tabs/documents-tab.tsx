"use client";

import * as React from "react";
import { FileText, Download, Eye, Upload, File, ImageIcon, FileBadge } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { format } from "date-fns";
import type { Document } from "@/types";

const DOC_TYPE_CONFIG: Record<string, { label: string; icon: React.ElementType; color: string }> = {
  passport:  { label: "Passport",  icon: FileBadge,  color: "bg-sky-50 text-sky-700 border-sky-200" },
  photo:     { label: "Photo",     icon: ImageIcon,  color: "bg-pink-50 text-pink-700 border-pink-200" },
  visa:      { label: "Visa",      icon: FileText,   color: "bg-violet-50 text-violet-700 border-violet-200" },
  medical:   { label: "Medical",   icon: FileText,   color: "bg-emerald-50 text-emerald-700 border-emerald-200" },
  contract:  { label: "Contract",  icon: FileText,   color: "bg-amber-50 text-amber-700 border-amber-200" },
  cv:        { label: "CV",        icon: File,       color: "bg-slate-50 text-slate-700 border-slate-200" },
  other:     { label: "Other",     icon: FileText,   color: "bg-zinc-50 text-zinc-700 border-zinc-200" },
};

function getDocConfig(type: string) {
  return DOC_TYPE_CONFIG[type.toLowerCase()] ?? DOC_TYPE_CONFIG.other;
}

function formatFileSize(bytes?: number) {
  if (!bytes) return null;
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

interface DocumentsTabProps {
  documents: Document[];
  canEdit: boolean;
}

export function DocumentsTab({ documents, canEdit }: DocumentsTabProps) {
  if (documents.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-16 text-center">
        <div className="h-12 w-12 rounded-full bg-muted flex items-center justify-center">
          <FileText className="h-5 w-5 text-muted-foreground" />
        </div>
        <p className="text-sm text-muted-foreground">No documents uploaded yet</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      {documents.map((doc) => {
        const config = getDocConfig(doc.document_type);
        const Icon = config.icon;
        const size = formatFileSize(doc.file_size);
        let uploadedAt: string | null = null;
        try {
          if (doc.uploaded_at) uploadedAt = format(new Date(doc.uploaded_at), "MMM d, yyyy");
        } catch { /* noop */ }

        return (
          <div
            key={doc.id}
            className="flex items-center gap-4 rounded-2xl border border-border/60 bg-white p-4 shadow-sm hover:shadow-md transition-shadow"
          >
            {/* Icon */}
            <div
              className={cn(
                "flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border",
                config.color,
              )}
            >
              <Icon className="h-5 w-5" />
            </div>

            {/* Info */}
            <div className="min-w-0 flex-1">
              <p className="text-sm font-semibold text-foreground truncate">{doc.file_name}</p>
              <div className="flex items-center gap-2 mt-0.5">
                <span
                  className={cn(
                    "inline-flex rounded-full border px-2 py-0.5 text-[10px] font-semibold",
                    config.color,
                  )}
                >
                  {config.label}
                </span>
                {size && <span className="text-[11px] text-muted-foreground">{size}</span>}
                {uploadedAt && <span className="text-[11px] text-muted-foreground">{uploadedAt}</span>}
              </div>
            </div>

            {/* Actions */}
            <div className="flex items-center gap-1.5 shrink-0">
              <Button
                size="icon"
                variant="ghost"
                className="h-8 w-8 rounded-xl"
                asChild
              >
                <a href={doc.file_url} target="_blank" rel="noopener noreferrer" title="Preview">
                  <Eye className="h-3.5 w-3.5" />
                </a>
              </Button>
              <Button
                size="icon"
                variant="ghost"
                className="h-8 w-8 rounded-xl"
                asChild
              >
                <a href={doc.file_url} download={doc.file_name} title="Download">
                  <Download className="h-3.5 w-3.5" />
                </a>
              </Button>
            </div>
          </div>
        );
      })}

      {canEdit && (
        <div className="mt-2 rounded-2xl border-2 border-dashed border-border/60 bg-muted/20 p-6 text-center hover:border-primary/40 hover:bg-primary/[0.02] transition-colors cursor-pointer">
          <Upload className="h-6 w-6 text-muted-foreground mx-auto mb-2" />
          <p className="text-sm text-muted-foreground font-medium">Drop files here or click to upload</p>
          <p className="text-[11px] text-muted-foreground/70 mt-1">PDF, JPG, PNG supported</p>
        </div>
      )}
    </div>
  );
}
