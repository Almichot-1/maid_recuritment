"use client"

import * as React from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { AlertCircle, Eye, EyeOff, Loader2, Mail } from "lucide-react"

import { loginSchema, type LoginInput } from "@/lib/validations"
import { getLoginErrorMessage, useLogin } from "@/hooks/use-auth"
import { useAuthStore } from "@/stores/auth-store"
import { getRoleHomePath } from "@/lib/role-home"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form"
import { Input } from "@/components/ui/input"

export default function LoginPage() {
  const router = useRouter()
  const { mutate: login, isPending } = useLogin()
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const user = useAuthStore((state) => state.user)
  const [showPassword, setShowPassword] = React.useState(false)

  React.useEffect(() => {
    if (isAuthenticated) {
      router.replace(getRoleHomePath(user?.role))
    }
  }, [isAuthenticated, router, user?.role])

  const form = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  })
  const watchedEmail = form.watch("email")
  const watchedPassword = form.watch("password")

  React.useEffect(() => {
    if (form.formState.errors.root?.message) {
      form.clearErrors("root")
    }
  }, [watchedEmail, watchedPassword, form])

  function onSubmit(data: LoginInput) {
    form.clearErrors("root")
    login(data, {
      onError: (error) => {
        form.setError("root", {
          type: "server",
          message: getLoginErrorMessage(error),
        })
      },
    })
  }

  if (isAuthenticated) {
    return null
  }

  return (
    <Card className="animated-border w-full max-w-md overflow-hidden border-muted shadow-glow">
      <CardHeader className="space-y-2 pb-6 text-center">
        <CardTitle className="text-2xl font-bold tracking-tight">Maid Recruitment Platform</CardTitle>
        <CardDescription>Enter your credentials to access the platform</CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-5">
            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email</FormLabel>
                  <FormControl>
                    <div className="relative">
                      <Mail className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
                      <Input
                        placeholder="name@example.com"
                        className="bg-background pl-9"
                        autoComplete="username"
                        {...field}
                      />
                    </div>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

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
                        placeholder="Password"
                        className="bg-background pr-10"
                        autoComplete="current-password"
                        {...field}
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="absolute right-0 top-0 h-full px-3 py-2 text-muted-foreground transition-colors hover:bg-transparent hover:text-foreground"
                        onClick={() => setShowPassword(!showPassword)}
                      >
                        {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                        <span className="sr-only">Toggle password visibility</span>
                      </Button>
                    </div>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="flex items-center space-x-2 pt-1">
              <Checkbox id="remember" />
              <label
                htmlFor="remember"
                className="cursor-pointer text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
              >
                Remember me
              </label>
            </div>

            <div className="flex justify-end">
              <Link href="/forgot-password" className="text-sm font-medium text-primary transition-colors hover:underline">
                Forgot password?
              </Link>
            </div>

            {form.formState.errors.root?.message ? (
              <div className="rounded-2xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive shadow-sm" role="alert">
                <div className="flex items-start gap-3">
                  <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                  <span>{form.formState.errors.root.message}</span>
                </div>
              </div>
            ) : null}

            <Button type="submit" className="mt-2 w-full" disabled={isPending}>
              {isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Signing in...
                </>
              ) : (
                "Sign In"
              )}
            </Button>
          </form>
        </Form>
      </CardContent>
      <CardFooter className="flex flex-col space-y-4 border-t pt-4 text-center text-sm text-muted-foreground">
        <div>
          Need an account?{" "}
          <Link href="/register" className="font-semibold text-primary transition-colors hover:underline">
            Register
          </Link>
        </div>
      </CardFooter>
    </Card>
  )
}
