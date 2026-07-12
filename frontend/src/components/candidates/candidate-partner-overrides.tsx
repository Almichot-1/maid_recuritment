"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"
import { Check, Loader2, Save } from "lucide-react"

import { Candidate, CandidatePairOverride } from "@/types"
import { usePairingContext } from "@/hooks/use-pairings"
import { useSetPairOverride } from "@/hooks/use-candidates"

import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

const COUNTRIES = [
  "Saudi Arabia",
  "United Arab Emirates",
  "Kuwait",
  "Qatar",
  "Bahrain",
  "Oman",
  "Lebanon",
  "Jordan",
]

const overrideSchema = z.object({
  country_applied: z.string().min(1, "Country is required").max(100),
  salary_offered: z.string().min(1, "Salary is required").max(100),
})

type OverrideFormValues = z.infer<typeof overrideSchema>

interface CandidatePartnerOverridesProps {
  candidate: Candidate
}

export function CandidatePartnerOverrides({ candidate }: CandidatePartnerOverridesProps) {
  const { context } = usePairingContext()

  const [savedStates, setSavedStates] = React.useState<Record<string, boolean>>({})

  if (!context || !context.workspaces.length) {
    return null
  }

  return (
    <div className="space-y-4">
      {context.workspaces.map((ws) => {
        const override = candidate.pair_overrides?.find((o) => o.pairing_id === ws.id)
        const partnerName = ws.partner_agency?.company_name || ws.partner_agency?.full_name || ws.partner_agency?.email || "Partner"
        return (
          <PartnerOverrideRow
            key={ws.id}
            pairingId={ws.id}
            partnerName={partnerName}
            candidateId={candidate.id}
            override={override}
            onSaved={() => {
              setSavedStates((prev) => ({ ...prev, [ws.id]: true }))
              setTimeout(() => {
                setSavedStates((prev) => ({ ...prev, [ws.id]: false }))
              }, 2000)
            }}
            showSaved={savedStates[ws.id] || false}
          />
        )
      })}
    </div>
  )
}

interface PartnerOverrideRowProps {
  pairingId: string
  partnerName: string
  candidateId: string
  override?: CandidatePairOverride
  onSaved: () => void
  showSaved: boolean
}

function PartnerOverrideRow({
  pairingId,
  partnerName,
  candidateId,
  override,
  onSaved,
  showSaved,
}: PartnerOverrideRowProps) {
  const { mutateAsync: setOverride, isPending } = useSetPairOverride(candidateId)

  const [logoUrl, setLogoUrl] = React.useState(override?.logo_url || "")

  const form = useForm<OverrideFormValues>({
    resolver: zodResolver(overrideSchema),
    defaultValues: {
      country_applied: override?.country_applied || "",
      salary_offered: override?.salary_offered || "",
    },
  })

  React.useEffect(() => {
    form.reset({
      country_applied: override?.country_applied || "",
      salary_offered: override?.salary_offered || "",
    })
    setLogoUrl(override?.logo_url || "")
  }, [override, form])

  const onSubmit = async (values: OverrideFormValues) => {
    await setOverride({
      pairing_id: pairingId,
      country_applied: values.country_applied,
      salary_offered: values.salary_offered,
      logo_url: logoUrl || undefined,
    })
    onSaved()
  }

  return (
    <div className="rounded-lg border border-border/70 bg-card p-4 space-y-3">
      <div className="flex items-center justify-between">
        <p className="font-semibold text-sm">{partnerName}</p>
        {showSaved && (
          <span className="flex items-center gap-1 text-xs text-green-600">
            <Check className="h-3 w-3" />
            Saved
          </span>
        )}
      </div>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-3">
          <div className="grid gap-3 sm:grid-cols-2">
            <div className="space-y-1.5">
              <Label htmlFor={`country-${pairingId}`}>Country Applied</Label>
              <Select
                value={form.watch("country_applied")}
                onValueChange={(val) => form.setValue("country_applied", val)}
              >
                <SelectTrigger id={`country-${pairingId}`}>
                  <SelectValue placeholder="Select country..." />
                </SelectTrigger>
                <SelectContent>
                  {COUNTRIES.map((c) => (
                    <SelectItem key={c} value={c}>{c}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {form.formState.errors.country_applied && (
                <p className="text-xs text-destructive">{form.formState.errors.country_applied.message}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor={`salary-${pairingId}`}>Salary Offered</Label>
              <Input
                id={`salary-${pairingId}`}
                placeholder="e.g., 1000 SR, 400 USD"
                {...form.register("salary_offered")}
              />
              {form.formState.errors.salary_offered && (
                <p className="text-xs text-destructive">{form.formState.errors.salary_offered.message}</p>
              )}
            </div>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor={`logo-${pairingId}`}>Partner Logo URL (optional)</Label>
            <Input
              id={`logo-${pairingId}`}
              placeholder="https://example.com/logo.png"
              value={logoUrl}
              onChange={(e) => setLogoUrl(e.target.value)}
            />
            <p className="text-xs text-muted-foreground">
              Overrides the workspace default logo for this candidate&apos;s CV.
            </p>
          </div>
        <div className="flex justify-end">
          <Button type="submit" size="sm" disabled={isPending}>
            {isPending ? (
              <>
                <Loader2 className="h-3 w-3 mr-1 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="h-3 w-3 mr-1" />
                Save
              </>
            )}
          </Button>
        </div>
      </form>
    </div>
  )
}
