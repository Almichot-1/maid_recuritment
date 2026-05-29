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
  Link2,
  LogOut,
  Menu,
  Settings2,
  UserCog,
  Users,
} from "lucide-react"

import { AdminRole } from "@/types"
import { useAdminLogout, useCurrentAdmin } from "@/hooks/use-admin-auth"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet"
import { LocaleSwitcher } from "@/components/shared/locale-switcher"
import { Logo } from "@/components/shared/logo"
import { useI18n } from "@/lib/i18n"
import { cn } from "@/lib/utils"

type AdminNavItem = {
  href: string
  label: string
  icon: React.ElementType
  roles?: AdminRole[]
}

function isVisible(item: AdminNavItem, role?: AdminRole) {
  if (!item.roles?.length) {
    return true
  }
  return role ? item.roles.includes(role) : false
}

function AdminNavLinks({
  items,
  onNavigate,
}: {
  items: AdminNavItem[]
  onNavigate?: () => void
}) {
  const pathname = usePathname()
  const { admin } = useCurrentAdmin()

  return (
    <nav className="space-y-1">
      {items
        .filter((item) => isVisible(item, admin?.role))
        .map((item) => {
          const active = pathname === item.href || (item.href !== "/admin/dashboard" && pathname.startsWith(item.href))
          return (
            <Link
              key={item.href}
              href={item.href}
              onClick={onNavigate}
              className={cn(
                "flex items-center gap-3 border border-transparent px-4 py-3 text-sm font-bold uppercase tracking-[0.05em] transition-colors",
                active
                  ? "border-primary bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:border-border hover:bg-muted/30 hover:text-foreground"
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
  const { isRTL, t } = useI18n()

  const navItems = React.useMemo<AdminNavItem[]>(
    () => [
      { href: "/admin/dashboard", label: t("admin.navDashboard"), icon: BarChart3 },
      { href: "/admin/agencies/pending", label: t("admin.navPending"), icon: FileClock },
      { href: "/admin/agencies", label: t("admin.navAgencies"), icon: Building2 },
      { href: "/admin/pairings", label: t("admin.navPairings"), icon: Link2 },
      { href: "/admin/candidates", label: t("admin.navCandidates"), icon: Users },
      { href: "/admin/selections", label: t("admin.navSelections"), icon: CheckCheck },
      { href: "/admin/reports", label: t("admin.navReports"), icon: ClipboardList },
      { href: "/admin/settings", label: t("admin.navSettings"), icon: Settings2, roles: [AdminRole.SUPER_ADMIN] },
      { href: "/admin/audit-logs", label: t("admin.navAudit"), icon: BellRing },
      { href: "/admin/admins", label: t("admin.navAdmins"), icon: UserCog, roles: [AdminRole.SUPER_ADMIN] },
    ],
    [t]
  )

  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="relative mx-auto flex min-h-screen max-w-[1760px]">
        <aside className={cn("hidden w-[300px] flex-col border-r border-border bg-card px-5 py-6 lg:flex", isRTL && "order-2 border-r-0 border-l")}>
          <div className="space-y-4 border border-border bg-background p-5">
            <Logo size="sm" />
            <div className="space-y-1">
              <p className="section-kicker">{t("admin.mode")}</p>
              <p className="text-sm text-muted-foreground">{t("admin.controlBody")}</p>
            </div>
            <LocaleSwitcher />
          </div>

          <div className="mt-6">
            <AdminNavLinks items={navItems} />
          </div>

          <div className="mt-auto border border-border bg-background p-4">
            <div className="flex items-start gap-3">
              <div className="flex h-10 w-10 items-center justify-center border border-foreground bg-foreground text-sm font-bold text-background">
                {admin?.full_name?.charAt(0) || "A"}
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-bold text-foreground">{admin?.full_name || "Admin"}</p>
                <p className="truncate text-xs text-muted-foreground">{admin?.email}</p>
              </div>
            </div>
            <div className="mt-3 flex items-center justify-between gap-3">
              <Badge variant="outline">{(admin?.role || "admin").replace("_", " ")}</Badge>
              <Button variant="ghost" size="icon" onClick={logout} aria-label={t("common.logout")}>
                <LogOut className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </aside>

        <div className="flex min-h-screen flex-1 flex-col">
          <header className="sticky top-0 z-30 border-b border-border bg-background">
            <div className="flex items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
              <div className="flex items-center gap-3">
                <Sheet>
                  <SheetTrigger asChild>
                    <Button variant="outline" size="icon" className="lg:hidden" aria-label="Open navigation menu">
                      <Menu className="h-4 w-4" />
                    </Button>
                  </SheetTrigger>
                  <SheetContent
                    side={isRTL ? "right" : "left"}
                    className="w-full max-w-none border-0 bg-background px-5 py-6"
                  >
                    <div className="space-y-4 border-b border-border pb-4">
                      <Logo size="sm" />
                      <LocaleSwitcher />
                    </div>
                    <div className="mt-4">
                      <AdminNavLinks items={navItems} />
                    </div>
                  </SheetContent>
                </Sheet>
                <div className="space-y-1">
                  <p className="route-stamp text-[10px] text-muted-foreground">{t("header.adminStamp")}</p>
                  <h1 className="font-display text-3xl text-foreground">{t("admin.controlCenter")}</h1>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="hidden lg:block">
                  <LocaleSwitcher compact />
                </div>
                <Badge variant="outline" className="hidden sm:inline-flex">
                  {t("admin.mode")}
                </Badge>
                <div className="hidden text-right sm:block">
                  <p className="text-sm font-bold text-foreground">{admin?.full_name}</p>
                  <p className="text-xs text-muted-foreground">{admin?.role?.replace("_", " ")}</p>
                </div>
                <Button variant="outline" size="sm" className="gap-2" onClick={logout}>
                  <LogOut className="h-4 w-4" />
                  <span className="hidden sm:inline">{t("common.logout")}</span>
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
