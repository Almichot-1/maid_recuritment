"use client"

import * as React from "react"
import { BellRing, KeyRound, ShieldCheck, UserCog } from "lucide-react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminEmptyState, AdminStatCard, AdminSurface, AdminToolbar } from "@/components/admin/admin-ui"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAdminAuditLogs } from "@/hooks/use-admin-portal"
import { formatDateTime, titleize } from "@/lib/admin-utils"

export default function AdminAuditLogsPage() {
  const [action, setAction] = React.useState("all")
  const [targetType, setTargetType] = React.useState("all")
  const [adminSearch, setAdminSearch] = React.useState("")

  const { data: logs = [], isLoading } = useAdminAuditLogs({
    action: action === "all" ? undefined : action,
    target_type: targetType === "all" ? undefined : targetType,
  })
  const loginEvents = logs.filter((log) => log.action === "admin_login").length
  const adminTargetEvents = logs.filter((log) => log.target_type === "admin").length
  const agencyTargetEvents = logs.filter((log) => log.target_type === "agency").length

  const filtered = React.useMemo(
    () =>
      logs.filter((log) => {
        if (!adminSearch) {
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

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Audit Logs"
        description="Immutable operator activity across approvals, admin access, and sensitive management actions."
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminStatCard label="Audit events" value={filtered.length} detail="Visible in the current result set" icon={BellRing} />
        <AdminStatCard label="Logins" value={loginEvents} detail="Admin sign-in activity" icon={KeyRound} />
        <AdminStatCard label="Agency actions" value={agencyTargetEvents} detail="Actions against agency records" icon={ShieldCheck} />
        <AdminStatCard label="Admin actions" value={adminTargetEvents} detail="Actions against operator accounts" icon={UserCog} />
      </div>

      <AdminToolbar className="grid gap-4 p-5 lg:grid-cols-[1.2fr_0.8fr_0.8fr]">
          <Input value={adminSearch} onChange={(event) => setAdminSearch(event.target.value)} placeholder="Search admin or action" className="bg-white dark:bg-slate-950" />
          <Select value={action} onValueChange={setAction}>
            <SelectTrigger className="bg-white dark:bg-slate-950">
              <SelectValue placeholder="Action type" />
            </SelectTrigger>
            <SelectContent>
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
            <SelectTrigger className="bg-white dark:bg-slate-950">
              <SelectValue placeholder="Target type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All targets</SelectItem>
              <SelectItem value="agency">Agency</SelectItem>
              <SelectItem value="admin">Admin</SelectItem>
            </SelectContent>
          </Select>
      </AdminToolbar>

      <AdminSurface className="overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Admin</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Target</TableHead>
                <TableHead>IP Address</TableHead>
                <TableHead>Timestamp</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={5} className="py-10 text-center text-sm text-slate-500 dark:text-slate-400">Loading audit logs...</TableCell>
                </TableRow>
              ) : null}
              {!isLoading && !filtered.length ? (
                <TableRow>
                  <TableCell colSpan={5} className="p-6">
                    <AdminEmptyState
                      title="No audit entries matched"
                      description="Try removing the action or target filters to widen the timeline and inspect more events."
                    />
                  </TableCell>
                </TableRow>
              ) : null}
              {filtered.map((log) => (
                <TableRow key={log.id}>
                  <TableCell>
                    <div>
                      <p className="font-medium text-slate-950 dark:text-slate-100">{log.admin_name || log.admin_id}</p>
                      <p className="text-xs text-slate-500 dark:text-slate-400">{log.admin_id}</p>
                    </div>
                  </TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{titleize(log.action)}</TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{titleize(log.target_type || "system")}</TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{log.ip_address || "N/A"}</TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{formatDateTime(log.created_at)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
      </AdminSurface>
    </div>
  )
}
