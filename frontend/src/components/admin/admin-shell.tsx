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

import { ThemeToggle } from "@/components/shared/theme-toggle"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet"
import { useAdminLogout, useCurrentAdmin } from "@/hooks/use-admin-auth"
import { cn } from "@/lib/utils"
import { AdminRole } from "@/types"

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

const controlButtonClass =
  "border-slate-200 bg-white/85 text-slate-700 hover:bg-white hover:text-slate-950 dark:border-slate-700 dark:bg-slate-900/90 dark:text-slate-100 dark:hover:bg-slate-800 dark:hover:text-white"

function isVisible(item: AdminNavItem, role?: AdminRole) {
  if (!item.roles?.length) {
    return true
  }
  return role ? item.roles.includes(role) : false
}

function matchesAdminPath(pathname: string, href: string) {
  return pathname === href || pathname.startsWith(`${href}/`)
}

function AdminNavLinks({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname()
  const { admin } = useCurrentAdmin()
  const visibleItems = navItems.filter((item) => isVisible(item, admin?.role))
  const activeHref =
    [...visibleItems]
      .sort((left, right) => right.href.length - left.href.length)
      .find((item) => matchesAdminPath(pathname, item.href))?.href ?? ""

  return (
    <nav className="space-y-1">
      {visibleItems.map((item) => {
          const active = item.href === activeHref

          return (
            <Link
              key={item.href}
              href={item.href}
              onClick={onNavigate}
              className={cn(
                "group flex items-center gap-3 rounded-2xl border px-4 py-3 text-sm font-medium transition-all",
                active
                  ? "border-amber-300/80 bg-amber-300 text-slate-950 shadow-lg shadow-amber-500/15 dark:border-amber-300/40 dark:bg-amber-300 dark:shadow-amber-500/20"
                  : "border-transparent text-slate-600 hover:border-slate-200 hover:bg-white/85 hover:text-slate-950 dark:text-slate-300 dark:hover:border-slate-800 dark:hover:bg-slate-900/80 dark:hover:text-white"
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
    <div className="min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top_left,_rgba(251,191,36,0.12),_transparent_22%),radial-gradient(circle_at_bottom_right,_rgba(14,165,233,0.10),_transparent_20%),linear-gradient(180deg,rgba(248,250,252,0.98),rgba(241,245,249,0.98),rgba(226,232,240,0.98))] text-slate-950 dark:bg-[radial-gradient(circle_at_top_left,_rgba(251,191,36,0.14),_transparent_22%),radial-gradient(circle_at_bottom_right,_rgba(14,165,233,0.12),_transparent_20%),linear-gradient(180deg,#020617_0%,#0f172a_46%,#111827_100%)] dark:text-slate-100">
      <div className="pointer-events-none absolute inset-x-0 top-0 h-40 bg-[linear-gradient(180deg,rgba(255,255,255,0.72),transparent)] dark:bg-[linear-gradient(180deg,rgba(255,255,255,0.04),transparent)]" />

      <div className="relative mx-auto flex min-h-screen max-w-[1760px]">
        <aside className="hidden w-[290px] flex-col border-r border-slate-200/80 bg-white/65 px-5 py-6 text-slate-950 backdrop-blur-xl dark:border-slate-800 dark:bg-slate-950/90 dark:text-slate-50 lg:flex">
          <div className="rounded-3xl border border-slate-200 bg-[radial-gradient(circle_at_top_left,_rgba(251,191,36,0.22),_transparent_30%),linear-gradient(135deg,rgba(255,255,255,0.98),rgba(248,250,252,0.94),rgba(226,232,240,0.92))] p-5 shadow-[0_24px_48px_-32px_rgba(15,23,42,0.24)] dark:border-slate-800 dark:bg-[radial-gradient(circle_at_top_left,_rgba(251,191,36,0.22),_transparent_30%),linear-gradient(135deg,#0f172a,#111827_55%,#020617)] dark:shadow-2xl dark:shadow-black/25">
            <div className="flex items-center gap-3">
              <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-amber-400 text-slate-950">
                <Shield className="h-5 w-5" />
              </div>
              <div>
                <div className="text-sm font-semibold uppercase tracking-[0.2em] text-amber-700 dark:text-amber-300">Operator Console</div>
                <p className="text-xs text-slate-500 dark:text-slate-400">Platform administration portal</p>
              </div>
            </div>
          </div>

          <div className="mt-6">
            <AdminNavLinks />
          </div>

          <div className="mt-auto rounded-3xl border border-slate-200 bg-white/82 p-4 shadow-[0_20px_40px_-30px_rgba(15,23,42,0.22)] dark:border-slate-800 dark:bg-slate-900/80 dark:shadow-xl dark:shadow-black/20">
            <div className="flex items-start gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-slate-200 text-sm font-semibold text-amber-700 dark:bg-slate-800 dark:text-amber-300">
                {admin?.full_name?.charAt(0) || "A"}
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-semibold text-slate-950 dark:text-white">{admin?.full_name || "Admin"}</p>
                <p className="truncate text-xs text-slate-500 dark:text-slate-400">{admin?.email}</p>
              </div>
            </div>
            <div className="mt-3 flex items-center justify-between gap-3">
              <Badge className="rounded-full bg-amber-100 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-amber-700 hover:bg-amber-100 dark:bg-slate-800 dark:text-amber-300 dark:hover:bg-slate-800">
                {(admin?.role || "admin").replace("_", " ")}
              </Badge>
              <Button variant="ghost" size="icon" className="text-slate-500 hover:bg-slate-100 hover:text-slate-950 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-white" onClick={logout}>
                <LogOut className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </aside>

        <div className="flex min-h-screen flex-1 flex-col">
          <header className="sticky top-0 z-30 border-b border-slate-200/80 bg-white/70 shadow-[0_16px_38px_-30px_rgba(15,23,42,0.18)] backdrop-blur-xl dark:border-slate-800/80 dark:bg-[linear-gradient(180deg,rgba(2,6,23,0.88),rgba(15,23,42,0.8))] dark:shadow-[0_16px_38px_-28px_rgba(2,6,23,0.75)]">
            <div className="flex items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
              <div className="flex items-center gap-3">
                <Sheet>
                  <SheetTrigger asChild>
                    <Button variant="outline" size="icon" className={cn(controlButtonClass, "lg:hidden")}>
                      <Menu className="h-4 w-4" />
                    </Button>
                  </SheetTrigger>
                  <SheetContent side="left" className="w-[300px] border-slate-200 bg-white px-5 py-6 text-slate-950 dark:border-slate-800 dark:bg-slate-950 dark:text-slate-50">
                    <div className="mb-6 flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-amber-400 text-slate-950">
                        <Flag className="h-5 w-5" />
                      </div>
                      <div>
                        <div className="text-sm font-semibold uppercase tracking-[0.2em] text-amber-700 dark:text-amber-300">Admin Mode</div>
                        <p className="text-xs text-slate-500 dark:text-slate-400">Platform operators only</p>
                      </div>
                    </div>

                    <div className="mb-4">
                      <ThemeToggle showLabel className={cn("w-full justify-between", controlButtonClass)} />
                    </div>

                    <AdminNavLinks />
                  </SheetContent>
                </Sheet>

                <div className="space-y-1">
                  <div className="text-xs font-semibold uppercase tracking-[0.24em] text-amber-700 dark:text-amber-300">Control Center</div>
                  <p className="text-sm text-slate-500 dark:text-slate-400">Oversight across Ethiopian and Foreign agencies</p>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <ThemeToggle className={controlButtonClass} />
                <Badge variant="outline" className="hidden rounded-full border-amber-400/30 bg-amber-400/10 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.2em] text-amber-700 dark:text-amber-300 sm:inline-flex">
                  Admin Mode
                </Badge>
                <div className="hidden text-right sm:block">
                  <p className="text-sm font-semibold text-slate-950 dark:text-slate-100">{admin?.full_name}</p>
                  <p className="text-xs text-slate-500 dark:text-slate-400">{admin?.role?.replace("_", " ")}</p>
                </div>
                <Button variant="outline" size="sm" className={cn("gap-2", controlButtonClass)} onClick={logout}>
                  <LogOut className="h-4 w-4" />
                  <span className="hidden sm:inline">Logout</span>
                </Button>
              </div>
            </div>
          </header>

          <main className="flex-1 px-4 py-6 text-slate-950 dark:text-slate-100 sm:px-6 lg:px-8">
            <div className="space-y-6">{children}</div>
          </main>
        </div>
      </div>
    </div>
  )
}
