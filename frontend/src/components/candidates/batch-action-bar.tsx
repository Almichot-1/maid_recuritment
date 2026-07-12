"use client"

import * as React from "react"
import {
  Lock,
  Unlock,
  Share2,
  Send,
  Download,
  RefreshCw,
  Pencil,
  Trash2,
  X,
  Loader2,
  ChevronDown,
  MessageCircle,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuLabel,
} from "@/components/ui/dropdown-menu"

interface BatchActionBarProps {
  selectedCount: number
  totalCount: number
  isEthiopianAgent: boolean
  onClear: () => void
  onSelectAll: () => void
  onShare: () => void
  onPublish: () => void
  onBulkCvDownload: () => void
  onBulkCvRegenerate: () => void
  onBulkOverride: () => void
  onDelete: () => void
  onLock?: () => void
  onUnlock?: () => void
  onWhatsappShare?: () => void
  isPublishing?: boolean
  isDeleting?: boolean
}

export function BatchActionBar({
  selectedCount,
  totalCount,
  isEthiopianAgent,
  onClear,
  onSelectAll,
  onShare,
  onPublish,
  onBulkCvDownload,
  onBulkCvRegenerate,
  onBulkOverride,
  onDelete,
  onLock,
  onUnlock,
  onWhatsappShare,
  isPublishing = false,
  isDeleting = false,
}: BatchActionBarProps) {
  if (selectedCount === 0) return null

  const isAllSelected = selectedCount === totalCount
  const isForeignAgent = !isEthiopianAgent

  return (
    <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 animate-in slide-in-from-bottom-4 duration-300">
      <div className="flex items-center gap-3 rounded-2xl border bg-background text-foreground px-5 py-4 shadow-lg">
        <Badge variant="secondary" className="text-sm px-3 py-1">
          {selectedCount} candidate{selectedCount !== 1 ? "s" : ""} selected
        </Badge>

        {!isAllSelected && (
          <button
            onClick={onSelectAll}
            className="text-sm font-medium text-primary underline-offset-4 hover:underline"
          >
            Select All
          </button>
        )}

        <div className="h-6 w-px bg-border" />

        <div className="flex items-center gap-2">
          {/* Foreign Agent batch actions */}
          {isForeignAgent && onLock && (
            <Button variant="outline" size="sm" onClick={onLock}>
              <Lock className="h-4 w-4" />
              Hold
            </Button>
          )}
          {isForeignAgent && onUnlock && (
            <Button variant="outline" size="sm" onClick={onUnlock}>
              <Unlock className="h-4 w-4" />
              Release
            </Button>
          )}
          {isForeignAgent && onWhatsappShare && (
            <Button variant="outline" size="sm" onClick={onWhatsappShare}>
              <MessageCircle className="h-4 w-4" />
              WhatsApp
            </Button>
          )}

          {/* Ethiopian Agent batch actions */}
          {isEthiopianAgent && (
            <Button
              variant="outline"
              size="sm"
              onClick={onShare}
            >
              <Share2 className="h-4 w-4" />
              Share
            </Button>
          )}

          {isEthiopianAgent && (
            <Button
              variant="outline"
              size="sm"
              onClick={onPublish}
              disabled={isPublishing}
            >
              {isPublishing ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Send className="h-4 w-4" />
              )}
              Publish
            </Button>
          )}

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm">
                <ChevronDown className="h-4 w-4" />
                Bulk CV
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
              <DropdownMenuLabel>Bulk CV Actions</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={onBulkCvDownload}>
                <Download className="h-4 w-4" />
                Download CVs (ZIP)
              </DropdownMenuItem>
              {isEthiopianAgent && (
                <DropdownMenuItem onClick={onBulkCvRegenerate}>
                  <RefreshCw className="h-4 w-4" />
                  Regenerate CVs
                </DropdownMenuItem>
              )}
              {isEthiopianAgent && (
                <DropdownMenuItem onClick={onBulkOverride}>
                  <Pencil className="h-4 w-4" />
                  Set Partner Overrides
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>

          {isEthiopianAgent && (
            <Button
              variant="destructive"
              size="sm"
              onClick={onDelete}
              disabled={isDeleting}
            >
              {isDeleting ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Trash2 className="h-4 w-4" />
              )}
              Delete
            </Button>
          )}
        </div>

        <div className="h-6 w-px bg-border" />

        <Button
          variant="ghost"
          size="icon"
          onClick={onClear}
          className="h-8 w-8 text-muted-foreground hover:text-foreground"
        >
          <X className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
