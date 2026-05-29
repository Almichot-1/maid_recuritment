import { UserCheck, CheckCircle2, XCircle, ArrowRight, Clock, Bell, ShieldAlert, PlaneTakeoff, PlaneLanding } from "lucide-react"

export function getNotificationIcon(type: string) {
  switch (type) {
    case "selection":
      return UserCheck
    case "approval":
      return CheckCircle2
    case "rejection":
      return XCircle
    case "status_update":
      return ArrowRight
    case "expiry":
      return Clock
    case "expiry_warning":
      return Clock
    case "passport_expiry_warning":
      return ShieldAlert
    case "flight_booked":
      return PlaneTakeoff
    case "arrived":
      return PlaneLanding
    default:
      return Bell
  }
}

export function getNotificationColor(type: string) {
  switch (type) {
    case "selection":
      return "bg-primary/10 text-primary"
    case "approval":
      return "bg-primary/15 text-primary"
    case "rejection":
      return "bg-destructive/10 text-destructive"
    case "status_update":
      return "bg-accent text-accent-foreground"
    case "expiry":
    case "expiry_warning":
    case "passport_expiry_warning":
      return "bg-muted text-muted-foreground"
    case "flight_booked":
      return "bg-secondary text-secondary-foreground"
    case "arrived":
      return "bg-primary/10 text-primary"
    default:
      return "bg-muted text-muted-foreground"
  }
}

export function getNotificationTypeLabel(type: string) {
  switch (type) {
    case "selection":
      return "Selection"
    case "approval":
      return "Approval"
    case "rejection":
      return "Rejection"
    case "status_update":
      return "Status update"
    case "expiry":
      return "Expiry"
    case "expiry_warning":
      return "Expiry warning"
    case "passport_expiry_warning":
      return "Passport warning"
    case "flight_booked":
      return "Flight booked"
    case "arrived":
      return "Arrived"
    default:
      return "Notification"
  }
}
