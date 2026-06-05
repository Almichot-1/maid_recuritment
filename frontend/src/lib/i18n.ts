/**
 * Internationalization (i18n) utility
 * Currently provides a simple translation interface
 * Can be extended with full i18n library in the future
 */

type TranslationKey = string

interface TranslationParams {
  [key: string]: string | number
}

const translations: Record<TranslationKey, string> = {
  // Dashboard translations
  "dashboard.foreignLabel": "Foreign Employer Portal",
  "dashboard.foreignTitle": "Find the right candidates",
  "dashboard.foreignBody": "Browse profiles, select candidates, and track recruitment progress with your Ethiopian partner agency.",
  "dashboard.activePartner": "Active partner: {{name}}",
  "dashboard.browsePool": "Browse candidate pool",
  "dashboard.openSelections": "View open selections",
  "dashboard.quickPulse": "Quick pulse",
  "dashboard.availableNow": "Available now",
  "dashboard.pendingApprovals": "Pending approvals",
  "dashboard.approvedProfiles": "Approved profiles",
}

/**
 * Replace template variables in translation strings
 * Example: "Hello {{name}}" with {name: "John"} => "Hello John"
 */
function interpolate(template: string, params?: TranslationParams): string {
  if (!params) return template
  
  return Object.entries(params).reduce((result, [key, value]) => {
    return result.replace(new RegExp(`\\{\\{${key}\\}\\}`, 'g'), String(value))
  }, template)
}

/**
 * Translate a key with optional parameters
 */
export function translate(key: TranslationKey, params?: TranslationParams): string {
  const template = translations[key] || key
  return interpolate(template, params)
}

/**
 * Hook for using translations in React components
 */
export function useI18n() {
  const t = (key: TranslationKey, params?: TranslationParams) => translate(key, params)
  
  return { t }
}

/**
 * Add translations dynamically (useful for extending translations)
 */
export function addTranslations(newTranslations: Record<TranslationKey, string>) {
  Object.assign(translations, newTranslations)
}
