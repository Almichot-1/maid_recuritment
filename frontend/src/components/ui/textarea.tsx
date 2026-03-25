import * as React from "react"

import { cn } from "@/lib/utils"

const Textarea = React.forwardRef<
  HTMLTextAreaElement,
  React.ComponentProps<"textarea">
>(({ className, ...props }, ref) => {
  return (
    <textarea
      className={cn(
        "flex min-h-[60px] w-full rounded-xl border border-input bg-background/95 px-3.5 py-2.5 text-base text-foreground shadow-sm transition-all duration-200 placeholder:text-muted-foreground hover:border-ring/35 hover:bg-background focus-visible:border-ring/45 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-ring/15 focus-visible:ring-offset-0 disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
        className
      )}
      ref={ref}
      {...props}
    />
  )
})
Textarea.displayName = "Textarea"

export { Textarea }
