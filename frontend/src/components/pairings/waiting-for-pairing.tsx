"use client"

import * as React from "react"
import { Clock3, ShieldCheck, UsersRound } from "lucide-react"

import { PageHeader } from "@/components/layout/page-header"
import { Card, CardContent } from "@/components/ui/card"

export function WaitingForPairingState() {
  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader
        heading="Waiting for Partner Assignment"
        text="Your agency account is approved, but an admin still needs to connect you to at least one partner workspace before candidate, selection, and tracking activity can begin."
      />

      <Card className="overflow-hidden border-0 bg-[radial-gradient(circle_at_top_left,_rgba(14,165,233,0.16),_transparent_28%),radial-gradient(circle_at_top_right,_rgba(16,185,129,0.16),_transparent_24%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.98))] text-white shadow-xl">
        <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_320px]">
          <div className="space-y-4">
            <div className="inline-flex w-fit items-center rounded-full border border-white/10 bg-white/10 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.24em] text-sky-100">
              Private Workspaces
            </div>
            <div className="space-y-2">
              <h2 className="text-3xl font-semibold tracking-tight">We are holding your workspace until the pairing is approved</h2>
              <p className="max-w-2xl text-sm text-slate-200/90">
                Each Ethiopian and foreign relationship now runs inside its own private workspace. That keeps candidate visibility controlled and prevents agencies from seeing each other&apos;s pipelines.
              </p>
            </div>
          </div>

          <div className="grid gap-3">
            <StatusTile
              icon={<ShieldCheck className="h-5 w-5" />}
              title="Account approved"
              description="Your login is active and ready."
            />
            <StatusTile
              icon={<UsersRound className="h-5 w-5" />}
              title="Workspace pending"
              description="An admin still needs to assign your partner relationship."
            />
            <StatusTile
              icon={<Clock3 className="h-5 w-5" />}
              title="Next step"
              description="Once the workspace is created, your dashboard unlocks automatically."
            />
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function StatusTile({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode
  title: string
  description: string
}) {
  return (
    <div className="rounded-2xl border border-white/10 bg-white/10 p-4 backdrop-blur">
      <div className="flex items-center gap-3">
        <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-white/10 text-white">{icon}</div>
        <div>
          <p className="font-semibold text-white">{title}</p>
          <p className="text-sm text-slate-200/80">{description}</p>
        </div>
      </div>
    </div>
  )
}
