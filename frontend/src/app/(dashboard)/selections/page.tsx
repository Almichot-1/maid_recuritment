"use client"

import * as React from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { ChevronRight, Clock3, Home, Inbox, Search, ShieldCheck, XCircle } from "lucide-react"

import { useCurrentUser } from "@/hooks/use-auth"
import { usePairingContext } from "@/hooks/use-pairings"
import { useMySelections } from "@/hooks/use-selections"
import { PageHeader } from "@/components/layout/page-header"
import { SelectionList } from "@/components/selections/selection-list"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { SelectionStatus } from "@/types"

type SelectionTab = "active" | "approved" | "rejected" | "expired" | "all"

function pickInitialTab(selectionsLengthByStatus: Record<SelectionTab, number>) {
  if (selectionsLengthByStatus.active > 0) return "active"
  if (selectionsLengthByStatus.approved > 0) return "approved"
  if (selectionsLengthByStatus.rejected > 0) return "rejected"
  if (selectionsLengthByStatus.expired > 0) return "expired"
  return "all"
}

export default function SelectionsPage() {
  const router = useRouter()
  const { isEthiopianAgent } = useCurrentUser()
  const { activeWorkspace } = usePairingContext()
  const { data: selections, isLoading, refetch } = useMySelections()

  const activeSelections = React.useMemo(
    () => selections?.filter((selection) => selection.status === SelectionStatus.PENDING) || [],
    [selections]
  )
  const approvedSelections = React.useMemo(
    () => selections?.filter((selection) => selection.status === SelectionStatus.APPROVED) || [],
    [selections]
  )
  const rejectedSelections = React.useMemo(
    () => selections?.filter((selection) => selection.status === SelectionStatus.REJECTED) || [],
    [selections]
  )
  const expiredSelections = React.useMemo(
    () => selections?.filter((selection) => selection.status === SelectionStatus.EXPIRED) || [],
    [selections]
  )

  const counts = React.useMemo(
    () => ({
      active: activeSelections.length,
      approved: approvedSelections.length,
      rejected: rejectedSelections.length,
      expired: expiredSelections.length,
      all: selections?.length || 0,
    }),
    [activeSelections.length, approvedSelections.length, rejectedSelections.length, expiredSelections.length, selections?.length]
  )

  const [activeTab, setActiveTab] = React.useState<SelectionTab>(() => pickInitialTab({ ...counts, all: 0 }))

  React.useEffect(() => {
    const interval = setInterval(() => {
      if (activeSelections.length > 0) {
        refetch()
      }
    }, 30000)

    return () => clearInterval(interval)
  }, [activeSelections.length, refetch])

  React.useEffect(() => {
    const nextTab = pickInitialTab(counts)
    if (counts[activeTab] === 0 && counts.all > 0) {
      setActiveTab(nextTab)
    }
    if (counts.all === 0) {
      setActiveTab("all")
    }
  }, [activeTab, counts])

  const breadcrumbs = (
    <nav className="mb-6 flex items-center text-sm font-medium text-muted-foreground">
      <Link href="/dashboard" className="flex items-center transition-all hover:text-primary">
        <Home className="mr-1.5 h-4 w-4" />
        Dashboard
      </Link>
      <ChevronRight className="mx-1 h-4 w-4 opacity-50" />
      <span className="font-semibold text-foreground">My Selections</span>
    </nav>
  )

  const emptyState = isEthiopianAgent ? (
    <div className="flex flex-col items-center justify-center py-16 space-y-4">
      <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted">
        <Inbox className="h-10 w-10 text-muted-foreground" />
      </div>
      <div className="space-y-2 text-center">
        <h3 className="text-lg font-semibold">No selections yet</h3>
        <p className="max-w-md text-muted-foreground">
          Once a foreign employer selects one of your candidates, the approval and tracking flow will appear here.
        </p>
      </div>
    </div>
  ) : (
    <div className="flex flex-col items-center justify-center py-16 space-y-4">
      <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted">
        <Search className="h-10 w-10 text-muted-foreground" />
      </div>
      <div className="space-y-2 text-center">
        <h3 className="text-lg font-semibold">No selections yet</h3>
        <p className="max-w-md text-muted-foreground">
          Your selected candidates and their approval or recruitment tracking states will appear here automatically.
        </p>
      </div>
      <Button onClick={() => router.push("/candidates")}>Browse Candidates</Button>
    </div>
  )

  const summaryCards = [
    {
      label: "Awaiting approval",
      value: counts.active,
      icon: <Clock3 className="h-5 w-5 text-amber-600" />,
      tone: "bg-amber-100 text-amber-900",
    },
    {
      label: "Fully approved",
      value: counts.approved,
      icon: <ShieldCheck className="h-5 w-5 text-emerald-600" />,
      tone: "bg-emerald-100 text-emerald-900",
    },
    {
      label: "Rejected",
      value: counts.rejected,
      icon: <XCircle className="h-5 w-5 text-rose-600" />,
      tone: "bg-rose-100 text-rose-900",
    },
  ]

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500 pb-10">
      {breadcrumbs}

      <PageHeader
        heading="My Selections"
        text={`Track candidate approvals, lock windows, and recruitment progress inside your current private workspace${activeWorkspace ? ` with ${activeWorkspace.partner_agency.company_name || activeWorkspace.partner_agency.full_name}` : ""}.`}
        action={
          !isEthiopianAgent ? (
            <Button asChild>
              <Link href="/candidates">Select more candidates</Link>
            </Button>
          ) : undefined
        }
      />

      <div className="grid gap-4 md:grid-cols-3">
        {summaryCards.map((card) => (
          <Card key={card.label} className="shadow-sm">
            <CardContent className="flex items-center justify-between p-5">
              <div>
                <p className="text-sm text-muted-foreground">{card.label}</p>
                <p className="mt-1 text-3xl font-semibold">{card.value}</p>
              </div>
              <div className={`flex h-12 w-12 items-center justify-center rounded-2xl ${card.tone}`}>
                {card.icon}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as SelectionTab)} className="space-y-6">
        <TabsList className="grid h-auto w-full grid-cols-2 gap-2 p-1 md:grid-cols-5">
          <SelectionTabTrigger value="active" label="Active" count={counts.active} tone="bg-amber-500" />
          <SelectionTabTrigger value="approved" label="Approved" count={counts.approved} tone="bg-emerald-500" />
          <SelectionTabTrigger value="rejected" label="Rejected" count={counts.rejected} tone="bg-rose-500" />
          <SelectionTabTrigger value="expired" label="Expired" count={counts.expired} tone="bg-slate-500" />
          <SelectionTabTrigger value="all" label="All" count={counts.all} />
        </TabsList>

        <TabsContent value="active">
          {counts.active > 0 ? <SelectionList selections={activeSelections} isLoading={isLoading} /> : emptyState}
        </TabsContent>

        <TabsContent value="approved">
          {counts.approved > 0 ? (
            <SelectionList selections={approvedSelections} isLoading={isLoading} />
          ) : (
            <EmptyPanel text="No fully approved selections yet." />
          )}
        </TabsContent>

        <TabsContent value="rejected">
          {counts.rejected > 0 ? (
            <SelectionList selections={rejectedSelections} isLoading={isLoading} />
          ) : (
            <EmptyPanel text="No rejected selections." />
          )}
        </TabsContent>

        <TabsContent value="expired">
          {counts.expired > 0 ? (
            <SelectionList selections={expiredSelections} isLoading={isLoading} />
          ) : (
            <EmptyPanel text="No expired selections." />
          )}
        </TabsContent>

        <TabsContent value="all">
          {counts.all > 0 ? <SelectionList selections={selections || []} isLoading={isLoading} /> : emptyState}
        </TabsContent>
      </Tabs>
    </div>
  )
}

function SelectionTabTrigger({
  value,
  label,
  count,
  tone = "",
}: {
  value: SelectionTab
  label: string
  count: number
  tone?: string
}) {
  return (
    <TabsTrigger value={value} className="py-2.5">
      {label}
      {count > 0 ? <Badge className={`ml-2 text-white hover:text-white ${tone || "bg-primary"}`}>{count}</Badge> : null}
    </TabsTrigger>
  )
}

function EmptyPanel({ text }: { text: string }) {
  return (
    <div className="rounded-2xl border border-dashed bg-card/40 py-14 text-center text-sm text-muted-foreground">
      {text}
    </div>
  )
}
