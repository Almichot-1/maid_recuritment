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
import { usePairingContext } from "@/hooks/use-pairings";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
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
import { VerifiedOcrInput } from "@/components/ui/verified-ocr-input";
import { DocumentUpload } from "./document-upload";
import { ExperienceEntry, PassportData, CandidatePairOverride } from "@/types";

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

const DEFAULT_FORM_VALUES: CandidateFormValues = {
  full_name: "",
  nationality: "Ethiopian",
  date_of_birth: "",
  age: undefined,
  place_of_birth: "",
  passport_number: "",
  gender: "",
  issue_date: "",
  expiry_date: "",
  experience_abroad: [],
  religion: "",
  marital_status: "",
  children_count: 0,
  education_level: "Secondary",
  skills: ["Cleaning", "Childcare", "Elderly Care", "Laundry", "Ironing"],
  languages: [],
  remark: "",
};

const LANGUAGES_OPTIONS = ["Arabic", "English", "Amharic", "French", "Swahili"];
const PROFICIENCY_OPTIONS = ["Basic", "Intermediate", "Fluent"];
const COUNTRIES_OPTIONS = [
  "Saudi Arabia",
  "United Arab Emirates",
  "Qatar",
  "Kuwait",
  "Oman",
  "Bahrain",
  "Lebanon",
  "Jordan",
  "Yemen",
]
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

const NATIONALITY_OPTIONS = [
  "Ethiopian",
  "Eritrean",
  "Kenyan",
  "Somali",
  "Sudanese",
  "South Sudanese",
  "Djiboutian",
  "Other",
];

const ISO_TO_NATIONALITY: Record<string, string> = {
  ETH: "Ethiopian",
  ERI: "Eritrean",
  KEN: "Kenyan",
  SOM: "Somali",
  SDN: "Sudanese",
  SSD: "South Sudanese",
  DJI: "Djiboutian",
};

const GENDER_OPTIONS = ["Female", "Male"];

const PASSPORT_OCR_PREVIEW_MAX_DIMENSION = 1800;
const PASSPORT_OCR_PREVIEW_QUALITY = 0.88;

type PassportPreviewResult = {
  passport: PassportData;
  requestKey: string;
};

interface CandidateFormProps {
  candidateId?: string;
  mode?: "create" | "edit";
  initialData?: Partial<CandidateFormValues>;
  initialDocuments?: Partial<Record<"passport" | "photo" | "video", File | null>>;
  onSubmit: (
    data: CandidateInput,
    context?: { submitter: "default" | "create_another"; partnerOverrides?: PartnerOverrideEntry[] },
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
  passport_number?: string;
  gender?: string;
  issue_date?: string;
  expiry_date?: string;
  issuing_authority?: string;
  experience_abroad?: ExperienceEntry[];
  religion?: string;
  marital_status?: string;
  children_count?: number | string;
  education_level?: string;
  experience_years?: number | string;
  country_of_experience?: string;
  skills: string[];
  languages: Array<{ language: string; proficiency: string }>;
  remark?: string;
};

export type PartnerOverrideEntry = {
  pairing_id: string;
  country_applied: string;
  salary_offered: string;
  logo_url?: string;
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

  const { context: pairingContext } = usePairingContext();
  const [partnerOverrides, setPartnerOverrides] = React.useState<Record<string, { country: string; salary: string; logo_url?: string }>>({});
  const [partnerLogoUrls, setPartnerLogoUrls] = React.useState<Record<string, string>>({});
  const { mutateAsync: parsePassport, isPending: isParsingPassport } =
    useParsePassport(candidateId);
  const [documentResetKey, setDocumentResetKey] = React.useState(0);
  const [isPassportProcessing, setIsPassportProcessing] = React.useState(false);
  const passportPreviewCacheRef = React.useRef(
    new Map<string, PassportData>(),
  );
  const passportPreviewInflightRef = React.useRef(
    new Map<string, Promise<PassportPreviewResult>>(),
  );
  const activePassportRequestKeyRef = React.useRef<string | null>(null);
  const activePassportAbortRef = React.useRef<AbortController | null>(null);
  const passportRequestSequenceRef = React.useRef(0);
  const [ocrDetectedValues, setOcrDetectedValues] = React.useState<Record<string, string>>({});
  // Debounce timer: prevents firing OCR for every rapid file change.
  // Only the file selected after 400 ms of inactivity triggers the API call.
  const passportDebounceTimerRef = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  const mergedDefaultValues = React.useMemo<CandidateFormValues>(() => {
    if (!initialData) return { ...DEFAULT_FORM_VALUES };
    const merged = { ...DEFAULT_FORM_VALUES };
    for (const [key, value] of Object.entries(initialData)) {
      if (value === null || value === undefined) continue;
      if (typeof value === "string" && value.trim() === "") continue;
      if (Array.isArray(value) && value.length === 0) continue;
      (merged as Record<string, unknown>)[key] = value;
    }
    return merged;
  }, [initialData]);

  const form = useForm<CandidateFormValues, undefined, CandidateInput>({
    resolver: zodResolver(candidateSchema) as never,
    defaultValues: mergedDefaultValues,
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "languages",
  });
  const { fields: experienceFields, append: appendExperience, remove: removeExperience } = useFieldArray({
    control: form.control,
    name: "experience_abroad",
  });
  const watchedDateOfBirth = form.watch("date_of_birth");
  const watchedAge = form.watch("age");
  const watchedNationality = form.watch("nationality");

  React.useEffect(() => {
    if (watchedNationality) {
      const currentPoB = form.getValues("place_of_birth");
      if (!currentPoB) {
        const map: Record<string, string> = {
          "Ethiopian": "Ethiopia",
          "Eritrean": "Eritrea",
          "Kenyan": "Kenya",
          "Somali": "Somalia",
          "Sudanese": "Sudan",
          "South Sudanese": "South Sudan",
          "Djiboutian": "Djibouti",
        };
        const mapped = map[watchedNationality] || watchedNationality;
        form.setValue("place_of_birth", mapped, {
          shouldDirty: true,
          shouldValidate: true,
        });
      }
    }
  }, [watchedNationality, form]);

  const isEditing = mode === "edit";
  const addLanguageEntry = () =>
    append({
      language: LANGUAGES_OPTIONS[0],
      proficiency: PROFICIENCY_OPTIONS[0],
    });

  // Initialize per-partner overrides
  React.useEffect(() => {
    if (!pairingContext?.workspaces?.length) return

    const initial: Record<string, { country: string; salary: string }> = {}

    if (mode === "create") {
      // Create mode: pre-fill from workspace defaults
      for (const ws of pairingContext.workspaces) {
        if (ws.status === "active" || ws.status === "approved") {
          const sal = ws.default_salary
          const curr = ws.default_currency
          const salaryVal = sal && curr ? `${sal} ${curr}` : (sal || curr || "")
          initial[ws.id] = {
            country: ws.default_country || "",
            salary: salaryVal,
          }
        }
      }
    } else if (mode === "edit" && initialData) {
      // Edit mode: pre-fill from existing pair_overrides
      const overrides = (initialData as unknown as { pair_overrides?: CandidatePairOverride[] }).pair_overrides
      for (const ws of pairingContext.workspaces) {
        const ov = overrides?.find((o) => o.pairing_id === ws.id)
        initial[ws.id] = {
          country: ov?.country_applied || "",
          salary: ov?.salary_offered || "",
        }
      }
    }

    setPartnerOverrides(initial)
  }, [mode, pairingContext?.workspaces, initialData])

  React.useEffect(() => {
    if (!onDraftChange) {
      return;
    }

    const subscription = form.watch((values) => {
      onDraftChange({
        ...DEFAULT_FORM_VALUES,
        ...values,
        skills: values.skills || [],
        languages:
          values.languages && values.languages.length > 0
            ? values.languages.map((language) => ({
                language: language.language || LANGUAGES_OPTIONS[0],
                proficiency: language.proficiency || PROFICIENCY_OPTIONS[0],
              }))
            : DEFAULT_FORM_VALUES.languages,
        experience_abroad:
          values.experience_abroad && values.experience_abroad.length > 0
            ? values.experience_abroad.map((item) => ({
                country: item.country || "",
                years: item.years || 0,
              }))
            : [],
      });
    });

    return () => subscription.unsubscribe();
  }, [form, onDraftChange]);

  React.useEffect(() => {
    const derivedAge = calculateAgeFromDate(watchedDateOfBirth);
    if (derivedAge === null) {
      return;
    }
    const currentAge =
      typeof watchedAge === "number"
        ? watchedAge
        : typeof watchedAge === "string" && watchedAge.trim()
          ? Number(watchedAge)
          : null;
    if (currentAge === null || Number.isNaN(currentAge) || currentAge !== derivedAge) {
      form.setValue("age", derivedAge, {
        shouldDirty: true,
        shouldValidate: true,
      });
    }
  }, [form, watchedAge, watchedDateOfBirth]);

  const resetForNextCandidate = React.useCallback(() => {
    form.reset(DEFAULT_FORM_VALUES);
    setDocumentResetKey((current) => current + 1);
    onDocumentChange?.("passport", null);
    onDocumentChange?.("photo", null);
    onDocumentChange?.("video", null);
    onClearDraft?.();
  }, [form, onClearDraft, onDocumentChange]);

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
    
    // Validate all active workspaces have country and salary filled
    const activeWorkspaces = pairingContext?.workspaces?.filter(
      (ws) => ws.status === "active" || ws.status === "approved"
    ) || []
    const hasEmptyOverrides = activeWorkspaces.some((ws) => {
      const val = partnerOverrides[ws.id]
      return !val?.country || !val?.salary
    })
    if (hasEmptyOverrides) {
      toast.error("Set country and salary for all partners before saving.")
      return
    }

    const overridesArr = Object.entries(partnerOverrides)
      .filter(([, val]) => val.country && val.salary)
      .map(([pairing_id, val]) => ({
        pairing_id,
        country_applied: val.country,
        salary_offered: val.salary,
        logo_url: partnerLogoUrls[pairing_id] || undefined,
      }))
    
    const result = await onSubmit(data, { submitter: submitMode, partnerOverrides: overridesArr.length > 0 ? overridesArr : undefined });

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
    const detected: Record<string, string> = {};

    if (parsed.holder_name?.trim()) {
      form.setValue("full_name", parsed.holder_name.trim(), {
        shouldDirty: true,
        shouldValidate: true,
      });
    }

    if (parsed.nationality?.trim()) {
      const iso = parsed.nationality.trim().toUpperCase();
      const nationality = ISO_TO_NATIONALITY[iso] || iso;
      form.setValue("nationality", nationality, {
        shouldDirty: true,
        shouldValidate: true,
      });
    }

    const normalizedDateOfBirth = normalizeDateInputValue(parsed.date_of_birth);
    if (normalizedDateOfBirth) {
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
      const val = parsed.place_of_birth.trim();
      form.setValue("place_of_birth", val, {
        shouldDirty: true,
        shouldValidate: true,
      });
      detected.place_of_birth = val;
    }

    if (parsed.passport_number?.trim()) {
      const val = parsed.passport_number.trim();
      form.setValue("passport_number", val, {
        shouldDirty: true,
        shouldValidate: true,
      });
      detected.passport_number = val;
    }
    const normalizedIssueDate = normalizeDateInputValue(parsed.issue_date);
    if (normalizedIssueDate) {
      form.setValue("issue_date", normalizedIssueDate, {
        shouldDirty: true,
        shouldValidate: true,
      });
      detected.issue_date = normalizedIssueDate;
    }
    const normalizedExpiryDate = normalizeDateInputValue(parsed.expiry_date);
    if (normalizedExpiryDate) {
      form.setValue("expiry_date", normalizedExpiryDate, {
        shouldDirty: true,
        shouldValidate: true,
      });
      detected.expiry_date = normalizedExpiryDate;
    }
    if (parsed.gender?.trim()) {
      const genderValue = parsed.gender.trim().toUpperCase() === "M" ? "Male" : parsed.gender.trim().toUpperCase() === "F" ? "Female" : parsed.gender.trim();
      form.setValue("gender", genderValue, {
        shouldDirty: true,
        shouldValidate: true,
      });
    }
    if (parsed.issuing_authority?.trim()) {
      form.setValue("issuing_authority", parsed.issuing_authority.trim(), {
        shouldDirty: true,
        shouldValidate: true,
      });
    }

    if (Object.keys(detected).length > 0) {
      setOcrDetectedValues(detected);
    }
  }, [form]);

  const handlePassportFileSelected = React.useCallback((file: File | null) => {
    // Clear any pending debounced call whenever the file changes.
    if (passportDebounceTimerRef.current !== null) {
      clearTimeout(passportDebounceTimerRef.current);
      passportDebounceTimerRef.current = null;
    }

    if (!file) {
      setIsPassportProcessing(false);
      passportRequestSequenceRef.current += 1;
      activePassportAbortRef.current?.abort();
      activePassportAbortRef.current = null;
      activePassportRequestKeyRef.current = null;
      return;
    }

    const isImageType = file.type.startsWith("image/");
    const isImageExt = /\.(jpg|jpeg|png|webp|bmp|tiff?)$/i.test(file.name);
    if (!isImageType && !isImageExt) {
      setIsPassportProcessing(false);
      passportRequestSequenceRef.current += 1;
      activePassportAbortRef.current?.abort();
      activePassportAbortRef.current = null;
      activePassportRequestKeyRef.current = null;
      return;
    }

    const runExtraction = async () => {
      setIsPassportProcessing(true);
      const requestKey = buildPassportPreviewRequestKey(file);
      const requestSequence = passportRequestSequenceRef.current + 1;
      passportRequestSequenceRef.current = requestSequence;

      if (activePassportRequestKeyRef.current && activePassportRequestKeyRef.current !== requestKey) {
        activePassportAbortRef.current?.abort();
      }

      if (passportPreviewCacheRef.current.has(requestKey)) {
        const cached = passportPreviewCacheRef.current.get(requestKey);
        if (cached) {
          activePassportRequestKeyRef.current = requestKey;
          applyPassportAutofill(cached);
        }
        return;
      }

      const existingRequest = passportPreviewInflightRef.current.get(requestKey);
      if (existingRequest) {
        activePassportRequestKeyRef.current = requestKey;
        try {
          const result = await existingRequest;
          if (
            passportRequestSequenceRef.current !== requestSequence ||
            activePassportRequestKeyRef.current !== result.requestKey
          ) {
            return;
          }
          applyPassportAutofill(result.passport);
        } catch {
          // The shared request already handles user-facing errors.
        }
        return;
      }

      const abortController = new AbortController();
      activePassportAbortRef.current = abortController;
      activePassportRequestKeyRef.current = requestKey;

      const parseRequest = (async () => {
        const previewFile = await buildPassportOCRPreviewFile(file);
        const result = await parsePassport({
          file: previewFile,
          signal: abortController.signal,
        });

        passportPreviewCacheRef.current.set(requestKey, result.passport);
        return {
          passport: result.passport,
          requestKey,
        };
      })();

      passportPreviewInflightRef.current.set(requestKey, parseRequest);

      try {
        const result = await parseRequest;
        if (
          passportRequestSequenceRef.current !== requestSequence ||
          activePassportRequestKeyRef.current !== result.requestKey
        ) {
          return;
        }
        applyPassportAutofill(result.passport);
      } catch {
        // The mutation already surfaces the error to the user.
      } finally {
        setIsPassportProcessing(false);
        const inflightRequest = passportPreviewInflightRef.current.get(requestKey);
        if (inflightRequest === parseRequest) {
          passportPreviewInflightRef.current.delete(requestKey);
        }

        if (activePassportAbortRef.current === abortController) {
          activePassportAbortRef.current = null;
        }
      }
    };

    // Debounce: wait 400 ms before firing OCR so rapid file swaps only
    // trigger a single API call — important on the free-tier server.
    passportDebounceTimerRef.current = setTimeout(() => {
      passportDebounceTimerRef.current = null;
      void runExtraction();
    }, 400);
  }, [applyPassportAutofill, parsePassport]);

  React.useEffect(() => {
    const activePassportAbortRefValue = activePassportAbortRef.current;
    const passportPreviewInflightRefValue = passportPreviewInflightRef.current;
    return () => {
      // Cancel any in-flight debounce timer on unmount.
      if (passportDebounceTimerRef.current !== null) {
        clearTimeout(passportDebounceTimerRef.current);
      }
      activePassportAbortRefValue?.abort();
      passportPreviewInflightRefValue.clear();
    };
  }, []);

  return (
    <Form {...form}>
      <form id="candidate-form" onSubmit={handleFormSubmit} className="w-full">
        <div className="space-y-6 pb-28">
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
                    <div className="flex flex-col rounded-2xl border border-border/70 bg-muted/20 p-4">
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
                      {isPassportProcessing && (
                        <div className="mt-4 flex items-center justify-center gap-2 rounded-xl border border-blue-200 bg-blue-50 p-3 text-sm font-medium text-blue-700 dark:border-blue-900/50 dark:bg-blue-900/20 dark:text-blue-300">
                          <Loader2 className="h-4 w-4 animate-spin" />
                          <span>Processing passport OCR...</span>
                        </div>
                      )}
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
                        <Select
                          onValueChange={field.onChange}
                          value={field.value || undefined}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="Select nationality" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            {NATIONALITY_OPTIONS.map((option) => (
                              <SelectItem key={option} value={option}>
                                {option}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormDescription>
                          Select the nationality for the CV details section.
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
                            readOnly
                            className="bg-muted/50 cursor-default"
                            tabIndex={-1}
                            ref={field.ref}
                          />
                        </FormControl>
                        <FormDescription>
                          Auto-calculated from date of birth. Change DOB to
                          update.
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
                        <FormControl>
                          <VerifiedOcrInput
                            id="place_of_birth"
                            label="Place of birth"
                            placeholder="Auto-filled from passport OCR"
                            ocrDetectedValue={ocrDetectedValues.place_of_birth}
                            {...field}
                            value={field.value ?? ""}
                          />
                        </FormControl>
                        <FormDescription>
                          Auto-filled from passport OCR — edit if needed.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="passport_number"
                    render={({ field }) => (
                      <FormItem>
                        <FormControl>
                          <VerifiedOcrInput
                            id="passport_number"
                            label="Passport number"
                            placeholder="e.g., EQ1817015"
                            ocrDetectedValue={ocrDetectedValues.passport_number}
                            {...field}
                            value={field.value ?? ""}
                          />
                        </FormControl>
                        <FormDescription>
                          Manually correct any OCR misreads here.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="gender"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Gender</FormLabel>
                        <Select
                          onValueChange={field.onChange}
                          value={field.value || undefined}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="Select gender" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            {GENDER_OPTIONS.map((option) => (
                              <SelectItem key={option} value={option}>
                                {option}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormDescription>
                          Auto-filled from passport OCR.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="issue_date"
                    render={({ field }) => (
                      <FormItem>
                        <FormControl>
                          <VerifiedOcrInput
                            id="issue_date"
                            label="Passport issue date"
                            type="date"
                            ocrDetectedValue={ocrDetectedValues.issue_date}
                            {...field}
                            value={field.value ?? ""}
                          />
                        </FormControl>
                        <FormDescription>
                          Auto-filled from passport OCR.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="expiry_date"
                    render={({ field }) => (
                      <FormItem>
                        <FormControl>
                          <VerifiedOcrInput
                            id="expiry_date"
                            label="Passport expiry date"
                            type="date"
                            ocrDetectedValue={ocrDetectedValues.expiry_date}
                            {...field}
                            value={field.value ?? ""}
                          />
                        </FormControl>
                        <FormDescription>
                          Manually correct any OCR misread of the expiry date.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <div className="md:col-span-2 xl:col-span-3 space-y-3">
                    <div className="flex items-center justify-between">
                      <label className="text-base font-medium">Experience abroad</label>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => appendExperience({ country: "", years: 0 })}
                      >
                        <Plus className="h-4 w-4 mr-1" />
                        Add experience
                      </Button>
                    </div>
                    {experienceFields.map((field, index) => (
                      <div key={field.id} className="flex gap-3 items-start">
                        <FormField
                          control={form.control}
                          name={`experience_abroad.${index}.country`}
                          render={({ field }) => (
                            <FormItem className="flex-1">
                              <Select onValueChange={field.onChange} value={field.value || ""}>
                                <FormControl>
                                  <SelectTrigger>
                                    <SelectValue placeholder="Select country" />
                                  </SelectTrigger>
                                </FormControl>
                                <SelectContent>
                                  {COUNTRIES_OPTIONS.map((c) => (
                                    <SelectItem key={c} value={c}>{c}</SelectItem>
                                  ))}
                                </SelectContent>
                              </Select>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                        <FormField
                          control={form.control}
                          name={`experience_abroad.${index}.years`}
                          render={({ field }) => (
                            <FormItem className="w-24">
                              <FormControl>
                                <Input
                                  type="number"
                                  min={0}
                                  max={50}
                                  placeholder="Years"
                                  {...field}
                                  value={field.value ?? 0}
                                  onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                                />
                              </FormControl>
                              <FormMessage />
                            </FormItem>
                          )}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="mt-0 shrink-0"
                          onClick={() => removeExperience(index)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    ))}
                    {experienceFields.length === 0 && (
                      <p className="text-sm text-muted-foreground">No experience abroad added yet.</p>
                    )}
                  </div>

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

                </div>
              </CardContent>
            </Card>

            <Card className="border-border/70 shadow-sm">
              <CardHeader>
                <SectionTitle
                  icon={<BriefcaseBusiness className="h-5 w-5" />}
                  title="Job details"
                  description="Destination country and salary offered for this placement."
                />
              </CardHeader>
              <CardContent className="space-y-5">
                {pairingContext?.workspaces && pairingContext.workspaces.length > 0 && (
                  <div className="space-y-4">
                    <div>
                      <h4 className="text-sm font-semibold text-foreground">Job details per partner</h4>
                      <p className="text-xs text-muted-foreground mt-1">
                        Set a unique country and salary for each foreign partner.
                      </p>
                    </div>
                    <div className="space-y-3">
                      {pairingContext.workspaces
                        .filter((ws) => ws.status === "active" || ws.status === "approved")
                        .map((ws) => {
                          const val = partnerOverrides[ws.id] || { country: "", salary: "" }
                          return (
                            <div key={ws.id} className="space-y-3 rounded-lg border border-border/70 bg-muted/20 p-4">
                              <div className="grid gap-3 sm:grid-cols-[1fr_1.5fr_1.5fr]">
                                <div className="flex items-center text-sm font-medium text-foreground">
                                  {ws.partner_agency?.company_name || ws.partner_agency?.full_name || ws.partner_agency?.email || "Partner"}
                                </div>
                                <select
                                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                  value={val.country}
                                  onChange={(e) =>
                                    setPartnerOverrides((prev) => ({
                                      ...prev,
                                      [ws.id]: { ...prev[ws.id], country: e.target.value },
                                    }))
                                  }
                                >
                                  <option value="">Select country</option>
                                  <option value="Saudi Arabia">Saudi Arabia</option>
                                  <option value="United Arab Emirates">United Arab Emirates</option>
                                  <option value="Kuwait">Kuwait</option>
                                  <option value="Qatar">Qatar</option>
                                  <option value="Bahrain">Bahrain</option>
                                  <option value="Oman">Oman</option>
                                  <option value="Lebanon">Lebanon</option>
                                  <option value="Jordan">Jordan</option>
                                </select>
                                <Input
                                  placeholder="e.g., 1000 SR, 400 USD"
                                  value={val.salary}
                                  onChange={(e) =>
                                    setPartnerOverrides((prev) => ({
                                      ...prev,
                                      [ws.id]: { ...prev[ws.id], salary: e.target.value },
                                    }))
                                  }
                                />
                              </div>
                              <Input
                                placeholder="Partner logo URL (optional)"
                                value={partnerLogoUrls[ws.id] || ""}
                                onChange={(e) =>
                                  setPartnerLogoUrls((prev) => ({
                                    ...prev,
                                    [ws.id]: e.target.value,
                                  }))
                                }
                              />
                            </div>
                          )
                        })}
                    </div>
                  </div>
                )}
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
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex flex-wrap gap-2">
                        {SKILLS_OPTIONS.map((skill) => {
                          const checked = field.value?.includes(skill);
                          return (
                            <button
                              key={skill}
                              type="button"
                              onClick={() => {
                                const next = checked
                                  ? (field.value || []).filter((s) => s !== skill)
                                  : [...(field.value || []), skill];
                                field.onChange(next);
                              }}
                              className={`inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-sm font-medium transition-all ${
                                checked
                                  ? "border-primary/60 bg-primary/10 text-primary shadow-sm"
                                  : "border-border/70 bg-muted/30 text-muted-foreground hover:border-primary/40 hover:text-foreground"
                              }`}
                            >
                              <span
                                className={`h-2 w-2 rounded-full transition-colors ${
                                  checked ? "bg-primary" : "bg-muted-foreground/30"
                                }`}
                              />
                              {skill}
                            </button>
                          );
                        })}
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

            <Card className="border-border/70 shadow-sm">
              <CardHeader>
                <SectionTitle
                  icon={<UserSquare2 className="h-5 w-5" />}
                  title="Additional notes"
                  description="Add an optional, brief remark that will be visible on the candidate's CV."
                />
              </CardHeader>
              <CardContent>
                <FormField
                  control={form.control}
                  name="remark"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Remark (Optional)</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="e.g., Highly recommended, beginner, fast learner..."
                          {...field}
                          value={field.value ?? ""}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </CardContent>
            </Card>
        </div>

      </form>

      <div className="fixed bottom-0 left-0 right-0 z-50 border-t border-border/70 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 shadow-[0_-4px_20px_rgba(0,0,0,0.08)]">
        <div className="mx-auto flex max-w-6xl items-center justify-end gap-3 px-3 py-4 sm:px-5 lg:px-6">
          {!isEditing ? (
            <Button
              type="submit"
              form="candidate-form"
              size="lg"
              variant="outline"
              className="min-w-[180px]"
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
            form="candidate-form"
            size="lg"
            className="min-w-[180px] shadow-md"
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
      </div>
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

function buildPassportPreviewRequestKey(file: File) {
  return [
    file.name,
    file.size,
    file.lastModified,
    file.type,
  ].join(":");
}

async function buildPassportOCRPreviewFile(file: File) {
  if (typeof window === "undefined" || !file.type.startsWith("image/")) {
    return file;
  }

  if (!("createImageBitmap" in window)) {
    return file;
  }

  try {
    const bitmap = await createImageBitmap(file);
    try {
      const longestSide = Math.max(bitmap.width, bitmap.height);
      const scale =
        longestSide > PASSPORT_OCR_PREVIEW_MAX_DIMENSION
          ? PASSPORT_OCR_PREVIEW_MAX_DIMENSION / longestSide
          : 1;

      if (scale >= 1 && file.size <= 2*1024*1024) {
        return file;
      }

      const canvas = document.createElement("canvas");
      canvas.width = Math.max(1, Math.round(bitmap.width * scale));
      canvas.height = Math.max(1, Math.round(bitmap.height * scale));

      const context = canvas.getContext("2d");
      if (!context) {
        return file;
      }

      context.drawImage(bitmap, 0, 0, canvas.width, canvas.height);

      const blob = await new Promise<Blob | null>((resolve) => {
        canvas.toBlob(resolve, "image/jpeg", PASSPORT_OCR_PREVIEW_QUALITY);
      });

      if (!blob) {
        return file;
      }

      return new File(
        [blob],
        replaceFileExtension(file.name, ".jpg"),
        {
          type: "image/jpeg",
          lastModified: file.lastModified,
        },
      );
    } finally {
      bitmap.close();
    }
  } catch {
    return file;
  }
}

function replaceFileExtension(fileName: string, nextExtension: string) {
  const normalizedExtension = nextExtension.startsWith(".")
    ? nextExtension
    : `.${nextExtension}`;
  const lastDotIndex = fileName.lastIndexOf(".");
  if (lastDotIndex <= 0) {
    return `${fileName}${normalizedExtension}`;
  }
  return `${fileName.slice(0, lastDotIndex)}${normalizedExtension}`;
}
