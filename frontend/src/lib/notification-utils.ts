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
      return "bg-blue-100 dark:bg-blue-950/30 text-blue-600 dark:text-blue-400"
    case "approval":
      return "bg-green-100 dark:bg-green-950/30 text-green-600 dark:text-green-400"
    case "rejection":
      return "bg-red-100 dark:bg-red-950/30 text-red-600 dark:text-red-400"
    case "status_update":
      return "bg-purple-100 dark:bg-purple-950/30 text-purple-600 dark:text-purple-400"
    case "expiry":
      return "bg-orange-100 dark:bg-orange-950/30 text-orange-600 dark:text-orange-400"
    case "expiry_warning":
      return "bg-orange-100 dark:bg-orange-950/30 text-orange-600 dark:text-orange-400"
    case "passport_expiry_warning":
      return "bg-amber-100 dark:bg-amber-950/30 text-amber-600 dark:text-amber-400"
    case "flight_booked":
      return "bg-sky-100 dark:bg-sky-950/30 text-sky-600 dark:text-sky-400"
    case "arrived":
      return "bg-emerald-100 dark:bg-emerald-950/30 text-emerald-600 dark:text-emerald-400"
    default:
      return "bg-gray-100 dark:bg-gray-950/30 text-gray-600 dark:text-gray-400"
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
      return "Status Update"
    case "expiry":
      return "Expiry"
    case "expiry_warning":
      return "Expiry Warning"
    case "passport_expiry_warning":
      return "Passport Warning"
    case "flight_booked":
      return "Flight Booked"
    case "arrived":
      return "Arrived"
    default:
      return "Notification"
  }
}
