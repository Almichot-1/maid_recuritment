"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface VerifiedOcrInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string
  ocrDetectedValue?: string
}

const VerifiedOcrInput = React.forwardRef<HTMLInputElement, VerifiedOcrInputProps>(
  ({ label, ocrDetectedValue, value, className, id, ...props }, ref) => {
    const isOcrMatched =
      !!ocrDetectedValue &&
      typeof value === "string" &&
      value === ocrDetectedValue

    return (
      <div className="flex flex-col gap-1.5 w-full">
        <div className="flex items-center justify-between">
          <label htmlFor={id} className="text-sm font-medium text-foreground">
            {label}
          </label>
          {isOcrMatched && (
            <span className="inline-flex items-center gap-1 rounded-full border border-amber-500/20 bg-amber-500/10 px-2 py-0.5 text-[10px] font-medium text-amber-600 dark:text-amber-400">
              <span className="h-1.5 w-1.5 rounded-full bg-amber-500 animate-pulse" />
              Auto-filled from OCR
            </span>
          )}
        </div>
        <input
          id={id}
          ref={ref}
          value={value}
          className={cn(
            "flex h-10 w-full rounded-md border bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
            isOcrMatched && "border-amber-500/40 focus-visible:ring-amber-500/40",
            className
          )}
          {...props}
        />
      </div>
    )
  }
)
VerifiedOcrInput.displayName = "VerifiedOcrInput"

export { VerifiedOcrInput }
