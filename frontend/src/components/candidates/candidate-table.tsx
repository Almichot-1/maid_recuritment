"use client"

import * as React from "react"
import Link from "next/link"
import { Eye, PencilLine, Share2, Sparkles } from "lucide-react"

import { useCurrentUser } from "@/hooks/use-auth"
import { Candidate, CandidateStatus } from "@/types"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { CandidateShareDialog } from "@/components/candidates/candidate-share-dialog"
import { SelectCandidateDialog } from "@/components/selections/select-candidate-dialog"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

interface CandidateTableProps {
  candidates: Candidate[]
}

export function CandidateTable({ candidates }: CandidateTableProps) {
  const { user, isEthiopianAgent, isForeignAgent } = useCurrentUser()
  const [candidateToSelect, setCandidateToSelect] = React.useState<Candidate | null>(null)
  const [candidateToShare, setCandidateToShare] = React.useState<Candidate | null>(null)

  return (
    <>
      <Card className="overflow-hidden rounded-2xl border bg-card shadow-sm">
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Candidate</TableHead>
                <TableHead className="hidden sm:table-cell">Age</TableHead>
                <TableHead className="hidden sm:table-cell">Experience</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="hidden md:table-cell">Languages</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {candidates.map((candidate) => {
                const isOwner = isEthiopianAgent && candidate.created_by === user?.id
                const canEdit = isOwner && (candidate.status === CandidateStatus.DRAFT || candidate.status === CandidateStatus.AVAILABLE)
                const canSelect = isForeignAgent && candidate.status === CandidateStatus.AVAILABLE

                return (
                  <TableRow key={candidate.id}>
                    <TableCell>
                      <div className="space-y-1">
                        <p className="font-medium">{candidate.full_name}</p>
                        <p className="text-xs text-muted-foreground">
                          {candidate.skills.slice(0, 2).join(", ") || "Profile ready for review"}
                        </p>
                      </div>
                    </TableCell>
                    <TableCell className="hidden sm:table-cell">{candidate.age ?? "N/A"}</TableCell>
                    <TableCell className="hidden sm:table-cell">{candidate.experience_years ?? 0} yrs</TableCell>
                    <TableCell>
                      <StatusBadge status={candidate.status} />
                    </TableCell>
                    <TableCell className="hidden md:table-cell">{candidate.languages.slice(0, 2).join(", ") || "None"}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1 sm:gap-2 flex-wrap">
                        {canSelect ? (
                          <Button size="sm" onClick={() => setCandidateToSelect(candidate)} className="bg-green-600 hover:bg-green-700">
                            <Sparkles className="sm:mr-2 h-4 w-4" />
                            <span className="hidden sm:inline">Select</span>
                          </Button>
                        ) : null}

                        {canEdit ? (
                          <Button size="sm" variant="outline" asChild>
                            <Link href={`/candidates/${candidate.id}/edit`}>
                              <PencilLine className="sm:mr-2 h-4 w-4" />
                              <span className="hidden sm:inline">Edit</span>
                            </Link>
                          </Button>
                        ) : null}

                        {isOwner ? (
                          <Button size="sm" variant="secondary" onClick={() => setCandidateToShare(candidate)}>
                            <Share2 className="sm:mr-2 h-4 w-4" />
                            <span className="hidden sm:inline">Share</span>
                          </Button>
                        ) : null}

                        <Button variant="outline" size="sm" asChild>
                          <Link href={`/candidates/${candidate.id}`}>
                            <Eye className="sm:mr-2 h-4 w-4" />
                            <span className="hidden sm:inline">{isOwner ? "Open" : "View"}</span>
                          </Link>
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </div>
      </Card>

      {candidateToSelect ? (
        <SelectCandidateDialog
          candidate={candidateToSelect}
          open={!!candidateToSelect}
          onOpenChange={(open) => {
            if (!open) {
              setCandidateToSelect(null)
            }
          }}
        />
      ) : null}

      {candidateToShare ? (
        <CandidateShareDialog
          candidate={candidateToShare}
          open={!!candidateToShare}
          onOpenChange={(open) => {
            if (!open) {
              setCandidateToShare(null)
            }
          }}
        />
      ) : null}
    </>
  )
}

function StatusBadge({ status }: { status: CandidateStatus }) {
  const classNameByStatus: Record<CandidateStatus, string> = {
    [CandidateStatus.DRAFT]: "bg-slate-500 text-white hover:bg-slate-500",
    [CandidateStatus.AVAILABLE]: "bg-green-500 text-white hover:bg-green-500",
    [CandidateStatus.LOCKED]: "bg-amber-500 text-white hover:bg-amber-500",
    [CandidateStatus.UNDER_REVIEW]: "bg-blue-500 text-white hover:bg-blue-500",
    [CandidateStatus.APPROVED]: "bg-emerald-600 text-white hover:bg-emerald-600",
    [CandidateStatus.IN_PROGRESS]: "bg-indigo-500 text-white hover:bg-indigo-500",
    [CandidateStatus.COMPLETED]: "bg-zinc-700 text-white hover:bg-zinc-700",
    [CandidateStatus.REJECTED]: "bg-red-500 text-white hover:bg-red-500",
  }

  return <Badge className={classNameByStatus[status]}>{status.replaceAll("_", " ")}</Badge>
}
