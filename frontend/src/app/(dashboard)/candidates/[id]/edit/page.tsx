"use client"

import * as React from "react"
import Link from "next/link"
import { useParams, useRouter } from "next/navigation"
import { ChevronRight, Eye, FileText, Home, ImagePlus, Loader2, PencilLine, UploadCloud, Video, XCircle } from "lucide-react"
import { toast } from "sonner"

import { CandidateForm } from "@/components/candidates/candidate-form"
import { DocumentUpload } from "@/components/candidates/document-upload"
import { PageHeader } from "@/components/layout/page-header"
import { useCurrentUser } from "@/hooks/use-auth"
import { useCandidate, useUpdateCandidate, useUploadDocument } from "@/hooks/use-candidates"
import { CandidateStatus, Document } from "@/types"
import { CandidateInput } from "@/lib/validations"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"

export default function EditCandidatePage() {
  const params = useParams()
  const router = useRouter()
  const candidateID = String(params.id || "")
  const { user, isEthiopianAgent, isLoading: isRoleLoading } = useCurrentUser()
  const { data: candidate, isLoading: isCandidateLoading, error } = useCandidate(candidateID)
  const { mutateAsync: updateCandidate, isPending } = useUpdateCandidate(candidateID)
  const { mutate: uploadDocument, isPending: isUploadingDocument } = useUploadDocument(candidateID)

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
    age: candidate.age,
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
        age: data.age,
        experience_years: data.experience_years,
        skills: data.skills,
        languages: data.languages.map((item) => item.language),
      })
      router.push(`/candidates/${candidateID}`)
    } catch {}
  }

  const getDocument = (documentType: string) => candidate.documents.find((document) => document.document_type === documentType)

  return (
    <div className="space-y-6 pb-10">
      {breadcrumbs}

      <PageHeader
        heading="Edit Candidate"
        text="Update the candidate profile while it is still in a draft or available state, then replace any passport, photo, or video files from the same page."
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

        <CandidateForm
          initialData={initialData}
          onSubmit={handleSubmit}
          isLoading={isPending}
          showDocuments={false}
        />

        <div className="mt-8 border-t border-border/70 pt-8">
          <div className="mb-5 space-y-2">
            <h2 className="text-xl font-semibold tracking-tight">Replace documents</h2>
            <p className="text-sm text-muted-foreground">
              Uploading a new passport, photo, or video here makes that newest file the active one used by the profile and CV.
            </p>
          </div>

          <div className="grid gap-5 xl:grid-cols-3">
            <EditableDocumentCard
              key={getDocument("passport")?.id || "passport-empty"}
              title="Passport document"
              icon={<FileText className="h-5 w-5" />}
              currentDocument={getDocument("passport")}
              emptyLabel="No passport uploaded yet"
              description="Upload a new PDF, JPG, or PNG passport file to replace the current one."
              accept={{
                "application/pdf": [".pdf"],
                "image/jpeg": [".jpg", ".jpeg"],
                "image/png": [".png"],
              }}
              disabled={isUploadingDocument}
              onUpload={(file) => uploadDocument({ file, type: "passport" })}
            />

            <EditableDocumentCard
              key={getDocument("photo")?.id || "photo-empty"}
              title="Full body photo"
              icon={<ImagePlus className="h-5 w-5" />}
              currentDocument={getDocument("photo")}
              emptyLabel="No full-body photo uploaded yet"
              description="Upload a cleaner or newer full-body photo whenever you want to refresh the profile or CV."
              accept={{
                "image/jpeg": [".jpg", ".jpeg"],
                "image/png": [".png"],
              }}
              disabled={isUploadingDocument}
              onUpload={(file) => uploadDocument({ file, type: "photo" })}
            />

            <EditableDocumentCard
              key={getDocument("video")?.id || "video-empty"}
              title="Video interview"
              icon={<Video className="h-5 w-5" />}
              currentDocument={getDocument("video")}
              emptyLabel="No video interview uploaded yet"
              description="Optional, but you can replace it here if you recorded a better introduction."
              accept={{ "video/mp4": [".mp4"] }}
              disabled={isUploadingDocument}
              onUpload={(file) => uploadDocument({ file, type: "video" })}
            />
          </div>
        </div>
      </div>
    </div>
  )
}

function EditableDocumentCard({
  title,
  icon,
  currentDocument,
  emptyLabel,
  description,
  accept,
  disabled,
  onUpload,
}: {
  title: string
  icon: React.ReactNode
  currentDocument?: Document
  emptyLabel: string
  description: string
  accept: Record<string, string[]>
  disabled?: boolean
  onUpload: (file: File) => void
}) {
  const isImage = !!currentDocument?.file_url && /\.(png|jpe?g)$/i.test(currentDocument.file_url)

  return (
    <Card className="border-border/70 shadow-sm">
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-base">
          {icon}
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {currentDocument ? (
          <div className="space-y-3 rounded-2xl border border-border/70 bg-muted/20 p-4">
            {isImage ? (
              <img
                src={currentDocument.file_url}
                alt={currentDocument.file_name}
                className="h-40 w-full rounded-xl object-cover border border-border/60 bg-background"
              />
            ) : (
              <div className="flex h-40 w-full items-center justify-center rounded-xl border border-dashed border-border/70 bg-background text-muted-foreground">
                <div className="flex flex-col items-center gap-2 text-center">
                  <FileText className="h-8 w-8" />
                  <span className="text-sm font-medium">{currentDocument.file_name}</span>
                </div>
              </div>
            )}

            <div className="flex items-center justify-between gap-3">
              <div className="min-w-0">
                <p className="truncate text-sm font-semibold">{currentDocument.file_name}</p>
                <p className="text-xs text-muted-foreground">Newest upload is active now</p>
              </div>
              <Button size="sm" variant="outline" asChild>
                <a href={currentDocument.file_url} target="_blank" rel="noopener noreferrer">
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

        <div className="rounded-2xl border border-primary/15 bg-primary/5 p-4">
          <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-primary">
            <UploadCloud className="h-4 w-4" />
            Upload a replacement
          </div>
          <DocumentUpload
            documentType={title}
            title={`Replace ${title.toLowerCase()}`}
            description={description}
            accept={accept}
            maxSize={title === "Video interview" ? 52428800 : 10485760}
            mode="instant"
            disabled={disabled}
            onUpload={onUpload}
          />
        </div>
      </CardContent>
    </Card>
  )
}
