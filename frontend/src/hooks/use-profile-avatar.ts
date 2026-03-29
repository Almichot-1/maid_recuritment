"use client"

import * as React from "react"

import { useCurrentUser } from "@/hooks/use-auth"

const MAX_PROFILE_PHOTO_SIZE_BYTES = 2 * 1024 * 1024
const SUPPORTED_PROFILE_PHOTO_TYPES = [
  "image/png",
  "image/jpeg",
  "image/webp",
]
const PROFILE_AVATAR_EVENT = "profile-avatar-updated"

type StoredProfileAvatar = {
  avatar_data_url?: string
  updated_at?: string
}

function getStorageKey(userID?: string) {
  return userID ? `profile_avatar:${userID}` : null
}

function emitAvatarUpdate() {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event(PROFILE_AVATAR_EVENT))
  }
}

function readStoredAvatar(userID?: string): StoredProfileAvatar {
  const key = getStorageKey(userID)
  if (!key || typeof window === "undefined") {
    return {}
  }

  try {
    const stored = localStorage.getItem(key)
    return stored ? (JSON.parse(stored) as StoredProfileAvatar) : {}
  } catch {
    return {}
  }
}

function readFileAsDataURL(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ""))
    reader.onerror = () => reject(new Error("Failed to read profile photo"))
    reader.readAsDataURL(file)
  })
}

export function useProfileAvatar() {
  const { user } = useCurrentUser()
  const [avatar, setAvatar] = React.useState<StoredProfileAvatar>({})

  React.useEffect(() => {
    setAvatar(readStoredAvatar(user?.id))
  }, [user?.id])

  React.useEffect(() => {
    if (typeof window === "undefined" || !user?.id) {
      return
    }

    const syncAvatar = () => {
      setAvatar(readStoredAvatar(user.id))
    }

    const onStorage = (event: StorageEvent) => {
      const key = getStorageKey(user.id)
      if (!key || event.key === key) {
        syncAvatar()
      }
    }

    window.addEventListener("storage", onStorage)
    window.addEventListener(PROFILE_AVATAR_EVENT, syncAvatar)

    return () => {
      window.removeEventListener("storage", onStorage)
      window.removeEventListener(PROFILE_AVATAR_EVENT, syncAvatar)
    }
  }, [user?.id])

  const saveAvatar = React.useCallback(
    async (file: File) => {
      const key = getStorageKey(user?.id)
      if (!key) {
        throw new Error("Sign in first to save a profile photo.")
      }
      if (file.size > MAX_PROFILE_PHOTO_SIZE_BYTES) {
        throw new Error("Profile photo must be smaller than 2 MB.")
      }
      if (!SUPPORTED_PROFILE_PHOTO_TYPES.includes(file.type)) {
        throw new Error("Please upload a PNG, JPG, or WEBP image.")
      }

      const avatarDataURL = await readFileAsDataURL(file)
      const nextAvatar: StoredProfileAvatar = {
        avatar_data_url: avatarDataURL,
        updated_at: new Date().toISOString(),
      }

      localStorage.setItem(key, JSON.stringify(nextAvatar))
      setAvatar(nextAvatar)
      emitAvatarUpdate()
    },
    [user?.id],
  )

  const clearAvatar = React.useCallback(() => {
    const key = getStorageKey(user?.id)
    if (!key) {
      return
    }

    localStorage.removeItem(key)
    setAvatar({})
    emitAvatarUpdate()
  }, [user?.id])

  return {
    avatarDataURL: avatar.avatar_data_url || "",
    hasAvatar: !!avatar.avatar_data_url,
    updatedAt: avatar.updated_at,
    maxProfilePhotoSizeBytes: MAX_PROFILE_PHOTO_SIZE_BYTES,
    saveAvatar,
    clearAvatar,
  }
}
