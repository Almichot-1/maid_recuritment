"use client"

import * as React from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"
import { Edit2, Loader2, Save, X } from "lucide-react"

import { Candidate, CandidatePairOverride } from "@/types"
import { usePairingContext } from "@/hooks/use-pairings"
import { useSetPairOverride } from "@/hooks/use-candidates"

import { Button } from "@/components/ui/button"
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion"
import { Badge } from "@/components/ui/badge"

const overrideSchema = z.object({
  country_applied: z.string().max(100),
  salary_offered: z.string().max(100),
})

type OverrideFormValues = z.infer<typeof overrideSchema>

interface CandidatePartnerOverridesProps {
  candidate: Candidate
}

export function CandidatePartnerOverrides({ candidate }: CandidatePartnerOverridesProps) {
  const { context } = usePairingContext()
  const { mutateAsync: setOverride } = useSetPairOverride(candidate.id)
  
  const [editingPairingId, setEditingPairingId] = React.useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = React.useState(false)

  const form = useForm<OverrideFormValues>({
    resolver: zodResolver(overrideSchema),
    defaultValues: {
      country_applied: "",
      salary_offered: "",
    },
  })

  if (!context || !context.workspaces.length) {
    return null
  }

  const handleEdit = (pairingId: string, currentOverride?: CandidatePairOverride) => {
    setEditingPairingId(pairingId)
    form.reset({
      country_applied: currentOverride?.country_applied || candidate.country_applied || "",
      salary_offered: currentOverride?.salary_offered || candidate.salary_offered || "",
    })
  }

  const handleCancel = () => {
    setEditingPairingId(null)
    form.reset()
  }

  const onSubmit = async (values: OverrideFormValues) => {
    if (!editingPairingId) return
    try {
      setIsSubmitting(true)
      await setOverride({
        pairing_id: editingPairingId,
        country_applied: values.country_applied,
        salary_offered: values.salary_offered,
      })
      setEditingPairingId(null)
    } finally {
      setIsSubmitting(false)
    }
  }

  const getWorkspaceName = (companyName?: string, fullName?: string) => {
    return companyName?.trim() || fullName?.trim() || "Unknown Partner"
  }

  return (
    <Accordion type="single" collapsible className="w-full">
      {context.workspaces.map((ws) => {
        const override = candidate.pair_overrides?.find((o) => o.pairing_id === ws.id)
        const isEditing = editingPairingId === ws.id
        const partnerName = getWorkspaceName(
          ws.partner_agency.company_name,
          ws.partner_agency.full_name
        )
        const hasOverride = !!override && (!!override.country_applied || !!override.salary_offered)

        return (
          <AccordionItem key={ws.id} value={ws.id} className="border rounded-lg mb-2">
            <AccordionTrigger className="px-4 hover:no-underline">
              <div className="flex items-center gap-2 flex-1 text-left">
                <span className="font-medium">{partnerName}</span>
                {hasOverride && (
                  <Badge variant="secondary" className="text-xs">
                    Custom
                  </Badge>
                )}
              </div>
            </AccordionTrigger>
            <AccordionContent className="px-4 pb-4">
              {isEditing ? (
                <Form {...form}>
                  <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                    <FormField
                      control={form.control}
                      name="country_applied"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Country Applied</FormLabel>
                          <FormControl>
                            <Input placeholder="e.g., Saudi Arabia" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <FormField
                      control={form.control}
                      name="salary_offered"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Salary Offered</FormLabel>
                          <FormControl>
                            <Input placeholder="e.g., 1500 SAR" {...field} />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <div className="flex gap-2">
                      <Button
                        type="button"
                        variant="outline"
                        onClick={handleCancel}
                        disabled={isSubmitting}
                      >
                        <X className="h-4 w-4 mr-2" />
                        Cancel
                      </Button>
                      <Button type="submit" disabled={isSubmitting}>
                        {isSubmitting ? (
                          <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        ) : (
                          <Save className="h-4 w-4 mr-2" />
                        )}
                        Save Override
                      </Button>
                    </div>
                  </form>
                </Form>
              ) : (
                <div className="space-y-3">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <p className="text-sm font-medium text-muted-foreground mb-1">
                        Country Applied
                      </p>
                      <p className={`text-sm ${override?.country_applied ? "font-semibold text-primary" : ""}`}>
                        {override?.country_applied || candidate.country_applied || "—"}
                      </p>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-muted-foreground mb-1">
                        Salary Offered
                      </p>
                      <p className={`text-sm ${override?.salary_offered ? "font-semibold text-primary" : ""}`}>
                        {override?.salary_offered || candidate.salary_offered || "—"}
                      </p>
                    </div>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleEdit(ws.id, override)}
                  >
                    <Edit2 className="h-4 w-4 mr-2" />
                    {hasOverride ? "Edit Override" : "Add Override"}
                  </Button>
                </div>
              )}
            </AccordionContent>
          </AccordionItem>
        )
      })}
    </Accordion>
  )
}
