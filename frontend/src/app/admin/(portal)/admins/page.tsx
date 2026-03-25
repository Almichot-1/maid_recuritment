"use client"

import * as React from "react"
import { toast } from "sonner"
import { KeyRound, ShieldCheck, ShieldMinus, UserCog } from "lucide-react"

import { AdminPageHeader } from "@/components/admin/admin-page-header"
import { AdminStatusBadge } from "@/components/admin/admin-status-badge"
import { AdminEmptyState, AdminStatCard, AdminSurface } from "@/components/admin/admin-ui"
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
    temporary_password: string
    mfa_secret: string
    provisioning_url: string
    invitation_warning?: string
  }>(null)
  const activeAdmins = admins.filter((admin) => admin.is_active).length
  const suspendedAdmins = admins.length - activeAdmins
  const superAdmins = admins.filter((admin) => admin.role === AdminRole.SUPER_ADMIN).length

  if (!isSuperAdmin) {
    return (
      <Card className="border-amber-200 bg-amber-50 dark:border-amber-400/20 dark:bg-amber-400/10">
        <CardContent className="p-6 text-sm text-amber-900 dark:text-amber-100">
          Admin management is restricted to Super Admin accounts.
        </CardContent>
      </Card>
    )
  }

  const handleCreateAdmin = async () => {
    const response = await createAdmin({ email, full_name: fullName, role })
    setCreatedCredentials({
      temporary_password: response.temporary_password,
      mfa_secret: response.mfa_secret,
      provisioning_url: response.provisioning_url,
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
        action={<Button onClick={() => setOpen(true)}>Add Admin</Button>}
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <AdminStatCard label="Operator accounts" value={admins.length} detail="Visible to super admins" icon={UserCog} />
        <AdminStatCard label="Active" value={activeAdmins} detail="Can access the admin portal" icon={ShieldCheck} />
        <AdminStatCard label="Suspended" value={suspendedAdmins} detail="Temporarily disabled logins" icon={ShieldMinus} />
        <AdminStatCard label="Super admins" value={superAdmins} detail="Full platform control" icon={KeyRound} />
      </div>

      {createdCredentials ? (
        <Card className="border-amber-300/40 bg-amber-50/90 dark:border-amber-400/20 dark:bg-amber-400/10">
          <CardContent className="space-y-2 p-5 text-sm text-amber-950">
            <p className="font-semibold">Initial admin credentials</p>
            <p>Temporary password: <span className="font-mono">{createdCredentials.temporary_password}</span></p>
            <p>MFA secret: <span className="font-mono">{createdCredentials.mfa_secret}</span></p>
            <p className="break-all">Provisioning URL: <span className="font-mono">{createdCredentials.provisioning_url}</span></p>
            {createdCredentials.invitation_warning ? (
              <p className="text-amber-800 dark:text-amber-200">Invitation warning: {createdCredentials.invitation_warning}</p>
            ) : null}
          </CardContent>
        </Card>
      ) : null}

      <AdminSurface className="overflow-hidden">
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
                  <TableCell colSpan={6} className="py-10 text-center text-sm text-slate-500 dark:text-slate-400">Loading admins...</TableCell>
                </TableRow>
              ) : null}
              {!isLoading && !admins.length ? (
                <TableRow>
                  <TableCell colSpan={6} className="p-6">
                    <AdminEmptyState
                      title="No admins found"
                      description="Create the next admin account from here and the credentials pack will appear immediately."
                    />
                  </TableCell>
                </TableRow>
              ) : null}
              {admins.map((admin) => (
                <TableRow key={admin.id}>
                  <TableCell>
                    <div>
                      <p className="font-medium text-slate-950 dark:text-slate-100">{admin.full_name}</p>
                      <p className="text-xs text-slate-500 dark:text-slate-400">{admin.email}</p>
                    </div>
                  </TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{titleize(admin.role)}</TableCell>
                  <TableCell><AdminStatusBadge status={admin.is_active ? "active" : "suspended"} /></TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{formatDateTime(admin.last_login || undefined)}</TableCell>
                  <TableCell className="text-slate-600 dark:text-slate-300">{formatDateTime(admin.created_at)}</TableCell>
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
      </AdminSurface>

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
