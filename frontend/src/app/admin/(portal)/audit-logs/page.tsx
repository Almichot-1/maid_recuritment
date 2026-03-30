"use client"

import * as React from "react"
import { Activity, Building2, LogIn, ShieldCheck, Smartphone } from "lucide-react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useAdminAgencyLogins, useAdminAuditLogs } from "@/hooks/use-admin-portal"
import { formatDateTime, formatRelative, titleize } from "@/lib/admin-utils"
import { UserRole } from "@/types"

const emptyLoginSummary = {
  total_login_events: 0,
  active_sessions: 0,
  ethiopian_login_events: 0,
  foreign_login_events: 0,
}

export default function AdminAuditLogsPage() {
  const [tab, setTab] = React.useState("operator")
  const [action, setAction] = React.useState("all")
  const [targetType, setTargetType] = React.useState("all")
  const [adminSearch, setAdminSearch] = React.useState("")
  const [agencyRole, setAgencyRole] = React.useState<"all" | UserRole>("all")
  const [agencySearch, setAgencySearch] = React.useState("")

  const { data: logs = [], isLoading: logsLoading } = useAdminAuditLogs({
    action: action === "all" ? undefined : action,
    target_type: targetType === "all" ? undefined : targetType,
  })

  const { data: agencyLoginData, isLoading: loginsLoading } = useAdminAgencyLogins({
    role: agencyRole,
    search: agencySearch.trim() || undefined,
  })

  const filteredLogs = React.useMemo(
    () =>
      logs.filter((log) => {
        if (!adminSearch.trim()) {
          return true
        }
        const needle = adminSearch.toLowerCase()
        return (
          log.admin_name.toLowerCase().includes(needle) ||
          log.action.toLowerCase().includes(needle) ||
          log.target_type.toLowerCase().includes(needle)
        )
      }),
    [adminSearch, logs]
  )

  const agencyLogins = agencyLoginData?.logins ?? []
  const loginSummary = agencyLoginData?.summary ?? emptyLoginSummary

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Audit Center"
        description="Track operator actions and real agency sign-ins from one admin surface, with mobile-friendly views for fast reviews on the go."
      />

      <Tabs value={tab} onValueChange={setTab} className="space-y-6">
        <TabsList className="grid h-auto w-full grid-cols-2 rounded-3xl border border-slate-800 bg-slate-950/80 p-1">
          <TabsTrigger
            value="operator"
            className="rounded-[20px] px-4 py-3 text-sm font-semibold text-slate-300 data-[state=active]:bg-slate-900 data-[state=active]:text-white data-[state=active]:shadow-none"
          >
            <Activity className="mr-2 h-4 w-4" />
            Operator Activity
          </TabsTrigger>
          <TabsTrigger
            value="agency"
            className="rounded-[20px] px-4 py-3 text-sm font-semibold text-slate-300 data-[state=active]:bg-slate-900 data-[state=active]:text-white data-[state=active]:shadow-none"
          >
            <LogIn className="mr-2 h-4 w-4" />
            Agency Sign-ins
          </TabsTrigger>
        </TabsList>

        <TabsContent value="operator" className="space-y-6">
          <Card className="border-slate-800 bg-slate-950/75 text-slate-50 shadow-2xl shadow-black/10">
            <CardContent className="grid gap-4 p-5 lg:grid-cols-[1.2fr_0.8fr_0.8fr]">
              <Input
                value={adminSearch}
                onChange={(event) => setAdminSearch(event.target.value)}
                placeholder="Search admin or action"
                className="border-slate-700 bg-slate-900 text-slate-50 placeholder:text-slate-500"
              />
              <Select value={action} onValueChange={setAction}>
                <SelectTrigger className="border-slate-700 bg-slate-900 text-slate-50">
                  <SelectValue placeholder="Action type" />
                </SelectTrigger>
                <SelectContent className="border-slate-800 bg-slate-950 text-slate-50">
                  <SelectItem value="all">All actions</SelectItem>
                  <SelectItem value="admin_login">Admin login</SelectItem>
                  <SelectItem value="approve_agency">Approve agency</SelectItem>
                  <SelectItem value="reject_agency">Reject agency</SelectItem>
                  <SelectItem value="update_agency_status">Update agency status</SelectItem>
                  <SelectItem value="create_admin">Create admin</SelectItem>
                  <SelectItem value="update_admin">Update admin</SelectItem>
                </SelectContent>
              </Select>
              <Select value={targetType} onValueChange={setTargetType}>
                <SelectTrigger className="border-slate-700 bg-slate-900 text-slate-50">
                  <SelectValue placeholder="Target type" />
                </SelectTrigger>
                <SelectContent className="border-slate-800 bg-slate-950 text-slate-50">
                  <SelectItem value="all">All targets</SelectItem>
                  <SelectItem value="agency">Agency</SelectItem>
                  <SelectItem value="admin">Admin</SelectItem>
                </SelectContent>
              </Select>
            </CardContent>
          </Card>

          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <MetricCard
              icon={ShieldCheck}
              label="Matched entries"
              value={filteredLogs.length}
              detail="Operator actions after current filters"
            />
            <MetricCard
              icon={Activity}
              label="Admin logins"
              value={filteredLogs.filter((log) => log.action === "admin_login").length}
              detail="Authentication events inside the current result set"
            />
            <MetricCard
              icon={Building2}
              label="Agency actions"
              value={filteredLogs.filter((log) => log.target_type === "agency").length}
              detail="Approvals, rejections, and status updates"
            />
            <MetricCard
              icon={LogIn}
              label="Recent activity"
              value={filteredLogs[0] ? formatRelative(filteredLogs[0].created_at) : "N/A"}
              detail="Freshest operator event"
            />
          </div>

          <div className="grid gap-4 md:hidden">
            {logsLoading ? (
              <EmptyCard message="Loading operator activity..." />
            ) : null}
            {!logsLoading && !filteredLogs.length ? (
              <EmptyCard message="No audit logs matched the current filters." />
            ) : null}
            {filteredLogs.map((log) => (
              <Card key={log.id} className="border-slate-800 bg-slate-950/75 text-slate-50 shadow-xl shadow-black/10">
                <CardContent className="space-y-4 p-5">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="font-semibold text-white">{log.admin_name || log.admin_id}</p>
                      <p className="text-xs text-slate-400">{log.admin_id}</p>
                    </div>
                    <AdminStatusBadge status={log.target_type || "system"} className="shrink-0" />
                  </div>
                  <div className="space-y-2 text-sm text-slate-300">
                    <InfoRow label="Action" value={titleize(log.action)} />
                    <InfoRow label="IP Address" value={log.ip_address || "N/A"} />
                    <InfoRow label="When" value={formatDateTime(log.created_at)} />
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          <Card className="hidden border-slate-800 bg-slate-950/75 text-slate-50 shadow-2xl shadow-black/10 md:block">
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow className="border-slate-800 hover:bg-transparent">
                    <TableHead className="text-slate-400">Admin</TableHead>
                    <TableHead className="text-slate-400">Action</TableHead>
                    <TableHead className="text-slate-400">Target</TableHead>
                    <TableHead className="text-slate-400">IP Address</TableHead>
                    <TableHead className="text-slate-400">Timestamp</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {logsLoading ? (
                    <TableRow className="border-slate-800">
                      <TableCell colSpan={5} className="py-10 text-center text-sm text-slate-400">
                        Loading operator activity...
                      </TableCell>
                    </TableRow>
                  ) : null}
                  {!logsLoading && !filteredLogs.length ? (
                    <TableRow className="border-slate-800">
                      <TableCell colSpan={5} className="py-10 text-center text-sm text-slate-400">
                        No audit logs matched the current filters.
                      </TableCell>
                    </TableRow>
                  ) : null}
                  {filteredLogs.map((log) => (
                    <TableRow key={log.id} className="border-slate-800/80 hover:bg-slate-900/60">
                      <TableCell>
                        <div>
                          <p className="font-medium text-white">{log.admin_name || log.admin_id}</p>
                          <p className="text-xs text-slate-500">{log.admin_id}</p>
                        </div>
                      </TableCell>
                      <TableCell className="text-slate-300">{titleize(log.action)}</TableCell>
                      <TableCell className="text-slate-300">{titleize(log.target_type || "system")}</TableCell>
                      <TableCell className="text-slate-300">{log.ip_address || "N/A"}</TableCell>
                      <TableCell className="text-slate-300">{formatDateTime(log.created_at)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="agency" className="space-y-6">
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <MetricCard
              icon={LogIn}
              label="Total login events"
              value={loginSummary.total_login_events}
              detail="All recorded agency sign-ins"
            />
            <MetricCard
              icon={ShieldCheck}
              label="Active sessions"
              value={loginSummary.active_sessions}
              detail="Sessions still valid right now"
            />
            <MetricCard
              icon={Building2}
              label="Ethiopian logins"
              value={loginSummary.ethiopian_login_events}
              detail="Agency-side access from Ethiopian operators"
            />
            <MetricCard
              icon={Smartphone}
              label="Foreign logins"
              value={loginSummary.foreign_login_events}
              detail="Agency-side access from foreign operators"
            />
          </div>

          <Card className="border-slate-800 bg-slate-950/75 text-slate-50 shadow-2xl shadow-black/10">
            <CardHeader className="space-y-1">
              <CardTitle className="text-lg text-white">Agency sign-in activity</CardTitle>
              <p className="text-sm text-slate-400">
                Review who signed in, from which device, and whether the session is still active.
              </p>
            </CardHeader>
            <CardContent className="grid gap-4 lg:grid-cols-[1.3fr_0.7fr]">
              <Input
                value={agencySearch}
                onChange={(event) => setAgencySearch(event.target.value)}
                placeholder="Search agency, contact, or email"
                className="border-slate-700 bg-slate-900 text-slate-50 placeholder:text-slate-500"
              />
              <Select value={agencyRole} onValueChange={(value) => setAgencyRole(value as "all" | UserRole)}>
                <SelectTrigger className="border-slate-700 bg-slate-900 text-slate-50">
                  <SelectValue placeholder="Agency role" />
                </SelectTrigger>
                <SelectContent className="border-slate-800 bg-slate-950 text-slate-50">
                  <SelectItem value="all">All agencies</SelectItem>
                  <SelectItem value={UserRole.ETHIOPIAN_AGENT}>Ethiopian agencies</SelectItem>
                  <SelectItem value={UserRole.FOREIGN_AGENT}>Foreign agencies</SelectItem>
                </SelectContent>
              </Select>
            </CardContent>
          </Card>

          <div className="grid gap-4 md:hidden">
            {loginsLoading ? (
              <EmptyCard message="Loading agency sign-ins..." />
            ) : null}
            {!loginsLoading && !agencyLogins.length ? (
              <EmptyCard message="No agency sign-ins matched the current filters." />
            ) : null}
            {agencyLogins.map((login) => (
              <Card key={login.session_id} className="border-slate-800 bg-slate-950/75 text-slate-50 shadow-xl shadow-black/10">
                <CardContent className="space-y-4 p-5">
                  <div className="flex items-start justify-between gap-3">
                    <div className="space-y-1">
                      <p className="font-semibold text-white">{login.agency_name}</p>
                      <p className="text-sm text-slate-400">{login.contact_name || login.email}</p>
                    </div>
                    <Badge
                      variant="outline"
                      className={login.is_active
                        ? "rounded-full border-emerald-400/40 bg-emerald-400/10 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide text-emerald-200"
                        : "rounded-full border-slate-700 bg-slate-900 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide text-slate-300"}
                    >
                      {login.is_active ? "Active" : "Ended"}
                    </Badge>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Badge variant="outline" className="rounded-full border-amber-400/30 bg-amber-400/10 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide text-amber-200">
                      {titleize(login.role)}
                    </Badge>
                    <Badge variant="outline" className="rounded-full border-slate-700 bg-slate-900 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide text-slate-300">
                      {login.device_label}
                    </Badge>
                  </div>
                  <div className="space-y-2 text-sm text-slate-300">
                    <InfoRow label="Email" value={login.email} />
                    <InfoRow label="IP Address" value={login.ip_address || "N/A"} />
                    <InfoRow label="Signed in" value={formatDateTime(login.logged_in_at)} />
                    <InfoRow label="Last seen" value={formatDateTime(login.last_seen_at)} />
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          <Card className="hidden border-slate-800 bg-slate-950/75 text-slate-50 shadow-2xl shadow-black/10 md:block">
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableRow className="border-slate-800 hover:bg-transparent">
                    <TableHead className="text-slate-400">Agency</TableHead>
                    <TableHead className="text-slate-400">Role</TableHead>
                    <TableHead className="text-slate-400">Device</TableHead>
                    <TableHead className="text-slate-400">IP Address</TableHead>
                    <TableHead className="text-slate-400">Logged in</TableHead>
                    <TableHead className="text-slate-400">Last seen</TableHead>
                    <TableHead className="text-right text-slate-400">Session</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {loginsLoading ? (
                    <TableRow className="border-slate-800">
                      <TableCell colSpan={7} className="py-10 text-center text-sm text-slate-400">
                        Loading agency sign-ins...
                      </TableCell>
                    </TableRow>
                  ) : null}
                  {!loginsLoading && !agencyLogins.length ? (
                    <TableRow className="border-slate-800">
                      <TableCell colSpan={7} className="py-10 text-center text-sm text-slate-400">
                        No agency sign-ins matched the current filters.
                      </TableCell>
                    </TableRow>
                  ) : null}
                  {agencyLogins.map((login) => (
                    <TableRow key={login.session_id} className="border-slate-800/80 hover:bg-slate-900/60">
                      <TableCell>
                        <div className="space-y-1">
                          <p className="font-medium text-white">{login.agency_name}</p>
                          <p className="text-xs text-slate-500">{login.contact_name || login.email}</p>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline" className="rounded-full border-amber-400/30 bg-amber-400/10 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide text-amber-200">
                          {titleize(login.role)}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-slate-300">
                        <div>
                          <p>{login.device_label}</p>
                          <p className="text-xs text-slate-500">
                            {login.browser_name} / {login.os_name}
                          </p>
                        </div>
                      </TableCell>
                      <TableCell className="text-slate-300">{login.ip_address || "N/A"}</TableCell>
                      <TableCell className="text-slate-300">{formatDateTime(login.logged_in_at)}</TableCell>
                      <TableCell className="text-slate-300">{formatDateTime(login.last_seen_at)}</TableCell>
                      <TableCell className="text-right">
                        <Badge
                          variant="outline"
                          className={login.is_active
                            ? "rounded-full border-emerald-400/40 bg-emerald-400/10 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide text-emerald-200"
                            : "rounded-full border-slate-700 bg-slate-900 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide text-slate-300"}
                        >
                          {login.is_active ? "Active" : "Ended"}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}

function MetricCard({
  icon: Icon,
  label,
  value,
  detail,
}: {
  icon: React.ElementType
  label: string
  value: number | string
  detail: string
}) {
  return (
    <Card className="border-slate-800 bg-slate-950/75 text-slate-50 shadow-2xl shadow-black/10">
      <CardContent className="flex items-start justify-between gap-4 p-5">
        <div className="space-y-2">
          <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-400">{label}</p>
          <p className="text-2xl font-semibold tracking-tight text-white">{value}</p>
          <p className="text-sm text-slate-400">{detail}</p>
        </div>
        <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl border border-slate-800 bg-slate-900 text-amber-300">
          <Icon className="h-5 w-5" />
        </div>
      </CardContent>
    </Card>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="text-slate-500">{label}</span>
      <span className="text-right text-slate-200">{value}</span>
    </div>
  )
}

function EmptyCard({ message }: { message: string }) {
  return (
    <Card className="border-slate-800 bg-slate-950/75 text-slate-50 shadow-xl shadow-black/10">
      <CardContent className="p-6 text-center text-sm text-slate-400">{message}</CardContent>
    </Card>
  )
}
