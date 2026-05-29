import Link from "next/link"
import { APP_NAME } from "@/constants/branding"
import { cn } from "@/lib/utils"

type LogoSize = "sm" | "md" | "lg"

interface LogoProps {
  size?: LogoSize
  showText?: boolean
  href?: string
  className?: string
}

const sizeConfig = {
  sm: {
    mark: "h-9 w-9 text-xs",
    text: "text-lg",
    container: "gap-3",
  },
  md: {
    mark: "h-11 w-11 text-sm",
    text: "text-xl",
    container: "gap-3",
  },
  lg: {
    mark: "h-12 w-12 text-sm",
    text: "text-2xl",
    container: "gap-4",
  },
}

export function Logo({ size = "md", showText = true, href = "/", className }: LogoProps) {
  const config = sizeConfig[size]

  const content = (
    <div className={cn("flex items-center", config.container, className)}>
      <div className={cn("flex items-center justify-center border border-foreground bg-foreground text-background", config.mark)}>
        <span className="route-stamp tracking-[0.24em] text-background">RM</span>
      </div>
      {showText && (
        <div className="flex flex-col">
          <span className={cn("font-display leading-none text-foreground", config.text)}>{APP_NAME}</span>
          <span className="route-stamp text-[10px] tracking-[0.22em] text-muted-foreground">AGENCY WORKSPACE</span>
        </div>
      )}
    </div>
  )

  if (href) {
    return (
      <Link href={href} className="transition-opacity hover:opacity-80">
        {content}
      </Link>
    )
  }

  return content
}
