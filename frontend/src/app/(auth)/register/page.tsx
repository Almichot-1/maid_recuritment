"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Eye, EyeOff, Loader2, Globe, Users } from "lucide-react"

import { registerSchema, type RegisterFormInput } from "@/lib/validations"
import { useRegister } from "@/hooks/use-auth"
import { useAuthStore } from "@/stores/auth-store"
import { UserRole } from "@/types"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Checkbox } from "@/components/ui/checkbox"
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form"
import { cn } from "@/lib/utils"

function calculateStrength(password: string): number {
  if (!password) return 0;
  let strength = 0;
  if (password.length >= 8) strength += 1;
  if (/\d/.test(password)) strength += 1;
  if (/[!@#$%^&*(),.?":{}|<>]/.test(password)) strength += 1;
  return strength;
}

export default function RegisterPage() {
  const router = useRouter();
  const { mutate: register, isPending } = useRegister();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  
  const [showPassword, setShowPassword] = React.useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = React.useState(false);

  React.useEffect(() => {
    if (isAuthenticated) {
      router.push("/dashboard");
    }
  }, [isAuthenticated, router]);

  const form = useForm<RegisterFormInput>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: "",
      password: "",
      confirmPassword: "",
      full_name: "",
      role: UserRole.ETHIOPIAN_AGENT,
      company_name: "",
      acceptTerms: false,
    },
  });

  const passwordVal = form.watch("password");
  const strength = calculateStrength(passwordVal);

  function onSubmit(data: RegisterFormInput) {
    register(data);
  }

  if (isAuthenticated) return null;

  return (
    <Card className="w-full max-w-2xl shadow-lg border-muted my-8">
      <CardHeader className="space-y-2 text-center pb-6">
        <CardTitle className="text-2xl font-bold tracking-tight">Create an Account</CardTitle>
        <CardDescription>Register your agency and submit it for admin approval</CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5">
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="full_name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Full Name</FormLabel>
                    <FormControl>
                      <Input placeholder="John Doe" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="company_name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Company Name</FormLabel>
                    <FormControl>
                      <Input placeholder="Agency LLC" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email Address</FormLabel>
                  <FormControl>
                    <Input placeholder="name@example.com" type="email" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password</FormLabel>
                    <FormControl>
                      <div className="relative">
                        <Input
                          type={showPassword ? "text" : "password"}
                          placeholder="••••••••"
                          className="pr-10"
                          {...field}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent text-muted-foreground"
                          onClick={() => setShowPassword(!showPassword)}
                        >
                          {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                        </Button>
                      </div>
                    </FormControl>
                    {/* Password Strength Indicator */}
                    <div className="mt-2 flex h-1.5 w-full gap-1 overflow-hidden rounded-full bg-slate-100 dark:bg-slate-800">
                      <div className={cn("h-full flex-1 transition-all", strength >= 1 ? (strength === 1 ? "bg-red-500" : strength === 2 ? "bg-yellow-500" : "bg-green-500") : "bg-transparent")} />
                      <div className={cn("h-full flex-1 transition-all", strength >= 2 ? (strength === 2 ? "bg-yellow-500" : "bg-green-500") : "bg-transparent")} />
                      <div className={cn("h-full flex-1 transition-all", strength >= 3 ? "bg-green-500" : "bg-transparent")} />
                    </div>
                    <p className="text-xs text-muted-foreground mt-1 min-h-[16px]">
                      {strength === 0 && ""}
                      {strength === 1 && "Weak - Add numbers and special characters"}
                      {strength === 2 && "Medium - Add special characters"}
                      {strength === 3 && "Strong password"}
                    </p>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="confirmPassword"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Confirm Password</FormLabel>
                    <FormControl>
                      <div className="relative">
                        <Input
                          type={showConfirmPassword ? "text" : "password"}
                          placeholder="••••••••"
                          className="pr-10"
                          {...field}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent text-muted-foreground"
                          onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                        >
                          {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                        </Button>
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="role"
              render={({ field }) => (
                <FormItem className="space-y-3 pt-2">
                  <FormLabel>Account Type</FormLabel>
                  <FormControl>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {/* Ethiopian Agency Card */}
                      <div
                        className={cn(
                          "relative flex cursor-pointer rounded-lg border bg-card p-4 shadow-sm hover:border-primary transition-all",
                          field.value === UserRole.ETHIOPIAN_AGENT ? "border-primary ring-1 ring-primary" : "border-muted"
                        )}
                        onClick={() => field.onChange(UserRole.ETHIOPIAN_AGENT)}
                      >
                        <div className="flex items-start space-x-3">
                          <Users className={cn("mt-0.5 h-5 w-5", field.value === UserRole.ETHIOPIAN_AGENT ? "text-primary" : "text-muted-foreground")} />
                          <div className="space-y-1">
                            <p className="text-sm font-medium leading-none">Ethiopian Agency</p>
                            <p className="text-xs text-muted-foreground">I manage and recruit candidates</p>
                          </div>
                        </div>
                      </div>

                      {/* Foreign Agency Card */}
                      <div
                        className={cn(
                          "relative flex cursor-pointer rounded-lg border bg-card p-4 shadow-sm hover:border-primary transition-all",
                          field.value === UserRole.FOREIGN_AGENT ? "border-primary ring-1 ring-primary" : "border-muted"
                        )}
                        onClick={() => field.onChange(UserRole.FOREIGN_AGENT)}
                      >
                        <div className="flex items-start space-x-3">
                          <Globe className={cn("mt-0.5 h-5 w-5", field.value === UserRole.FOREIGN_AGENT ? "text-primary" : "text-muted-foreground")} />
                          <div className="space-y-1">
                            <p className="text-sm font-medium leading-none">Foreign Agency (Jordan)</p>
                            <p className="text-xs text-muted-foreground">I browse and select candidates</p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="acceptTerms"
              render={({ field }) => (
                <FormItem className="flex flex-row items-start space-x-3 space-y-0 pt-4">
                  <FormControl>
                    <Checkbox
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                  <div className="space-y-1 leading-none">
                    <FormLabel className="cursor-pointer font-normal">
                      I accept the terms and conditions
                    </FormLabel>
                  </div>
                </FormItem>
              )}
            />

            <Button type="submit" className="w-full" disabled={isPending}>
              {isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Submitting registration...
                </>
              ) : (
                "Submit For Review"
              )}
            </Button>
          </form>
        </Form>
      </CardContent>
      <CardFooter className="flex flex-col space-y-4 text-center text-sm text-muted-foreground border-t pt-4">
        <div>
          Already have an account?{" "}
          <Link href="/login" className="font-semibold text-primary hover:underline transition-colors">
            Login
          </Link>
        </div>
      </CardFooter>
    </Card>
  )
}
