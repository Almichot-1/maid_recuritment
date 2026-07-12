export function shareOnWhatsApp(text: string) {
  const encoded = encodeURIComponent(text)
  window.open(`https://wa.me/?text=${encoded}`, "_blank")
}

export interface WhatsAppCandidate {
  id: string
  full_name: string
  age?: number | null
  experience_years?: number | null
  cv_pdf_url?: string | null
  nationality?: string | null
  skills?: string[]
}

export function buildCandidateMessage(candidate: WhatsAppCandidate): string {
  const origin = typeof window !== "undefined" ? window.location.origin : ""
  const lines: string[] = []
  lines.push(`*Candidate: ${candidate.full_name}*`)
  if (candidate.age != null) lines.push(`Age: ${candidate.age} years`)
  if (candidate.experience_years != null) lines.push(`Experience: ${candidate.experience_years} years`)
  if (candidate.nationality) lines.push(`Nationality: ${candidate.nationality}`)
  if (candidate.skills && candidate.skills.length > 0) {
    lines.push(`Skills: ${candidate.skills.slice(0, 5).join(", ")}`)
  }
  if (candidate.cv_pdf_url) {
    lines.push(`CV: ${candidate.cv_pdf_url}`)
  } else {
    lines.push(`Profile: ${origin}/candidates/${candidate.id}`)
  }
  return lines.join("\n")
}

export function buildBatchCandidateMessage(candidates: WhatsAppCandidate[]): string {
  const parts: string[] = []
  candidates.forEach((c, i) => {
    parts.push(`${i + 1}. ${c.full_name}${c.age != null ? ` - ${c.age}yrs` : ""}${c.experience_years != null ? `, ${c.experience_years}yrs exp` : ""}${c.cv_pdf_url ? `\n   CV: ${c.cv_pdf_url}` : ""}`)
  })
  return `*Candidates:*\n${parts.join("\n")}`
}
