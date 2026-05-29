import { RegistrationPendingCard } from "@/components/auth/registration-pending-card"

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
  const email = resolvedSearchParams?.email || "your email"
  const roleLabel =
    resolvedSearchParams?.role === "foreign_agent" ? "foreign agency" : "Ethiopian agency"

  return (
    <RegistrationPendingCard companyName={companyName} email={email} roleLabel={roleLabel} />
  )
}
