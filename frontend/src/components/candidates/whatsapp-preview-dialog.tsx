"use client"

import * as React from "react"
import { MessageCircle } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Textarea } from "@/components/ui/textarea"
import { shareOnWhatsApp, WhatsAppCandidate } from "@/lib/whatsapp"

interface WhatsAppPreviewDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  candidates: WhatsAppCandidate[]
}

export function WhatsAppPreviewDialog({ open, onOpenChange, candidates }: WhatsAppPreviewDialogProps) {
  const [message, setMessage] = React.useState("")

  React.useEffect(() => {
    if (open && candidates.length > 0) {
      const parts: string[] = []
      candidates.forEach((c, i) => {
        parts.push(`${i + 1}. ${c.full_name}${c.age != null ? ` - ${c.age}yrs` : ""}${c.experience_years != null ? `, ${c.experience_years}yrs exp` : ""}${c.cv_pdf_url ? `\n   CV: ${c.cv_pdf_url}` : ""}`)
      })
      setMessage(`*Candidates:*\n${parts.join("\n")}`)
    }
  }, [open, candidates])

  const handleSend = () => {
    shareOnWhatsApp(message)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Share via WhatsApp</DialogTitle>
          <DialogDescription>
            {candidates.length > 0
              ? `Review the message below for ${candidates.length} candidate${candidates.length !== 1 ? "s" : ""}. You can edit before sending.`
              : "Prepare your message"}
          </DialogDescription>
        </DialogHeader>

        <div className="py-4">
          <Textarea
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            rows={10}
            className="w-full font-mono text-sm"
          />
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSend} className="bg-green-600 hover:bg-green-700">
            <MessageCircle className="mr-2 h-4 w-4" />
            Open WhatsApp
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
