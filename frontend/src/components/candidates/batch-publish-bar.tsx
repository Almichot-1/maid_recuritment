"use client"

import * as React from "react"
import { Loader2, Rocket, X } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"

interface BatchPublishBarProps {
  selectedCount: number
  onPublish: () => void
  onClear: () => void
  isPublishing?: boolean
}

export function BatchPublishBar({
  selectedCount,
  onPublish,
  onClear,
  isPublishing = false,
}: BatchPublishBarProps) {
  if (selectedCount === 0) return null

  return (
    <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 animate-in slide-in-from-bottom-4 duration-300">
      <div className="bg-gradient-to-r from-emerald-500 to-teal-500 text-white rounded-full shadow-lg px-6 py-4 flex items-center gap-4">
        <Badge variant="secondary" className="bg-white/20 text-white border-0 text-lg px-3 py-1">
          {selectedCount}
        </Badge>
        
        <span className="font-medium">
          {selectedCount === 1 ? "candidate" : "candidates"} selected
        </span>

        <div className="flex gap-2">
          <Button
            size="sm"
            variant="ghost"
            className="text-white hover:bg-white/20"
            onClick={onClear}
            disabled={isPublishing}
          >
            <X className="h-4 w-4 mr-1" />
            Clear
          </Button>

          <Button
            size="sm"
            className="bg-white text-emerald-600 hover:bg-white/90"
            onClick={onPublish}
            disabled={isPublishing}
          >
            {isPublishing ? (
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <Rocket className="h-4 w-4 mr-2" />
            )}
            Publish All
          </Button>
        </div>
      </div>
    </div>
  )
}
