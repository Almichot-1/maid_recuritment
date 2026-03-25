import * as React from "react"
import { ThemeToggle } from "@/components/shared/theme-toggle"

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative flex min-h-screen items-center justify-center bg-background p-4 transition-colors duration-300">
      <div className="absolute right-4 top-4">
        <ThemeToggle />
      </div>
      {children}
    </div>
  )
}
