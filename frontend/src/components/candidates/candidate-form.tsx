"use client";

import * as React from "react";
import { FieldErrors, useFieldArray, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  BriefcaseBusiness,
  Languages,
  Loader2,
  Plus,
  ShieldCheck,
  Trash2,
  UserSquare2,
} from "lucide-react";
import { toast } from "sonner";

import { useParsePassport } from "@/hooks/use-passport-ocr";
import { CandidateInput, candidateSchema } from "@/lib/validations";
import { useAgencyBranding } from "@/hooks/use-agency-branding";
import { useCurrentUser } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DocumentUpload } from "./document-upload";
import { PassportData } from "@/types";

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
];

const LANGUAGES_OPTIONS = ["Arabic", "English", "Amharic", "French", "Swahili"];
const PROFICIENCY_OPTIONS = ["Basic", "Intermediate", "Fluent"];
const RELIGION_OPTIONS = ["Muslim", "Christian", "Other"];
const MARITAL_STATUS_OPTIONS = ["Single", "Married", "Divorced", "Widowed"];
const EDUCATION_LEVEL_OPTIONS = [
  "Elementary",
  "Secondary",
  "High School",
  "Diploma",
  "Degree",
  "Other",
];

interface CandidateFormProps {
  candidateId?: string;
  mode?: "create" | "edit";
  initialData?: Partial<CandidateFormValues>;
  initialDocuments?: Partial<Record<"passport" | "photo" | "video", File | null>>;
  onSubmit: (
    data: CandidateInput,
    context?: { submitter: "default" | "create_another" },
  ) => Promise<{ resetForm?: boolean } | void> | { resetForm?: boolean } | void;
  isLoading?: boolean;
  onDocumentChange?: (
    documentType: "passport" | "photo" | "video",
    file: File | null,
  ) => void;
  onDraftChange?: (draft: CandidateFormValues) => void;
  onClearDraft?: () => void;
  showDocuments?: boolean;
  resetSignal?: number;
}

export type CandidateFormValues = {
  full_name: string;
  nationality?: string;
  date_of_birth?: string;
  age?: number | string;
  place_of_birth?: string;
  religion?: string;
  marital_status?: string;
  children_count?: number | string;
  education_level?: string;
  experience_years?: number | string;
  skills: string[];
  languages: Array<{ language: string; proficiency: string }>;
};

function SectionTitle({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
}) {
  return (
    <div className="flex items-start gap-4">
      <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-slate-950 text-white shadow-sm">
        {icon}
      </div>
      <div className="space-y-1">
        <h3 className="text-lg font-semibold tracking-tight text-foreground">
          {title}
        </h3>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
    </div>
  );
}

export function CandidateForm({
  candidateId,
  mode = "create",
  initialData,
  initialDocuments,
  onSubmit,
  isLoading,
  onDocumentChange,
  onDraftChange,
  onClearDraft,
  showDocuments = true,
  resetSignal = 0,
}: CandidateFormProps) {
  const { user } = useCurrentUser();
  const { hasLogo } = useAgencyBranding();
  const { mutateAsync: parsePassport, isPending: isParsingPassport } =
    useParsePassport(candidateId);
  const [documentResetKey, setDocumentResetKey] = React.useState(0);
  const [ageAutoSyncEnabled, setAgeAutoSyncEnabled] = React.useState(true);

  const blankFormValues = React.useMemo<CandidateFormValues>(
    () => ({
      full_name: "",
      nationality: "",
      date_of_birth: "",
      age: undefined,
      place_of_birth: "",
      religion: "",
      marital_status: "",
      children_count: undefined,
      education_level: "",
      experience_years: undefined,
      skills: [],
      languages: [
        { language: LANGUAGES_OPTIONS[0], proficiency: PROFICIENCY_OPTIONS[0] },
      ],
    }),
    [],
  );

  const form = useForm<CandidateFormValues, undefined, CandidateInput>({
    resolver: zodResolver(candidateSchema) as never,
    defaultValues: initialData || blankFormValues,
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "languages",
  });
  const watchedDateOfBirth = form.watch("date_of_birth");
  const watchedAge = form.watch("age");

  const isEditing = mode === "edit";
  const addLanguageEntry = () =>
    append({
      language: LANGUAGES_OPTIONS[0],
      proficiency: PROFICIENCY_OPTIONS[0],
    });

  React.useEffect(() => {
    if (!onDraftChange) {
      return;
    }

    const subscription = form.watch((values) => {
      onDraftChange({
        ...blankFormValues,
        ...values,
        skills: values.skills || [],
        languages:
          values.languages && values.languages.length > 0
            ? values.languages.map((language) => ({
                language: language.language || LANGUAGES_OPTIONS[0],
                proficiency: language.proficiency || PROFICIENCY_OPTIONS[0],
              }))
            : blankFormValues.languages,
      });
    });

    return () => subscription.unsubscribe();
  }, [blankFormValues, form, onDraftChange]);

  React.useEffect(() => {
    const derivedAge = calculateAgeFromDate(watchedDateOfBirth);
    if (derivedAge === null || !ageAutoSyncEnabled) {
      return;
    }

    const currentAge =
      typeof watchedAge === "number"
        ? watchedAge
        : typeof watchedAge === "string" && watchedAge.trim()
          ? Number(watchedAge)
          : null;

    if (currentAge === null || Number.isNaN(currentAge)) {
      form.setValue("age", derivedAge, {
        shouldDirty: true,
        shouldValidate: true,
      });
    }
  }, [ageAutoSyncEnabled, form, watchedAge, watchedDateOfBirth]);

  const resetForNextCandidate = React.useCallback(() => {
    form.reset(blankFormValues);
    setAgeAutoSyncEnabled(true);
    setDocumentResetKey((current) => current + 1);
    onDocumentChange?.("passport", null);
    onDocumentChange?.("photo", null);
    onDocumentChange?.("video", null);
    onClearDraft?.();
  }, [blankFormValues, form, onClearDraft, onDocumentChange]);

  React.useEffect(() => {
    if (resetSignal <= 0) {
      return
    }
    resetForNextCandidate()
  }, [resetForNextCandidate, resetSignal])

  const handleFormSubmit = form.handleSubmit(async (data, event) => {
    const submitter = (event?.nativeEvent as SubmitEvent | undefined)
      ?.submitter as HTMLButtonElement | null;
    const submitMode =
      submitter?.dataset.submitMode === "create_another"
        ? "create_another"
        : "default";
    const result = await onSubmit(data, { submitter: submitMode });

    if (!isEditing && submitMode === "create_another" && result?.resetForm) {
      resetForNextCandidate();
    }
  }, (errors: FieldErrors<CandidateFormValues>) => {
    const message = isEditing
      ? "Please fix the highlighted fields before saving changes."
      : "Please fix the highlighted fields before creating the candidate.";

    toast.error(message);

    window.requestAnimationFrame(() => {
      const invalidElement = document.querySelector<HTMLElement>(
        "[aria-invalid='true']",
      );

      invalidElement?.scrollIntoView({
        behavior: "smooth",
        block: "center",
      });
      invalidElement?.focus();
    });

    if (Object.keys(errors).length === 0) {
      return;
    }
  });

  const applyPassportAutofill = React.useCallback((parsed: PassportData) => {
    if (parsed.holder_name?.trim()) {
      form.setValue("full_name", parsed.holder_name.trim(), {
        shouldDirty: true,
        shouldValidate: true,
      });
    }

    if (parsed.nationality?.trim()) {
      form.setValue("nationality", parsed.nationality.trim().toUpperCase(), {
        shouldDirty: true,
        shouldValidate: true,
      });
    }

    const normalizedDateOfBirth = normalizeDateInputValue(parsed.date_of_birth);
    if (normalizedDateOfBirth) {
      setAgeAutoSyncEnabled(true);
      form.setValue("date_of_birth", normalizedDateOfBirth, {
        shouldDirty: true,
        shouldValidate: true,
      });

      const derivedAge = calculateAgeFromDate(normalizedDateOfBirth);
      if (derivedAge !== null) {
        form.setValue("age", derivedAge, {
          shouldDirty: true,
          shouldValidate: true,
        });
      }
    }

    if (parsed.place_of_birth?.trim()) {
      form.setValue("place_of_birth", parsed.place_of_birth.trim(), {
        shouldDirty: true,
        shouldValidate: true,
      });
    }
  }, [form]);

  const handlePassportFileSelected = React.useCallback((file: File | null) => {
    if (!file || !file.type.startsWith("image/")) {
      return;
    }

    const runExtraction = async () => {
      try {
        const parsed = await parsePassport(file);
        applyPassportAutofill(parsed);
      } catch {
        // The mutation already surfaces the error to the user.
      }
    };

    window.setTimeout(() => {
      void runExtraction();
    }, 80);
  }, [applyPassportAutofill, parsePassport]);

  return (
    <Form {...form}>
      <form onSubmit={handleFormSubmit} className="w-full">
        <div className="space-y-6">
            <Card className="overflow-hidden border-border/70 bg-[radial-gradient(circle_at_top_left,_rgba(251,191,36,0.22),_transparent_32%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.96))] text-white shadow-xl">
              <CardContent className="space-y-5 p-5 sm:p-6">
                <div className="space-y-4">
                  <Badge className="w-fit rounded-full border-0 bg-white/15 px-3 py-1 text-[11px] uppercase tracking-[0.24em] text-amber-200 hover:bg-white/15">
                    Candidate workflow
                  </Badge>
                  <div className="space-y-2">
                    <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">
                      {isEditing
                        ? "Refine and update this candidate profile"
                        : "Create a polished candidate profile in one pass"}
                    </h2>
                    <p className="max-w-2xl text-sm text-slate-200/90">
                      Fill the profile, capture the documents, and keep the
                      submission clean enough for selection and post-approval
                      tracking.
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
                      Passport details fill automatically from the image
                    </span>
                  </div>
                </div>

                <div className="rounded-3xl border border-white/10 bg-white/10 p-5 backdrop-blur">
                  <p className="text-xs uppercase tracking-[0.24em] text-amber-200/90">
                    Agency branding
                  </p>
                  <p className="mt-2 truncate text-sm font-semibold text-white">
                    {user?.company_name || "Agency profile"}
                  </p>
                  <p className="mt-1 text-xs text-slate-200/80">
                    {hasLogo
                      ? "Your saved branding will still be used inside generated CVs."
                      : "You can still add a reusable logo later from Settings."}
                  </p>
                </div>
              </CardContent>
            </Card>

            {showDocuments ? (
              <Card className="border-border/70 shadow-sm">
                <CardHeader>
                  <SectionTitle
                    icon={<ShieldCheck className="h-5 w-5" />}
                    title="Documents and media"
                    description="Start with the passport and full-body photo so the profile is ready for OCR, CV generation, and faster employer review."
                  />
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="grid gap-5 md:grid-cols-2">
                    <div className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                      <DocumentUpload
                        key={`passport-${documentResetKey}`}
                        initialFile={initialDocuments?.passport || null}
                        documentType="passport"
                        title="Passport document"
                        description="PDF, JPG, or PNG. Passport images auto-fill the matching applicant fields."
                        accept={{
                          "application/pdf": [".pdf"],
                          "image/jpeg": [".jpg", ".jpeg"],
                          "image/png": [".png"],
                        }}
                        maxSize={10485760}
                        onUpload={(file) => {
                          onDocumentChange?.("passport", file);
                          handlePassportFileSelected(file);
                        }}
                        onRemove={() => {
                          onDocumentChange?.("passport", null);
                        }}
                      />
                    </div>

                    <div className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                      <DocumentUpload
                        key={`photo-${documentResetKey}`}
                        initialFile={initialDocuments?.photo || null}
                        documentType="photo"
                        title="Full body photo"
                        description="Use a clean photo with good light and a clear full-body view."
                        accept={{
                          "image/jpeg": [".jpg", ".jpeg"],
                          "image/png": [".png"],
                        }}
                        maxSize={10485760}
                        onUpload={(file) => {
                          onDocumentChange?.("photo", file);
                        }}
                        onRemove={() => {
                          onDocumentChange?.("photo", null);
                        }}
                      />
                    </div>

                    <div className="rounded-2xl border border-border/70 bg-muted/20 p-4 md:col-span-2">
                      <DocumentUpload
                        key={`video-${documentResetKey}`}
                        initialFile={initialDocuments?.video || null}
                        documentType="video"
                        title="Video interview"
                        description="Optional, but helpful for faster employer review."
                        accept={{ "video/mp4": [".mp4"] }}
                        maxSize={52428800}
                        onUpload={(file) => {
                          onDocumentChange?.("video", file);
                        }}
                        onRemove={() => {
                          onDocumentChange?.("video", null);
                        }}
                      />
                    </div>
                  </div>
                </CardContent>
              </Card>
            ) : null}

            <Card className="border-border/70 shadow-sm">
              <CardHeader>
                <SectionTitle
                  icon={<UserSquare2 className="h-5 w-5" />}
                  title="Personal profile"
                  description="Fill the exact applicant details that should appear in the CV and employer-facing profile."
                />
              </CardHeader>
              <CardContent className="space-y-6">
                <FormField
                  control={form.control}
                  name="full_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Full name</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="For example: Aster Demeke"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="rounded-2xl border border-border/70 bg-muted/20 px-4 py-3 text-sm text-muted-foreground">
                  Uploading a passport image fills the matching applicant fields automatically, but every value below stays editable before you save the candidate.
                  {isParsingPassport ? (
                    <span className="mt-2 flex items-center gap-2 font-medium text-foreground">
                      <Loader2 className="h-4 w-4 animate-spin text-primary" />
                      Reading the passport image now.
                    </span>
                  ) : null}
                </div>

                <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
                  <FormField
                    control={form.control}
                    name="nationality"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Nationality</FormLabel>
                        <FormControl>
                          <Input
                            placeholder="For example: ETH or Ethiopian"
                            {...field}
                            value={field.value ?? ""}
                          />
                        </FormControl>
                        <FormDescription>
                          Type the exact value you want printed in the CV
                          applicants detail table.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="date_of_birth"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Date of birth</FormLabel>
                        <FormControl>
                          <Input
                            type="date"
                            {...field}
                            value={field.value ?? ""}
                          />
                        </FormControl>
                        <FormDescription>
                          The age field below will help itself once a valid date
                          is entered.
                        </FormDescription>
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
                            value={
                              typeof field.value === "number" ||
                              typeof field.value === "string"
                                ? field.value
                                : ""
                            }
                            onBlur={field.onBlur}
                            onChange={(event) => {
                              setAgeAutoSyncEnabled(false);
                              field.onChange(event);
                            }}
                            ref={field.ref}
                          />
                        </FormControl>
                        <FormDescription>
                          Edit this if the agency wants a specific CV age
                          display.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="place_of_birth"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Place of birth</FormLabel>
                        <FormControl>
                          <Input
                            placeholder="For example: Legambo"
                            {...field}
                            value={field.value ?? ""}
                          />
                        </FormControl>
                        <FormDescription>
                          Use the exact town or city name the employer should
                          see.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="religion"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Religion</FormLabel>
                        <Select
                          onValueChange={field.onChange}
                          value={field.value || undefined}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="Select religion" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            {RELIGION_OPTIONS.map((option) => (
                              <SelectItem key={option} value={option}>
                                {option}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormDescription>
                          Pick the value you want in the applicants detail
                          section.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="marital_status"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Marital status</FormLabel>
                        <Select
                          onValueChange={field.onChange}
                          value={field.value || undefined}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="Select marital status" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            {MARITAL_STATUS_OPTIONS.map((option) => (
                              <SelectItem key={option} value={option}>
                                {option}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormDescription>
                          Choose the exact marital status that should print in
                          the CV.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="children_count"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Children</FormLabel>
                        <FormControl>
                          <Input
                            type="number"
                            min={0}
                            name={field.name}
                            value={
                              typeof field.value === "number" ||
                              typeof field.value === "string"
                                ? field.value
                                : ""
                            }
                            onBlur={field.onBlur}
                            onChange={field.onChange}
                            ref={field.ref}
                          />
                        </FormControl>
                        <FormDescription>
                          Enter the actual number you want shown instead of
                          leaving it blank.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="education_level"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Education level</FormLabel>
                        <Select
                          onValueChange={field.onChange}
                          value={field.value || undefined}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="Select education level" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            {EDUCATION_LEVEL_OPTIONS.map((option) => (
                              <SelectItem key={option} value={option}>
                                {option}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormDescription>
                          This appears exactly in the applicants detail section
                          of the CV.
                        </FormDescription>
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
                            value={
                              typeof field.value === "number" ||
                              typeof field.value === "string"
                                ? field.value
                                : ""
                            }
                            onBlur={field.onBlur}
                            onChange={field.onChange}
                            ref={field.ref}
                          />
                        </FormControl>
                        <FormDescription>
                          How long she has worked in similar roles
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
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
                              const checked = field.value?.includes(skill);
                              return (
                                <FormItem className="rounded-2xl border border-border/70 bg-muted/20 p-4 transition-colors hover:bg-muted/40">
                                  <div className="flex items-center gap-3">
                                    <FormControl>
                                      <Checkbox
                                        checked={checked}
                                        onCheckedChange={(nextChecked) => {
                                          return nextChecked
                                            ? field.onChange([
                                                ...(field.value || []),
                                                skill,
                                              ])
                                            : field.onChange(
                                                (field.value || []).filter(
                                                  (value) => value !== skill,
                                                ),
                                              );
                                        }}
                                      />
                                    </FormControl>
                                    <FormLabel className="cursor-pointer text-sm font-semibold">
                                      {skill}
                                    </FormLabel>
                                  </div>
                                </FormItem>
                              );
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
                  <Button
                    type="button"
                    variant="outline"
                    className="border-dashed md:shrink-0"
                    onClick={addLanguageEntry}
                  >
                    <Plus className="mr-2 h-4 w-4" />
                    Add another
                  </Button>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                {fields.map((field, index) => (
                  <div
                    key={field.id}
                    className="rounded-2xl border border-border/70 bg-muted/20 p-4"
                  >
                    <div className="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] md:items-end">
                      <FormField
                        control={form.control}
                        name={`languages.${index}.language`}
                        render={({ field }) => (
                          <FormItem>
                            <FormLabel>Language</FormLabel>
                            <Select
                              onValueChange={field.onChange}
                              defaultValue={field.value}
                            >
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
                            <Select
                              onValueChange={field.onChange}
                              defaultValue={field.value}
                            >
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
              {isLoading ? (
                <Loader2 className="mr-2 h-5 w-5 animate-spin" />
              ) : (
                <Plus className="mr-2 h-5 w-5" />
              )}
              {isLoading ? "Saving..." : "Create & Add Another"}
            </Button>
          ) : null}
          <Button
            type="submit"
            size="lg"
            className="min-w-[220px]"
            data-submit-mode="default"
            disabled={isLoading}
          >
            {isLoading ? (
              <Loader2 className="mr-2 h-5 w-5 animate-spin" />
            ) : null}
            {isLoading
              ? isEditing
                ? "Saving Changes..."
                : "Creating Candidate..."
              : isEditing
                ? "Save Changes"
                : "Create Candidate"}
          </Button>
        </div>
      </form>
    </Form>
  );
}

function normalizeDateInputValue(value?: string) {
  if (!value) {
    return "";
  }

  const directMatch = value.match(/^(\d{4})-(\d{2})-(\d{2})/);
  if (directMatch) {
    return `${directMatch[1]}-${directMatch[2]}-${directMatch[3]}`;
  }

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return "";
  }

  return parsed.toISOString().slice(0, 10);
}

function calculateAgeFromDate(value?: string) {
  if (!value) {
    return null;
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return null;
  }

  const today = new Date();
  let age = today.getFullYear() - date.getFullYear();
  const monthDelta = today.getMonth() - date.getMonth();
  if (
    monthDelta < 0 ||
    (monthDelta === 0 && today.getDate() < date.getDate())
  ) {
    age -= 1;
  }
  return age >= 0 ? age : null;
}
