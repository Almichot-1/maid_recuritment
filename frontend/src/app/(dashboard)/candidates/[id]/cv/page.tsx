"use client"

import * as React from "react"
import Link from "next/link"
import { useParams } from "next/navigation"
import {
  AlertCircle,
  ArrowLeft,
  CheckCircle2,
  ChevronRight,
  Download,
  ExternalLink,
  FileText,
  Home,
  Loader2,
  Sparkles,
} from "lucide-react"

import { useCurrentUser } from "@/hooks/use-auth"
import { useAgencyBranding } from "@/hooks/use-agency-branding"
import { useCandidate, useGenerateCV } from "@/hooks/use-candidates"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"

export default function CandidateCVPage() {
  const params = useParams()
  const candidateId = String(params.id || "")
  const { user, isEthiopianAgent } = useCurrentUser()
  const { hasLogo, logoDataURL, isLoaded: isBrandingLoaded } = useAgencyBranding()
  const { data: candidate, isLoading, error } = useCandidate(candidateId)
  const { mutate: generateCV, isPending: isGeneratingCV } = useGenerateCV(candidateId)
  const [hasStartedPreparation, setHasStartedPreparation] = React.useState(false)

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
          <h1 className="text-2xl font-bold">CV page not available</h1>
          <p className="max-w-md text-muted-foreground">
            The candidate could not be loaded, or you no longer have access to this profile.
          </p>
        </div>
        <Button asChild>
          <Link href="/candidates">Back to candidates</Link>
        </Button>
      </div>
    )
  }

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
      <Link href={`/candidates/${candidate.id}`} className="transition-colors hover:text-primary">
        {candidate.full_name}
      </Link>
      <ChevronRight className="mx-1 h-4 w-4 opacity-50" />
      <span className="font-semibold text-foreground">CV Download</span>
    </nav>
  )

  const statusBadge = candidate.cv_pdf_url ? (
    <Badge className="bg-emerald-500 text-white hover:bg-emerald-500">Ready</Badge>
  ) : !isBrandingLoaded && isEthiopianAgent ? (
    <Badge className="bg-slate-700 text-white hover:bg-slate-700">Loading branding</Badge>
  ) : isGeneratingCV ? (
    <Badge className="bg-sky-500 text-white hover:bg-sky-500">Preparing</Badge>
  ) : missingRequiredDocuments.length > 0 ? (
    <Badge className="bg-amber-500 text-white hover:bg-amber-500">Missing documents</Badge>
  ) : (
    <Badge className="bg-slate-700 text-white hover:bg-slate-700">Waiting</Badge>
  )

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 pb-10">
      {breadcrumbs}

      <Card className="overflow-hidden border-0 bg-[radial-gradient(circle_at_top_left,_rgba(14,165,233,0.16),_transparent_30%),radial-gradient(circle_at_top_right,_rgba(16,185,129,0.18),_transparent_24%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.96))] text-white shadow-xl">
        <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_280px]">
          <div className="space-y-4">
            <Badge className="w-fit rounded-full border-0 bg-white/15 px-3 py-1 text-[11px] uppercase tracking-[0.24em] text-sky-200 hover:bg-white/15">
              Candidate CV
            </Badge>
            <div className="space-y-2">
              <h1 className="text-3xl font-semibold tracking-tight">Download {candidate.full_name}&apos;s CV</h1>
              <p className="max-w-2xl text-sm text-slate-200/90">
                This page keeps CV preparation and download in one place, so the candidate detail page stays focused on profile data and workflow updates.
              </p>
            </div>
            <div className="flex flex-wrap gap-3">
              <Button variant="outline" className="border-white/20 bg-white/10 text-white hover:bg-white/15 hover:text-white" asChild>
                <Link href={`/candidates/${candidate.id}`}>
                  <ArrowLeft className="mr-2 h-4 w-4" />
                  Back to candidate
                </Link>
              </Button>
              {candidate.cv_pdf_url ? (
                <>
                  <Button className="bg-white text-slate-950 hover:bg-slate-100" asChild>
                    <a href={candidate.cv_pdf_url} download>
                      <Download className="mr-2 h-4 w-4" />
                      Download PDF
                    </a>
                  </Button>
                  <Button variant="outline" className="border-white/20 bg-white/10 text-white hover:bg-white/15 hover:text-white" asChild>
                    <a href={candidate.cv_pdf_url} target="_blank" rel="noopener noreferrer">
                      <ExternalLink className="mr-2 h-4 w-4" />
                      Open in new tab
                    </a>
                  </Button>
                  {isEthiopianAgent ? (
                    <Button
                      variant="outline"
                      className="border-white/20 bg-white/10 text-white hover:bg-white/15 hover:text-white"
                      onClick={triggerCVBuild}
                      disabled={isGeneratingCV}
                    >
                      {isGeneratingCV ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Sparkles className="mr-2 h-4 w-4" />}
                      Refresh Layout
                    </Button>
                  ) : null}
                </>
              ) : null}
            </div>
          </div>

          <div className="rounded-3xl border border-white/10 bg-white/10 p-5 backdrop-blur">
            <p className="text-xs uppercase tracking-[0.24em] text-sky-200">Current state</p>
            <div className="mt-4 space-y-4">
              <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                <p className="text-xs uppercase tracking-[0.2em] text-slate-300">Candidate</p>
                <p className="mt-2 text-2xl font-semibold text-white">{candidate.full_name}</p>
              </div>
              <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                <p className="text-xs uppercase tracking-[0.2em] text-slate-300">Status</p>
                <div className="mt-2">{statusBadge}</div>
              </div>
              <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                <p className="text-xs uppercase tracking-[0.2em] text-slate-300">Branding</p>
                <p className="mt-2 text-sm font-medium text-white">
                  {hasLogo ? "Your saved logo will be used in the PDF header." : "Using the text-only agency header."}
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {candidate.cv_pdf_url ? (
        <div className="grid gap-6 xl:grid-cols-[minmax(0,1.35fr)_320px]">
          <Card className="overflow-hidden shadow-sm">
            <CardHeader>
              <CardTitle>PDF Preview</CardTitle>
            </CardHeader>
            <CardContent>
              <iframe
                src={candidate.cv_pdf_url}
                title={`${candidate.full_name} CV`}
                className="h-[75vh] w-full rounded-2xl border bg-background"
              />
            </CardContent>
          </Card>

          <Card className="shadow-sm">
            <CardHeader>
              <CardTitle>Download options</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="rounded-2xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-900">
                The final CV is ready. You can download it directly or open it in a separate browser tab.
              </div>
              {isEthiopianAgent ? (
                <div className="rounded-2xl border border-sky-200 bg-sky-50 p-4 text-sm text-sky-900">
                  Open this page again or use Refresh Layout after changing the logo so the PDF picks up the latest branding and colors.
                </div>
              ) : null}
              <Button className="w-full" asChild>
                <a href={candidate.cv_pdf_url} download>
                  <Download className="mr-2 h-4 w-4" />
                  Download CV
                </a>
              </Button>
              {isEthiopianAgent ? (
                <Button variant="outline" className="w-full" onClick={triggerCVBuild} disabled={isGeneratingCV}>
                  {isGeneratingCV ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Sparkles className="mr-2 h-4 w-4" />}
                  Refresh Layout
                </Button>
              ) : null}
              <Button variant="outline" className="w-full" asChild>
                <a href={candidate.cv_pdf_url} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="mr-2 h-4 w-4" />
                  Open in new tab
                </a>
              </Button>
            </CardContent>
          </Card>
        </div>
      ) : missingRequiredDocuments.length > 0 ? (
        <Card className="shadow-sm">
          <CardHeader>
            <CardTitle>CV is blocked until required files are ready</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
              Upload {missingRequiredDocuments.join(" and ")} on the candidate detail page first. After that, this page will prepare the CV automatically.
            </div>
            <Button asChild>
              <Link href={`/candidates/${candidate.id}`}>
                <FileText className="mr-2 h-4 w-4" />
                Go back to candidate files
              </Link>
            </Button>
          </CardContent>
        </Card>
      ) : isGeneratingCV ? (
        <Card className="shadow-sm">
          <CardContent className="py-16">
            <div className="mx-auto flex max-w-lg flex-col items-center space-y-4 text-center">
              <div className="flex h-20 w-20 items-center justify-center rounded-full bg-sky-100 text-sky-700">
                <Loader2 className="h-10 w-10 animate-spin" />
              </div>
              <div className="space-y-2">
                <h2 className="text-2xl font-semibold tracking-tight">Preparing the CV</h2>
                <p className="text-muted-foreground">
                  The PDF is being generated now with the latest layout, colors, and agency branding. Stay on this page for a moment and it will become downloadable automatically.
                </p>
              </div>
              <div className="w-full rounded-2xl border border-border/70 bg-muted/20 p-4 text-left">
                <div className="flex items-center gap-3">
                  <CheckCircle2 className="h-5 w-5 text-emerald-600" />
                  <span className="text-sm font-medium">Profile details verified</span>
                </div>
                <div className="mt-3 flex items-center gap-3">
                  <Loader2 className="h-5 w-5 animate-spin text-sky-600" />
                  <span className="text-sm font-medium">Building the final PDF document</span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      ) : (
        <Card className="shadow-sm">
          <CardContent className="py-16">
            <div className="mx-auto flex max-w-lg flex-col items-center space-y-4 text-center">
              <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted">
                <FileText className="h-10 w-10 text-muted-foreground" />
              </div>
              <div className="space-y-2">
                <h2 className="text-2xl font-semibold tracking-tight">CV not prepared yet</h2>
                <p className="text-muted-foreground">
                  The Ethiopian agency has not generated the PDF yet. Once it is ready, this page will show the preview and download actions.
                </p>
              </div>
              <Button variant="outline" asChild>
                <Link href={`/candidates/${candidate.id}`}>Back to candidate</Link>
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
