import * as React from "react"

import { ThemeToggle } from "@/components/shared/theme-toggle"

export default function AdminAuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative min-h-screen bg-[radial-gradient(circle_at_top_left,rgba(251,191,36,0.18),transparent_22%),radial-gradient(circle_at_bottom_right,rgba(14,165,233,0.12),transparent_24%),linear-gradient(180deg,rgba(255,255,255,0.98),rgba(248,250,252,0.96),rgba(226,232,240,0.96))] px-4 py-10 text-slate-950 dark:bg-[radial-gradient(circle_at_top_left,#2f2413_0%,#0f172a_38%,#020617_100%)] dark:text-white">
      <div className="absolute right-4 top-4 z-10">
        <ThemeToggle className="border-slate-200 bg-white/85 text-slate-700 hover:bg-white dark:border-white/10 dark:bg-slate-950/60 dark:text-white dark:hover:bg-slate-900" />
      </div>
      <div className="mx-auto grid min-h-[calc(100vh-5rem)] max-w-6xl items-center gap-10 lg:grid-cols-[1.1fr_0.9fr]">
        <div className="hidden space-y-6 lg:block">
          <div className="inline-flex items-center rounded-full border border-amber-400/30 bg-amber-400/10 px-4 py-2 text-xs font-semibold uppercase tracking-[0.28em] text-amber-700 dark:text-amber-200">
            Platform Operator Access
          </div>
          <div className="space-y-4">
            <h1 className="max-w-2xl text-5xl font-semibold leading-tight tracking-tight">
              Admin Portal for full oversight across agencies, candidates, approvals, and audit history.
            </h1>
            <p className="max-w-2xl text-lg text-slate-600 dark:text-slate-300">
              This workspace is isolated from the agency portal and is reserved for platform operators handling approvals, compliance, and operational monitoring.
            </p>
          </div>
          <div className="grid max-w-2xl gap-4 sm:grid-cols-3">
            <div className="rounded-3xl border border-slate-200 bg-white/72 p-5 shadow-[0_18px_40px_-34px_rgba(15,23,42,0.24)] dark:border-white/10 dark:bg-white/5 dark:shadow-none">
              <p className="text-sm font-semibold text-slate-950 dark:text-white">Separate admin auth</p>
              <p className="mt-2 text-sm text-slate-600 dark:text-slate-300">Dedicated admin credentials, shorter sessions, and MFA at login.</p>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white/72 p-5 shadow-[0_18px_40px_-34px_rgba(15,23,42,0.24)] dark:border-white/10 dark:bg-white/5 dark:shadow-none">
              <p className="text-sm font-semibold text-slate-950 dark:text-white">Audit first</p>
              <p className="mt-2 text-sm text-slate-600 dark:text-slate-300">Every sensitive action is tracked to the admin account and timestamp.</p>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white/72 p-5 shadow-[0_18px_40px_-34px_rgba(15,23,42,0.24)] dark:border-white/10 dark:bg-white/5 dark:shadow-none">
              <p className="text-sm font-semibold text-slate-950 dark:text-white">Approval gate</p>
              <p className="mt-2 text-sm text-slate-600 dark:text-slate-300">New Ethiopian and Foreign agencies stay blocked until reviewed.</p>
            </div>
          </div>
        </div>
        <div className="flex justify-center lg:justify-end">{children}</div>
      </div>
    </div>
  )
}
