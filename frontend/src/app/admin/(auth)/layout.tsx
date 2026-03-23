import * as React from "react"

export default function AdminAuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,#2f2413_0%,#0f172a_38%,#020617_100%)] px-4 py-10 text-white">
      <div className="mx-auto grid min-h-[calc(100vh-5rem)] max-w-6xl items-center gap-10 lg:grid-cols-[1.1fr_0.9fr]">
        <div className="hidden space-y-6 lg:block">
          <div className="inline-flex items-center rounded-full border border-amber-400/30 bg-amber-400/10 px-4 py-2 text-xs font-semibold uppercase tracking-[0.28em] text-amber-200">
            Platform Operator Access
          </div>
          <div className="space-y-4">
            <h1 className="max-w-2xl text-5xl font-semibold leading-tight tracking-tight">
              Admin Portal for full oversight across agencies, candidates, approvals, and audit history.
            </h1>
            <p className="max-w-2xl text-lg text-slate-300">
              This workspace is isolated from the agency portal and is reserved for platform operators handling approvals, compliance, and operational monitoring.
            </p>
          </div>
          <div className="grid max-w-2xl gap-4 sm:grid-cols-3">
            <div className="rounded-3xl border border-white/10 bg-white/5 p-5">
              <p className="text-sm font-semibold text-white">Separate admin auth</p>
              <p className="mt-2 text-sm text-slate-300">Dedicated admin credentials, shorter sessions, and MFA at login.</p>
            </div>
            <div className="rounded-3xl border border-white/10 bg-white/5 p-5">
              <p className="text-sm font-semibold text-white">Audit first</p>
              <p className="mt-2 text-sm text-slate-300">Every sensitive action is tracked to the admin account and timestamp.</p>
            </div>
            <div className="rounded-3xl border border-white/10 bg-white/5 p-5">
              <p className="text-sm font-semibold text-white">Approval gate</p>
              <p className="mt-2 text-sm text-slate-300">New Ethiopian and Foreign agencies stay blocked until reviewed.</p>
            </div>
          </div>
        </div>
        <div className="flex justify-center lg:justify-end">{children}</div>
      </div>
    </div>
  )
}
