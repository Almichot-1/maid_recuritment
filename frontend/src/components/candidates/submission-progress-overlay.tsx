"use client"

import { Loader2 } from "lucide-react"

import { Card, CardContent } from "@/components/ui/card"

interface SubmissionProgressOverlayProps {
  open: boolean
  title: string
  description?: string
}

export function SubmissionProgressOverlay({
  open,
  title,
  description,
}: SubmissionProgressOverlayProps) {
  if (!open) {
    return null
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 px-4 backdrop-blur-sm">
      <Card className="w-full max-w-md shadow-lg">
        <CardContent className="flex flex-col items-center gap-4 py-10 text-center">
          <Loader2 className="h-10 w-10 animate-spin text-primary" />
          <div className="space-y-1">
            <h2 className="text-lg font-semibold text-foreground">{title}</h2>
            {description ? (
              <p className="text-sm text-muted-foreground">{description}</p>
            ) : null}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
