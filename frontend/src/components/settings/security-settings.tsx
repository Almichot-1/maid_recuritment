"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { formatDistanceToNow } from "date-fns"
import { AlertCircle, CheckCircle2, Loader2, LogOut, ShieldCheck, Smartphone, Trash2 } from "lucide-react"

import { useChangePassword, useLogoutAllDevices, PasswordChangeData } from "@/hooks/use-settings"
import { useBrowserSessions } from "@/hooks/use-browser-sessions"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

const passwordSchema = z.object({
  current_password: z.string().min(1, "Current password is required"),
  new_password: z.string().min(8, "Password must be at least 8 characters"),
  confirm_password: z.string(),
}).refine((data) => data.new_password === data.confirm_password, {
  message: "Passwords don't match",
  path: ["confirm_password"],
})

export function SecuritySettings() {
  const { mutate: changePassword, isPending: isChangingPassword } = useChangePassword()
  const { mutate: logoutAllSessions, isPending: isLoggingOut } = useLogoutAllDevices()
  const { sessions, currentSessionID, removeSession } = useBrowserSessions()

  const form = useForm<PasswordChangeData & { confirm_password: string }>({
    resolver: zodResolver(passwordSchema),
    defaultValues: {
      current_password: "",
      new_password: "",
      confirm_password: "",
    },
  })

  const newPassword = form.watch("new_password")

  const getPasswordStrength = (password: string) => {
    if (!password) return { strength: 0, label: "", color: "" }
    
    let strength = 0
    if (password.length >= 8) strength++
    if (password.length >= 12) strength++
    if (/[a-z]/.test(password) && /[A-Z]/.test(password)) strength++
    if (/\d/.test(password)) strength++
    if (/[^a-zA-Z0-9]/.test(password)) strength++

    if (strength <= 2) return { strength, label: "Weak", color: "bg-red-500" }
    if (strength <= 3) return { strength, label: "Fair", color: "bg-orange-500" }
    if (strength <= 4) return { strength, label: "Good", color: "bg-yellow-500" }
    return { strength, label: "Strong", color: "bg-green-500" }
  }

  const passwordStrength = getPasswordStrength(newPassword)

  const onSubmit = (data: PasswordChangeData & { confirm_password: string }) => {
    changePassword({
      current_password: data.current_password,
      new_password: data.new_password,
    }, {
      onSuccess: () => {
        form.reset()
      },
    })
  }

  return (
    <div className="space-y-6">
      <div className="rounded-2xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-900 dark:border-emerald-900/60 dark:bg-emerald-950/20 dark:text-emerald-200">
        Password changes save through the live API, and the session list below now tracks the active browser sessions you have opened on this device.
      </div>

      <Card className="overflow-hidden border-border/70 shadow-sm">
        <CardHeader>
          <CardTitle>Change Password</CardTitle>
          <CardDescription>
            Update your password to keep your account secure.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              {/* Current Password */}
              <FormField
                control={form.control}
                name="current_password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Current Password</FormLabel>
                    <FormControl>
                      <Input type="password" placeholder="Enter current password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* New Password */}
              <FormField
                control={form.control}
                name="new_password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>New Password</FormLabel>
                    <FormControl>
                      <Input type="password" placeholder="Enter new password" {...field} />
                    </FormControl>
                    {newPassword && (
                      <div className="space-y-2 mt-2">
                        <div className="flex items-center gap-2">
                          <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                            <div
                              className={cn("h-full transition-all", passwordStrength.color)}
                              style={{ width: `${(passwordStrength.strength / 5) * 100}%` }}
                            />
                          </div>
                          <span className="text-xs font-medium">{passwordStrength.label}</span>
                        </div>
                        <p className="text-xs text-muted-foreground">
                          Use 8+ characters with a mix of letters, numbers & symbols
                        </p>
                      </div>
                    )}
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* Confirm Password */}
              <FormField
                control={form.control}
                name="confirm_password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Confirm New Password</FormLabel>
                    <FormControl>
                      <Input type="password" placeholder="Confirm new password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* Submit Button */}
              <Button type="submit" disabled={isChangingPassword}>
                {isChangingPassword && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Update Password
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>

      <Card className="overflow-hidden border-border/70 shadow-sm">
        <CardHeader>
          <CardTitle>Active Sessions</CardTitle>
          <CardDescription>
            Review the sessions this browser has remembered for your account and clear the ones you no longer want.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-3">
            {sessions.length === 0 ? (
              <div className="rounded-2xl border border-dashed border-border/70 bg-muted/20 p-4 text-sm text-muted-foreground">
                No remembered browser sessions yet.
              </div>
            ) : (
              sessions.map((session) => {
                const isCurrent = session.id === currentSessionID
                const Icon = isCurrent ? ShieldCheck : Smartphone

                return (
                  <div
                    key={session.id}
                    className="flex flex-col gap-4 rounded-2xl border border-border/70 bg-muted/20 p-4 sm:flex-row sm:items-center sm:justify-between"
                  >
                    <div className="flex items-start gap-3">
                      <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-background shadow-sm">
                        <Icon className="h-5 w-5 text-primary" />
                      </div>
                      <div className="space-y-1">
                        <div className="flex flex-wrap items-center gap-2">
                          <p className="font-medium">{session.device_label}</p>
                          {isCurrent ? (
                            <Badge className="bg-green-500 text-white hover:bg-green-600">
                              <CheckCircle2 className="mr-1 h-3 w-3" />
                              Current
                            </Badge>
                          ) : null}
                        </div>
                        <p className="text-sm text-muted-foreground">
                          {session.browser_name} • {session.os_name}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          Last active {formatDistanceToNow(new Date(session.last_active_at), { addSuffix: true })}
                        </p>
                      </div>
                    </div>

                    {!isCurrent ? (
                      <Button
                        type="button"
                        variant="ghost"
                        className="self-start text-destructive hover:text-destructive sm:self-center"
                        onClick={() => removeSession(session.id)}
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        Remove
                      </Button>
                    ) : null}
                  </div>
                )
              })
            )}
          </div>

          <Separator />

          <div className="space-y-3">
            <div className="flex items-start gap-3 rounded-2xl border border-amber-200 bg-amber-50 p-3 dark:border-amber-800 dark:bg-amber-950/20">
              <AlertCircle className="mt-0.5 h-5 w-5 shrink-0 text-amber-600" />
              <div className="text-sm text-amber-800 dark:text-amber-200">
                <p className="mb-1 font-medium">Clear this browser</p>
                <p className="text-xs">
                  This signs out the current browser session and removes the remembered session history stored on this browser.
                </p>
              </div>
            </div>
            <Button
              variant="outline"
              className="w-full sm:w-auto"
              onClick={() => logoutAllSessions()}
              disabled={isLoggingOut}
            >
              {isLoggingOut ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <LogOut className="mr-2 h-4 w-4" />
              )}
              Sign Out and Clear Sessions
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
