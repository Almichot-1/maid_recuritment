"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Camera, CheckCircle2, ImagePlus, Loader2, Trash2, Upload, User } from "lucide-react"
import { toast } from "sonner"

import { useCurrentUser } from "@/hooks/use-auth"
import { useAgencyBranding } from "@/hooks/use-agency-branding"
import { useProfileAvatar } from "@/hooks/use-profile-avatar"
import { useUpdateProfile, ProfileUpdateData } from "@/hooks/use-settings"
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
import { Label } from "@/components/ui/label"
import { Badge } from "@/components/ui/badge"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"

const profileSchema = z.object({
  full_name: z.string().min(2, "Name must be at least 2 characters"),
  company_name: z.string().min(2, "Company name must be at least 2 characters"),
})

export function ProfileSettings() {
  const { user } = useCurrentUser()
  const avatarInputRef = React.useRef<HTMLInputElement | null>(null)
  const logoInputRef = React.useRef<HTMLInputElement | null>(null)
  const { mutate: updateProfile, isPending } = useUpdateProfile()
  const { hasLogo, logoDataURL, saveLogo, clearLogo } = useAgencyBranding()
  const { avatarDataURL, hasAvatar, saveAvatar, clearAvatar } = useProfileAvatar()

  const form = useForm<ProfileUpdateData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      full_name: user?.full_name || "",
      company_name: user?.company_name || "",
    },
  })

  React.useEffect(() => {
    if (!user) {
      return
    }

    form.reset({
      full_name: user.full_name || "",
      company_name: user.company_name || "",
    })
  }, [form, user])

  const onSubmit = (data: ProfileUpdateData) => {
    updateProfile(data)
  }

  const handleLogoChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) {
      return
    }

    try {
      await saveLogo(file)
      toast.success("Agency logo saved for future candidate submissions on this device")
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Failed to save agency logo")
    } finally {
      event.target.value = ""
    }
  }

  if (!user) return null

  return (
    <Card className="overflow-hidden border-border/70 shadow-sm">
      <CardHeader>
        <CardTitle>Profile Information</CardTitle>
        <CardDescription>
          Update your personal information, profile photo, and reusable agency branding.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
          <div className="rounded-3xl border border-border/70 bg-muted/25 p-5">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
              <Avatar className="h-24 w-24 border border-border/70 shadow-sm">
                <AvatarImage
                  src={
                    avatarDataURL ||
                    `https://api.dicebear.com/7.x/initials/svg?seed=${user.full_name}`
                  }
                  alt={user.full_name}
                />
                <AvatarFallback>
                  <User className="h-12 w-12" />
                </AvatarFallback>
              </Avatar>

              <div className="flex-1 space-y-3">
                <div>
                  <p className="text-sm font-semibold">Profile photo</p>
                  <p className="text-sm text-muted-foreground">
                    Saved on this device and reused across the app immediately.
                  </p>
                </div>

                <div className="flex flex-wrap gap-2">
                  <Button type="button" variant="outline" onClick={() => avatarInputRef.current?.click()}>
                    <Camera className="mr-2 h-4 w-4" />
                    {hasAvatar ? "Change Photo" : "Upload Photo"}
                  </Button>
                  {hasAvatar ? (
                    <Button
                      type="button"
                      variant="ghost"
                      className="text-destructive hover:text-destructive"
                      onClick={() => {
                        clearAvatar()
                        toast.success("Profile photo removed")
                      }}
                    >
                      <Trash2 className="mr-2 h-4 w-4" />
                      Remove
                    </Button>
                  ) : null}
                </div>

                <p className="text-xs text-muted-foreground">
                  PNG, JPG, or WEBP. Max size 2 MB.
                </p>
              </div>
            </div>

            <input
              ref={avatarInputRef}
              type="file"
              accept=".jpg,.jpeg,.png,.webp,image/jpeg,image/png,image/webp"
              className="hidden"
              onChange={async (event) => {
                const file = event.target.files?.[0]
                if (!file) {
                  return
                }

                try {
                  await saveAvatar(file)
                  toast.success("Profile photo saved on this device")
                } catch (error) {
                  toast.error(error instanceof Error ? error.message : "Failed to save profile photo")
                } finally {
                  event.target.value = ""
                }
              }}
            />
          </div>

          <div className="rounded-3xl border border-border/70 bg-muted/25 p-5">
            <div className="flex flex-col gap-5 lg:justify-between">
              <div className="space-y-2">
                <p className="text-sm font-semibold">Agency logo</p>
                <p className="text-sm text-muted-foreground">
                  Upload this once and we will reuse it while you add candidates on this device.
                </p>
              </div>

              <div className="flex flex-col items-start gap-4 sm:flex-row sm:items-center">
                <div className="flex h-20 w-20 items-center justify-center overflow-hidden rounded-2xl border bg-background shadow-sm">
                  {hasLogo ? (
                    <img src={logoDataURL} alt={`${user.company_name || user.full_name} logo`} className="h-full w-full object-cover" />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-primary/10 via-background to-amber-100 text-primary">
                      <ImagePlus className="h-8 w-8" />
                    </div>
                  )}
                </div>

                <div className="flex flex-wrap gap-2">
                  <Button type="button" variant="outline" onClick={() => logoInputRef.current?.click()}>
                    <Upload className="mr-2 h-4 w-4" />
                    {hasLogo ? "Change Logo" : "Upload Logo"}
                  </Button>
                  {hasLogo ? (
                    <Button
                      type="button"
                      variant="ghost"
                      className="text-destructive hover:text-destructive"
                      onClick={() => {
                        clearLogo()
                        toast.success("Agency logo removed")
                      }}
                    >
                      <Trash2 className="mr-2 h-4 w-4" />
                      Remove
                    </Button>
                  ) : null}
                </div>
              </div>
            </div>

            <input
              ref={logoInputRef}
              type="file"
              accept=".jpg,.jpeg,.png,image/jpeg,image/png"
              className="hidden"
              onChange={handleLogoChange}
            />
          </div>
        </div>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            {/* Full Name */}
            <FormField
              control={form.control}
              name="full_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Full Name</FormLabel>
                  <FormControl>
                    <Input placeholder="John Doe" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Email (Read-only) */}
            <div className="space-y-2">
              <Label>Email Address</Label>
              <div className="flex items-center gap-2">
                <Input value={user.email} disabled className="flex-1" />
                <Badge className="bg-green-500 hover:bg-green-600 text-white">
                  <CheckCircle2 className="h-3 w-3 mr-1" />
                  Verified
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground">
                Your email address is verified and cannot be changed.
              </p>
            </div>

            {/* Company Name */}
            <FormField
              control={form.control}
              name="company_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Company Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Acme Inc." {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {/* Role (Read-only) */}
            <div className="space-y-2">
              <Label>Role</Label>
              <div>
                <Badge variant="secondary" className="text-sm">
                  {user.role === "ethiopian_agent" ? "Ethiopian Agent" : "Foreign Agent"}
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground">
                Your role is assigned by the system administrator.
              </p>
            </div>

            {/* Save Button */}
            <Button type="submit" disabled={isPending}>
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Save Changes
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  )
}
