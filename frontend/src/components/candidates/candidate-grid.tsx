"use client"

import { Candidate } from "@/types"
import { CandidateCard } from "./candidate-card"

interface CandidateGridProps {
  candidates: Candidate[]
  selectable?: boolean
  selectedIds?: Set<string>
  onSelectionChange?: (candidateId: string, selected: boolean) => void
}

export function CandidateGrid({ candidates, selectable = false, selectedIds = new Set(), onSelectionChange }: CandidateGridProps) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 animate-in fade-in duration-300">
      {candidates.map((candidate) => (
        <CandidateCard 
          key={candidate.id} 
          candidate={candidate}
          selectable={selectable}
          selected={selectedIds.has(candidate.id)}
          onSelectionChange={onSelectionChange}
        />
      ))}
    </div>
  )
}
