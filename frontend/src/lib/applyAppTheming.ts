import type { Features, GetStatusResponse, InitResponse } from '../api/client'
import { useCustomTheme } from 'picocrank/vue/composables/useCustomTheme.js'

export function themeControlFromFeatures(features: Features | undefined) {
  const raw = (features?.themeControl || 'user').trim()
  return raw === 'system' ? 'system' : 'user'
}

export function featuresFromShell(
  init: InitResponse | null | undefined,
  status: GetStatusResponse | null | undefined,
): Features {
  return init?.features ?? status?.features ?? {}
}

export function themeColorSchemeSwitcherEnabledFromFeatures(features: Features | undefined) {
  return features?.themeColorSchemeSwitcherEnabled === true
}

export function applyAppTheming(features: Features | undefined) {
  const { themePreference, setTheme, clearTheme } = useCustomTheme()

  const control = themeControlFromFeatures(features)
  const systemTheme = (features?.themeName || '').trim()

  if (control === 'system') {
    if (systemTheme) {
      setTheme(systemTheme)
    } else {
      clearTheme()
    }
    return { control, appliedTheme: systemTheme }
  }

  const stored = (themePreference.value || '').trim()
  if (stored) {
    setTheme(stored)
    return { control, appliedTheme: stored }
  }
  if (systemTheme) {
    setTheme(systemTheme)
    return { control, appliedTheme: systemTheme }
  }
  clearTheme()
  return { control, appliedTheme: '' }
}
