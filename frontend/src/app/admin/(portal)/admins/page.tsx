"use client"

import * as React from "react"
import { toast } from "sonner"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useCurrentAdmin } from "@/hooks/use-admin-auth"
import { useAdminUsers, useCreateAdminUser, useUpdateAdminUser } from "@/hooks/use-admin-portal"
import { AdminRole } from "@/types"
import { formatDateTime, titleize } from "@/lib/admin-utils"

export default function AdminManagementPage() {
  const { isSuperAdmin } = useCurrentAdmin()
  const { data: admins = [], isLoading } = useAdminUsers(isSuperAdmin)
  const { mutateAsync: createAdmin, isPending: creating } = useCreateAdminUser()
  const { mutateAsync: updateAdmin, isPending: updating } = useUpdateAdminUser()

  const [open, setOpen] = React.useState(false)
  const [email, setEmail] = React.useState("")
  const [fullName, setFullName] = React.useState("")
  const [role, setRole] = React.useState<AdminRole>(AdminRole.SUPPORT_ADMIN)
  const [createdCredentials, setCreatedCredentials] = React.useState<null | {
    setup_url: string
    invitation_warning?: string
  }>(null)

  if (!isSuperAdmin) {
    return (
      <Card className="border-amber-200 bg-amber-50">
        <CardContent className="p-6 text-sm text-amber-900">
          Admin management is restricted to Super Admin accounts.
        </CardContent>
      </Card>
    )
  }

  const handleCreateAdmin = async () => {
    const response = await createAdmin({ email, full_name: fullName, role })
    setCreatedCredentials({
      setup_url: response.setup_url,
      invitation_warning: response.invitation_warning,
    })
    toast.success("Admin account created.")
    setEmail("")
    setFullName("")
    setRole(AdminRole.SUPPORT_ADMIN)
    setOpen(false)
  }

  const handleToggleAdmin = async (id: string, isActive: boolean) => {
    await updateAdmin({ id, is_active: !isActive })
    toast.success(isActive ? "Admin account suspended." : "Admin account reactivated.")
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Admin Management"
        description="Create, review, and update the platform operator accounts that can enter the admin portal."
        action={<Button className="bg-slate-950 hover:bg-slate-800" onClick={() => setOpen(true)}>Add Admin</Button>}
      />

      {createdCredentials ? (
        <Card className="border-amber-200 bg-amber-50">
          <CardContent className="space-y-2 p-5 text-sm text-amber-950">
            <p className="font-semibold">One-time admin setup link</p>
            <p className="break-all">Setup URL: <span className="font-mono">{createdCredentials.setup_url}</span></p>
            {createdCredentials.invitation_warning ? (
              <p className="text-amber-800">Invitation warning: {createdCredentials.invitation_warning}</p>
            ) : null}
          </CardContent>
        </Card>
      ) : null}

      <Card className="border-slate-200 bg-white/90">
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Admin</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Last Login</TableHead>
                <TableHead>Created</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                <TableRow>
                  <TableCell colSpan={6} className="py-10 text-center text-sm text-slate-500">Loading admins...</TableCell>
                </TableRow>
              ) : null}
              {!isLoading && !admins.length ? (
                <TableRow>
                  <TableCell colSpan={6} className="py-10 text-center text-sm text-slate-500">No admins found.</TableCell>
                </TableRow>
              ) : null}
              {admins.map((admin) => (
                <TableRow key={admin.id}>
                  <TableCell>
                    <div>
                      <p className="font-medium text-slate-950">{admin.full_name}</p>
                      <p className="text-xs text-slate-500">{admin.email}</p>
                    </div>
                  </TableCell>
                  <TableCell className="text-slate-600">{titleize(admin.role)}</TableCell>
                  <TableCell><AdminStatusBadge status={admin.is_active ? "active" : "suspended"} /></TableCell>
                  <TableCell className="text-slate-600">{formatDateTime(admin.last_login || undefined)}</TableCell>
                  <TableCell className="text-slate-600">{formatDateTime(admin.created_at)}</TableCell>
                  <TableCell>
                    <div className="flex items-center justify-end gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={updating}
                        onClick={() => handleToggleAdmin(admin.id, admin.is_active)}
                      >
                        {admin.is_active ? "Suspend" : "Activate"}
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add new admin</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <Input placeholder="Full name" value={fullName} onChange={(event) => setFullName(event.target.value)} />
            <Input placeholder="Email address" type="email" value={email} onChange={(event) => setEmail(event.target.value)} />
            <Select value={role} onValueChange={(value) => setRole(value as AdminRole)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={AdminRole.SUPPORT_ADMIN}>Support Admin</SelectItem>
                <SelectItem value={AdminRole.SUPER_ADMIN}>Super Admin</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
            <Button disabled={creating} onClick={handleCreateAdmin}>Create Admin</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
