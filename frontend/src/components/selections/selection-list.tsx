"use client"

import * as React from "react"
import { Loader2, Search } from "lucide-react"

import { Selection } from "@/types"
import { SelectionCard } from "./selection-card"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

interface SelectionListProps {
  selections: Selection[]
  isLoading?: boolean
  sortBy?: "newest" | "expiring"
  onSortByChange?: (sortBy: "newest" | "expiring") => void
}

export function SelectionList({ selections, isLoading, sortBy: propSortBy = "newest", onSortByChange }: SelectionListProps) {
  const [searchQuery, setSearchQuery] = React.useState("")
  const [sortBy, setSortBy] = React.useState<"newest" | "expiring">(propSortBy)

  // Update local sortBy when prop changes
  React.useEffect(() => {
    setSortBy(propSortBy)
  }, [propSortBy])

  const handleSortByChange = (value: "newest" | "expiring") => {
    setSortBy(value)
    onSortByChange?.(value)
  }

  // Filter by search query
  const filteredSelections = React.useMemo(() => {
    let filtered = selections

    if (searchQuery) {
      filtered = filtered.filter((selection) =>
        selection.candidate?.full_name
          ?.toLowerCase()
          .includes(searchQuery.toLowerCase())
      )
    }

    return filtered
  }, [selections, searchQuery])

  // Sort selections (API-based)
  const sortedSelections = React.useMemo(() => {
    // Selections are already sorted by the API based on sortBy parameter
    // This just filters for display - sorting is handled server-side
    return [...filteredSelections]
  }, [filteredSelections])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search by candidate name..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        <Select value={sortBy} onValueChange={(value) => handleSortByChange(value as "newest" | "expiring")}>
          <SelectTrigger className="w-full sm:w-[180px]">
            <SelectValue placeholder="Sort by" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="newest">Newest First</SelectItem>
            <SelectItem value="expiring">Expiring Soon</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Selection Cards */}
      {sortedSelections.length > 0 ? (
        <div className="space-y-3">
          {sortedSelections.map((selection) => (
            <SelectionCard key={selection.id} selection={selection} />
          ))}
        </div>
      ) : (
        <div className="text-center py-12 text-muted-foreground">
          {searchQuery ? (
            <p>No selections found for {searchQuery}</p>
          ) : (
            <p>No selections found</p>
          )}
        </div>
      )}
    </div>
  )
}
