"use client"

import * as React from "react"
import { KeyRound, Loader2, Save, ShieldAlert } from "lucide-react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import { useAdminChangePassword, useCurrentAdmin } from "@/hooks/use-admin-auth"
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
    <div className="flex items-start justify-between gap-4 rounded-2xl border border-slate-200 p-4">
      <div className="space-y-1">
        <Label htmlFor={id} className="text-sm font-semibold text-slate-950">
          {title}
        </Label>
        <p className="text-sm text-slate-500">{description}</p>
      </div>
      <Switch id={id} checked={checked} onCheckedChange={onCheckedChange} />
    </div>
  )
}

export default function AdminSettingsPage() {
  const { isSuperAdmin } = useCurrentAdmin()
  const { data, isLoading } = useAdminPlatformSettings(isSuperAdmin)
  const updateSettings = useUpdateAdminPlatformSettings()
  const changePassword = useAdminChangePassword()
  const [form, setForm] = React.useState<PlatformSettings | null>(null)
  const [currentPassword, setCurrentPassword] = React.useState("")
  const [newPassword, setNewPassword] = React.useState("")
  const [confirmPassword, setConfirmPassword] = React.useState("")

  React.useEffect(() => {
    if (data) {
      setForm(data)
    }
  }, [data])

  if (!isSuperAdmin) {
    return (
      <Card className="border-amber-200 bg-amber-50">
        <CardContent className="flex items-center gap-3 p-6 text-sm text-amber-900">
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

  const savePassword = async () => {
    if (!currentPassword || !newPassword || !confirmPassword) {
      toast({
        title: "Missing password fields",
        description: "Fill the current password, the new password, and the confirmation field.",
        variant: "destructive",
      })
      return
    }

    if (newPassword !== confirmPassword) {
      toast({
        title: "Passwords do not match",
        description: "The new password and the confirmation must be identical.",
        variant: "destructive",
      })
      return
    }

    try {
      await changePassword.mutateAsync({
        current_password: currentPassword,
        new_password: newPassword,
      })
      setCurrentPassword("")
      setNewPassword("")
      setConfirmPassword("")
      toast({
        title: "Password updated",
        description: "Your admin password has been changed successfully.",
      })
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Please check the current password and try again."
      toast({
        title: "Failed to change password",
        description: message,
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
          <Button className="gap-2 bg-slate-950 hover:bg-slate-800" disabled={!form || !isDirty || updateSettings.isPending} onClick={saveSettings}>
            {updateSettings.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            Save changes
          </Button>
        }
      />

      {isLoading || !form ? (
        <Card className="border-slate-200 bg-white/90">
          <CardContent className="flex items-center gap-3 p-6 text-sm text-slate-500">
            <Loader2 className="h-4 w-4 animate-spin" />
            Loading platform settings...
          </CardContent>
        </Card>
      ) : (
        <>
          <Card className="border-slate-200 bg-white/90">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg text-slate-950">
                <KeyRound className="h-5 w-5 text-amber-400" />
                Admin password
              </CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)_auto] xl:items-end">
              <div className="space-y-2">
                <Label htmlFor="admin-current-password" className="text-sm font-semibold text-slate-950">
                  Current password
                </Label>
                <Input
                  id="admin-current-password"
                  type="password"
                  value={currentPassword}
                  onChange={(event) => setCurrentPassword(event.target.value)}
                  className="bg-white"
                  placeholder="Enter current password"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="admin-new-password" className="text-sm font-semibold text-slate-950">
                  New password
                </Label>
                <Input
                  id="admin-new-password"
                  type="password"
                  value={newPassword}
                  onChange={(event) => setNewPassword(event.target.value)}
                  className="bg-white"
                  placeholder="At least 12 chars with complexity"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="admin-confirm-password" className="text-sm font-semibold text-slate-950">
                  Confirm new password
                </Label>
                <Input
                  id="admin-confirm-password"
                  type="password"
                  value={confirmPassword}
                  onChange={(event) => setConfirmPassword(event.target.value)}
                  className="bg-white"
                  placeholder="Repeat the new password"
                />
              </div>
              <Button
                className="gap-2 bg-amber-400 text-slate-950 hover:bg-amber-300"
                disabled={changePassword.isPending}
                onClick={savePassword}
              >
                {changePassword.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <KeyRound className="h-4 w-4" />}
                Change password
              </Button>
            </CardContent>
          </Card>

          <div className="grid gap-6 xl:grid-cols-2">
            <Card className="border-slate-200 bg-white/90">
              <CardHeader>
                <CardTitle className="text-lg text-slate-950">Recruitment rules</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="selection-lock-duration" className="text-sm font-semibold text-slate-950">
                    Selection lock duration
                  </Label>
                  <Select
                    value={String(form.selection_lock_duration_hours)}
                    onValueChange={(value) => setField("selection_lock_duration_hours", Number(value))}
                  >
                    <SelectTrigger id="selection-lock-duration" className="bg-white">
                      <SelectValue placeholder="Select lock duration" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="12">12 hours</SelectItem>
                      <SelectItem value="24">24 hours</SelectItem>
                      <SelectItem value="48">48 hours</SelectItem>
                    </SelectContent>
                  </Select>
                  <p className="text-sm text-slate-500">Applies to all new foreign-agency selections.</p>
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
            </Card>

            <Card className="border-slate-200 bg-white/90">
              <CardHeader>
                <CardTitle className="text-lg text-slate-950">Agency onboarding and notifications</CardTitle>
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
            </Card>
          </div>

          <Card className="border-slate-200 bg-white/90">
            <CardHeader>
              <CardTitle className="text-lg text-slate-950">Maintenance mode</CardTitle>
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
                <Label htmlFor="maintenance-message" className="text-sm font-semibold text-slate-950">
                  Maintenance message
                </Label>
                <Input
                  id="maintenance-message"
                  value={form.maintenance_message}
                  onChange={(event) => setField("maintenance_message", event.target.value)}
                  className="bg-white"
                />
              </div>
            </CardContent>
          </Card>

          <Card className="border-slate-200 bg-white/90">
            <CardHeader>
              <CardTitle className="text-lg text-slate-950">Email templates</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 xl:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="approval-template" className="text-sm font-semibold text-slate-950">
                  Agency approval
                </Label>
                <Textarea
                  id="approval-template"
                  value={form.agency_approval_email_template}
                  onChange={(event) => setField("agency_approval_email_template", event.target.value)}
                  className="min-h-40 bg-white"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="rejection-template" className="text-sm font-semibold text-slate-950">
                  Agency rejection
                </Label>
                <Textarea
                  id="rejection-template"
                  value={form.agency_rejection_email_template}
                  onChange={(event) => setField("agency_rejection_email_template", event.target.value)}
                  className="min-h-40 bg-white"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="selection-template" className="text-sm font-semibold text-slate-950">
                  Selection notifications
                </Label>
                <Textarea
                  id="selection-template"
                  value={form.selection_notification_email_template}
                  onChange={(event) => setField("selection_notification_email_template", event.target.value)}
                  className="min-h-40 bg-white"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="expiry-template" className="text-sm font-semibold text-slate-950">
                  Expiry notifications
                </Label>
                <Textarea
                  id="expiry-template"
                  value={form.expiry_notification_email_template}
                  onChange={(event) => setField("expiry_notification_email_template", event.target.value)}
                  className="min-h-40 bg-white"
                />
              </div>
              <div className="rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-4 text-sm text-slate-500 xl:col-span-2">
                Supported variables: <code>{"{company_name}"}</code>, <code>{"{full_name}"}</code>, <code>{"{candidate_name}"}</code>, <code>{"{reason}"}</code>, <code>{"{message}"}</code>, <code>{"{title}"}</code>.
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  )
}
