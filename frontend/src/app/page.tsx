import Link from "next/link"
import { ArrowRight, Clock, Eye, Bell, Users, TrendingUp } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Logo } from "@/components/shared/logo"
import { ThemeToggle } from "@/components/shared/theme-toggle"
import { APP_NAME, COMPANY_INFO } from "@/constants/branding"

export default function LandingPage() {
  return (
    <div className="flex min-h-screen flex-col">
      {/* Header */}
      <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container flex h-16 items-center justify-between">
          <Logo size="md" />
          <nav className="flex items-center gap-4">
            <ThemeToggle className="hidden sm:inline-flex" />
            <Link href="/login">
              <Button variant="ghost">Login</Button>
            </Link>
            <Link href="/register">
              <Button>Get Started</Button>
            </Link>
          </nav>
        </div>
      </header>

      {/* Hero Section */}
      <section className="container flex flex-col items-center justify-center gap-8 py-16 md:py-24 lg:py-32">
        <div className="flex max-w-[980px] flex-col items-center gap-4 text-center">
          <h1 className="text-4xl font-bold leading-tight tracking-tighter md:text-6xl lg:text-7xl">
            {COMPANY_INFO.tagline}
          </h1>
          <p className="max-w-[750px] text-lg text-muted-foreground sm:text-xl">
            {COMPANY_INFO.description}
          </p>
          <div className="flex flex-col gap-4 sm:flex-row mt-4">
            <Link href="/register?role=ethiopian_agent">
              <Button size="lg" className="w-full sm:w-auto">
                Register as Ethiopian Agency
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </Link>
            <Link href="/register?role=foreign_agent">
              <Button size="lg" variant="outline" className="w-full sm:w-auto">
                Register as Foreign Agency
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </Link>
          </div>
        </div>

        {/* Hero Image Placeholder */}
        <div className="relative mt-8 w-full max-w-5xl">
          <div className="aspect-video rounded-xl border bg-muted/50 flex items-center justify-center">
            <Users className="h-24 w-24 text-muted-foreground/20" />
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="container py-16 md:py-24">
        <div className="mx-auto flex max-w-[980px] flex-col items-center gap-4 text-center mb-12">
          <h2 className="text-3xl font-bold leading-tight tracking-tighter md:text-5xl">
            Everything You Need
          </h2>
          <p className="max-w-[750px] text-lg text-muted-foreground">
            Streamline your recruitment process with powerful features designed for efficiency and transparency.
          </p>
        </div>

        <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col items-center text-center gap-4">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                  <Eye className="h-6 w-6 text-primary" />
                </div>
                <h3 className="font-semibold">Browse & Select</h3>
                <p className="text-sm text-muted-foreground">
                  Browse candidate profiles with detailed CVs, photos, and video interviews.
                </p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col items-center text-center gap-4">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                  <Clock className="h-6 w-6 text-primary" />
                </div>
                <h3 className="font-semibold">Secure Selection</h3>
                <p className="text-sm text-muted-foreground">
                  24-hour lock ensures exclusive access when you select a candidate.
                </p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col items-center text-center gap-4">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                  <TrendingUp className="h-6 w-6 text-primary" />
                </div>
                <h3 className="font-semibold">Transparent Tracking</h3>
                <p className="text-sm text-muted-foreground">
                  Track every step of the recruitment process from selection to deployment.
                </p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col items-center text-center gap-4">
                <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                  <Bell className="h-6 w-6 text-primary" />
                </div>
                <h3 className="font-semibold">Instant Notifications</h3>
                <p className="text-sm text-muted-foreground">
                  Stay updated on selections, approvals, and status changes in real-time.
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </section>

      {/* How It Works Section */}
      <section className="container py-16 md:py-24 bg-muted/50">
        <div className="mx-auto flex max-w-[980px] flex-col items-center gap-4 text-center mb-12">
          <h2 className="text-3xl font-bold leading-tight tracking-tighter md:text-5xl">
            How It Works
          </h2>
          <p className="max-w-[750px] text-lg text-muted-foreground">
            Simple, transparent process from start to finish.
          </p>
        </div>

        <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-4 max-w-6xl mx-auto">
          <div className="flex flex-col items-center text-center gap-4">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary text-primary-foreground text-2xl font-bold">
              1
            </div>
            <h3 className="font-semibold">Create Profiles</h3>
            <p className="text-sm text-muted-foreground">
              Ethiopian agencies create detailed candidate profiles with documents and videos.
            </p>
          </div>

          <div className="flex flex-col items-center text-center gap-4">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary text-primary-foreground text-2xl font-bold">
              2
            </div>
            <h3 className="font-semibold">Browse & Select</h3>
            <p className="text-sm text-muted-foreground">
              Foreign agencies browse available candidates and make selections.
            </p>
          </div>

          <div className="flex flex-col items-center text-center gap-4">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary text-primary-foreground text-2xl font-bold">
              3
            </div>
            <h3 className="font-semibold">Mutual Approval</h3>
            <p className="text-sm text-muted-foreground">
              Both parties review and approve the selection within 24 hours.
            </p>
          </div>

          <div className="flex flex-col items-center text-center gap-4">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary text-primary-foreground text-2xl font-bold">
              4
            </div>
            <h3 className="font-semibold">Track Progress</h3>
            <p className="text-sm text-muted-foreground">
              Monitor recruitment progress together through every step.
            </p>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="container py-16 md:py-24">
        <div className="mx-auto flex max-w-[980px] flex-col items-center gap-4 text-center">
          <h2 className="text-3xl font-bold leading-tight tracking-tighter md:text-5xl">
            Ready to Get Started?
          </h2>
          <p className="max-w-[750px] text-lg text-muted-foreground mb-4">
            Join hundreds of agencies already using {APP_NAME} for efficient recruitment.
          </p>
          <Link href="/register">
            <Button size="lg">
              Create Your Account
              <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t py-8 md:py-12">
        <div className="container flex flex-col gap-8">
          <div className="flex flex-col items-center justify-between gap-4 md:flex-row">
            <Logo size="sm" />
            <nav className="flex gap-6">
              <Link href="/login" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                Login
              </Link>
              <Link href="/register" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                Register
              </Link>
              <Link href="mailto:contact@recruitmatch.com" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                Contact
              </Link>
            </nav>
          </div>
          <div className="flex flex-col items-center justify-between gap-4 md:flex-row">
            <p className="text-sm text-muted-foreground">
              (c) {COMPANY_INFO.year} {COMPANY_INFO.name}. All rights reserved.
            </p>
          </div>
        </div>
      </footer>
    </div>
  )
}

