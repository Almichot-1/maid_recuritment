"use client"

import * as React from "react"
import { Clock } from "lucide-react"
import { cn } from "@/lib/utils"

interface LockCountdownProps {
  expiresAt: string
  onExpired?: () => void
  className?: string
  showIcon?: boolean
}

export function LockCountdown({ expiresAt, onExpired, className, showIcon = true }: LockCountdownProps) {
  const [timeRemaining, setTimeRemaining] = React.useState<string>("")
  const [colorClass, setColorClass] = React.useState<string>("")
  const [isExpired, setIsExpired] = React.useState(false)

  React.useEffect(() => {
    const calculateTimeRemaining = () => {
      const now = new Date().getTime()
      const expiry = new Date(expiresAt).getTime()
      const diff = expiry - now

      if (diff <= 0) {
        setTimeRemaining("Expired")
        setColorClass("text-red-600 dark:text-red-400")
        setIsExpired(true)
        if (onExpired && !isExpired) {
          onExpired()
        }
        return
      }

      const hours = Math.floor(diff / (1000 * 60 * 60))
      const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
      const seconds = Math.floor((diff % (1000 * 60)) / 1000)

      setTimeRemaining(`${hours}h ${minutes}m ${seconds}s remaining`)

      // Set color based on time remaining
      if (hours >= 12) {
        setColorClass("text-green-600 dark:text-green-400")
      } else if (hours >= 6) {
        setColorClass("text-yellow-600 dark:text-yellow-400")
      } else if (hours >= 1) {
        setColorClass("text-orange-600 dark:text-orange-400")
      } else {
        setColorClass("text-red-600 dark:text-red-400 animate-pulse")
      }
    }

    calculateTimeRemaining()
    const interval = setInterval(calculateTimeRemaining, 1000)

    return () => clearInterval(interval)
  }, [expiresAt, onExpired, isExpired])

  return (
    <div className={cn("flex items-center gap-2 font-medium", colorClass, className)}>
      {showIcon && <Clock className="h-4 w-4" />}
      <span className="text-sm">{timeRemaining}</span>
    </div>
  )
}
