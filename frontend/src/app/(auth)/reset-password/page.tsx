import { ResetPasswordForm } from "@/components/auth/reset-password-form"

interface ResetPasswordPageProps {
  searchParams?: {
    email?: string | string[]
  }
}

export default function ResetPasswordPage({ searchParams }: ResetPasswordPageProps) {
  const initialEmail = Array.isArray(searchParams?.email) ? searchParams?.email[0] ?? "" : searchParams?.email ?? ""
  return <ResetPasswordForm initialEmail={initialEmail} />
}
