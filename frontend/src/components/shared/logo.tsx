import Link from "next/link"
import { APP_NAME, LOGO_DARK_URL, LOGO_URL } from "@/constants/branding"
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
    icon: "h-7 w-7",
    text: "text-lg",
    container: "gap-2",
  },
  md: {
    icon: "h-9 w-9",
    text: "text-xl",
    container: "gap-2",
  },
  lg: {
    icon: "h-12 w-12",
    text: "text-2xl",
    container: "gap-3",
  },
}

export function Logo({ size = "md", showText = true, href = "/", className }: LogoProps) {
  const config = sizeConfig[size]

  const content = (
    <div className={cn("flex items-center", config.container, className)}>
      <span className={cn("inline-flex shrink-0 items-center justify-center", config.icon)}>
        <img
          src={LOGO_URL}
          alt={`${APP_NAME} logo`}
          className="h-full w-full object-contain dark:hidden"
          decoding="async"
        />
        <img
          src={LOGO_DARK_URL}
          alt={`${APP_NAME} logo`}
          className="hidden h-full w-full object-contain dark:block"
          decoding="async"
        />
      </span>
      {showText && (
        <span className={cn("font-bold text-foreground", config.text)}>
          {APP_NAME}
        </span>
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
