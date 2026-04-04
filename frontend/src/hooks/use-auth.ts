import { AxiosError } from "axios"
import { useMutation } from "@tanstack/react-query"
import { useRouter } from "next/navigation"
import { toast } from "sonner"
import api from "@/lib/api"
import { useAuthStore } from "@/stores/auth-store"
import { ForgotPasswordRequestInput, ForgotPasswordResetInput, LoginInput, RegisterFormInput } from "@/lib/validations"
import { AccountStatus, User, UserRole } from "@/types"
import { getRoleHomePath } from "@/lib/role-home"

interface AuthResponse {
  user: User
}

interface RegisterResponse {
  message: string
  user: User
}

interface GenericMessageResponse {
  message: string
  user?: User
}

interface ApiErrorResponse {
  error?: string
  message?: string
  account_status?: AccountStatus
}

export function getLoginErrorMessage(error: unknown): string {
  const response = (error as AxiosError<ApiErrorResponse>).response

  if (response?.data?.account_status === AccountStatus.PENDING_APPROVAL) {
    return "Your account is under review. An admin must approve your agency before you can sign in."
  }

  if (response?.status === 401 || response?.status === 403) {
    if (response?.data?.error === "email not verified") {
      return "Please verify your email first. Check your inbox for the verification link."
    }
    return "The email or password you entered is not correct. Please check it and try again."
  }

  return (
    response?.data?.error ||
    response?.data?.message ||
    "We could not sign you in right now. Please try again."
  )
}

export function useLogin() {
  const router = useRouter()
  const setAuth = useAuthStore((state) => state.setAuth)

  return useMutation({
    mutationFn: async (data: LoginInput) => {
      const response = await api.post<AuthResponse>("/auth/login", data)
      return response.data
    },
    onSuccess: (data) => {
      setAuth(data.user)
      toast.success("Successfully logged in")
      router.replace(getRoleHomePath(data.user.role))
    },
    onError: (error) => {
      const message = getLoginErrorMessage(error)
      const response = (error as AxiosError<ApiErrorResponse>).response?.data
      if (response?.account_status === AccountStatus.PENDING_APPROVAL) {
        toast.error(message)
      }
    },
  })
}

export function useRegister() {
  const router = useRouter()

  return useMutation({
    mutationFn: async (formData: RegisterFormInput) => {
      const data = {
        email: formData.email,
        password: formData.password,
        full_name: formData.full_name,
        role: formData.role,
        company_name: formData.company_name,
      }

      const response = await api.post<RegisterResponse>("/auth/register", data)
      return response.data
    },
    onSuccess: (_, variables) => {
      toast.success("Registration created. Please verify your email to continue.")
      const params = new URLSearchParams({
        email: variables.email,
        company_name: variables.company_name,
        role: variables.role,
      })
      router.push(`/register/pending?${params.toString()}`)
    },
  })
}

export function useVerifyEmail() {
  const router = useRouter()

  return useMutation({
    mutationFn: async (token: string) => {
      const response = await api.post<GenericMessageResponse>("/auth/verify-email", { token })
      return response.data
    },
    onSuccess: (data) => {
      toast.success(data.message)
      router.replace("/login")
    },
    onError: (error: unknown) => {
      const response = (error as AxiosError<ApiErrorResponse>).response
      toast.error(
        response?.data?.error ||
        response?.data?.message ||
        "We could not verify your email right now. Please try again."
      )
    },
  })
}

export function useResendVerification() {
  return useMutation({
    mutationFn: async (email: string) => {
      const response = await api.post<GenericMessageResponse>("/auth/resend-verification", { email })
      return response.data
    },
    onSuccess: (data) => {
      toast.success(data.message)
    },
    onError: (error: unknown) => {
      const response = (error as AxiosError<ApiErrorResponse>).response
      toast.error(
        response?.data?.error ||
        response?.data?.message ||
        "We could not resend the verification email right now. Please try again."
      )
    },
  })
}

export function useLogout() {
  const router = useRouter()
  const logout = useAuthStore((state) => state.logout)

  return async () => {
    try {
      await api.post("/auth/logout")
    } catch {
      // handled by interceptor when needed
    } finally {
      logout()
      router.push("/login")
      toast.info("Logged out successfully")
    }
  }
}

export function useRequestPasswordReset() {
  return useMutation({
    mutationFn: async (data: ForgotPasswordRequestInput) => {
      const response = await api.post<GenericMessageResponse>("/auth/forgot-password/request", data)
      return response.data
    },
    onSuccess: (data) => {
      toast.success(data.message)
    },
    onError: () => {
      toast.error("We could not send a reset code right now. Please try again.")
    },
  })
}

export function useResetForgottenPassword() {
  const router = useRouter()

  return useMutation({
    mutationFn: async (data: ForgotPasswordResetInput) => {
      const response = await api.post<GenericMessageResponse>("/auth/forgot-password/reset", {
        email: data.email,
        code: data.code,
        new_password: data.new_password,
      })
      return response.data
    },
    onSuccess: (data) => {
      toast.success(data.message)
      router.push("/login")
    },
    onError: (error: unknown) => {
      const response = (error as AxiosError<ApiErrorResponse>).response
      toast.error(
        response?.data?.error ||
          response?.data?.message ||
          "We could not reset your password right now. Please try again."
      )
    },
  })
}

export function useCurrentUser() {
  const user = useAuthStore((state) => state.user)
  const isLoading = false
  
  return {
    user,
    isEthiopianAgent: user?.role === UserRole.ETHIOPIAN_AGENT,
    isForeignAgent: user?.role === UserRole.FOREIGN_AGENT,
    isLoading,
  }
}
