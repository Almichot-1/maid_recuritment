"use client"

import * as React from "react"
import { Languages } from "lucide-react"

import { useI18n, localeOptions, type Locale } from "@/lib/i18n"
import { cn } from "@/lib/utils"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

type LocaleSwitcherProps = {
  className?: string
  compact?: boolean
}

export function LocaleSwitcher({ className, compact = false }: LocaleSwitcherProps) {
  const { locale, setLocale, t } = useI18n()

  const handleChange = React.useCallback(
    (value: string) => {
      React.startTransition(() => {
        setLocale(value as Locale)
      })
    },
    [setLocale]
  )

  return (
    <div className={cn("flex items-center gap-2", className)}>
      {!compact ? (
        <span className="section-kicker whitespace-nowrap">{t("common.language")}</span>
      ) : (
        <Languages className="h-4 w-4 text-muted-foreground" />
      )}
      <Select value={locale} onValueChange={handleChange}>
        <SelectTrigger
          aria-label={t("common.language")}
          className={cn(
            "bg-card text-foreground",
            compact ? "h-9 min-w-[128px]" : "min-w-[180px]"
          )}
        >
          <SelectValue placeholder={t("preferences.selectLanguage")} />
        </SelectTrigger>
        <SelectContent>
          {localeOptions.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.nativeLabel}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}
