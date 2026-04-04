import { PendingClient } from "./pending-client"

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
  const email = resolvedSearchParams?.email || "the email you registered with"
  const companyName = resolvedSearchParams?.company_name || "Your agency"
  const roleLabel = resolvedSearchParams?.role === "foreign_agent" ? "Foreign agency" : "Ethiopian agency"

  return <PendingClient email={email} companyName={companyName} roleLabel={roleLabel} />
}

