import { starapp } from '../api/client'

let sidebarApplier: ((enabled: boolean) => void) | null = null
let themeToggleApplier: ((enabled: boolean) => void) | null = null

export function registerSidebarApplier(fn: (enabled: boolean) => void) {
  sidebarApplier = fn
}

export function registerThemeToggleApplier(fn: (enabled: boolean) => void) {
  themeToggleApplier = fn
}

export function applyUserLanguage(language: string | undefined) {
  if (typeof document === 'undefined') return
  const next = (language || '').trim()
  if (next) {
    document.documentElement.lang = next
  }
}

export function applyUserSidebar(enabled: boolean) {
  sidebarApplier?.(Boolean(enabled))
}

export function applyUserThemeToggle(enabled: boolean) {
  themeToggleApplier?.(enabled === true)
}

export async function loadAndApplyUserPreferences() {
  try {
    const res = await starapp.getUserPreferences()
    applyUserLanguage(res.language)
    applyUserSidebar(res.sidebarEnabled)
    applyUserThemeToggle(res.themeToggleEnabled)
    return res
  } catch (e) {
    console.warn('Failed to load user preferences', e)
    return null
  }
}

export function languageLabel(code: string): string {
  switch (code) {
    case 'en':
      return 'English'
    default:
      return code
  }
}
