import { z } from 'zod';
import { UserRole } from '@/types';

const optionalTrimmedString = z.preprocess((value) => {
  if (typeof value !== 'string') {
    return value;
  }
  const trimmed = value.trim();
  return trimmed === '' ? undefined : trimmed;
}, z.string().optional());

const optionalNumber = z.preprocess((value) => {
  if (value === '' || value === null || value === undefined) {
    return undefined;
  }
  return value;
}, z.coerce.number().optional());

export const loginSchema = z.object({
  email: z.string().email('Please enter a valid email address.'),
  password: z.string().min(1, 'Password is required.'),
});

export const forgotPasswordRequestSchema = z.object({
  email: z.string().email('Please enter a valid email address.'),
});

export const forgotPasswordResetSchema = z.object({
  email: z.string().email('Please enter a valid email address.'),
  code: z.string().length(6, 'Enter the 6-digit code from your email.'),
  new_password: z.string().min(8, 'Password must be at least 8 characters long.'),
  confirmPassword: z.string(),
}).refine((data) => data.new_password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

export const registerSchema = z.object({
  email: z.string().email('Please enter a valid email address.'),
  password: z.string().min(8, 'Password must be at least 8 characters long.'),
  confirmPassword: z.string(),
  full_name: z.string().min(2, 'Full name must be at least 2 characters long.'),
  role: z.nativeEnum(UserRole, {
    message: 'Please select a valid role.',
  }),
  company_name: z.string().min(2, 'Company name must be at least 2 characters long.'),
  acceptTerms: z.boolean().refine(val => val === true, {
    message: 'You must accept the terms and conditions',
  })
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

export const candidateSchema = z.object({
  full_name: z.string().min(2, 'Full name is required.'),
  nationality: optionalTrimmedString,
  date_of_birth: optionalTrimmedString.refine((value) => {
    if (!value) {
      return true;
    }
    return !Number.isNaN(new Date(value).getTime());
  }, 'Please enter a valid date of birth.'),
  age: z.coerce.number().min(18, 'Candidate must be at least 18 years old.').max(65, 'Candidate must be at most 65 years old.'),
  place_of_birth: optionalTrimmedString,
  religion: optionalTrimmedString,
  marital_status: optionalTrimmedString,
  children_count: optionalNumber.refine((value) => value === undefined || value >= 0, 'Children count cannot be negative.'),
  education_level: optionalTrimmedString,
  experience_years: z.coerce.number().min(0).max(30),
  skills: z.array(z.string()).min(1, 'At least one skill must be selected.'),
  languages: z.array(z.object({
    language: z.string(),
    proficiency: z.string()
  })).min(1, 'At least one language must be selected.'),
});

export type LoginInput = z.infer<typeof loginSchema>;
export type ForgotPasswordRequestInput = z.infer<typeof forgotPasswordRequestSchema>;
export type ForgotPasswordResetInput = z.infer<typeof forgotPasswordResetSchema>;
export type RegisterFormInput = z.infer<typeof registerSchema>;
export type CandidateInput = z.infer<typeof candidateSchema>;

export interface RegisterInput {
  email: string;
  password: string;
  full_name: string;
  role: UserRole;
  company_name: string;
}
