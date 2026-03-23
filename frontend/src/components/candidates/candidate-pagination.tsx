"use client"

import * as React from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { ChevronLeft, ChevronRight } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

interface CandidatePaginationProps {
  currentPage: number
  pageSize: number
  total: number
}

export function CandidatePagination({
  currentPage,
  pageSize,
  total,
}: CandidatePaginationProps) {
  const router = useRouter()
  const searchParams = useSearchParams()

  const totalPages = Math.ceil(total / pageSize)
  const startItem = (currentPage - 1) * pageSize + 1
  const endItem = Math.min(currentPage * pageSize, total)

  const updatePage = (page: number) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set("page", page.toString())
    router.push(`?${params.toString()}`)
  }

  const updatePageSize = (size: string) => {
    const params = new URLSearchParams(searchParams.toString())
    params.set("page_size", size)
    params.set("page", "1") // Reset to first page
    router.push(`?${params.toString()}`)
  }

  if (total === 0) {
    return null
  }

  return (
    <div className="flex flex-col sm:flex-row items-center justify-between gap-4 py-4">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <span>
          Showing {startItem} to {endItem} of {total} candidates
        </span>
      </div>

      <div className="flex items-center gap-6">
        {/* Page Size Selector */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">Per page:</span>
          <Select value={pageSize.toString()} onValueChange={updatePageSize}>
            <SelectTrigger className="w-[70px] h-9">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="12">12</SelectItem>
              <SelectItem value="24">24</SelectItem>
              <SelectItem value="48">48</SelectItem>
              <SelectItem value="96">96</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Page Navigation */}
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => updatePage(currentPage - 1)}
            disabled={currentPage === 1}
          >
            <ChevronLeft className="h-4 w-4" />
            <span className="sr-only">Previous page</span>
          </Button>

          <div className="flex items-center gap-1">
            {/* Show first page */}
            {currentPage > 2 && (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => updatePage(1)}
                  className="w-9"
                >
                  1
                </Button>
                {currentPage > 3 && (
                  <span className="text-muted-foreground px-1">...</span>
                )}
              </>
            )}

            {/* Show previous page */}
            {currentPage > 1 && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => updatePage(currentPage - 1)}
                className="w-9"
              >
                {currentPage - 1}
              </Button>
            )}

            {/* Current page */}
            <Button variant="default" size="sm" className="w-9">
              {currentPage}
            </Button>

            {/* Show next page */}
            {currentPage < totalPages && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => updatePage(currentPage + 1)}
                className="w-9"
              >
                {currentPage + 1}
              </Button>
            )}

            {/* Show last page */}
            {currentPage < totalPages - 1 && (
              <>
                {currentPage < totalPages - 2 && (
                  <span className="text-muted-foreground px-1">...</span>
                )}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => updatePage(totalPages)}
                  className="w-9"
                >
                  {totalPages}
                </Button>
              </>
            )}
          </div>

          <Button
            variant="outline"
            size="sm"
            onClick={() => updatePage(currentPage + 1)}
            disabled={currentPage === totalPages}
          >
            <ChevronRight className="h-4 w-4" />
            <span className="sr-only">Next page</span>
          </Button>
        </div>
      </div>
    </div>
  )
}
