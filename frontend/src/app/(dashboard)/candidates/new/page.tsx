"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import { ChevronRight, Home, Loader2, UserPlus } from "lucide-react"
import Link from "next/link"

import { useCurrentUser } from "@/hooks/use-auth"
import { publishCandidateById, uploadCandidateDocumentFile, useCreateCandidate } from "@/hooks/use-candidates"
import { CandidateForm } from "@/components/candidates/candidate-form"
import { SubmissionProgressOverlay } from "@/components/candidates/submission-progress-overlay"
import { CandidateInput } from "@/lib/validations"
import { PageHeader } from "@/components/layout/page-header"

type PendingDocuments = {
  passport: File | null
  photo: File | null
  video: File | null
}

type SubmissionStage = "idle" | "creating" | "uploading" | "publishing" | "finalizing"
type SubmissionIntent = "default" | "create_another"

const DOCUMENT_LABELS: Record<keyof PendingDocuments, string> = {
  passport: "passport document",
  photo: "full-body photo",
  video: "video interview",
}

export default function NewCandidatePage() {
  const router = useRouter()
  const { isEthiopianAgent, isLoading: isRoleLoading } = useCurrentUser()
  const { mutateAsync: createCandidate, isPending } = useCreateCandidate()
  const [pendingDocuments, setPendingDocuments] = React.useState<PendingDocuments>({
    passport: null,
    photo: null,
    video: null,
  })
  const [submissionStage, setSubmissionStage] = React.useState<SubmissionStage>("idle")
  const [submissionIntent, setSubmissionIntent] = React.useState<SubmissionIntent>("default")
  const [activeUpload, setActiveUpload] = React.useState<keyof PendingDocuments | null>(null)
  const [uploadProgress, setUploadProgress] = React.useState<Record<keyof PendingDocuments, number>>({
    passport: 0,
    photo: 0,
    video: 0,
  })

  // Guard protecting Foreign Agents from accessing this component
  React.useEffect(() => {
    if (!isRoleLoading && !isEthiopianAgent) {
      toast.error("Only Ethiopian agents can create candidates")
      router.push("/candidates")
    }
  }, [isEthiopianAgent, isRoleLoading, router])

  if (isRoleLoading || !isEthiopianAgent) {
    return (
      <div className="flex h-[50vh] w-full items-center justify-center animate-in fade-in duration-500">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
      </div>
    )
  }

  const breadcrumbs = (
    <nav className="flex items-center text-sm font-medium text-muted-foreground mb-6">
      <Link href="/dashboard" className="transition-all hover:text-primary flex items-center bg-muted/50 px-2 py-1 rounded-md">
        <Home className="mr-1.5 h-4 w-4" />
        Dashboard
      </Link>
      <ChevronRight className="h-4 w-4 mx-1 opacity-50" />
      <Link href="/candidates" className="transition-colors hover:text-primary flex items-center">
        Candidates
      </Link>
      <ChevronRight className="h-4 w-4 mx-1 opacity-50" />
      <span className="text-foreground font-semibold flex items-center">
        <UserPlus className="h-4 w-4 mr-1.5" />
        Add New
      </span>
    </nav>
  )

  const handleDocumentChange = (documentType: keyof PendingDocuments, file: File | null) => {
    setPendingDocuments((current) => ({
      ...current,
      [documentType]: file,
    }))

    setUploadProgress((current) => ({
      ...current,
      [documentType]: file ? 0 : 0,
    }))
  }

  const handleSubmit = async (
    data: CandidateInput,
    { submitter = "default" }: { submitter?: SubmissionIntent } = {}
  ) => {
    const candidateData = {
      ...data,
      languages: data.languages.map((language) => language.language),
    }
    const queuedDocuments = (Object.entries(pendingDocuments).filter(([, file]) => !!file) as Array<[keyof PendingDocuments, File]>)

    try {
      setSubmissionIntent(submitter)
      setSubmissionStage("creating")
      const response = await createCandidate(candidateData)
      const candidateID = response.candidate.id

      if (queuedDocuments.length > 0) {
        setSubmissionStage("uploading")

        for (const [documentType, file] of queuedDocuments) {
          setActiveUpload(documentType)
          await uploadCandidateDocumentFile(candidateID, {
            file,
            type: documentType,
            onProgress: (progress) => {
              setUploadProgress((current) => ({
                ...current,
                [documentType]: progress,
              }))
            },
          })
          setUploadProgress((current) => ({
            ...current,
            [documentType]: 100,
          }))
        }

        setActiveUpload(null)
        toast.success("Candidate and documents saved successfully")
      } else {
        toast.info("Candidate created without documents")
      }

      if (submitter === "create_another") {
        setSubmissionStage("publishing")
        await publishCandidateById(candidateID)
      }

      setSubmissionStage("finalizing")
      await new Promise((resolve) => setTimeout(resolve, 450))
      if (submitter === "create_another") {
        setPendingDocuments({
          passport: null,
          photo: null,
          video: null,
        })
        setUploadProgress({
          passport: 0,
          photo: 0,
          video: 0,
        })
        setActiveUpload(null)
        setSubmissionStage("idle")
        toast.success("Candidate saved, published, and the form is ready for another profile.")
        return { resetForm: true }
      }

      router.push(`/candidates/${candidateID}`)
    } catch {
      setSubmissionStage("idle")
      setActiveUpload(null)
      setSubmissionIntent("default")
    }
  }

  const overlaySteps = [
    {
      label: "Create candidate profile",
      description:
        submissionStage === "creating"
          ? "Saving the candidate record and preparing a private workspace for this agency."
          : submissionStage === "idle"
            ? "The profile details will be saved first."
            : "Candidate profile saved successfully.",
      status:
        submissionStage === "creating"
          ? "active"
          : submissionStage === "idle"
            ? "pending"
            : "complete",
    },
    {
      label: "Upload selected documents",
      description:
        Object.values(pendingDocuments).some(Boolean)
          ? activeUpload
            ? `Uploading ${DOCUMENT_LABELS[activeUpload]}${uploadProgress[activeUpload] ? ` (${uploadProgress[activeUpload]}%)` : ""}.`
            : submissionStage === "publishing" || submissionStage === "finalizing"
              ? "All queued documents have been uploaded."
              : "Passport, photo, and optional video will be attached right after the profile is created."
          : "No additional files were queued for this submission.",
      status:
        !Object.values(pendingDocuments).some(Boolean)
          ? submissionStage === "idle"
            ? "pending"
            : "complete"
          : submissionStage === "uploading"
            ? "active"
            : submissionStage === "publishing" || submissionStage === "finalizing"
              ? "complete"
              : submissionStage === "idle" || submissionStage === "creating"
                ? "pending"
                : "complete",
    },
    {
      label: submissionIntent === "create_another" ? "Publish and prepare next form" : "Open candidate workspace",
      description:
        submissionStage === "publishing"
          ? "Publishing this candidate so it is immediately available before the form resets."
          : submissionStage === "finalizing"
          ? submissionIntent === "create_another"
            ? "Resetting the form so you can add the next candidate right away."
            : "Finishing the handoff and opening the candidate detail page."
          : submissionIntent === "create_another"
            ? "Once everything is saved, this candidate will be published and the form will clear for the next profile."
            : "Once everything is ready, you will land on the candidate page automatically.",
      status:
        submissionStage === "publishing" || submissionStage === "finalizing"
          ? "active"
          : submissionStage === "idle"
            ? "pending"
            : "pending",
    },
  ] as const

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 pb-10">
      <SubmissionProgressOverlay
        open={submissionStage !== "idle"}
        title={submissionIntent === "create_another" ? "Saving candidate" : "Creating candidate"}
        description={
          submissionIntent === "create_another"
            ? "The system is saving the current profile, processing the files, and preparing the next blank form."
            : "The system is saving the profile and processing the files you added."
        }
        steps={overlaySteps.map((step) => ({ ...step }))}
        footer="Please keep this tab open while the candidate profile finishes saving."
      />

      {breadcrumbs}
      <PageHeader 
        heading="Add Candidate" 
        text="Create a candidate profile, reuse your saved agency branding, and upload the supporting documents used throughout the recruitment flow."
      />
      
      <div className="relative mx-auto max-w-[1480px] rounded-2xl border border-border/60 bg-card p-2 shadow-sm md:p-6">
        <CandidateForm 
          onSubmit={handleSubmit} 
          isLoading={isPending || submissionStage !== "idle"} 
          onDocumentChange={handleDocumentChange}
        />
      </div>
    </div>
  )
}
