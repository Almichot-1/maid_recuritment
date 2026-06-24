"use client"

import * as React from "react"
import { Loader2, Upload } from "lucide-react"
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

interface PartnerDefaultsDialogProps {
  workspace: WorkspaceSummary | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

const COUNTRIES = [
  "Saudi Arabia",
  "UAE",
  "Kuwait",
  "Qatar",
  "Bahrain",
  "Oman",
  "Lebanon",
  "Jordan",
]

const CURRENCIES = ["SAR", "AED", "KWD", "QAR", "OMR", "BHD", "JOD", "USD", "EUR"]

export function PartnerDefaultsDialog({
  workspace,
  open,
  onOpenChange,
}: PartnerDefaultsDialogProps) {
  const [defaultCountry, setDefaultCountry] = React.useState("")
  const [defaultSalary, setDefaultSalary] = React.useState("")
  const [defaultCurrency, setDefaultCurrency] = React.useState("")
  const [logoFile, setLogoFile] = React.useState<File | null>(null)
  const [logoPreview, setLogoPreview] = React.useState<string | null>(null)
  const fileInputRef = React.useRef<HTMLInputElement>(null)

  const updateDefaults = useUpdatePairingDefaults()
  const updateLogo = useUpdatePairingLogo()

  React.useEffect(() => {
    if (workspace) {
      setDefaultCountry(workspace.default_country || "")
      setDefaultSalary(workspace.default_salary || "")
      setDefaultCurrency(workspace.default_currency || "")
      setLogoPreview(workspace.partner_logo_url || null)
      setLogoFile(null)
    }
  }, [workspace])

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

  const handleSave = async () => {
    if (!workspace) return

    try {
      await updateDefaults.mutateAsync({
        id: workspace.id,
        default_country: defaultCountry,
        default_currency: defaultCurrency,
        default_salary: defaultSalary || undefined,
      })

      if (logoFile) {
        await updateLogo.mutateAsync({
          id: workspace.id,
          file: logoFile,
        })
      }

      toast.success("Partner defaults saved successfully")
      onOpenChange(false)
    } catch {
      toast.error("Failed to save partner defaults")
    }
  }

  const isSaving = updateDefaults.isPending || updateLogo.isPending

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Edit CV Defaults</DialogTitle>
          <DialogDescription>
            Configure default CV settings for{" "}
            <strong>
              {workspace?.partner_agency?.company_name ||
                workspace?.partner_agency?.full_name ||
                "this partner"}
            </strong>
            .
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="country">Default Country</Label>
            <Select value={defaultCountry} onValueChange={setDefaultCountry}>
              <SelectTrigger id="country">
                <SelectValue placeholder="Select a country" />
              </SelectTrigger>
              <SelectContent>
                {COUNTRIES.map((country) => (
                  <SelectItem key={country} value={country}>
                    {country}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="salary">Default Salary Amount</Label>
            <Input
              id="salary"
              type="number"
              min={0}
              placeholder="e.g., 2000"
              value={defaultSalary}
              onChange={(e) => setDefaultSalary(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="currency">Default Currency</Label>
            <Select value={defaultCurrency} onValueChange={setDefaultCurrency}>
              <SelectTrigger id="currency">
                <SelectValue placeholder="Select currency" />
              </SelectTrigger>
              <SelectContent>
                {CURRENCIES.map((currency) => (
                  <SelectItem key={currency} value={currency}>
                    {currency}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>Partner Logo (Optional)</Label>
            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => fileInputRef.current?.click()}
                className="flex-1"
              >
                <Upload className="mr-2 h-4 w-4" />
                {logoFile ? "Change Logo" : "Upload Logo"}
              </Button>
              {logoPreview && (
                <div className="h-10 w-10 overflow-hidden rounded border">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={logoPreview}
                    alt="Logo preview"
                    className="h-full w-full object-contain"
                  />
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
              This logo will appear on generated CVs for this partner.
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isSaving}
          >
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={isSaving}>
            {isSaving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
