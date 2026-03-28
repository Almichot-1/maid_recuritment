import Link from "next/link"
import { Clock3, ShieldCheck, Mail } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

interface PendingSearchParams {
  email?: string
  company_name?: string
  role?: string
}

interface PendingPageProps {
  searchParams?: Promise<PendingSearchParams>
}

export default async function RegistrationPendingPage({ searchParams }: PendingPageProps) {
  const resolvedSearchParams = await Promise.resolve(searchParams)
  const companyName = resolvedSearchParams?.company_name || "Your agency"
  const email = resolvedSearchParams?.email || "the email you registered with"
  const roleLabel =
    resolvedSearchParams?.role === "foreign_agent" ? "Foreign agency" : "Ethiopian agency"

  return (
    <Card className="w-full max-w-2xl border-slate-200 shadow-xl">
      <CardHeader className="space-y-4 pb-4 text-center">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-100 text-amber-700">
          <Clock3 className="h-8 w-8" />
        </div>
        <div className="space-y-2">
          <CardTitle className="text-3xl font-semibold tracking-tight">Your account is under review</CardTitle>
          <CardDescription className="text-base">
            We received the registration for <span className="font-medium text-foreground">{companyName}</span>.
            An admin needs to approve this {roleLabel.toLowerCase()} before anyone can sign in.
          </CardDescription>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
            <ShieldCheck className="mb-3 h-5 w-5 text-emerald-600" />
            <h2 className="font-medium text-slate-950">Manual verification</h2>
            <p className="mt-2 text-sm text-slate-600">
              The admin team checks each agency before platform access is activated.
            </p>
          </div>
          <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
            <Mail className="mb-3 h-5 w-5 text-sky-600" />
            <h2 className="font-medium text-slate-950">Approval email</h2>
            <p className="mt-2 text-sm text-slate-600">
              We&apos;ll notify <span className="font-medium">{email}</span> as soon as the account is approved.
            </p>
          </div>
          <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
            <Clock3 className="mb-3 h-5 w-5 text-amber-600" />
            <h2 className="font-medium text-slate-950">What to expect</h2>
            <p className="mt-2 text-sm text-slate-600">
              Until approval, login stays blocked and the agency dashboard remains unavailable.
            </p>
          </div>
        </div>

        <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-5 text-sm text-slate-600">
          If you need to update registration details, contact platform support before attempting another registration.
        </div>

        <div className="flex flex-col gap-3 sm:flex-row">
          <Link href="/login" className="flex-1">
            <Button className="w-full">Go To Login</Button>
          </Link>
          <Link href="/" className="flex-1">
            <Button variant="outline" className="w-full">Back To Homepage</Button>
          </Link>
        </div>
      </CardContent>
    </Card>
  )
}
