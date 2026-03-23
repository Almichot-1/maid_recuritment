"use client"

import * as React from "react"
import { useFieldArray, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import {
  BadgeCheck,
  BriefcaseBusiness,
  Building2,
  Globe2,
  ImagePlus,
  Languages,
  Loader2,
  Plus,
  ShieldCheck,
  Sparkles,
  Trash2,
  UserSquare2,
} from "lucide-react"
import { z } from "zod"

import { CandidateInput, candidateSchema } from "@/lib/validations"
import { useAgencyBranding } from "@/hooks/use-agency-branding"
import { useCurrentUser } from "@/hooks/use-auth"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { DocumentUpload } from "./document-upload"

const SKILLS_OPTIONS = [
  "Cooking",
  "Cleaning",
  "Childcare",
  "Elderly Care",
  "Laundry",
  "Ironing",
  "Driving",
  "First Aid",
  "Pet Care",
]

const LANGUAGES_OPTIONS = ["Arabic", "English", "Amharic", "French", "Swahili"]
const PROFICIENCY_OPTIONS = ["Basic", "Intermediate", "Fluent"]

interface CandidateFormProps {
  initialData?: Partial<CandidateFormValues>
  onSubmit: (
    data: CandidateInput,
    context?: { submitter: "default" | "create_another" }
  ) => Promise<{ resetForm?: boolean } | void> | { resetForm?: boolean } | void
  isLoading?: boolean
  onDocumentChange?: (documentType: "passport" | "photo" | "video", file: File | null) => void
  showDocuments?: boolean
}

type CandidateFormValues = z.input<typeof candidateSchema>

function SectionTitle({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode
  title: string
  description: string
}) {
  return (
    <div className="flex items-start gap-4">
      <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-slate-950 text-white shadow-sm">
        {icon}
      </div>
      <div className="space-y-1">
        <h3 className="text-lg font-semibold tracking-tight text-foreground">{title}</h3>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
    </div>
  )
}

export function CandidateForm({
  initialData,
  onSubmit,
  isLoading,
  onDocumentChange,
  showDocuments = true,
}: CandidateFormProps) {
  const { user } = useCurrentUser()
  const { hasLogo, logoDataURL } = useAgencyBranding()
  const [documentResetKey, setDocumentResetKey] = React.useState(0)

  const blankFormValues = React.useMemo<CandidateFormValues>(
    () => ({
      full_name: "",
      age: undefined,
      experience_years: undefined,
      skills: [],
      languages: [{ language: LANGUAGES_OPTIONS[0], proficiency: PROFICIENCY_OPTIONS[0] }],
    }),
    []
  )

  const form = useForm<CandidateFormValues, undefined, CandidateInput>({
    resolver: zodResolver(candidateSchema),
    defaultValues: initialData || blankFormValues,
  })

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "languages",
  })
  const [documentPresence, setDocumentPresence] = React.useState({
    passport: false,
    photo: false,
    video: false,
  })

  const selectedSkills = form.watch("skills") || []
  const watchedLanguages = form.watch("languages") || []
  const fullName = form.watch("full_name")

  const requiredDocumentChecklist = [
    { label: "Passport", done: documentPresence.passport || !showDocuments },
    { label: "Full body photo", done: documentPresence.photo || !showDocuments },
    { label: "Video interview (optional)", done: documentPresence.video || !showDocuments },
  ]

  const isEditing = !!initialData
  const addLanguageEntry = () => append({ language: LANGUAGES_OPTIONS[0], proficiency: PROFICIENCY_OPTIONS[0] })
  const resetForNextCandidate = React.useCallback(() => {
    form.reset(blankFormValues)
    setDocumentPresence({
      passport: false,
      photo: false,
      video: false,
    })
    setDocumentResetKey((current) => current + 1)
    onDocumentChange?.("passport", null)
    onDocumentChange?.("photo", null)
    onDocumentChange?.("video", null)
  }, [blankFormValues, form, onDocumentChange])

  const handleFormSubmit = form.handleSubmit(async (data, event) => {
    const submitter = (event?.nativeEvent as SubmitEvent | undefined)?.submitter as HTMLButtonElement | null
    const submitMode = submitter?.dataset.submitMode === "create_another" ? "create_another" : "default"
    const result = await onSubmit(data, { submitter: submitMode })

    if (!isEditing && submitMode === "create_another" && result?.resetForm) {
      resetForNextCandidate()
    }
  })

  return (
    <Form {...form}>
      <form onSubmit={handleFormSubmit} className="w-full">
        <div className="grid gap-6 xl:grid-cols-[minmax(0,1.45fr)_360px]">
          <div className="space-y-6">
            <Card className="overflow-hidden border-border/70 bg-[radial-gradient(circle_at_top_left,_rgba(251,191,36,0.22),_transparent_32%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.96))] text-white shadow-xl">
              <CardContent className="grid gap-6 p-6 lg:grid-cols-[minmax(0,1fr)_220px]">
                <div className="space-y-4">
                  <Badge className="w-fit rounded-full border-0 bg-white/15 px-3 py-1 text-[11px] uppercase tracking-[0.24em] text-amber-200 hover:bg-white/15">
                    Candidate workflow
                  </Badge>
                  <div className="space-y-2">
                    <h2 className="text-3xl font-semibold tracking-tight">
                      {isEditing ? "Refine and update this candidate profile" : "Create a polished candidate profile in one pass"}
                    </h2>
                    <p className="max-w-2xl text-sm text-slate-200/90">
                      Fill the profile, capture the documents, and keep the submission clean enough for selection and post-approval tracking.
                    </p>
                  </div>
                  <div className="flex flex-wrap gap-3 text-xs text-slate-100/90">
                    <span className="rounded-full border border-white/15 bg-white/10 px-3 py-1.5">
                      Draft first, publish when ready
                    </span>
                    <span className="rounded-full border border-white/15 bg-white/10 px-3 py-1.5">
                      Passport and photo required for CV
                    </span>
                    <span className="rounded-full border border-white/15 bg-white/10 px-3 py-1.5">
                      Post-approval tracking starts automatically
                    </span>
                  </div>
                </div>

                <div className="rounded-3xl border border-white/10 bg-white/10 p-5 backdrop-blur">
                  <div className="flex items-center gap-4">
                    <div className="flex h-16 w-16 items-center justify-center overflow-hidden rounded-2xl border border-white/15 bg-white/10">
                      {hasLogo ? (
                        <img
                          src={logoDataURL}
                          alt={`${user?.company_name || user?.full_name || "Agency"} logo`}
                          className="h-full w-full object-cover"
                        />
                      ) : (
                        <ImagePlus className="h-8 w-8 text-amber-200" />
                      )}
                    </div>
                    <div className="min-w-0">
                      <p className="text-xs uppercase tracking-[0.24em] text-amber-200/90">Agency branding</p>
                      <p className="truncate text-sm font-semibold text-white">
                        {user?.company_name || "Agency profile"}
                      </p>
                      <p className="text-xs text-slate-200/80">
                        {hasLogo ? "Saved once and ready to reuse" : "Upload a reusable logo from Settings"}
                      </p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-border/70 shadow-sm">
              <CardHeader>
                <SectionTitle
                  icon={<UserSquare2 className="h-5 w-5" />}
                  title="Personal profile"
                  description="Capture the core identity details that foreign employers will compare first."
                />
              </CardHeader>
              <CardContent className="grid gap-5 md:grid-cols-2">
                <FormField
                  control={form.control}
                  name="full_name"
                  render={({ field }) => (
                    <FormItem className="md:col-span-2">
                      <FormLabel>Full name</FormLabel>
                      <FormControl>
                        <Input placeholder="For example: Aster Demeke" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="age"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Age</FormLabel>
                      <FormControl>
                        <Input
                          type="number"
                          min={18}
                          max={65}
                          name={field.name}
                          value={typeof field.value === "number" || typeof field.value === "string" ? field.value : ""}
                          onBlur={field.onBlur}
                          onChange={field.onChange}
                          ref={field.ref}
                        />
                      </FormControl>
                      <FormDescription>Allowed range: 18 to 65</FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="experience_years"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Years of experience</FormLabel>
                      <FormControl>
                        <Input
                          type="number"
                          min={0}
                          max={30}
                          name={field.name}
                          value={typeof field.value === "number" || typeof field.value === "string" ? field.value : ""}
                          onBlur={field.onBlur}
                          onChange={field.onChange}
                          ref={field.ref}
                        />
                      </FormControl>
                      <FormDescription>How long she has worked in similar roles</FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </CardContent>
            </Card>

            <Card className="border-border/70 shadow-sm">
              <CardHeader>
                <SectionTitle
                  icon={<BriefcaseBusiness className="h-5 w-5" />}
                  title="Skills and service strengths"
                  description="Select the practical service areas this candidate can confidently deliver."
                />
              </CardHeader>
              <CardContent className="space-y-5">
                <FormField
                  control={form.control}
                  name="skills"
                  render={() => (
                    <FormItem>
                      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                        {SKILLS_OPTIONS.map((skill) => (
                          <FormField
                            key={skill}
                            control={form.control}
                            name="skills"
                            render={({ field }) => {
                              const checked = field.value?.includes(skill)
                              return (
                                <FormItem className="rounded-2xl border border-border/70 bg-muted/20 p-4 transition-colors hover:bg-muted/40">
                                  <div className="flex items-center gap-3">
                                    <FormControl>
                                      <Checkbox
                                        checked={checked}
                                        onCheckedChange={(nextChecked) => {
                                          return nextChecked
                                            ? field.onChange([...(field.value || []), skill])
                                            : field.onChange((field.value || []).filter((value) => value !== skill))
                                        }}
                                      />
                                    </FormControl>
                                    <FormLabel className="cursor-pointer text-sm font-semibold">{skill}</FormLabel>
                                  </div>
                                </FormItem>
                              )
                            }}
                          />
                        ))}
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </CardContent>
            </Card>

            <Card className="border-border/70 shadow-sm">
              <CardHeader>
                <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
                  <SectionTitle
                    icon={<Languages className="h-5 w-5" />}
                    title="Languages"
                    description="Track the languages the candidate can use and the level she can handle confidently."
                  />
                  <Button type="button" variant="outline" className="border-dashed md:shrink-0" onClick={addLanguageEntry}>
                    <Plus className="mr-2 h-4 w-4" />
                    Add another
                  </Button>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                {fields.map((field, index) => (
                  <div key={field.id} className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                    <div className="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] md:items-end">
                      <FormField
                        control={form.control}
                        name={`languages.${index}.language`}
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Language</FormLabel>
                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                              <FormControl>
                                <SelectTrigger>
                                  <SelectValue placeholder="Select language" />
                                </SelectTrigger>
                              </FormControl>
                              <SelectContent>
                                {LANGUAGES_OPTIONS.map((option) => (
                                  <SelectItem key={option} value={option}>
                                    {option}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <FormField
                        control={form.control}
                        name={`languages.${index}.proficiency`}
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Proficiency</FormLabel>
                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                              <FormControl>
                                <SelectTrigger>
                                  <SelectValue placeholder="Select proficiency" />
                                </SelectTrigger>
                              </FormControl>
                              <SelectContent>
                                {PROFICIENCY_OPTIONS.map((option) => (
                                  <SelectItem key={option} value={option}>
                                    {option}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                            <FormMessage />
                          </FormItem>
                        )}
                      />

                      <Button
                        type="button"
                        variant="outline"
                        className="border-dashed text-destructive hover:text-destructive"
                        disabled={fields.length === 1}
                        onClick={() => remove(index)}
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        Remove
                      </Button>
                    </div>
                  </div>
                ))}

                <Button
                  type="button"
                  variant="outline"
                  className="border-dashed"
                  onClick={addLanguageEntry}
                >
                  <Plus className="mr-2 h-4 w-4" />
                  Add another language
                </Button>
              </CardContent>
            </Card>

            {showDocuments ? (
              <Card className="border-border/70 shadow-sm">
                <CardHeader>
                  <SectionTitle
                    icon={<ShieldCheck className="h-5 w-5" />}
                    title="Documents and media"
                    description="Upload the key recruitment files now so the profile is immediately useful after creation."
                  />
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="grid gap-5 xl:grid-cols-2">
                    <div className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                      <DocumentUpload
                        key={`passport-${documentResetKey}`}
                        documentType="passport"
                        title="Passport document"
                        description="PDF, JPG, or PNG. This is required before CV download."
                        accept={{
                          "application/pdf": [".pdf"],
                          "image/jpeg": [".jpg", ".jpeg"],
                          "image/png": [".png"],
                        }}
                        maxSize={10485760}
                        onUpload={(file) => {
                          setDocumentPresence((current) => ({ ...current, passport: true }))
                          onDocumentChange?.("passport", file)
                        }}
                        onRemove={() => {
                          setDocumentPresence((current) => ({ ...current, passport: false }))
                          onDocumentChange?.("passport", null)
                        }}
                      />
                    </div>

                    <div className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                      <DocumentUpload
                        key={`photo-${documentResetKey}`}
                        documentType="photo"
                        title="Full body photo"
                        description="Use a clean photo with good light and a clear full-body view."
                        accept={{
                          "image/jpeg": [".jpg", ".jpeg"],
                          "image/png": [".png"],
                        }}
                        maxSize={10485760}
                        onUpload={(file) => {
                          setDocumentPresence((current) => ({ ...current, photo: true }))
                          onDocumentChange?.("photo", file)
                        }}
                        onRemove={() => {
                          setDocumentPresence((current) => ({ ...current, photo: false }))
                          onDocumentChange?.("photo", null)
                        }}
                      />
                    </div>

                    <div className="rounded-2xl border border-border/70 bg-muted/20 p-4 xl:col-span-2">
                      <DocumentUpload
                        key={`video-${documentResetKey}`}
                        documentType="video"
                        title="Video interview"
                        description="Optional, but helpful for faster employer review."
                        accept={{ "video/mp4": [".mp4"] }}
                        maxSize={52428800}
                        onUpload={(file) => {
                          setDocumentPresence((current) => ({ ...current, video: true }))
                          onDocumentChange?.("video", file)
                        }}
                        onRemove={() => {
                          setDocumentPresence((current) => ({ ...current, video: false }))
                          onDocumentChange?.("video", null)
                        }}
                      />
                    </div>
                  </div>
                </CardContent>
              </Card>
            ) : null}
          </div>

          <div className="space-y-6 xl:sticky xl:top-6 xl:self-start">
            <Card className="overflow-hidden border-border/70 shadow-lg">
              <CardHeader className="bg-gradient-to-br from-amber-50 via-background to-slate-50">
                <CardTitle className="flex items-center gap-2 text-lg">
                  <Sparkles className="h-5 w-5 text-amber-600" />
                  Submission summary
                </CardTitle>
                <CardDescription>Keep the profile complete before you publish it to foreign agencies.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-5 p-6">
                <div className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                  <div className="flex items-center gap-4">
                    <div className="flex h-16 w-16 items-center justify-center overflow-hidden rounded-2xl border bg-background shadow-sm">
                      {hasLogo ? (
                        <img
                          src={logoDataURL}
                          alt={`${user?.company_name || user?.full_name || "Agency"} branding`}
                          className="h-full w-full object-cover"
                        />
                      ) : (
                        <Building2 className="h-7 w-7 text-muted-foreground" />
                      )}
                    </div>
                    <div className="min-w-0">
                      <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Agency</p>
                      <p className="truncate text-sm font-semibold">{user?.company_name || "Agency profile"}</p>
                      <p className="text-xs text-muted-foreground">
                        {hasLogo ? "Reusable logo is ready" : "Add a reusable logo from Settings"}
                      </p>
                    </div>
                  </div>
                </div>

                <div className="space-y-3">
                  <SummaryItem
                    label="Candidate name"
                    value={fullName?.trim() ? fullName : "Not filled yet"}
                  />
                  <SummaryItem
                    label="Skills selected"
                    value={selectedSkills.length ? `${selectedSkills.length} selected` : "Select at least one skill"}
                  />
                  <SummaryItem
                    label="Languages tracked"
                    value={watchedLanguages.length ? `${watchedLanguages.length} language entries` : "Add a language"}
                  />
                </div>

                <div className="space-y-3 rounded-2xl border border-border/70 bg-muted/20 p-4">
                  <div className="flex items-center gap-2">
                    <Globe2 className="h-4 w-4 text-slate-700" />
                    <p className="text-sm font-semibold">Before you submit</p>
                  </div>
                  {requiredDocumentChecklist.map((item) => (
                    <div key={item.label} className="flex items-center gap-2 text-sm">
                      <BadgeCheck className={`h-4 w-4 ${item.done ? "text-green-600" : "text-muted-foreground"}`} />
                      <span className={item.done ? "text-foreground" : "text-muted-foreground"}>{item.label}</span>
                    </div>
                  ))}
                  {!showDocuments ? (
                    <p className="text-xs text-muted-foreground">
                      Profile details are edited here. The latest supporting files can be managed from the surrounding candidate workspace.
                    </p>
                  ) : (
                    <p className="text-xs text-muted-foreground">
                      Uploaded files are attached right after the candidate record is saved.
                    </p>
                  )}
                </div>

              </CardContent>
            </Card>
          </div>
        </div>

        <div className="mt-8 flex flex-col-reverse gap-3 border-t border-border/70 pt-6 sm:flex-row sm:justify-end">
          {!isEditing ? (
            <Button
              type="submit"
              size="lg"
              variant="outline"
              className="min-w-[220px]"
              data-submit-mode="create_another"
              disabled={isLoading}
            >
              {isLoading ? <Loader2 className="mr-2 h-5 w-5 animate-spin" /> : <Plus className="mr-2 h-5 w-5" />}
              {isLoading ? "Saving..." : "Create & Add Another"}
            </Button>
          ) : null}
          <Button type="submit" size="lg" className="min-w-[220px]" data-submit-mode="default" disabled={isLoading}>
            {isLoading ? <Loader2 className="mr-2 h-5 w-5 animate-spin" /> : null}
            {isLoading ? (isEditing ? "Saving Changes..." : "Creating Candidate...") : isEditing ? "Save Changes" : "Create Candidate"}
          </Button>
        </div>
      </form>
    </Form>
  )
}

function SummaryItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="space-y-1">
      <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">{label}</p>
      <p className="text-sm font-semibold text-foreground">{value}</p>
    </div>
  )
}
