"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { AlertCircle, CheckCircle2, Loader2, LogOut, Monitor } from "lucide-react"

import { useChangePassword, useLogoutAllDevices, PasswordChangeData } from "@/hooks/use-settings"
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
  const { isPending: isLoggingOut } = useLogoutAllDevices()

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
      <div className="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900 dark:border-amber-900/60 dark:bg-amber-950/20 dark:text-amber-200">
        Password changes are now backed by the API. Logging out of all devices is still unavailable because the current JWT auth flow has no token revocation support yet.
      </div>

      {/* Change Password */}
      <Card>
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

      {/* Active Sessions */}
      <Card>
        <CardHeader>
          <CardTitle>Active Sessions</CardTitle>
          <CardDescription>
            Manage your active sessions across different devices.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Current Session */}
          <div className="flex items-center justify-between p-4 bg-muted/50 rounded-lg border">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-green-100 dark:bg-green-950/30">
                <Monitor className="h-5 w-5 text-green-600 dark:text-green-400" />
              </div>
              <div>
                <p className="font-medium">Current Session</p>
                <p className="text-sm text-muted-foreground">This device - Active now</p>
              </div>
            </div>
            <Badge className="bg-green-500 hover:bg-green-600 text-white">
              <CheckCircle2 className="h-3 w-3 mr-1" />
              Active
            </Badge>
          </div>

          <Separator />

          {/* Logout All Devices */}
          <div className="space-y-3">
            <div className="flex items-start gap-3 p-3 bg-amber-50 dark:bg-amber-950/20 rounded-lg border border-amber-200 dark:border-amber-800">
              <AlertCircle className="h-5 w-5 text-amber-600 shrink-0 mt-0.5" />
              <div className="text-sm text-amber-800 dark:text-amber-200">
                <p className="font-medium mb-1">Security Action</p>
                <p className="text-xs">
                  Logging out of all devices will end all active sessions. You will need to log in again on this device.
                </p>
              </div>
            </div>
            <Button
              variant="outline"
              disabled
              className="w-full sm:w-auto"
            >
              {isLoggingOut ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <LogOut className="mr-2 h-4 w-4" />
              )}
              Coming Soon
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
