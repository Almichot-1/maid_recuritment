"use client";

import * as React from "react";
import { MessageSquare } from "lucide-react";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";

interface NotesTabProps {
  candidateId: string;
  canEdit: boolean;
}

export function NotesTab({ canEdit }: NotesTabProps) {
  const [note, setNote] = React.useState("");

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col items-center justify-center gap-3 py-8 text-center rounded-2xl bg-muted/30 border border-dashed border-border/60">
        <div className="h-10 w-10 rounded-full bg-muted flex items-center justify-center">
          <MessageSquare className="h-4.5 w-4.5 text-muted-foreground" />
        </div>
        <p className="text-sm text-muted-foreground">No notes yet for this applicant.</p>
      </div>

      {canEdit && (
        <div className="flex flex-col gap-2">
          <Textarea
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="Write a note about this applicant…"
            className="min-h-[100px] rounded-2xl text-sm resize-none border-border/60"
          />
          <Button
            size="sm"
            disabled={!note.trim()}
            className="self-end rounded-xl"
          >
            Add Note
          </Button>
        </div>
      )}
    </div>
  );
}
