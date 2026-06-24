"use client"

import * as React from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { Plus, Grid3x3, List, CheckSquare, Download, Pencil, RefreshCw, Send, Trash2 } from "lucide-react"
import Link from "next/link"
import { toast } from "sonner"

import { useCurrentUser } from "@/hooks/use-auth"
import { useCandidates, useBatchPublishCandidates, useBatchDeleteCandidates } from "@/hooks/use-candidates"
import { usePairingContext } from "@/hooks/use-pairings"
import { PageHeader } from "@/components/layout/page-header"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { CandidateFilters } from "@/components/candidates/candidate-filters"
import { CandidateGrid } from "@/components/candidates/candidate-grid"
import { CandidateTable } from "@/components/candidates/candidate-table"
import { CandidatePagination } from "@/components/candidates/candidate-pagination"
import { BatchPublishBar } from "@/components/candidates/batch-publish-bar"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { BulkCvActionsDialog } from "@/components/candidates/bulk-cv-actions"
import { BulkSetOverrideDialog } from "@/components/candidates/bulk-set-override-dialog"
import { BulkPublishDialog } from "@/components/candidates/bulk-publish-dialog"

export default function CandidatesPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { isEthiopianAgent } = useCurrentUser()
  const { activeWorkspace, context } = usePairingContext()

  // View preference (grid or table)
  const [viewMode, setViewMode] = React.useState<"grid" | "table">("grid")

  // Sort options
  const SORT_OPTIONS = [
    { value: "created_at|desc", label: "Newest First" },
    { value: "created_at|asc", label: "Oldest First" },
    { value: "full_name|asc", label: "Name A-Z" },
    { value: "full_name|desc", label: "Name Z-A" },
    { value: "age|asc", label: "Youngest First" },
    { value: "age|desc", label: "Oldest First" },
    { value: "experience_years|desc", label: "Most Experienced" },
    { value: "experience_years|asc", label: "Least Experienced" },
    { value: "religion|asc", label: "Religion (Muslim First)" },
  ] as const

  const currentSortBy = searchParams.get("sort_by") || "created_at"
  const currentSortOrder = searchParams.get("sort_order") || "desc"
  const currentSort = `${currentSortBy}|${currentSortOrder}`

  const handleSortChange = (value: string) => {
    const [sortBy, sortOrder] = value.split("|")
    const params = new URLSearchParams(searchParams.toString())
    params.set("sort_by", sortBy)
    params.set("sort_order", sortOrder)
    params.set("page", "1")
    router.push(`?${params.toString()}`)
  }
  
  // Batch selection state
  const [selectionMode, setSelectionMode] = React.useState(false)
  const [selectedIds, setSelectedIds] = React.useState<Set<string>>(new Set())
  const [bulkCvDialogOpen, setBulkCvDialogOpen] = React.useState(false)
  const [bulkCvAction, setBulkCvAction] = React.useState<"regenerate" | "download">("regenerate")
  const [bulkOverrideDialogOpen, setBulkOverrideDialogOpen] = React.useState(false)
  const [bulkPublishDialogOpen, setBulkPublishDialogOpen] = React.useState(false)
  const [bulkDeleteDialogOpen, setBulkDeleteDialogOpen] = React.useState(false)

  const { mutateAsync: batchPublish, isPending: isPublishingBatch } = useBatchPublishCandidates()
  const { mutateAsync: batchDelete, isPending: isDeletingBatch } = useBatchDeleteCandidates()

  // Load view preference from localStorage
  React.useEffect(() => {
    const savedView = localStorage.getItem("candidates_view_mode")
    if (savedView === "grid" || savedView === "table") {
      setViewMode(savedView)
    }
  }, [])

  // Save view preference
  const handleViewChange = (mode: "grid" | "table") => {
    setViewMode(mode)
    localStorage.setItem("candidates_view_mode", mode)
  }

  // Build filters from URL params
  const filters = React.useMemo(() => {
    return {
      search: searchParams.get("search") || undefined,
      status: searchParams.get("status") || undefined,
      min_age: searchParams.get("min_age") ? Number(searchParams.get("min_age")) : undefined,
      max_age: searchParams.get("max_age") ? Number(searchParams.get("max_age")) : undefined,
      min_experience: searchParams.get("min_experience") ? Number(searchParams.get("min_experience")) : undefined,
      max_experience: searchParams.get("max_experience") ? Number(searchParams.get("max_experience")) : undefined,
      languages: searchParams.get("languages") || undefined,
      page: searchParams.get("page") ? Number(searchParams.get("page")) : 1,
      page_size: searchParams.get("page_size") ? Number(searchParams.get("page_size")) : 12,
      sort_by: searchParams.get("sort_by") || undefined,
      sort_order: searchParams.get("sort_order") || undefined,
    }
  }, [searchParams])

  const { data, isLoading } = useCandidates(filters)

  const handleSelectionChange = (candidateId: string, selected: boolean) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (selected) {
        next.add(candidateId)
      } else {
        next.delete(candidateId)
      }
      return next
    })
  }

  const handleSelectAll = () => {
    if (!data?.data) return
    setSelectedIds(new Set(data.data.map((c) => c.id)))
  }

  const handleBatchDelete = async () => {
    if (selectedIds.size === 0) return
    await batchDelete(Array.from(selectedIds))
    handleClearSelection()
  }

  const handleClearSelection = () => {
    setSelectedIds(new Set())
    setSelectionMode(false)
  }

  const handleBatchPublish = async () => {
    if (selectedIds.size === 0) return

    const workspaceIds = context?.workspaces.map((w) => w.id) || []
    
    if (workspaceIds.length === 0) {
      toast.error("No partner workspaces available")
      return
    }

    // Show bulk publish dialog for partner selection
    if (workspaceIds.length > 0) {
      setBulkPublishDialogOpen(true)
      return
    }

    // Otherwise, publish directly
    try {
      await batchPublish({
        candidateIds: Array.from(selectedIds),
        pairingIds: workspaceIds,
      })
      handleClearSelection()
    } catch {
      toast.error("Failed to batch publish candidates")
    }
  }

  const pageHeader = (
    <PageHeader
      heading={isEthiopianAgent ? "My Candidate Library" : "Browse Candidates"}
      text={
        isEthiopianAgent
          ? "Create and update your agency's candidates here, then share the right profiles with partner agencies."
          : `These are the candidates shared with your current partner workspace${activeWorkspace ? ` by ${activeWorkspace.partner_agency.company_name || activeWorkspace.partner_agency.full_name}` : ""}.`
      }
        action={
          <div className="flex flex-wrap gap-2">
            {data && data.data.length > 0 && !selectionMode && (
              <Button
                variant="outline"
                size="lg"
                onClick={() => setSelectionMode(true)}
              >
                <CheckSquare className="mr-2 h-5 w-5" />
                <span className="hidden sm:inline">
                  {isEthiopianAgent ? "Batch Publish" : "Batch Select"}
                </span>
              </Button>
            )}
            {selectionMode && (
              <div className="flex flex-wrap items-center gap-2">
                <Button variant="outline" size="lg" onClick={handleClearSelection}>
                  Cancel
                </Button>
                {selectedIds.size !== data?.data.length && (
                  <Button variant="ghost" size="sm" onClick={handleSelectAll}>
                    Select All ({data?.data.length || 0})
                  </Button>
                )}
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="lg">
                      Bulk CV
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuLabel>CV Operations</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={() => { setBulkCvAction("download"); setBulkCvDialogOpen(true); }}>
                      <Download className="mr-2 h-4 w-4" />
                      Download CVs (ZIP)
                    </DropdownMenuItem>
                    {isEthiopianAgent && (
                      <>
                        <DropdownMenuItem onClick={() => { setBulkCvAction("regenerate"); setBulkCvDialogOpen(true); }}>
                          <RefreshCw className="mr-2 h-4 w-4" />
                          Regenerate CVs
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => setBulkOverrideDialogOpen(true)}>
                          <Pencil className="mr-2 h-4 w-4" />
                          Set Partner Overrides
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={() => setBulkPublishDialogOpen(true)}>
                          <Send className="mr-2 h-4 w-4" />
                          Publish to Partners
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => setBulkDeleteDialogOpen(true)}
                        >
                          <Trash2 className="mr-2 h-4 w-4" />
                          Delete Selected
                        </DropdownMenuItem>
                      </>
                    )}
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            )}
            {!selectionMode && isEthiopianAgent && (
              <Button asChild size="lg" className="shadow-md">
                <Link href="/candidates/new">
                  <Plus className="sm:mr-2 h-5 w-5" />
                  <span className="hidden sm:inline">Add Candidate</span>
                </Link>
              </Button>
            )}
          </div>
        }
    />
  )

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      {pageHeader}

      <div className="space-y-4">
        {/* Filters and View Toggle */}
        <div className="flex flex-col lg:flex-row gap-4 items-start lg:items-center justify-between">
          <div className="flex-1 w-full min-w-0">
            <CandidateFilters />
          </div>
          
          <div className="flex items-center gap-2 w-full sm:w-auto">
            <Select value={currentSort} onValueChange={handleSortChange}>
              <SelectTrigger className="flex-1 sm:w-[180px] h-10 shadow-sm">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {SORT_OPTIONS.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <div className="flex items-center gap-2 border rounded-lg p-1 bg-background shadow-sm shrink-0">
              <Button
                variant={viewMode === "grid" ? "default" : "ghost"}
                size="sm"
                onClick={() => handleViewChange("grid")}
                className="h-8"
              >
                <Grid3x3 className="h-4 w-4" />
              </Button>
              <Button
                variant={viewMode === "table" ? "default" : "ghost"}
                size="sm"
                onClick={() => handleViewChange("table")}
                className="h-8"
              >
                <List className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>

        {/* Content */}
        {isLoading ? (
          <LoadingState viewMode={viewMode} />
        ) : !data || data.data.length === 0 ? (
          <EmptyState isEthiopianAgent={isEthiopianAgent} />
        ) : (
          <>
            {viewMode === "grid" ? (
              <CandidateGrid 
                candidates={data.data}
                selectable={selectionMode}
                selectedIds={selectedIds}
                onSelectionChange={handleSelectionChange}
              />
            ) : (
              <CandidateTable candidates={data.data} />
            )}
            
            <CandidatePagination
              currentPage={filters.page}
              pageSize={filters.page_size}
              total={data.meta.total}
            />
          </>
        )}
      </div>

      {/* Batch Publish Bar */}
      {isEthiopianAgent && selectionMode && (
        <BatchPublishBar
          selectedCount={selectedIds.size}
          onPublish={handleBatchPublish}
          onClear={handleClearSelection}
          isPublishing={isPublishingBatch}
        />
      )}

      <BulkCvActionsDialog
        open={bulkCvDialogOpen}
        onOpenChange={setBulkCvDialogOpen}
        candidateIds={Array.from(selectedIds)}
        workspaces={context?.workspaces || []}
        defaultAction={bulkCvAction}
      />
      <BulkSetOverrideDialog
        open={bulkOverrideDialogOpen}
        onOpenChange={setBulkOverrideDialogOpen}
        candidateIds={Array.from(selectedIds)}
        workspaces={context?.workspaces || []}
      />
      <BulkPublishDialog
        open={bulkPublishDialogOpen}
        onOpenChange={setBulkPublishDialogOpen}
        candidateIds={Array.from(selectedIds)}
        workspaces={context?.workspaces || []}
      />

      <Dialog open={bulkDeleteDialogOpen} onOpenChange={setBulkDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete {selectedIds.size} candidate(s)?</DialogTitle>
            <DialogDescription>
              This action cannot be undone. The candidates and all their associated data will be permanently deleted.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="gap-2">
            <Button variant="outline" onClick={() => setBulkDeleteDialogOpen(false)} disabled={isDeletingBatch}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              disabled={isDeletingBatch}
              onClick={handleBatchDelete}
            >
              {isDeletingBatch ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

function LoadingState({ viewMode }: { viewMode: "grid" | "table" }) {
  if (viewMode === "grid") {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {Array.from({ length: 8 }).map((_, i) => (
          <div key={i} className="border rounded-xl p-3 space-y-3 bg-card shadow-sm">
            <Skeleton className="w-full aspect-square rounded-lg" />
            <Skeleton className="h-6 w-3/4" />
            <Skeleton className="h-4 w-1/2" />
            <div className="flex gap-2">
              <Skeleton className="h-6 w-16" />
              <Skeleton className="h-6 w-16" />
              <Skeleton className="h-6 w-16" />
            </div>
            <Skeleton className="h-10 w-full" />
          </div>
        ))}
      </div>
    )
  }

  return (
    <div className="border rounded-xl bg-card shadow-sm overflow-hidden">
      <div className="p-4 space-y-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="flex items-center gap-4">
            <Skeleton className="h-12 w-12 rounded-full" />
            <Skeleton className="h-6 flex-1" />
            <Skeleton className="h-6 w-20" />
            <Skeleton className="h-6 w-24" />
            <Skeleton className="h-8 w-20" />
          </div>
        ))}
      </div>
    </div>
  )
}

function EmptyState({ isEthiopianAgent }: { isEthiopianAgent: boolean }) {
  return (
    <div className="flex flex-col items-center justify-center py-20 px-4 text-center border rounded-xl bg-card/50 backdrop-blur-sm">
      <div className="w-24 h-24 rounded-full bg-muted flex items-center justify-center mb-6">
        <svg
          className="w-12 h-12 text-muted-foreground"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.5}
            d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
          />
        </svg>
      </div>
      <h3 className="text-xl font-semibold mb-2">
        {isEthiopianAgent ? "No candidates yet" : "No candidates found"}
      </h3>
      <p className="text-muted-foreground mb-6 max-w-md">
        {isEthiopianAgent
          ? "Get started by adding your first candidate to the system"
          : "No candidates match your current filters. Try adjusting your search criteria"}
      </p>
      {isEthiopianAgent && (
        <Button asChild size="lg">
          <Link href="/candidates/new">
            <Plus className="mr-2 h-5 w-5" />
            Add Your First Candidate
          </Link>
        </Button>
      )}
    </div>
  )
}
