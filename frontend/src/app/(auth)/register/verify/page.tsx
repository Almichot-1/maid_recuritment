import { VerifyClient } from "./verify-client"

interface VerifySearchParams {
  token?: string
}

interface VerifyPageProps {
  searchParams?: Promise<VerifySearchParams>
}

export default async function VerifyEmailPage({ searchParams }: VerifyPageProps) {
  const resolvedSearchParams = await Promise.resolve(searchParams)
  const token = resolvedSearchParams?.token || ""

  return <VerifyClient token={token} />
}

