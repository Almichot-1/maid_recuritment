"use client"

import * as React from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import {
  BarChart3,
  BellRing,
  Building2,
  CheckCheck,
  ClipboardList,
  FileClock,
  Flag,
  Link2,
  LogOut,
  Menu,
  Settings2,
  Shield,
  UserCog,
  Users,
} from "lucide-react"

import { AdminRole } from "@/types"
import { useAdminLogout, useCurrentAdmin } from "@/hooks/use-admin-auth"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet"
import { cn } from "@/lib/utils"

type AdminNavItem = {
  href: string
  label: string
  icon: React.ElementType
  roles?: AdminRole[]
}

const navItems: AdminNavItem[] = [
  { href: "/admin/dashboard", label: "Dashboard", icon: BarChart3 },
  { href: "/admin/agencies/pending", label: "Pending Approvals", icon: FileClock },
  { href: "/admin/agencies", label: "Agencies", icon: Building2 },
  { href: "/admin/pairings", label: "Pair Workspaces", icon: Link2 },
  { href: "/admin/candidates", label: "Candidates", icon: Users },
  { href: "/admin/selections", label: "Selections", icon: CheckCheck },
  { href: "/admin/reports", label: "Reports", icon: ClipboardList },
  { href: "/admin/settings", label: "Platform Settings", icon: Settings2, roles: [AdminRole.SUPER_ADMIN] },
  { href: "/admin/audit-logs", label: "Audit Logs", icon: BellRing },
  { href: "/admin/admins", label: "Admin Management", icon: UserCog, roles: [AdminRole.SUPER_ADMIN] },
]

function isVisible(item: AdminNavItem, role?: AdminRole) {
  if (!item.roles?.length) {
    return true
  }
  return role ? item.roles.includes(role) : false
}

function AdminNavLinks({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname()
  const { admin } = useCurrentAdmin()

  return (
    <nav className="space-y-1">
      {navItems
        .filter((item) => isVisible(item, admin?.role))
        .map((item) => {
          const active = pathname === item.href || (item.href !== "/admin/dashboard" && pathname.startsWith(item.href))
          return (
            <Link
              key={item.href}
              href={item.href}
              onClick={onNavigate}
              className={cn(
                "group flex items-center gap-3 rounded-2xl px-4 py-3 text-sm font-medium transition-all",
                active
                  ? "bg-amber-400 text-slate-950 shadow-lg shadow-amber-500/20"
                  : "text-slate-300 hover:bg-slate-900 hover:text-white"
              )}
            >
              <item.icon className="h-4 w-4 shrink-0" />
              <span>{item.label}</span>
            </Link>
          )
        })}
    </nav>
  )
}

export function AdminShell({ children }: { children: React.ReactNode }) {
  const { admin } = useCurrentAdmin()
  const logout = useAdminLogout()

  return (
    <div className="min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top_left,_rgba(251,191,36,0.16),_transparent_24%),radial-gradient(circle_at_bottom_right,_rgba(14,165,233,0.1),_transparent_22%),linear-gradient(180deg,#fff9ef_0%,#f3f6fb_52%,#eff3f9_100%)] text-slate-950">
      <div className="pointer-events-none absolute inset-x-0 top-0 h-40 bg-[linear-gradient(180deg,rgba(15,23,42,0.06),transparent)]" />
      <div className="relative mx-auto flex min-h-screen max-w-[1760px]">
        <aside className="hidden w-[290px] flex-col border-r border-slate-800 bg-slate-950 px-5 py-6 text-slate-50 lg:flex">
          <div className="rounded-3xl border border-slate-800 bg-[radial-gradient(circle_at_top_left,_rgba(251,191,36,0.22),_transparent_30%),linear-gradient(135deg,#0f172a,#111827_55%,#020617)] p-5 shadow-2xl shadow-black/25">
            <div className="flex items-center gap-3">
              <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-amber-400 text-slate-950">
                <Shield className="h-5 w-5" />
              </div>
              <div>
                <div className="text-sm font-semibold uppercase tracking-[0.2em] text-amber-300">Operator Console</div>
                <p className="text-xs text-slate-400">Platform administration portal</p>
              </div>
            </div>
          </div>

          <div className="mt-6">
            <AdminNavLinks />
          </div>

          <div className="mt-auto rounded-3xl border border-slate-800 bg-slate-900/80 p-4 shadow-xl shadow-black/20">
            <div className="flex items-start gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-slate-800 text-sm font-semibold text-amber-300">
                {admin?.full_name?.charAt(0) || "A"}
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-semibold text-white">{admin?.full_name || "Admin"}</p>
                <p className="truncate text-xs text-slate-400">{admin?.email}</p>
              </div>
            </div>
            <div className="mt-3 flex items-center justify-between gap-3">
              <Badge className="rounded-full bg-slate-800 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-amber-300 hover:bg-slate-800">
                {(admin?.role || "admin").replace("_", " ")}
              </Badge>
              <Button variant="ghost" size="icon" className="text-slate-400 hover:bg-slate-800 hover:text-white" onClick={logout}>
                <LogOut className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </aside>

        <div className="flex min-h-screen flex-1 flex-col">
          <header className="sticky top-0 z-30 border-b border-slate-200/70 bg-white/78 backdrop-blur-xl">
            <div className="flex items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
              <div className="flex items-center gap-3">
                <Sheet>
                  <SheetTrigger asChild>
                    <Button variant="outline" size="icon" className="lg:hidden">
                      <Menu className="h-4 w-4" />
                    </Button>
                  </SheetTrigger>
                  <SheetContent side="left" className="w-[300px] border-slate-800 bg-slate-950 px-5 py-6 text-slate-50">
                    <div className="mb-6 flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-amber-400 text-slate-950">
                        <Flag className="h-5 w-5" />
                      </div>
                      <div>
                        <div className="text-sm font-semibold uppercase tracking-[0.2em] text-amber-300">Admin Mode</div>
                        <p className="text-xs text-slate-400">Platform operators only</p>
                      </div>
                    </div>
                    <AdminNavLinks />
                  </SheetContent>
                </Sheet>
                <div className="space-y-1">
                  <div className="text-xs font-semibold uppercase tracking-[0.24em] text-amber-700">Control Center</div>
                  <p className="text-sm text-slate-500">Oversight across Ethiopian and Foreign agencies</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Badge variant="outline" className="hidden rounded-full border-amber-200 bg-amber-50 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.2em] text-amber-700 sm:inline-flex">
                  Admin Mode
                </Badge>
                <div className="hidden text-right sm:block">
                  <p className="text-sm font-semibold text-slate-900">{admin?.full_name}</p>
                  <p className="text-xs text-slate-500">{admin?.role?.replace("_", " ")}</p>
                </div>
                <Button variant="outline" size="sm" className="gap-2" onClick={logout}>
                  <LogOut className="h-4 w-4" />
                  <span className="hidden sm:inline">Logout</span>
                </Button>
              </div>
            </div>
          </header>

          <main className="flex-1 px-4 py-6 sm:px-6 lg:px-8">
            <div className="space-y-6">{children}</div>
          </main>
        </div>
      </div>
    </div>
  )
}
