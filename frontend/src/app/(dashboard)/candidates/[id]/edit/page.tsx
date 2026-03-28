"use client"

import * as React from "react"
import Link from "next/link"
import { useParams, useRouter } from "next/navigation"
import { ChevronRight, Eye, FileText, Home, ImagePlus, Loader2, PencilLine, Video, XCircle } from "lucide-react"
import { toast } from "sonner"

import { CandidateForm } from "@/components/candidates/candidate-form"
import { PageHeader } from "@/components/layout/page-header"
import { useCurrentUser } from "@/hooks/use-auth"
import { uploadCandidateDocumentFile, useCandidate, useUpdateCandidate } from "@/hooks/use-candidates"
import { CandidateStatus, Document } from "@/types"
import { CandidateInput } from "@/lib/validations"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"

type PendingDocuments = {
  passport: File | null
  photo: File | null
  video: File | null
}

export default function EditCandidatePage() {
  const params = useParams()
  const router = useRouter()
  const candidateID = String(params.id || "")
  const { user, isEthiopianAgent, isLoading: isRoleLoading } = useCurrentUser()
  const { data: candidate, isLoading: isCandidateLoading, error } = useCandidate(candidateID)
  const { mutateAsync: updateCandidate, isPending } = useUpdateCandidate(candidateID)
  const [pendingDocuments, setPendingDocuments] = React.useState<PendingDocuments>({
    passport: null,
    photo: null,
    video: null,
  })
  const [isUploadingDocuments, setIsUploadingDocuments] = React.useState(false)

  const isOwner = !!candidate && !!user && candidate.created_by === user.id
  const canEdit = !!candidate && (candidate.status === CandidateStatus.DRAFT || candidate.status === CandidateStatus.AVAILABLE)

  React.useEffect(() => {
    if (!isRoleLoading && !isEthiopianAgent) {
      toast.error("Only Ethiopian agents can edit candidates")
      router.replace("/candidates")
    }
  }, [isEthiopianAgent, isRoleLoading, router])

  const breadcrumbs = (
    <nav className="mb-6 flex items-center text-sm font-medium text-muted-foreground">
      <Link href="/dashboard" className="flex items-center transition-all hover:text-primary">
        <Home className="mr-1.5 h-4 w-4" />
        Dashboard
      </Link>
      <ChevronRight className="mx-1 h-4 w-4 opacity-50" />
      <Link href="/candidates" className="transition-colors hover:text-primary">
        Candidates
      </Link>
      <ChevronRight className="mx-1 h-4 w-4 opacity-50" />
      <Link href={`/candidates/${candidateID}`} className="transition-colors hover:text-primary">
        Candidate
      </Link>
      <ChevronRight className="mx-1 h-4 w-4 opacity-50" />
      <span className="font-semibold text-foreground">Edit</span>
    </nav>
  )

  if (isRoleLoading || isCandidateLoading) {
    return (
      <div className="flex h-[50vh] items-center justify-center">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
      </div>
    )
  }

  if (error || !candidate || !isOwner) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center space-y-4 text-center">
        <div className="flex h-20 w-20 items-center justify-center rounded-full bg-destructive/10">
          <XCircle className="h-10 w-10 text-destructive" />
        </div>
        <h2 className="text-2xl font-bold">Candidate not available for editing</h2>
        <p className="max-w-md text-muted-foreground">
          This candidate either does not exist, belongs to another agency, or is no longer editable in the current workflow state.
        </p>
      </div>
    )
  }

  if (!canEdit) {
    return (
      <div className="space-y-6">
        {breadcrumbs}
        <PageHeader
          heading="Editing locked by workflow"
          text="This candidate is already in a protected stage of the recruitment process. Use the detail page to track progress instead."
        />
      </div>
    )
  }

  const initialData = {
    full_name: candidate.full_name,
    nationality: candidate.nationality || "",
    date_of_birth: candidate.date_of_birth || "",
    age: candidate.age,
    place_of_birth: candidate.place_of_birth || "",
    religion: candidate.religion || "",
    marital_status: candidate.marital_status || "",
    children_count: candidate.children_count,
    education_level: candidate.education_level || "",
    experience_years: candidate.experience_years,
    skills: candidate.skills,
    languages: candidate.languages.length
      ? candidate.languages.map((language) => ({ language, proficiency: "Intermediate" }))
      : [{ language: "English", proficiency: "Basic" }],
  }

  const handleSubmit = async (data: CandidateInput) => {
    try {
      await updateCandidate({
        full_name: data.full_name,
        nationality: data.nationality,
        date_of_birth: data.date_of_birth || undefined,
        age: data.age,
        place_of_birth: data.place_of_birth,
        religion: data.religion,
        marital_status: data.marital_status,
        children_count: data.children_count,
        education_level: data.education_level,
        experience_years: data.experience_years,
        skills: data.skills,
        languages: data.languages.map((item) => item.language),
      })

      const queuedDocuments = Object.entries(pendingDocuments).filter(([, file]) => !!file) as Array<
        [keyof PendingDocuments, File]
      >

      if (queuedDocuments.length > 0) {
        setIsUploadingDocuments(true)
        for (const [documentType, file] of queuedDocuments) {
          await uploadCandidateDocumentFile(candidateID, {
            file,
            type: documentType,
          })
        }
        toast.success("Candidate details and replacement files saved successfully.")
      }

      router.push(`/candidates/${candidateID}`)
    } catch {
      setIsUploadingDocuments(false)
      return
    } finally {
      setIsUploadingDocuments(false)
    }
  }

  const getDocument = (documentType: string) => candidate.documents.find((document) => document.document_type === documentType)

  return (
    <div className="space-y-6 pb-10">
      {breadcrumbs}

      <PageHeader
        heading="Edit Candidate"
        text="Update the candidate profile, replace any passport, photo, or video files in the same workflow, and let a new passport image refresh the key fields automatically."
      />

      <div className="rounded-2xl border border-border/70 bg-card p-3 shadow-sm md:p-6">
        <div className="mb-6 flex items-start gap-4 rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
          <PencilLine className="mt-0.5 h-5 w-5 shrink-0" />
          <div>
            <p className="font-semibold">Safe edit mode</p>
            <p className="mt-1 text-amber-800/90">
              You can edit profile details here and replace the current passport, full-body photo, or video without leaving this page.
            </p>
          </div>
        </div>

        <div className="mb-8 grid gap-5 xl:grid-cols-3">
          <CurrentDocumentCard
            title="Current passport"
            icon={<FileText className="h-5 w-5" />}
            document={getDocument("passport")}
            emptyLabel="No passport uploaded yet"
          />
          <CurrentDocumentCard
            title="Current full-body photo"
            icon={<ImagePlus className="h-5 w-5" />}
            document={getDocument("photo")}
            emptyLabel="No photo uploaded yet"
          />
          <CurrentDocumentCard
            title="Current video interview"
            icon={<Video className="h-5 w-5" />}
            document={getDocument("video")}
            emptyLabel="No video interview uploaded yet"
          />
        </div>

        <CandidateForm
          candidateId={candidateID}
          initialData={initialData}
          onSubmit={handleSubmit}
          isLoading={isPending || isUploadingDocuments}
          showDocuments
          onDocumentChange={(documentType, file) => {
            setPendingDocuments((current) => ({
              ...current,
              [documentType]: file,
            }))
          }}
        />
      </div>
    </div>
  )
}

function CurrentDocumentCard({
  title,
  icon,
  document,
  emptyLabel,
}: {
  title: string
  icon: React.ReactNode
  document?: Document
  emptyLabel: string
}) {
  const isImage = !!document?.file_url && /\.(png|jpe?g)$/i.test(document.file_url)

  return (
    <Card className="border-border/70 shadow-sm">
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-base">
          {icon}
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {document ? (
          <div className="space-y-3 rounded-2xl border border-border/70 bg-muted/20 p-4">
            {isImage ? (
              <img
                src={document.file_url}
                alt={document.file_name}
                className="h-40 w-full rounded-xl object-cover border border-border/60 bg-background"
              />
            ) : (
              <div className="flex h-40 w-full items-center justify-center rounded-xl border border-dashed border-border/70 bg-background text-muted-foreground">
                <div className="flex flex-col items-center gap-2 text-center">
                  <FileText className="h-8 w-8" />
                  <span className="text-sm font-medium">{document.file_name}</span>
                </div>
              </div>
            )}

            <div className="flex items-center justify-between gap-3">
              <div className="min-w-0">
                <p className="truncate text-sm font-semibold">{document.file_name}</p>
                <p className="text-xs text-muted-foreground">Newest upload is active now</p>
              </div>
              <Button size="sm" variant="outline" asChild>
                <a href={document.file_url} target="_blank" rel="noopener noreferrer">
                  <Eye className="mr-2 h-4 w-4" />
                  View
                </a>
              </Button>
            </div>
          </div>
        ) : (
          <div className="rounded-2xl border border-dashed border-border/70 bg-muted/10 px-4 py-6 text-center text-sm text-muted-foreground">
            {emptyLabel}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
