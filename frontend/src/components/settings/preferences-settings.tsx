"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { Globe, Loader2, Moon, Sun } from "lucide-react"
import { useTheme } from "next-themes"

import { useUpdatePreferences, PreferencesData } from "@/hooks/use-settings"
import { localeOptions, useI18n } from "@/lib/i18n"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from "@/components/ui/form"
import { Switch } from "@/components/ui/switch"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

export function PreferencesSettings() {
  const { theme, setTheme } = useTheme()
  const { mutate: updatePreferences, isPending } = useUpdatePreferences()
  const { locale, setLocale, t } = useI18n()
  const [mounted, setMounted] = React.useState(false)

  const form = useForm<PreferencesData>({
    defaultValues: {
      theme: (theme as PreferencesData["theme"]) || "system",
      language: locale,
      email_notifications: true,
      selection_alerts: true,
      status_update_alerts: true,
      approval_alerts: true,
    },
  })

  React.useEffect(() => {
    setMounted(true)
  }, [])

  React.useEffect(() => {
    if (typeof window === "undefined") {
      return
    }

    const saved = localStorage.getItem("user_preferences")
    if (!saved) {
      form.reset({
        ...form.getValues(),
        language: locale,
        theme: (theme as PreferencesData["theme"]) || "system",
      })
      return
    }

    try {
      form.reset(JSON.parse(saved) as PreferencesData)
    } catch {
      form.reset({
        ...form.getValues(),
        language: locale,
        theme: (theme as PreferencesData["theme"]) || "system",
      })
    }
  }, [form, locale, theme])

  const onSubmit = (data: PreferencesData) => {
    setTheme(data.theme)
    setLocale(data.language)
    updatePreferences(data)
  }

  return (
    <div className="space-y-6">
      <Card className="overflow-hidden">
        <CardHeader>
          <CardTitle>{t("preferences.appearanceTitle")}</CardTitle>
          <CardDescription>
            {t("preferences.appearanceBody")}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <FormField
                control={form.control}
                name="theme"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("preferences.theme")}</FormLabel>
                    <FormControl>
                      <RadioGroup
                        onValueChange={field.onChange}
                        value={field.value}
                        className="grid gap-3 sm:grid-cols-3"
                      >
                        <div>
                          <RadioGroupItem
                            value="light"
                            id="light"
                            className="peer sr-only"
                          />
                          <Label
                            htmlFor="light"
                            className="flex flex-col items-center justify-between rounded-md border-2 border-muted bg-popover p-4 hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-primary [&:has([data-state=checked])]:border-primary cursor-pointer"
                          >
                            <Sun className="mb-3 h-6 w-6" />
                            <span className="text-sm font-bold">{t("preferences.themeLight")}</span>
                          </Label>
                        </div>
                        <div>
                          <RadioGroupItem
                            value="dark"
                            id="dark"
                            className="peer sr-only"
                          />
                          <Label
                            htmlFor="dark"
                            className="flex flex-col items-center justify-between rounded-md border-2 border-muted bg-popover p-4 hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-primary [&:has([data-state=checked])]:border-primary cursor-pointer"
                          >
                            <Moon className="mb-3 h-6 w-6" />
                            <span className="text-sm font-bold">{t("preferences.themeDark")}</span>
                          </Label>
                        </div>
                        <div>
                          <RadioGroupItem
                            value="system"
                            id="system"
                            className="peer sr-only"
                          />
                          <Label
                            htmlFor="system"
                            className="flex flex-col items-center justify-between rounded-md border-2 border-muted bg-popover p-4 hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-primary [&:has([data-state=checked])]:border-primary cursor-pointer"
                          >
                            <Globe className="mb-3 h-6 w-6" />
                            <span className="text-sm font-bold">{t("preferences.themeSystem")}</span>
                          </Label>
                        </div>
                      </RadioGroup>
                    </FormControl>
                    <FormDescription>
                      {t("preferences.themeHelp")}
                    </FormDescription>
                  </FormItem>
                )}
              />

              <div className="grid gap-3 md:grid-cols-3">
                {[
                  {
                    key: "light",
                    label: t("preferences.themeLight"),
                    description: "Bright surfaces and sharp borders for daytime use.",
                  },
                  {
                    key: "dark",
                    label: t("preferences.themeDark"),
                    description: "Muted contrast for long review sessions.",
                  },
                  {
                    key: "system",
                    label: t("preferences.themeSystem"),
                    description: "Automatically follows the device appearance.",
                  },
                ].map((option) => {
                  const isActive = form.watch("theme") === option.key
                  return (
                    <div
                      key={option.key}
                      className={`border p-4 transition-colors ${
                        isActive
                          ? "border-primary bg-primary/5"
                          : "border-border bg-muted/20"
                      }`}
                    >
                      <p className="text-sm font-bold">{option.label}</p>
                      <p className="mt-1 text-xs text-muted-foreground">
                        {option.description}
                      </p>
                    </div>
                  )
                })}
              </div>

              <div className="border border-border bg-muted/25 p-4 text-sm text-muted-foreground">
                {t("preferences.currentTheme")}:{" "}
                <span className="font-semibold text-foreground">
                  {mounted ? ((theme as string) || form.getValues("theme")) : form.getValues("theme")}
                </span>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>

      <Card className="overflow-hidden">
        <CardHeader>
          <CardTitle>{t("preferences.notificationsTitle")}</CardTitle>
          <CardDescription>
            {t("preferences.notificationsBody")}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <div className="space-y-4">
                {/* Email Notifications */}
                <FormField
                  control={form.control}
                  name="email_notifications"
                  render={({ field }) => (
                    <FormItem className="flex items-center justify-between gap-4 border border-border bg-muted/20 p-4">
                      <div className="space-y-0.5">
                        <FormLabel className="text-base">{t("preferences.emailNotifications")}</FormLabel>
                        <FormDescription>
                          {t("preferences.emailNotificationsBody")}
                        </FormDescription>
                      </div>
                      <FormControl>
                        <Switch
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>
                    </FormItem>
                  )}
                />

                {/* Selection Alerts */}
                <FormField
                  control={form.control}
                  name="selection_alerts"
                  render={({ field }) => (
                    <FormItem className="flex items-center justify-between gap-4 border border-border bg-muted/20 p-4">
                      <div className="space-y-0.5">
                        <FormLabel className="text-base">{t("preferences.selectionAlerts")}</FormLabel>
                        <FormDescription>
                          {t("preferences.selectionAlertsBody")}
                        </FormDescription>
                      </div>
                      <FormControl>
                        <Switch
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>
                    </FormItem>
                  )}
                />

                {/* Status Update Alerts */}
                <FormField
                  control={form.control}
                  name="status_update_alerts"
                  render={({ field }) => (
                    <FormItem className="flex items-center justify-between gap-4 border border-border bg-muted/20 p-4">
                      <div className="space-y-0.5">
                        <FormLabel className="text-base">{t("preferences.statusAlerts")}</FormLabel>
                        <FormDescription>
                          {t("preferences.statusAlertsBody")}
                        </FormDescription>
                      </div>
                      <FormControl>
                        <Switch
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>
                    </FormItem>
                  )}
                />

                {/* Approval Alerts */}
                <FormField
                  control={form.control}
                  name="approval_alerts"
                  render={({ field }) => (
                    <FormItem className="flex items-center justify-between gap-4 border border-border bg-muted/20 p-4">
                      <div className="space-y-0.5">
                        <FormLabel className="text-base">{t("preferences.approvalAlerts")}</FormLabel>
                        <FormDescription>
                          {t("preferences.approvalAlertsBody")}
                        </FormDescription>
                      </div>
                      <FormControl>
                        <Switch
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>
                    </FormItem>
                  )}
                />
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>

      <Card className="overflow-hidden">
        <CardHeader>
          <CardTitle>{t("preferences.languageTitle")}</CardTitle>
          <CardDescription>
            {t("preferences.languageBody")}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <div className="space-y-2">
              <Label>{t("preferences.languageField")}</Label>
              <FormField
                control={form.control}
                name="language"
                render={({ field }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger>
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
                )}
              />
              <p className="text-xs text-muted-foreground">
                {localeOptions.find((option) => option.value === form.watch("language"))?.nativeLabel}
              </p>
            </div>
          </Form>
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button onClick={form.handleSubmit(onSubmit)} disabled={isPending} className="min-w-[180px]">
          {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          {t("common.savePreferences")}
        </Button>
      </div>
    </div>
  )
}
