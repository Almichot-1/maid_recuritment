"use client";

import * as React from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  AlertTriangle,
  BadgeCheck,
  BriefcaseBusiness,
  Globe2,
  Languages,
  Loader2,
  Plus,
  ShieldCheck,
  Sparkles,
  Trash2,
  UserSquare2,
} from "lucide-react";

import { useParsePassport, usePassportData } from "@/hooks/use-passport-ocr";
import { CandidateInput, candidateSchema } from "@/lib/validations";
import { useAgencyBranding } from "@/hooks/use-agency-branding";
import { useCurrentUser } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
  initialData?: Partial<CandidateFormValues>;
  onSubmit: (
    data: CandidateInput,
    context?: { submitter: "default" | "create_another" },
  ) => Promise<{ resetForm?: boolean } | void> | { resetForm?: boolean } | void;
  isLoading?: boolean;
  onDocumentChange?: (
    documentType: "passport" | "photo" | "video",
    file: File | null,
  ) => void;
  showDocuments?: boolean;
}

type CandidateFormValues = {
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
  initialData,
  onSubmit,
  isLoading,
  onDocumentChange,
  showDocuments = true,
}: CandidateFormProps) {
  const { user } = useCurrentUser();
  const { hasLogo } = useAgencyBranding();
  const { data: passportData } = usePassportData(
    candidateId,
    Boolean(candidateId),
  );
  const { mutateAsync: parsePassport, isPending: isParsingPassport } =
    useParsePassport(candidateId);
  const [documentResetKey, setDocumentResetKey] = React.useState(0);
  const [selectedPassportFile, setSelectedPassportFile] =
    React.useState<File | null>(null);
  const [passportPreview, setPassportPreview] =
    React.useState<PassportData | null>(null);
  const [lastAutoParsedPassportKey, setLastAutoParsedPassportKey] =
    React.useState<string | null>(null);
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
  const [documentPresence, setDocumentPresence] = React.useState({
    passport: false,
    photo: false,
    video: false,
  });

  const selectedSkills = form.watch("skills") || [];
  const watchedLanguages = form.watch("languages") || [];
  const fullName = form.watch("full_name");
  const watchedNationality = form.watch("nationality");
  const watchedDateOfBirth = form.watch("date_of_birth");
  const watchedAge = form.watch("age");
  const watchedPlaceOfBirth = form.watch("place_of_birth");
  const watchedReligion = form.watch("religion");
  const watchedMaritalStatus = form.watch("marital_status");
  const watchedChildrenCount = form.watch("children_count");
  const watchedEducationLevel = form.watch("education_level");
  const activePassportPreview = passportPreview || passportData || null;

  const requiredDocumentChecklist = [
    { label: "Passport", done: documentPresence.passport || !showDocuments },
    {
      label: "Full body photo",
      done: documentPresence.photo || !showDocuments,
    },
    {
      label: "Video interview (optional)",
      done: documentPresence.video || !showDocuments,
    },
  ];

  const isEditing = !!initialData;
  const addLanguageEntry = () =>
    append({
      language: LANGUAGES_OPTIONS[0],
      proficiency: PROFICIENCY_OPTIONS[0],
    });

  React.useEffect(() => {
    if (!passportData) {
      return;
    }
    setPassportPreview(passportData);
  }, [passportData]);

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
    setDocumentPresence({
      passport: false,
      photo: false,
      video: false,
    });
    setSelectedPassportFile(null);
    setPassportPreview(null);
    setLastAutoParsedPassportKey(null);
    setDocumentResetKey((current) => current + 1);
    onDocumentChange?.("passport", null);
    onDocumentChange?.("photo", null);
    onDocumentChange?.("video", null);
  }, [blankFormValues, form, onDocumentChange]);

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
  });

  const handleExtractPassportData = React.useCallback(async () => {
    if (
      !selectedPassportFile ||
      !selectedPassportFile.type.startsWith("image/")
    ) {
      return;
    }

    const parsed = await parsePassport(selectedPassportFile);
    setPassportPreview(parsed);

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
  }, [form, parsePassport, selectedPassportFile]);

  React.useEffect(() => {
    if (
      !selectedPassportFile ||
      !selectedPassportFile.type.startsWith("image/")
    ) {
      return;
    }
    const fileKey = `${selectedPassportFile.name}:${selectedPassportFile.size}:${selectedPassportFile.lastModified}`;
    if (lastAutoParsedPassportKey === fileKey || isParsingPassport) {
      return;
    }

    setLastAutoParsedPassportKey(fileKey);
    const idleWindow = window as Window & {
      requestIdleCallback?: (
        callback: () => void,
        options?: { timeout: number },
      ) => number;
      cancelIdleCallback?: (handle: number) => void;
    };
    let timer: number | null = null;
    let idleHandle: number | null = null;
    const runExtraction = () => {
      void handleExtractPassportData();
    };

    if (typeof idleWindow.requestIdleCallback === "function") {
      idleHandle = idleWindow.requestIdleCallback(runExtraction, {
        timeout: 500,
      });
    } else {
      timer = window.setTimeout(runExtraction, 120);
    }

    return () => {
      if (
        idleHandle !== null &&
        typeof idleWindow.cancelIdleCallback === "function"
      ) {
        idleWindow.cancelIdleCallback(idleHandle);
      }
      if (timer !== null) {
        window.clearTimeout(timer);
      }
    };
  }, [
    handleExtractPassportData,
    isParsingPassport,
    lastAutoParsedPassportKey,
    selectedPassportFile,
  ]);

  const applicantDetailPreview = [
    {
      label: "Nationality",
      value: compactFormPreviewValue(watchedNationality),
    },
    {
      label: "Date Of Birth",
      value: formatApplicantPreviewDate(watchedDateOfBirth),
    },
    { label: "Age", value: formatApplicantPreviewAge(watchedAge) },
    {
      label: "Place Of Birth",
      value: compactFormPreviewValue(watchedPlaceOfBirth),
    },
    { label: "Religion", value: compactFormPreviewValue(watchedReligion) },
    {
      label: "Marital Status",
      value: compactFormPreviewValue(watchedMaritalStatus),
    },
    {
      label: "Children",
      value: compactCountPreviewValue(watchedChildrenCount),
    },
    {
      label: "Education Level",
      value: compactFormPreviewValue(watchedEducationLevel),
    },
  ];
  const applicantDetailCompletion = applicantDetailPreview.filter(
    (item) => item.value !== "Not filled yet",
  ).length;

  return (
    <Form {...form}>
      <form onSubmit={handleFormSubmit} className="w-full">
        <div className="grid gap-6 xl:grid-cols-[minmax(0,1.45fr)_360px]">
          <div className="space-y-6">
            <Card className="overflow-hidden border-border/70 bg-[radial-gradient(circle_at_top_left,_rgba(251,191,36,0.22),_transparent_32%),linear-gradient(135deg,rgba(15,23,42,0.98),rgba(30,41,59,0.96))] text-white shadow-xl">
              <CardContent className="space-y-5 p-6">
                <div className="space-y-4">
                  <Badge className="w-fit rounded-full border-0 bg-white/15 px-3 py-1 text-[11px] uppercase tracking-[0.24em] text-amber-200 hover:bg-white/15">
                    Candidate workflow
                  </Badge>
                  <div className="space-y-2">
                    <h2 className="text-3xl font-semibold tracking-tight">
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
                          setDocumentPresence((current) => ({
                            ...current,
                            passport: true,
                          }));
                          setSelectedPassportFile(file);
                          setLastAutoParsedPassportKey(null);
                          onDocumentChange?.("passport", file);
                        }}
                        onRemove={() => {
                          setDocumentPresence((current) => ({
                            ...current,
                            passport: false,
                          }));
                          setSelectedPassportFile(null);
                          setLastAutoParsedPassportKey(null);
                          onDocumentChange?.("passport", null);
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
                          setDocumentPresence((current) => ({
                            ...current,
                            photo: true,
                          }));
                          onDocumentChange?.("photo", file);
                        }}
                        onRemove={() => {
                          setDocumentPresence((current) => ({
                            ...current,
                            photo: false,
                          }));
                          onDocumentChange?.("photo", null);
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
                          setDocumentPresence((current) => ({
                            ...current,
                            video: true,
                          }));
                          onDocumentChange?.("video", file);
                        }}
                        onRemove={() => {
                          setDocumentPresence((current) => ({
                            ...current,
                            video: false,
                          }));
                          onDocumentChange?.("video", null);
                        }}
                      />
                    </div>
                  </div>

                  <div className="rounded-2xl border border-primary/15 bg-primary/5 p-4">
                    <div className="space-y-2">
                      <p className="text-sm font-semibold text-foreground">
                        Automatic passport extraction
                      </p>
                      <p className="text-sm text-muted-foreground">
                        As soon as you upload a clear passport image, the system
                        reads it automatically, fills the matching applicant
                        detail fields below, and shows the extracted passport
                        details here. You can still edit any filled value before
                        saving the candidate.
                      </p>
                    </div>

                    {selectedPassportFile &&
                    !selectedPassportFile.type.startsWith("image/") ? (
                      <div className="mt-4 rounded-xl border border-amber-300/40 bg-amber-50 px-4 py-3 text-sm text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/30 dark:text-amber-100">
                        Upload a passport image in JPG or PNG format if you want
                        OCR extraction. PDF passports can still be saved
                        normally for the CV.
                      </div>
                    ) : null}

                    {selectedPassportFile?.type.startsWith("image/") &&
                    isParsingPassport ? (
                      <div className="mt-4 flex items-center gap-3 rounded-xl border border-border/70 bg-background/80 px-4 py-4 text-sm text-muted-foreground">
                        <Loader2 className="h-4 w-4 animate-spin text-primary" />
                        Reading the passport image and filling the form now.
                      </div>
                    ) : null}

                    {activePassportPreview ? (
                      <PassportPreviewCard passport={activePassportPreview} />
                    ) : (
                      <div className="mt-4 rounded-xl border border-dashed border-border/70 bg-background/70 px-4 py-4 text-sm text-muted-foreground">
                        Upload a passport image and the extracted passport
                        fields will appear here automatically.
                      </div>
                    )}
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

                <div className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                  <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
                    <span className="rounded-full border border-border bg-background px-3 py-1.5">
                      {applicantDetailCompletion}/8 applicant rows filled
                    </span>
                    <span className="rounded-full border border-border bg-background px-3 py-1.5">
                      Passport image fills DOB, age, nationality, and place of birth
                    </span>
                    <span className="rounded-full border border-border bg-background px-3 py-1.5">
                      Every field stays editable before save
                    </span>
                  </div>
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

          <div className="space-y-6 xl:sticky xl:top-6 xl:self-start">
            <Card className="overflow-hidden border-border/70 shadow-lg">
              <CardHeader className="bg-gradient-to-br from-amber-50 via-background to-slate-50">
                <CardTitle className="flex items-center gap-2 text-lg">
                  <Sparkles className="h-5 w-5 text-amber-600" />
                  Submission summary
                </CardTitle>
                <CardDescription>
                  Keep the profile complete before you publish it to foreign
                  agencies.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-5 p-6">
                <div className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                  <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
                    Agency
                  </p>
                  <p className="mt-1 truncate text-sm font-semibold">
                    {user?.company_name || "Agency profile"}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {hasLogo
                      ? "Saved branding is ready for generated CVs."
                      : "You can add reusable branding later from Settings."}
                  </p>
                </div>

                <div className="space-y-3">
                  <SummaryItem
                    label="Candidate name"
                    value={fullName?.trim() ? fullName : "Not filled yet"}
                  />
                  <SummaryItem
                    label="Applicant detail rows"
                    value={`${applicantDetailCompletion}/8 ready for the CV`}
                  />
                  <SummaryItem
                    label="Skills selected"
                    value={
                      selectedSkills.length
                        ? `${selectedSkills.length} selected`
                        : "Select at least one skill"
                    }
                  />
                  <SummaryItem
                    label="Languages tracked"
                    value={
                      watchedLanguages.length
                        ? `${watchedLanguages.length} language entries`
                        : "Add a language"
                    }
                  />
                </div>

                <div className="space-y-3 rounded-2xl border border-border/70 bg-muted/20 p-4">
                  <div className="flex items-center gap-2">
                    <Globe2 className="h-4 w-4 text-slate-700" />
                    <p className="text-sm font-semibold">Before you submit</p>
                  </div>
                  {requiredDocumentChecklist.map((item) => (
                    <div
                      key={item.label}
                      className="flex items-center gap-2 text-sm"
                    >
                      <BadgeCheck
                        className={`h-4 w-4 ${item.done ? "text-green-600" : "text-muted-foreground"}`}
                      />
                      <span
                        className={
                          item.done
                            ? "text-foreground"
                            : "text-muted-foreground"
                        }
                      >
                        {item.label}
                      </span>
                    </div>
                  ))}
                  {!showDocuments ? (
                    <p className="text-xs text-muted-foreground">
                      Profile details are edited here. The latest supporting
                      files can be managed from the surrounding candidate
                      workspace.
                    </p>
                  ) : (
                    <p className="text-xs text-muted-foreground">
                      Uploaded files are attached right after the candidate
                      record is saved.
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

function SummaryItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="space-y-1">
      <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
        {label}
      </p>
      <p className="text-sm font-semibold text-foreground">{value}</p>
    </div>
  );
}

function PassportPreviewCard({ passport }: { passport: PassportData }) {
  const expiryDate = new Date(passport.expiry_date);
  const expiryWarning = isDateWithinMonths(expiryDate, 6);
  const calculatedAge = calculateAgeFromDate(passport.date_of_birth);

  return (
    <div className="mt-4 rounded-2xl border border-border/70 bg-background/85 p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">
            Extracted passport details
          </p>
          <p className="mt-1 text-sm font-semibold text-foreground">
            {passport.holder_name || "Unnamed passport holder"}
          </p>
        </div>
        <Badge
          variant="outline"
          className={
            expiryWarning ? "border-rose-300 bg-rose-50 text-rose-700" : ""
          }
        >
          Confidence {Math.round(passport.confidence)}
        </Badge>
      </div>

      <div className="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <PassportPreviewItem
          label="Passport No"
          value={passport.passport_number || "Not found"}
        />
        <PassportPreviewItem
          label="Holder name"
          value={passport.holder_name || "Not found"}
        />
        <PassportPreviewItem
          label="Nationality"
          value={passport.nationality || "Not found"}
        />
        <PassportPreviewItem
          label="Gender"
          value={passport.gender || "Not found"}
        />
        <PassportPreviewItem
          label="Date of birth"
          value={formatPassportValue(passport.date_of_birth)}
        />
        <PassportPreviewItem
          label="Age"
          value={
            calculatedAge !== null ? `${calculatedAge} years` : "Not found"
          }
        />
        <PassportPreviewItem
          label="Issue date"
          value={formatPassportValue(passport.issue_date)}
        />
        <PassportPreviewItem
          label="Expiry date"
          value={formatPassportValue(passport.expiry_date)}
          valueClassName={
            expiryWarning ? "text-rose-600 dark:text-rose-300" : undefined
          }
        />
        <PassportPreviewItem
          label="Country code"
          value={passport.country_code || "Not found"}
        />
      </div>

      <div className="mt-3 grid gap-3 md:grid-cols-2">
        <PassportPreviewItem
          label="Place of birth"
          value={passport.place_of_birth || "Not found"}
        />
        <PassportPreviewItem
          label="Issuing authority"
          value={passport.issuing_authority || "Not found"}
        />
      </div>

      {expiryWarning ? (
        <div className="mt-3 flex items-start gap-2 rounded-xl border border-rose-300/40 bg-rose-50 px-3 py-3 text-sm text-rose-900 dark:border-rose-900/50 dark:bg-rose-950/30 dark:text-rose-100">
          <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
          <span>
            This passport expires within 6 months, so the CV will highlight it
            as a warning.
          </span>
        </div>
      ) : null}
    </div>
  );
}

function compactFormPreviewValue(value: string | undefined) {
  const cleaned = value?.trim();
  return cleaned ? cleaned : "Not filled yet";
}

function compactCountPreviewValue(value: number | string | undefined) {
  if (typeof value === "number") {
    return String(value);
  }
  if (typeof value === "string" && value.trim()) {
    return value.trim();
  }
  return "Not filled yet";
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

function formatApplicantPreviewDate(value: string | undefined) {
  if (!value) {
    return "Not filled yet";
  }
  const parsed = new Date(`${value}T00:00:00`);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }
  return parsed
    .toLocaleDateString("en-GB", {
      day: "2-digit",
      month: "short",
      year: "2-digit",
    })
    .replace(/ /g, "-");
}

function formatApplicantPreviewAge(value: number | string | undefined) {
  if (typeof value === "number" && !Number.isNaN(value)) {
    return `${value} YRS`;
  }
  if (typeof value === "string" && value.trim()) {
    return `${value.trim()} YRS`;
  }
  return "Not filled yet";
}

function PassportPreviewItem({
  label,
  value,
  valueClassName,
}: {
  label: string;
  value: string;
  valueClassName?: string;
}) {
  return (
    <div className="rounded-xl border border-border/60 bg-muted/20 px-3 py-3">
      <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
        {label}
      </p>
      <p
        className={`mt-1 text-sm font-semibold text-foreground break-words ${valueClassName ?? ""}`}
      >
        {value}
      </p>
    </div>
  );
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

function formatPassportValue(value?: string) {
  if (!value) {
    return "Not found";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleDateString();
}

function isDateWithinMonths(date: Date, months: number) {
  if (Number.isNaN(date.getTime())) {
    return false;
  }

  const threshold = new Date();
  threshold.setMonth(threshold.getMonth() + months);
  return date.getTime() <= threshold.getTime();
}
