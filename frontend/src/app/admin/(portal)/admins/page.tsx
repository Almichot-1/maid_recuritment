"use client"

import * as React from "react"
import { usePathname } from "next/navigation"
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

type CreatedCredentials = {
  temporary_password: string
  mfa_secret: string
  provisioning_url: string
  invitation_warning?: string
}

export default function AdminManagementPage() {
  const pathname = usePathname()
  const { isSuperAdmin } = useCurrentAdmin()
  const { data: admins = [], isLoading } = useAdminUsers(isSuperAdmin)
  const { mutateAsync: createAdmin, isPending: creating } = useCreateAdminUser()
  const { mutateAsync: updateAdmin, isPending: updating } = useUpdateAdminUser()

  const [open, setOpen] = React.useState(false)
  const [email, setEmail] = React.useState("")
  const [fullName, setFullName] = React.useState("")
  const [role, setRole] = React.useState<AdminRole>(AdminRole.SUPPORT_ADMIN)
  const [createdCredentials, setCreatedCredentials] = React.useState<CreatedCredentials | null>(null)
  const [mfaRevealed, setMfaRevealed] = React.useState(false)

  React.useEffect(() => {
    setCreatedCredentials(null)
    setMfaRevealed(false)
  }, [pathname])

  React.useEffect(() => {
    if (!createdCredentials) {
      setMfaRevealed(false)
      return
    }

    const timeout = window.setTimeout(() => {
      setCreatedCredentials(null)
      setMfaRevealed(false)
    }, 120_000)

    return () => window.clearTimeout(timeout)
  }, [createdCredentials])

  React.useEffect(() => {
    if (!mfaRevealed) {
      return
    }

    const timeout = window.setTimeout(() => {
      setMfaRevealed(false)
    }, 30_000)

    return () => window.clearTimeout(timeout)
  }, [mfaRevealed])

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
      temporary_password: response.temporary_password,
      mfa_secret: response.mfa_secret,
      provisioning_url: response.provisioning_url,
      invitation_warning: response.invitation_warning,
    })
    setMfaRevealed(false)
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

  const dismissCredentials = () => {
    setCreatedCredentials(null)
    setMfaRevealed(false)
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="Admin Management"
        description="Create, review, and update the platform operator accounts that can enter the admin portal."
        action={<Button onClick={() => setOpen(true)}>Add Admin</Button>}
      />

      {createdCredentials ? (
        <Card className="border-amber-200 bg-amber-50">
          <CardContent className="space-y-3 p-5 text-sm text-amber-950">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <p className="font-semibold">Initial admin credentials</p>
              <Button variant="outline" size="sm" onClick={dismissCredentials}>
                Dismiss
              </Button>
            </div>
            <p>
              Temporary password:{" "}
              <span className="font-mono">{createdCredentials.temporary_password}</span>
            </p>
            <div className="space-y-2">
              <p>
                MFA secret:{" "}
                <span className="font-mono">
                  {mfaRevealed ? createdCredentials.mfa_secret : "••••••••••••••••"}
                </span>
              </p>
              <Button variant="outline" size="sm" onClick={() => setMfaRevealed((value) => !value)}>
                {mfaRevealed ? "Hide MFA secret" : "Reveal MFA secret"}
              </Button>
            </div>
            <p className="break-all">
              Provisioning URL:{" "}
              <span className="font-mono">{createdCredentials.provisioning_url}</span>
            </p>
            {createdCredentials.invitation_warning ? (
              <p className="text-amber-800">Invitation warning: {createdCredentials.invitation_warning}</p>
            ) : null}
            <p className="text-xs text-amber-800">
              This panel clears automatically after two minutes or when you leave this page.
            </p>
          </CardContent>
        </Card>
      ) : null}

      <Card className="border-border bg-card">
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
                  <TableCell colSpan={6} className="py-10 text-center text-sm text-muted-foreground">Loading admins...</TableCell>
                </TableRow>
              ) : null}
              {!isLoading && !admins.length ? (
                <TableRow>
                  <TableCell colSpan={6} className="py-10 text-center text-sm text-muted-foreground">No admins found.</TableCell>
                </TableRow>
              ) : null}
              {admins.map((admin) => (
                <TableRow key={admin.id}>
                  <TableCell>
                    <div>
                      <p className="font-medium text-foreground">{admin.full_name}</p>
                      <p className="text-xs text-muted-foreground">{admin.email}</p>
                    </div>
                  </TableCell>
                  <TableCell className="text-muted-foreground">{titleize(admin.role)}</TableCell>
                  <TableCell><AdminStatusBadge status={admin.is_active ? "active" : "suspended"} /></TableCell>
                  <TableCell className="text-muted-foreground">{formatDateTime(admin.last_login || undefined)}</TableCell>
                  <TableCell className="text-muted-foreground">{formatDateTime(admin.created_at)}</TableCell>
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
            <div className="space-y-2">
              <label htmlFor="admin-full-name" className="text-sm font-medium text-muted-foreground">
                Full name
              </label>
              <Input
                id="admin-full-name"
                placeholder="Jane Doe"
                value={fullName}
                onChange={(event) => setFullName(event.target.value)}
              />
            </div>
            <div className="space-y-2">
              <label htmlFor="admin-email" className="text-sm font-medium text-muted-foreground">
                Email address
              </label>
              <Input
                id="admin-email"
                placeholder="admin@example.com"
                type="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
              />
            </div>
            <div className="space-y-2">
              <label htmlFor="admin-role" className="text-sm font-medium text-muted-foreground">
                Role
              </label>
              <Select value={role} onValueChange={(value) => setRole(value as AdminRole)}>
                <SelectTrigger id="admin-role">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={AdminRole.SUPPORT_ADMIN}>Support Admin</SelectItem>
                  <SelectItem value={AdminRole.SUPER_ADMIN}>Super Admin</SelectItem>
                </SelectContent>
              </Select>
            </div>
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
