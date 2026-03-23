"use client"

import * as React from "react"
import { useSearchParams } from "next/navigation"
import { Plus, Grid3x3, List } from "lucide-react"
import Link from "next/link"

import { useCurrentUser } from "@/hooks/use-auth"
import { useCandidates } from "@/hooks/use-candidates"
import { usePairingContext } from "@/hooks/use-pairings"
import { PageHeader } from "@/components/layout/page-header"
import { Button } from "@/components/ui/button"
import { CandidateFilters } from "@/components/candidates/candidate-filters"
import { CandidateGrid } from "@/components/candidates/candidate-grid"
import { CandidateTable } from "@/components/candidates/candidate-table"
import { CandidatePagination } from "@/components/candidates/candidate-pagination"
import { Skeleton } from "@/components/ui/skeleton"

export default function CandidatesPage() {
  const searchParams = useSearchParams()
  const { isEthiopianAgent } = useCurrentUser()
  const { activeWorkspace } = usePairingContext()

  // View preference (grid or table)
  const [viewMode, setViewMode] = React.useState<"grid" | "table">("grid")

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
    }
  }, [searchParams])

  const { data, isLoading } = useCandidates(filters)

  const pageHeader = (
    <PageHeader
      heading={isEthiopianAgent ? "My Candidate Library" : "Browse Candidates"}
      text={
        isEthiopianAgent
          ? "Create and update your agency's candidates here, then share the right profiles with partner agencies."
          : `These are the candidates shared with your current partner workspace${activeWorkspace ? ` by ${activeWorkspace.partner_agency.company_name || activeWorkspace.partner_agency.full_name}` : ""}.`
      }
      action={
        isEthiopianAgent ? (
          <Button asChild size="lg" className="shadow-md">
            <Link href="/candidates/new">
              <Plus className="mr-2 h-5 w-5" />
              Add Candidate
            </Link>
          </Button>
        ) : undefined
      }
    />
  )

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      {pageHeader}

      <div className="space-y-4">
        {/* Filters and View Toggle */}
        <div className="flex flex-col lg:flex-row gap-4 items-start lg:items-center justify-between">
          <div className="flex-1 w-full">
            <CandidateFilters />
          </div>
          
          <div className="flex items-center gap-2 border rounded-lg p-1 bg-background shadow-sm">
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

        {/* Content */}
        {isLoading ? (
          <LoadingState viewMode={viewMode} />
        ) : !data || data.data.length === 0 ? (
          <EmptyState isEthiopianAgent={isEthiopianAgent} />
        ) : (
          <>
            {viewMode === "grid" ? (
              <CandidateGrid candidates={data.data} />
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
    </div>
  )
}

function LoadingState({ viewMode }: { viewMode: "grid" | "table" }) {
  if (viewMode === "grid") {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="border rounded-xl p-4 space-y-4 bg-card shadow-sm">
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
