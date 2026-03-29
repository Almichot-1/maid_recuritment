"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { Globe, Loader2, Moon, Sun } from "lucide-react"
import { useTheme } from "next-themes"

import { useUpdatePreferences, PreferencesData } from "@/hooks/use-settings"
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
  const [mounted, setMounted] = React.useState(false)

  const form = useForm<PreferencesData>({
    defaultValues: {
      theme: (theme as PreferencesData["theme"]) || "system",
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
        theme: (theme as PreferencesData["theme"]) || "system",
      })
      return
    }

    try {
      form.reset(JSON.parse(saved) as PreferencesData)
    } catch {
      form.reset({
        ...form.getValues(),
        theme: (theme as PreferencesData["theme"]) || "system",
      })
    }
  }, [form, theme])

  const onSubmit = (data: PreferencesData) => {
    // Update theme immediately
    setTheme(data.theme)
    
    // Save preferences to backend
    updatePreferences(data)
  }

  return (
    <div className="space-y-6">
      <Card className="overflow-hidden border-border/70 shadow-sm">
        <CardHeader>
          <CardTitle>Appearance</CardTitle>
          <CardDescription>
            Switch between light, dark, and system themes and keep the interface comfortable on mobile or desktop.
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
                    <FormLabel>Theme</FormLabel>
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
                            <span className="text-sm font-medium">Light</span>
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
                            <span className="text-sm font-medium">Dark</span>
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
                            <span className="text-sm font-medium">System</span>
                          </Label>
                        </div>
                      </RadioGroup>
                    </FormControl>
                    <FormDescription>
                      Select the theme for the application interface.
                    </FormDescription>
                  </FormItem>
                )}
              />

              <div className="grid gap-3 md:grid-cols-3">
                {[
                  {
                    key: "light",
                    label: "Light",
                    description: "Bright surfaces and soft borders for daytime use.",
                  },
                  {
                    key: "dark",
                    label: "Dark",
                    description: "Muted contrast and calmer glare for long sessions.",
                  },
                  {
                    key: "system",
                    label: "System",
                    description: "Automatically follows the device appearance.",
                  },
                ].map((option) => {
                  const isActive = form.watch("theme") === option.key
                  return (
                    <div
                      key={option.key}
                      className={`rounded-2xl border p-4 transition-all ${
                        isActive
                          ? "border-primary bg-primary/5 shadow-sm"
                          : "border-border/70 bg-muted/20"
                      }`}
                    >
                      <p className="text-sm font-semibold">{option.label}</p>
                      <p className="mt-1 text-xs text-muted-foreground">
                        {option.description}
                      </p>
                    </div>
                  )
                })}
              </div>

              <div className="rounded-2xl border border-border/70 bg-muted/25 p-4 text-sm text-muted-foreground">
                Current theme:{" "}
                <span className="font-semibold text-foreground">
                  {mounted ? ((theme as string) || "system") : "system"}
                </span>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>

      <Card className="overflow-hidden border-border/70 shadow-sm">
        <CardHeader>
          <CardTitle>Notifications</CardTitle>
          <CardDescription>
            Configure how you want to receive notifications.
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
                    <FormItem className="flex items-center justify-between gap-4 rounded-2xl border border-border/70 bg-muted/20 p-4">
                      <div className="space-y-0.5">
                        <FormLabel className="text-base">Email Notifications</FormLabel>
                        <FormDescription>
                          Receive email notifications for important updates
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
                    <FormItem className="flex items-center justify-between gap-4 rounded-2xl border border-border/70 bg-muted/20 p-4">
                      <div className="space-y-0.5">
                        <FormLabel className="text-base">Selection Alerts</FormLabel>
                        <FormDescription>
                          Get notified when candidates are selected
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
                    <FormItem className="flex items-center justify-between gap-4 rounded-2xl border border-border/70 bg-muted/20 p-4">
                      <div className="space-y-0.5">
                        <FormLabel className="text-base">Status Update Alerts</FormLabel>
                        <FormDescription>
                          Get notified about recruitment status changes
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
                    <FormItem className="flex items-center justify-between gap-4 rounded-2xl border border-border/70 bg-muted/20 p-4">
                      <div className="space-y-0.5">
                        <FormLabel className="text-base">Approval Alerts</FormLabel>
                        <FormDescription>
                          Get notified when approvals are needed or received
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

      <Card className="overflow-hidden border-border/70 shadow-sm">
        <CardHeader>
          <CardTitle>Language</CardTitle>
          <CardDescription>
            Select your preferred language for the application.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Application Language</Label>
              <Select defaultValue="en" disabled>
                <SelectTrigger>
                  <SelectValue placeholder="Select language" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="en">English</SelectItem>
                  <SelectItem value="ar" disabled>Arabic (Coming Soon)</SelectItem>
                  <SelectItem value="am" disabled>Amharic (Coming Soon)</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Additional languages will be available in future updates.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Save Button */}
      <div className="flex justify-end">
        <Button onClick={form.handleSubmit(onSubmit)} disabled={isPending} className="min-w-[180px]">
          {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Save Preferences
        </Button>
      </div>
    </div>
  )
}
