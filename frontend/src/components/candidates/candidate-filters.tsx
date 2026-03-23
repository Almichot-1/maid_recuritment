"use client"

import * as React from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { Search, X, Filter } from "lucide-react"

import { useCurrentUser } from "@/hooks/use-auth"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Badge } from "@/components/ui/badge"

const LANGUAGES_OPTIONS = ["Arabic", "English", "Amharic", "French", "Swahili"]
const ALL_VALUE = "all"
const EXPERIENCE_OPTIONS = [
  { label: "Any Experience", value: ALL_VALUE },
  { label: "1-2 years", value: "1-2" },
  { label: "3-5 years", value: "3-5" },
  { label: "5+ years", value: "5+" },
]

export function CandidateFilters() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { isEthiopianAgent } = useCurrentUser()

  const [search, setSearch] = React.useState(searchParams.get("search") || "")
  const [debouncedSearch, setDebouncedSearch] = React.useState(search)
  const minExperience = searchParams.get("min_experience")
  const maxExperience = searchParams.get("max_experience")
  const experienceValue = getExperienceFilterValue(minExperience, maxExperience)

  // Debounce search input
  React.useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search)
    }, 300)

    return () => clearTimeout(timer)
  }, [search])

  // Update URL when debounced search changes
  React.useEffect(() => {
    updateFilters({ search: debouncedSearch || undefined })
  }, [debouncedSearch])

  const updateFilters = (updates: Record<string, string | undefined>) => {
    const params = new URLSearchParams(searchParams.toString())
    
    Object.entries(updates).forEach(([key, value]) => {
      if (value) {
        params.set(key, value)
      } else {
        params.delete(key)
      }
    })

    // Reset to page 1 when filters change
    if (!updates.page) {
      params.set("page", "1")
    }

    router.push(`?${params.toString()}`)
  }

  const clearFilters = () => {
    setSearch("")
    router.push(window.location.pathname)
  }

  // Count active filters
  const activeFilterCount = React.useMemo(() => {
    let count = 0
    if (searchParams.get("search")) count++
    if (searchParams.get("status")) count++
    if (searchParams.get("min_age")) count++
    if (searchParams.get("max_age")) count++
    if (searchParams.get("languages")) count++
    if (searchParams.get("min_experience") || searchParams.get("max_experience")) count++
    return count
  }, [searchParams])

  // Status options based on role
  const statusOptions = isEthiopianAgent
    ? [
        { label: "All Statuses", value: ALL_VALUE },
        { label: "Available", value: "available" },
        { label: "In Progress", value: "in_progress" },
        { label: "Locked", value: "locked" },
        { label: "Under Review", value: "under_review" },
        { label: "Approved", value: "approved" },
        { label: "Rejected", value: "rejected" },
      ]
    : [
        { label: "All Statuses", value: ALL_VALUE },
        { label: "Available", value: "available" },
      ]

  return (
    <div className="space-y-4">
      <div className="flex flex-col lg:flex-row gap-3">
        {/* Search */}
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search by name..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9 pr-9 h-10 shadow-sm"
          />
          {search && (
            <button
              onClick={() => setSearch("")}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>

        {/* Status */}
        <Select
          value={searchParams.get("status") || ALL_VALUE}
          onValueChange={(value) => updateFilters({ status: value === ALL_VALUE ? undefined : value })}
        >
          <SelectTrigger className="w-full lg:w-[180px] h-10 shadow-sm">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            {statusOptions.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Age Range */}
        <div className="flex gap-2 items-center">
          <Input
            type="number"
            placeholder="Min age"
            min={18}
            max={65}
            value={searchParams.get("min_age") || ""}
            onChange={(e) => updateFilters({ min_age: e.target.value || undefined })}
            className="w-24 h-10 shadow-sm"
          />
          <span className="text-muted-foreground">-</span>
          <Input
            type="number"
            placeholder="Max age"
            min={18}
            max={65}
            value={searchParams.get("max_age") || ""}
            onChange={(e) => updateFilters({ max_age: e.target.value || undefined })}
            className="w-24 h-10 shadow-sm"
          />
        </div>

        {/* Languages */}
        <Select
          value={searchParams.get("languages") || ALL_VALUE}
          onValueChange={(value) => updateFilters({ languages: value === ALL_VALUE ? undefined : value })}
        >
          <SelectTrigger className="w-full lg:w-[160px] h-10 shadow-sm">
            <SelectValue placeholder="Language" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL_VALUE}>All Languages</SelectItem>
            {LANGUAGES_OPTIONS.map((lang) => (
              <SelectItem key={lang} value={lang}>
                {lang}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Experience */}
        <Select
          value={experienceValue}
          onValueChange={(value) => {
            const next =
              value === "1-2"
                ? { min_experience: "1", max_experience: "2" }
                : value === "3-5"
                  ? { min_experience: "3", max_experience: "5" }
                  : value === "5+"
                    ? { min_experience: "5", max_experience: undefined }
                    : { min_experience: undefined, max_experience: undefined }

            updateFilters(next)
          }}
        >
          <SelectTrigger className="w-full lg:w-[160px] h-10 shadow-sm">
            <SelectValue placeholder="Experience" />
          </SelectTrigger>
          <SelectContent>
            {EXPERIENCE_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {/* Clear Filters */}
        {activeFilterCount > 0 && (
          <Button
            variant="outline"
            size="sm"
            onClick={clearFilters}
            className="h-10 shadow-sm whitespace-nowrap"
          >
            <X className="mr-2 h-4 w-4" />
            Clear
            <Badge variant="secondary" className="ml-2 px-1.5 min-w-[20px] justify-center">
              {activeFilterCount}
            </Badge>
          </Button>
        )}
      </div>

      {/* Active Filters Display */}
      {activeFilterCount > 0 && (
        <div className="flex flex-wrap gap-2 items-center text-sm">
          <span className="text-muted-foreground flex items-center">
            <Filter className="h-3.5 w-3.5 mr-1.5" />
            Active filters:
          </span>
          {searchParams.get("search") && (
            <Badge variant="secondary" className="gap-1">
              Search: {searchParams.get("search")}
              <button
                onClick={() => {
                  setSearch("")
                  updateFilters({ search: undefined })
                }}
                className="hover:text-destructive"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          )}
          {searchParams.get("status") && (
            <Badge variant="secondary" className="gap-1">
              Status: {searchParams.get("status")}
              <button
                onClick={() => updateFilters({ status: undefined })}
                className="hover:text-destructive"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          )}
          {(searchParams.get("min_age") || searchParams.get("max_age")) && (
            <Badge variant="secondary" className="gap-1">
              Age: {searchParams.get("min_age") || "18"}-{searchParams.get("max_age") || "65"}
              <button
                onClick={() => updateFilters({ min_age: undefined, max_age: undefined })}
                className="hover:text-destructive"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          )}
          {searchParams.get("languages") && (
            <Badge variant="secondary" className="gap-1">
              Language: {searchParams.get("languages")}
              <button
                onClick={() => updateFilters({ languages: undefined })}
                className="hover:text-destructive"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          )}
          {(minExperience || maxExperience) && (
            <Badge variant="secondary" className="gap-1">
              Experience: {formatExperienceBadge(minExperience, maxExperience)}
              <button
                onClick={() => updateFilters({ min_experience: undefined, max_experience: undefined })}
                className="hover:text-destructive"
              >
                <X className="h-3 w-3" />
              </button>
            </Badge>
          )}
        </div>
      )}
    </div>
  )
}

function getExperienceFilterValue(minExperience: string | null, maxExperience: string | null) {
  if (minExperience === "1" && maxExperience === "2") {
    return "1-2"
  }
  if (minExperience === "3" && maxExperience === "5") {
    return "3-5"
  }
  if (minExperience === "5" && !maxExperience) {
    return "5+"
  }
  return ALL_VALUE
}

function formatExperienceBadge(minExperience: string | null, maxExperience: string | null) {
  if (minExperience && maxExperience) {
    return `${minExperience}-${maxExperience} years`
  }
  if (minExperience) {
    return `${minExperience}+ years`
  }
  if (maxExperience) {
    return `Up to ${maxExperience} years`
  }
  return "Any"
}
