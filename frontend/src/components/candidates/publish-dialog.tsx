"use client"

import * as React from "react"
import { Loader2, CheckCircle2, AlertCircle, Upload } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { WorkspaceSummary } from "@/types"
import { useUpdatePairingDefaults, useUpdatePairingLogo } from "@/hooks/use-pairings"

interface PublishDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  candidateName: string
  workspaces: WorkspaceSummary[]
  selectedWorkspace?: WorkspaceSummary
  onConfirm: () => void | Promise<void>
  isPublishing?: boolean
}

export function PublishDialog({
  open,
  onOpenChange,
  candidateName,
  selectedWorkspace,
  onConfirm,
  isPublishing = false,
}: PublishDialogProps) {
  const [setupRequired, setSetupRequired] = React.useState(false)
  const [defaultCountry, setDefaultCountry] = React.useState("")
  const [defaultCurrency, setDefaultCurrency] = React.useState("")
  const [logoFile, setLogoFile] = React.useState<File | null>(null)
  const [logoPreview, setLogoPreview] = React.useState<string | null>(null)
  const fileInputRef = React.useRef<HTMLInputElement>(null)

  const updateDefaults = useUpdatePairingDefaults()
  const updateLogo = useUpdatePairingLogo()

  // Check if workspace needs setup on dialog open
  React.useEffect(() => {
    if (open && selectedWorkspace) {
      const needsSetup = !selectedWorkspace.default_country || !selectedWorkspace.default_currency
      setSetupRequired(needsSetup)
      
      // Pre-fill existing values
      setDefaultCountry(selectedWorkspace.default_country || "")
      setDefaultCurrency(selectedWorkspace.default_currency || "")
      setLogoPreview(selectedWorkspace.partner_logo_url || null)
      setLogoFile(null)
    }
  }, [open, selectedWorkspace])

  const handleLogoChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      setLogoFile(file)
      const reader = new FileReader()
      reader.onloadend = () => {
        setLogoPreview(reader.result as string)
      }
      reader.readAsDataURL(file)
    }
  }

  const handleSaveDefaults = async () => {
    if (!selectedWorkspace) return

    if (!defaultCountry || !defaultCurrency) {
      toast.error("Please fill in all required fields")
      return
    }

    try {
      // Save defaults
      await updateDefaults.mutateAsync({
        id: selectedWorkspace.id,
        default_country: defaultCountry,
        default_currency: defaultCurrency,
      })

      // Save logo if provided
      if (logoFile) {
        await updateLogo.mutateAsync({
          id: selectedWorkspace.id,
          file: logoFile,
        })
      }

      setSetupRequired(false)
      toast.success("Partner defaults saved successfully")
    } catch {
      toast.error("Failed to save partner defaults")
    }
  }

  const handleConfirm = async () => {
    if (setupRequired) {
      await handleSaveDefaults()
      return
    }
    await onConfirm()
  }

  const isSaving = updateDefaults.isPending || updateLogo.isPending

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>
            {setupRequired ? "Setup Partner Defaults" : "Publish Candidate"}
          </DialogTitle>
          <DialogDescription>
            {setupRequired ? (
              <>
                Before publishing to <strong>{selectedWorkspace?.partner_agency.company_name || selectedWorkspace?.partner_agency.full_name}</strong>, 
                set up default CV values. These will be used for all future candidates unless overridden.
              </>
            ) : (
              <>
                Publishing will add <strong>{candidateName}</strong> to your library and auto-generate a CV using your partner defaults.
              </>
            )}
          </DialogDescription>
        </DialogHeader>

        {setupRequired ? (
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="country">
                Default Country <span className="text-destructive">*</span>
              </Label>
              <Input
                id="country"
                placeholder="e.g., Saudi Arabia"
                value={defaultCountry}
                onChange={(e) => setDefaultCountry(e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="currency">
                Default Currency <span className="text-destructive">*</span>
              </Label>
              <Select value={defaultCurrency} onValueChange={setDefaultCurrency}>
                <SelectTrigger id="currency">
                  <SelectValue placeholder="Select currency" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="SAR">SAR (Saudi Riyal)</SelectItem>
                  <SelectItem value="AED">AED (UAE Dirham)</SelectItem>
                  <SelectItem value="KWD">KWD (Kuwaiti Dinar)</SelectItem>
                  <SelectItem value="QAR">QAR (Qatari Riyal)</SelectItem>
                  <SelectItem value="OMR">OMR (Omani Rial)</SelectItem>
                  <SelectItem value="BHD">BHD (Bahraini Dinar)</SelectItem>
                  <SelectItem value="JOD">JOD (Jordanian Dinar)</SelectItem>
                  <SelectItem value="USD">USD (US Dollar)</SelectItem>
                  <SelectItem value="EUR">EUR (Euro)</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="logo">Partner Logo (Optional)</Label>
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => fileInputRef.current?.click()}
                  className="flex-1"
                >
                  <Upload className="h-4 w-4 mr-2" />
                  {logoFile ? "Change Logo" : "Upload Logo"}
                </Button>
                {logoPreview && (
                  <div className="h-10 w-10 border rounded overflow-hidden">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={logoPreview} alt="Logo preview" className="h-full w-full object-contain" />
                  </div>
                )}
              </div>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                className="hidden"
                onChange={handleLogoChange}
              />
              <p className="text-xs text-muted-foreground">
                This logo will appear on generated CVs for this partner
              </p>
            </div>

            <div className="flex items-start gap-2 p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-800">
              <AlertCircle className="h-4 w-4 text-blue-600 mt-0.5 flex-shrink-0" />
              <p className="text-xs text-blue-800 dark:text-blue-300">
                These defaults apply to all candidates you publish to this partner, but can be overridden per candidate.
              </p>
            </div>
          </div>
        ) : (
          <div className="py-4">
            <div className="flex items-center gap-3 p-4 bg-green-50 dark:bg-green-950/20 rounded-lg border border-green-200 dark:border-green-800">
              <CheckCircle2 className="h-5 w-5 text-green-600 flex-shrink-0" />
              <div className="text-sm text-green-800 dark:text-green-300">
                <p className="font-medium">Ready to publish</p>
                <p className="text-xs mt-1">CV will be auto-generated using your partner defaults</p>
              </div>
            </div>
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isSaving || isPublishing}>
            Cancel
          </Button>
          <Button onClick={handleConfirm} disabled={isSaving || isPublishing}>
            {(isSaving || isPublishing) && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {setupRequired ? "Save & Continue" : "Publish"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
