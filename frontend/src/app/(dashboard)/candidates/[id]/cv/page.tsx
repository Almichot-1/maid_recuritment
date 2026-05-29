"use client"

import * as React from "react"
import Link from "next/link"
import { useParams } from "next/navigation"
import {
  AlertCircle,
  ArrowLeft,
  Download,
  ExternalLink,
  FileText,
  Loader2,
  RefreshCw,
} from "lucide-react"
import { toast } from "sonner"

import { useCurrentUser } from "@/hooks/use-auth"
import { useAgencyBranding } from "@/hooks/use-agency-branding"
import { downloadCandidateCVFile, useCandidate, useGenerateCV } from "@/hooks/use-candidates"
import { PageHeader } from "@/components/layout/page-header"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"

export default function CandidateCVPage() {
  const params = useParams()
  const candidateId = String(params.id || "")
  const { user, isEthiopianAgent } = useCurrentUser()
  const { logoDataURL, isLoaded: isBrandingLoaded } = useAgencyBranding()
  const { data: candidate, isLoading, error } = useCandidate(candidateId)
  const { mutate: generateCV, isPending: isGeneratingCV } = useGenerateCV(candidateId)
  const [hasStartedPreparation, setHasStartedPreparation] = React.useState(false)
  const [isDownloading, setIsDownloading] = React.useState(false)

  const missingRequiredDocuments = React.useMemo(() => {
    if (!candidate) {
      return []
    }

    const documentTypes = new Set(candidate.documents.map((document) => document.document_type))
    return ["passport", "photo"].filter((documentType) => !documentTypes.has(documentType))
  }, [candidate])

  const canPrepareCV = isEthiopianAgent && missingRequiredDocuments.length === 0 && isBrandingLoaded
  const brandingPayload = React.useMemo(
    () => ({
      branding_logo_data_url: logoDataURL || undefined,
      company_name: user?.company_name || undefined,
    }),
    [logoDataURL, user?.company_name]
  )

  const triggerCVBuild = React.useCallback(() => {
    setHasStartedPreparation(true)
    generateCV(brandingPayload)
  }, [brandingPayload, generateCV])

  const handleDownload = React.useCallback(async () => {
    if (!candidate?.cv_pdf_url) {
      return
    }

    try {
      setIsDownloading(true)
      await downloadCandidateCVFile(candidate.id, candidate.full_name)
    } catch {
      toast.error("Failed to download CV")
    } finally {
      setIsDownloading(false)
    }
  }, [candidate])

  React.useEffect(() => {
    if (!candidate || !canPrepareCV || hasStartedPreparation || isGeneratingCV || !isEthiopianAgent) {
      return
    }

    triggerCVBuild()
  }, [candidate, canPrepareCV, hasStartedPreparation, isEthiopianAgent, isGeneratingCV, triggerCVBuild])

  if (isLoading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
      </div>
    )
  }

  if (error || !candidate) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4 text-center">
        <div className="flex h-20 w-20 items-center justify-center rounded-full bg-destructive/10">
          <AlertCircle className="h-10 w-10 text-destructive" />
        </div>
        <div className="space-y-2">
          <h1 className="text-2xl font-bold">CV not available</h1>
          <p className="max-w-md text-muted-foreground">
            This candidate could not be loaded or you no longer have access.
          </p>
        </div>
        <Button asChild>
          <Link href="/candidates">Back to candidates</Link>
        </Button>
      </div>
    )
  }

  const statusLabel = candidate.cv_pdf_url
    ? "Ready"
    : !isBrandingLoaded && isEthiopianAgent
      ? "Loading branding"
      : isGeneratingCV
        ? "Generating"
        : missingRequiredDocuments.length > 0
          ? "Missing files"
          : "Not ready"

  const statusVariant = candidate.cv_pdf_url
    ? "default"
    : missingRequiredDocuments.length > 0
      ? "secondary"
      : "outline"

  return (
    <div className="space-y-6 pb-10">
      <PageHeader
        heading={`${candidate.full_name} — CV`}
        text="Download or refresh the agency-branded PDF."
        action={
          <Button variant="outline" asChild>
            <Link href={`/candidates/${candidate.id}`}>
              <ArrowLeft className="mr-2 h-4 w-4" />
              Candidate
            </Link>
          </Button>
        }
      />

      <div className="flex flex-wrap items-center gap-3">
        <Badge variant={statusVariant}>{statusLabel}</Badge>
        {candidate.cv_pdf_url ? (
          <>
            <Button onClick={handleDownload} disabled={isDownloading}>
              <Download className="mr-2 h-4 w-4" />
              {isDownloading ? "Downloading…" : "Download PDF"}
            </Button>
            <Button variant="outline" asChild>
              <a href={candidate.cv_pdf_url} target="_blank" rel="noopener noreferrer">
                <ExternalLink className="mr-2 h-4 w-4" />
                Open
              </a>
            </Button>
            {isEthiopianAgent ? (
              <Button variant="outline" onClick={triggerCVBuild} disabled={isGeneratingCV}>
                {isGeneratingCV ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <RefreshCw className="mr-2 h-4 w-4" />
                )}
                Regenerate
              </Button>
            ) : null}
          </>
        ) : null}
      </div>

      {candidate.cv_pdf_url ? (
        <Card>
          <CardHeader>
            <CardTitle>Preview</CardTitle>
          </CardHeader>
          <CardContent>
            <iframe
              src={candidate.cv_pdf_url}
              title={`${candidate.full_name} CV`}
              className="h-[75vh] w-full rounded-lg border border-border bg-background"
            />
          </CardContent>
        </Card>
      ) : missingRequiredDocuments.length > 0 ? (
        <Card>
          <CardContent className="space-y-4 py-8">
            <p className="text-sm text-muted-foreground">
              Upload {missingRequiredDocuments.join(" and ")} on the candidate profile first.
            </p>
            <Button asChild>
              <Link href={`/candidates/${candidate.id}`}>
                <FileText className="mr-2 h-4 w-4" />
                Go to candidate
              </Link>
            </Button>
          </CardContent>
        </Card>
      ) : isGeneratingCV ? (
        <Card>
          <CardContent className="flex flex-col items-center gap-3 py-16 text-center">
            <Loader2 className="h-10 w-10 animate-spin text-primary" />
            <p className="text-sm text-muted-foreground">Generating PDF…</p>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent className="space-y-4 py-8 text-center">
            <FileText className="mx-auto h-10 w-10 text-muted-foreground" />
            <p className="text-sm text-muted-foreground">
              The CV has not been generated yet. It will appear here when ready.
            </p>
            <Button variant="outline" asChild>
              <Link href={`/candidates/${candidate.id}`}>Back to candidate</Link>
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
