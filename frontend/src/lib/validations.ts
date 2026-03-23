import { z } from 'zod';
import { UserRole } from '@/types';

export const loginSchema = z.object({
  email: z.string().email('Please enter a valid email address.'),
  password: z.string().min(1, 'Password is required.'),
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
  age: z.coerce.number().min(18, 'Candidate must be at least 18 years old.').max(65, 'Candidate must be at most 65 years old.'),
  experience_years: z.coerce.number().min(0).max(30),
  skills: z.array(z.string()).min(1, 'At least one skill must be selected.'),
  languages: z.array(z.object({
    language: z.string(),
    proficiency: z.string()
  })).min(1, 'At least one language must be selected.'),
});

export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterFormInput = z.infer<typeof registerSchema>;
export type CandidateInput = z.infer<typeof candidateSchema>;

export interface RegisterInput {
  email: string;
  password: string;
  full_name: string;
  role: UserRole;
  company_name: string;
}
