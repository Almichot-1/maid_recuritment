"use client"

import * as React from "react"
import { LaptopMinimal, MoonStar, SunMedium } from "lucide-react"
import { useTheme } from "next-themes"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"

type ThemeOption = "light" | "dark" | "system"

function getThemeIcon(theme?: string) {
  switch (theme) {
    case "light":
      return SunMedium
    case "dark":
      return MoonStar
    default:
      return LaptopMinimal
  }
}

function getThemeLabel(theme?: string) {
  switch (theme) {
    case "light":
      return "Light"
    case "dark":
      return "Dark"
    default:
      return "System"
  }
}

export function ThemeToggle({
  className,
  align = "end",
  showLabel = false,
}: {
  className?: string
  align?: "start" | "center" | "end"
  showLabel?: boolean
}) {
  const { theme, resolvedTheme, setTheme } = useTheme()
  const [mounted, setMounted] = React.useState(false)

  React.useEffect(() => {
    setMounted(true)
  }, [])

  const activeTheme = mounted ? (theme as ThemeOption | undefined) : undefined
  const Icon = getThemeIcon(activeTheme ?? resolvedTheme)
  const activeLabel = getThemeLabel(activeTheme)

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size={showLabel ? "sm" : "icon"}
          className={cn("gap-2", className)}
          aria-label="Toggle theme"
        >
          <Icon className="h-4 w-4" />
          {showLabel ? <span>{activeLabel}</span> : null}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align={align} className="w-44">
        <DropdownMenuLabel>Appearance</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuRadioGroup value={activeTheme ?? "system"} onValueChange={(value) => setTheme(value as ThemeOption)}>
          <DropdownMenuRadioItem value="light">
            <SunMedium className="h-4 w-4" />
            Light
          </DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="dark">
            <MoonStar className="h-4 w-4" />
            Dark
          </DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="system">
            <LaptopMinimal className="h-4 w-4" />
            System
          </DropdownMenuRadioItem>
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
