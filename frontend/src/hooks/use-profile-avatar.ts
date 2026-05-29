"use client"

import * as React from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import axios from "axios"

import api from "@/lib/api"
import { useCurrentUser } from "@/hooks/use-auth"
import { useAuthStore } from "@/stores/auth-store"
import { User } from "@/types"

const MAX_PROFILE_PHOTO_SIZE_BYTES = 2 * 1024 * 1024
const SUPPORTED_PROFILE_PHOTO_TYPES = [
  "image/png",
  "image/jpeg",
  "image/webp",
]

interface UserResponse {
  user: User
}

interface ApiErrorResponse {
  error?: string
  message?: string
}

export function useProfileAvatar() {
  const { user } = useCurrentUser()
  const updateUser = useAuthStore((state) => state.updateUser)
  const queryClient = useQueryClient()

  const uploadAvatar = useMutation({
    mutationFn: async (file: File) => {
      if (file.size > MAX_PROFILE_PHOTO_SIZE_BYTES) {
        throw new Error("Profile photo must be smaller than 2 MB.")
      }
      if (!SUPPORTED_PROFILE_PHOTO_TYPES.includes(file.type)) {
        throw new Error("Please upload a PNG, JPG, or WEBP image.")
      }

      const formData = new FormData()
      formData.append("file", file)
      const response = await api.post<UserResponse>("/users/avatar", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      })
      return response.data.user
    },
    onSuccess: (nextUser) => {
      updateUser(nextUser)
      queryClient.invalidateQueries({ queryKey: ["active-sessions"] })
      toast.success("Profile photo updated successfully")
    },
    onError: (error: unknown) => {
      const message = axios.isAxiosError<ApiErrorResponse>(error)
        ? error.response?.data?.error || error.response?.data?.message || error.message
        : error instanceof Error
          ? error.message
          : "Failed to upload profile photo"
      toast.error(message)
    },
  })

  const removeAvatar = useMutation({
    mutationFn: async () => {
      const response = await api.delete<UserResponse>("/users/avatar")
      return response.data.user
    },
    onSuccess: (nextUser) => {
      updateUser(nextUser)
      toast.success("Profile photo removed")
    },
    onError: (error: unknown) => {
      const message = axios.isAxiosError<ApiErrorResponse>(error)
        ? error.response?.data?.error || error.response?.data?.message || error.message
        : error instanceof Error
          ? error.message
          : "Failed to remove profile photo"
      toast.error(message)
    },
  })

  return {
    avatarDataURL: user?.avatar_url || "",
    hasAvatar: Boolean(user?.avatar_url),
    isUploading: uploadAvatar.isPending,
    isRemoving: removeAvatar.isPending,
    maxProfilePhotoSizeBytes: MAX_PROFILE_PHOTO_SIZE_BYTES,
    saveAvatar: React.useCallback(async (file: File) => {
      await uploadAvatar.mutateAsync(file)
    }, [uploadAvatar]),
    clearAvatar: React.useCallback(() => {
      removeAvatar.mutate()
    }, [removeAvatar]),
  }
}
