"use client"

import * as React from "react"
import { useCurrentUser } from "@/hooks/use-auth"

type CandidateDefaults = {
  country_applied?: string
  salary_offered?: string
}

function getStorageKeys(userID?: string) {
  if (!userID) return null
  return {
    countryApplied: `candidate_defaults:${userID}:country_applied`,
    salaryOffered: `candidate_defaults:${userID}:salary_offered`,
  }
}

export function useCandidateDefaults() {
  const { user } = useCurrentUser()
  const [defaults, setDefaults] = React.useState<CandidateDefaults>({})
  const [isLoaded, setIsLoaded] = React.useState(false)

  React.useEffect(() => {
    const keys = getStorageKeys(user?.id)
    if (!keys || typeof window === "undefined") {
      setDefaults({})
      setIsLoaded(true)
      return
    }

    try {
      const countryApplied = localStorage.getItem(keys.countryApplied) || undefined
      const salaryOffered = localStorage.getItem(keys.salaryOffered) || undefined
      setDefaults({ country_applied: countryApplied, salary_offered: salaryOffered })
    } catch {
      setDefaults({})
    } finally {
      setIsLoaded(true)
    }
  }, [user?.id])

  const saveDefaults = React.useCallback(
    (newDefaults: CandidateDefaults) => {
      const keys = getStorageKeys(user?.id)
      if (!keys) return

      try {
        if (newDefaults.country_applied) {
          localStorage.setItem(keys.countryApplied, newDefaults.country_applied)
        }
        if (newDefaults.salary_offered) {
          localStorage.setItem(keys.salaryOffered, newDefaults.salary_offered)
        }
        setDefaults(newDefaults)
      } catch (error) {
        console.error("Failed to save candidate defaults:", error)
      }
    },
    [user?.id]
  )

  const clearDefaults = React.useCallback(() => {
    const keys = getStorageKeys(user?.id)
    if (!keys) return

    try {
      localStorage.removeItem(keys.countryApplied)
      localStorage.removeItem(keys.salaryOffered)
      setDefaults({})
    } catch (error) {
      console.error("Failed to clear candidate defaults:", error)
    }
  }, [user?.id])

  return {
    isLoaded,
    defaults,
    saveDefaults,
    clearDefaults,
  }
}
