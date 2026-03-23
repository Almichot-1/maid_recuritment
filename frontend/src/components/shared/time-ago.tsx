"use client"

import * as React from "react"
import { formatDistanceToNow, format } from "date-fns"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"

interface TimeAgoProps {
  date: string | Date
  className?: string
}

export function TimeAgo({ date, className }: TimeAgoProps) {
  const [timeAgo, setTimeAgo] = React.useState("")

  React.useEffect(() => {
    const updateTimeAgo = () => {
      setTimeAgo(formatDistanceToNow(new Date(date), { addSuffix: true }))
    }

    updateTimeAgo()
    const interval = setInterval(updateTimeAgo, 60000) // Update every minute

    return () => clearInterval(interval)
  }, [date])

  const fullDate = format(new Date(date), "PPpp") // e.g., "Apr 29, 2023, 9:30:00 AM"

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <time dateTime={new Date(date).toISOString()} className={className}>
            {timeAgo}
          </time>
        </TooltipTrigger>
        <TooltipContent>
          <p>{fullDate}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
