"use client"

import * as React from "react"

import { LocaleSwitcher } from "@/components/shared/locale-switcher"
import { Logo } from "@/components/shared/logo"

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto flex min-h-screen w-full max-w-lg flex-col px-4 py-6 sm:px-6">
        <div className="flex items-center justify-between border-b border-border pb-4">
          <Logo size="sm" />
          <LocaleSwitcher compact />
        </div>
        <div className="flex flex-1 flex-col items-center justify-center py-10">{children}</div>
      </div>
    </div>
  )
}
