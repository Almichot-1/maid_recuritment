"use client"

import * as React from "react"

import { useCurrentUser } from "@/hooks/use-auth"

const MAX_LOGO_SIZE_BYTES = 700 * 1024
const SUPPORTED_LOGO_TYPES = ["image/png", "image/jpeg"]

type StoredAgencyBranding = {
  logo_data_url?: string
  updated_at?: string
}

function getStorageKey(userID?: string) {
  return userID ? `agency_branding:${userID}` : null
}

function readFileAsDataURL(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ""))
    reader.onerror = () => reject(new Error("Failed to read logo file"))
    reader.readAsDataURL(file)
  })
}

export function useAgencyBranding() {
  const { user } = useCurrentUser()
  const [branding, setBranding] = React.useState<StoredAgencyBranding>({})
  const [isLoaded, setIsLoaded] = React.useState(false)

  React.useEffect(() => {
    const key = getStorageKey(user?.id)
    if (!key || typeof window === "undefined") {
      setBranding({})
      setIsLoaded(true)
      return
    }

    try {
      const stored = localStorage.getItem(key)
      setBranding(stored ? JSON.parse(stored) : {})
    } catch {
      setBranding({})
    } finally {
      setIsLoaded(true)
    }
  }, [user?.id])

  const saveLogo = React.useCallback(
    async (file: File) => {
      const key = getStorageKey(user?.id)
      if (!key) {
        throw new Error("Sign in first to save agency branding.")
      }
      if (file.size > MAX_LOGO_SIZE_BYTES) {
        throw new Error("Logo must be smaller than 700 KB.")
      }
      if (!SUPPORTED_LOGO_TYPES.includes(file.type)) {
        throw new Error("Please upload a PNG or JPG image for the agency logo.")
      }

      const logoDataURL = await readFileAsDataURL(file)
      const nextBranding: StoredAgencyBranding = {
        logo_data_url: logoDataURL,
        updated_at: new Date().toISOString(),
      }

      localStorage.setItem(key, JSON.stringify(nextBranding))
      setBranding(nextBranding)
    },
    [user?.id]
  )

  const clearLogo = React.useCallback(() => {
    const key = getStorageKey(user?.id)
    if (!key) {
      return
    }

    localStorage.removeItem(key)
    setBranding({})
  }, [user?.id])

  return {
    isLoaded,
    hasLogo: !!branding.logo_data_url,
    logoDataURL: branding.logo_data_url || "",
    updatedAt: branding.updated_at,
    maxLogoSizeBytes: MAX_LOGO_SIZE_BYTES,
    saveLogo,
    clearLogo,
  }
}
