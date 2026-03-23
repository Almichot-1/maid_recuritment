"use client"

import { Candidate } from "@/types"
import { CandidateCard } from "./candidate-card"

interface CandidateGridProps {
  candidates: Candidate[]
}

export function CandidateGrid({ candidates }: CandidateGridProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 animate-in fade-in duration-300">
      {candidates.map((candidate) => (
        <CandidateCard key={candidate.id} candidate={candidate} />
      ))}
    </div>
  )
}
