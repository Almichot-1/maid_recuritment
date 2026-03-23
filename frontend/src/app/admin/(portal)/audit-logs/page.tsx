"use client"

import * as React from "react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { Card, CardContent } from "@/components/ui/card"
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

      <Card className="border-slate-200 bg-white/90">
        <CardContent className="grid gap-4 p-5 lg:grid-cols-[1.2fr_0.8fr_0.8fr]">
          <Input value={adminSearch} onChange={(event) => setAdminSearch(event.target.value)} placeholder="Search admin or action" className="bg-white" />
          <Select value={action} onValueChange={setAction}>
            <SelectTrigger className="bg-white">
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
            <SelectTrigger className="bg-white">
              <SelectValue placeholder="Target type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All targets</SelectItem>
              <SelectItem value="agency">Agency</SelectItem>
              <SelectItem value="admin">Admin</SelectItem>
            </SelectContent>
          </Select>
        </CardContent>
      </Card>

      <Card className="border-slate-200 bg-white/90">
        <CardContent className="p-0">
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
                  <TableCell colSpan={5} className="py-10 text-center text-sm text-slate-500">Loading audit logs...</TableCell>
                </TableRow>
              ) : null}
              {!isLoading && !filtered.length ? (
                <TableRow>
                  <TableCell colSpan={5} className="py-10 text-center text-sm text-slate-500">No audit logs matched the current filters.</TableCell>
                </TableRow>
              ) : null}
              {filtered.map((log) => (
                <TableRow key={log.id}>
                  <TableCell>
                    <div>
                      <p className="font-medium text-slate-950">{log.admin_name || log.admin_id}</p>
                      <p className="text-xs text-slate-500">{log.admin_id}</p>
                    </div>
                  </TableCell>
                  <TableCell className="text-slate-600">{titleize(log.action)}</TableCell>
                  <TableCell className="text-slate-600">{titleize(log.target_type || "system")}</TableCell>
                  <TableCell className="text-slate-600">{log.ip_address || "N/A"}</TableCell>
                  <TableCell className="text-slate-600">{formatDateTime(log.created_at)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
