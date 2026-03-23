"use client"

import * as React from "react"
import { ChevronRight, Home } from "lucide-react"
import Link from "next/link"

import { PageHeader } from "@/components/layout/page-header"
import { ProfileSettings } from "@/components/settings/profile-settings"
import { SecuritySettings } from "@/components/settings/security-settings"
import { PreferencesSettings } from "@/components/settings/preferences-settings"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

export default function SettingsPage() {
  const breadcrumbs = (
    <nav className="flex items-center text-sm font-medium text-muted-foreground mb-6">
      <Link href="/dashboard" className="transition-all hover:text-primary flex items-center">
        <Home className="mr-1.5 h-4 w-4" />
        Dashboard
      </Link>
      <ChevronRight className="h-4 w-4 mx-1 opacity-50" />
      <span className="text-foreground font-semibold">Settings</span>
    </nav>
  )

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 pb-10">
      {breadcrumbs}

      <PageHeader
        heading="Settings"
        text="Manage your account settings and preferences."
      />

      <Tabs defaultValue="profile" className="space-y-6">
        <TabsList className="grid w-full grid-cols-3 h-auto p-1">
          <TabsTrigger value="profile" className="py-2">
            Profile
          </TabsTrigger>
          <TabsTrigger value="security" className="py-2">
            Security
          </TabsTrigger>
          <TabsTrigger value="preferences" className="py-2">
            Preferences
          </TabsTrigger>
        </TabsList>

        <TabsContent value="profile" className="space-y-6">
          <ProfileSettings />
        </TabsContent>

        <TabsContent value="security" className="space-y-6">
          <SecuritySettings />
        </TabsContent>

        <TabsContent value="preferences" className="space-y-6">
          <PreferencesSettings />
        </TabsContent>
      </Tabs>
    </div>
  )
}
