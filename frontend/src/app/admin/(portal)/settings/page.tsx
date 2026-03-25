"use client"

import * as React from "react"
import { Building2, Clock3, Loader2, MoonStar, Save, ShieldAlert, ShieldCheck, Wrench } from "lucide-react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatCard, AdminSurface } from "@/components/admin/admin-ui"
import { ThemeToggle } from "@/components/shared/theme-toggle"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import { useCurrentAdmin } from "@/hooks/use-admin-auth"
import { toast } from "@/hooks/use-toast"
import { useAdminPlatformSettings, useUpdateAdminPlatformSettings } from "@/hooks/use-admin-portal"
import { PlatformSettings } from "@/types"

function SettingsToggle({
  id,
  title,
  description,
  checked,
  onCheckedChange,
}: {
  id: string
  title: string
  description: string
  checked: boolean
  onCheckedChange: (checked: boolean) => void
}) {
  return (
    <div className="flex items-start justify-between gap-4 rounded-2xl border border-slate-200/80 bg-white/80 p-4 dark:border-slate-800 dark:bg-slate-900/78">
      <div className="space-y-1">
        <Label htmlFor={id} className="text-sm font-semibold text-slate-950 dark:text-slate-100">
          {title}
        </Label>
        <p className="text-sm text-slate-500 dark:text-slate-400">{description}</p>
      </div>
      <Switch id={id} checked={checked} onCheckedChange={onCheckedChange} />
    </div>
  )
}

export default function AdminSettingsPage() {
  const { isSuperAdmin } = useCurrentAdmin()
  const { data, isLoading } = useAdminPlatformSettings(isSuperAdmin)
  const updateSettings = useUpdateAdminPlatformSettings()
  const [form, setForm] = React.useState<PlatformSettings | null>(null)

  React.useEffect(() => {
    if (data) {
      setForm(data)
    }
  }, [data])

  if (!isSuperAdmin) {
    return (
      <Card className="border-amber-200 bg-amber-50 dark:border-amber-400/20 dark:bg-amber-400/10">
        <CardContent className="flex items-center gap-3 p-6 text-sm text-amber-900 dark:text-amber-100">
          <ShieldAlert className="h-5 w-5" />
          Platform settings are restricted to Super Admin accounts.
        </CardContent>
      </Card>
    )
  }

  const isDirty = Boolean(data && form && JSON.stringify(data) !== JSON.stringify(form))

  const setField = <K extends keyof PlatformSettings>(field: K, value: PlatformSettings[K]) => {
    setForm((current) => (current ? { ...current, [field]: value } : current))
  }

  const saveSettings = async () => {
    if (!form) {
      return
    }

    try {
      await updateSettings.mutateAsync(form)
      toast({
        title: "Platform settings updated",
        description: "New admin controls have been saved and are now live.",
      })
    } catch (error) {
      toast({
        title: "Failed to save settings",
        description: error instanceof Error ? error.message : "Please try again.",
        variant: "destructive",
      })
    }
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Platform Settings"
        description="Control live platform behavior for approvals, expiry, maintenance, and admin-driven email policy."
        action={
          <Button className="gap-2" disabled={!form || !isDirty || updateSettings.isPending} onClick={saveSettings}>
            {updateSettings.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            Save changes
          </Button>
        }
      />

      {isLoading || !form ? (
        <AdminSurface>
          <CardContent className="flex items-center gap-3 p-6 text-sm text-slate-500 dark:text-slate-400">
            <Loader2 className="h-4 w-4 animate-spin" />
            Loading platform settings...
          </CardContent>
        </AdminSurface>
      ) : (
        <>
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <AdminStatCard label="Selection lock" value={`${form.selection_lock_duration_hours}h`} detail="Platform-wide selection hold duration" icon={Clock3} />
            <AdminStatCard label="Dual approval" value={form.require_both_approvals ? "On" : "Off"} detail="Both agencies must confirm the selection" icon={ShieldCheck} />
            <AdminStatCard label="Auto-approve agencies" value={form.auto_approve_agencies ? "On" : "Off"} detail="Skips the manual registration queue" icon={Building2} />
            <AdminStatCard label="Maintenance mode" value={form.maintenance_mode ? "Live" : "Off"} detail="Blocks agency access while admins stay online" icon={Wrench} />
          </div>

          <div className="grid gap-6 xl:grid-cols-2">
            <AdminSurface>
              <CardHeader>
                <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Recruitment rules</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="selection-lock-duration" className="text-sm font-semibold text-slate-950 dark:text-slate-100">
                    Selection lock duration
                  </Label>
                  <Select
                    value={String(form.selection_lock_duration_hours)}
                    onValueChange={(value) => setField("selection_lock_duration_hours", Number(value))}
                  >
                    <SelectTrigger id="selection-lock-duration">
                      <SelectValue placeholder="Select lock duration" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="12">12 hours</SelectItem>
                      <SelectItem value="24">24 hours</SelectItem>
                      <SelectItem value="48">48 hours</SelectItem>
                    </SelectContent>
                  </Select>
                  <p className="text-sm text-slate-500 dark:text-slate-400">Applies to all new foreign-agency selections.</p>
                </div>

                <SettingsToggle
                  id="require-both-approvals"
                  title="Require both parties to approve"
                  description="When off, the foreign agency approval is enough to move a recruitment into progress."
                  checked={form.require_both_approvals}
                  onCheckedChange={(checked) => setField("require_both_approvals", checked)}
                />

                <SettingsToggle
                  id="auto-expire-selections"
                  title="Auto-expire pending selections"
                  description="When off, the expiry worker stops releasing candidates automatically."
                  checked={form.auto_expire_selections}
                  onCheckedChange={(checked) => setField("auto_expire_selections", checked)}
                />
              </CardContent>
            </AdminSurface>

            <AdminSurface>
              <CardHeader>
                <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Agency onboarding and notifications</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <SettingsToggle
                  id="auto-approve-agencies"
                  title="Auto-approve agencies"
                  description="Use carefully. New agency registrations skip the manual pending-approval queue when this is enabled."
                  checked={form.auto_approve_agencies}
                  onCheckedChange={(checked) => setField("auto_approve_agencies", checked)}
                />

                <SettingsToggle
                  id="email-notifications-enabled"
                  title="Email notifications enabled"
                  description="Turns admin and agency emails on or off without affecting in-app notifications."
                  checked={form.email_notifications_enabled}
                  onCheckedChange={(checked) => setField("email_notifications_enabled", checked)}
                />
              </CardContent>
            </AdminSurface>
          </div>

          <AdminSurface>
            <CardHeader>
              <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Maintenance mode</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <SettingsToggle
                id="maintenance-mode"
                title="Enable maintenance mode"
                description="Agency routes and authentication are blocked while admins can still access the portal."
                checked={form.maintenance_mode}
                onCheckedChange={(checked) => setField("maintenance_mode", checked)}
              />
              <div className="space-y-2">
                <Label htmlFor="maintenance-message" className="text-sm font-semibold text-slate-950 dark:text-slate-100">
                  Maintenance message
                </Label>
                <Input
                  id="maintenance-message"
                  value={form.maintenance_message}
                  onChange={(event) => setField("maintenance_message", event.target.value)}
                />
              </div>
            </CardContent>
          </AdminSurface>

          <AdminSurface>
            <CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Appearance & operator experience</CardTitle>
                <p className="text-sm text-slate-500 dark:text-slate-400">
                  Light mode, dark mode, and system mode are now available across the website. Each user preference is stored locally in their browser.
                </p>
              </div>
              <ThemeToggle showLabel className="w-full justify-between sm:w-auto" />
            </CardHeader>
            <CardContent>
              <div className="rounded-2xl border border-slate-200/80 bg-white/80 p-4 dark:border-slate-800 dark:bg-slate-900/78">
                <div className="flex items-start gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-amber-100 text-amber-700 dark:bg-amber-400/15 dark:text-amber-300">
                    <MoonStar className="h-5 w-5" />
                  </div>
                  <div className="space-y-1">
                    <p className="font-semibold text-slate-950 dark:text-slate-100">Theme switching is live</p>
                    <p className="text-sm text-slate-500 dark:text-slate-400">
                      Users can switch appearance from the website header, mobile header, auth screens, and the admin portal shell without touching backend settings.
                    </p>
                  </div>
                </div>
              </div>
            </CardContent>
          </AdminSurface>

          <AdminSurface>
            <CardHeader>
              <CardTitle className="text-lg text-slate-950 dark:text-slate-50">Email templates</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 xl:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="approval-template" className="text-sm font-semibold text-slate-950 dark:text-slate-100">
                  Agency approval
                </Label>
                <Textarea
                  id="approval-template"
                  value={form.agency_approval_email_template}
                  onChange={(event) => setField("agency_approval_email_template", event.target.value)}
                  className="min-h-40"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="rejection-template" className="text-sm font-semibold text-slate-950 dark:text-slate-100">
                  Agency rejection
                </Label>
                <Textarea
                  id="rejection-template"
                  value={form.agency_rejection_email_template}
                  onChange={(event) => setField("agency_rejection_email_template", event.target.value)}
                  className="min-h-40"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="selection-template" className="text-sm font-semibold text-slate-950 dark:text-slate-100">
                  Selection notifications
                </Label>
                <Textarea
                  id="selection-template"
                  value={form.selection_notification_email_template}
                  onChange={(event) => setField("selection_notification_email_template", event.target.value)}
                  className="min-h-40"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="expiry-template" className="text-sm font-semibold text-slate-950 dark:text-slate-100">
                  Expiry notifications
                </Label>
                <Textarea
                  id="expiry-template"
                  value={form.expiry_notification_email_template}
                  onChange={(event) => setField("expiry_notification_email_template", event.target.value)}
                  className="min-h-40"
                />
              </div>
              <div className="rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-4 text-sm text-slate-500 dark:border-slate-700 dark:bg-slate-900/72 dark:text-slate-400 xl:col-span-2">
                Supported variables: <code>{"{company_name}"}</code>, <code>{"{full_name}"}</code>, <code>{"{candidate_name}"}</code>, <code>{"{reason}"}</code>, <code>{"{message}"}</code>, <code>{"{title}"}</code>.
              </div>
            </CardContent>
          </AdminSurface>
        </>
      )}
    </div>
  )
}
