import Link from "next/link"
import { Users } from "lucide-react"
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
    icon: "h-6 w-6",
    text: "text-lg",
    container: "gap-2",
  },
  md: {
    icon: "h-8 w-8",
    text: "text-xl",
    container: "gap-2",
  },
  lg: {
    icon: "h-10 w-10",
    text: "text-2xl",
    container: "gap-3",
  },
}

export function Logo({ size = "md", showText = true, href = "/", className }: LogoProps) {
  const config = sizeConfig[size]

  const content = (
    <div className={cn("flex items-center", config.container, className)}>
      <div className="flex items-center justify-center rounded-lg bg-primary p-1.5">
        <Users className={cn(config.icon, "text-primary-foreground")} />
      </div>
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
