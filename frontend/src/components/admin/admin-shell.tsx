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
  const pathname = usePathname()
  const logout = useAdminLogout()
  const visibleNavItems = React.useMemo(() => navItems.filter((item) => isVisible(item, admin?.role)), [admin?.role])

  return (
    <div className="admin-portal dark min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top_left,_rgba(245,158,11,0.14),_transparent_20%),radial-gradient(circle_at_top_right,_rgba(34,211,238,0.1),_transparent_24%),linear-gradient(180deg,#020617_0%,#08111f_38%,#0f172a_100%)] text-slate-50">
      <div className="pointer-events-none absolute inset-x-0 top-0 h-40 bg-[linear-gradient(180deg,rgba(148,163,184,0.08),transparent)]" />
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
          <header className="sticky top-0 z-30 border-b border-slate-800/80 bg-slate-950/88 backdrop-blur-xl">
            <div className="flex items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
              <div className="flex items-center gap-3">
                <Sheet>
                  <SheetTrigger asChild>
                    <Button variant="outline" size="icon" className="border-slate-700 bg-slate-900 text-slate-100 hover:bg-slate-800 lg:hidden">
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
                  <div className="text-xs font-semibold uppercase tracking-[0.24em] text-amber-300">Control Center</div>
                  <p className="text-sm text-slate-400">Oversight across Ethiopian and Foreign agencies</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <Badge variant="outline" className="hidden rounded-full border-amber-400/30 bg-amber-400/10 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.2em] text-amber-200 sm:inline-flex">
                  Admin Mode
                </Badge>
                <div className="hidden text-right sm:block">
                  <p className="text-sm font-semibold text-white">{admin?.full_name}</p>
                  <p className="text-xs text-slate-400">{admin?.role?.replace("_", " ")}</p>
                </div>
                <Button variant="outline" size="sm" className="gap-2 border-slate-700 bg-slate-900 text-slate-100 hover:bg-slate-800" onClick={logout}>
                  <LogOut className="h-4 w-4" />
                  <span className="hidden sm:inline">Logout</span>
                </Button>
              </div>
            </div>
            <div className="border-t border-slate-800/70 px-4 py-3 lg:hidden">
              <div className="-mx-1 overflow-x-auto px-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
                <div className="flex min-w-max gap-2">
                  {visibleNavItems.map((item) => {
                    const active = pathname === item.href || (item.href !== "/admin/dashboard" && pathname.startsWith(item.href))
                    return (
                      <Link
                        key={item.href}
                        href={item.href}
                        className={cn(
                          "inline-flex items-center gap-2 rounded-full border px-3 py-2 text-xs font-semibold transition-all",
                          active
                            ? "border-amber-300/60 bg-amber-400/15 text-amber-100"
                            : "border-slate-700 bg-slate-900/90 text-slate-300"
                        )}
                      >
                        <item.icon className="h-3.5 w-3.5 shrink-0" />
                        <span className="whitespace-nowrap">{item.label}</span>
                      </Link>
                    )
                  })}
                </div>
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
