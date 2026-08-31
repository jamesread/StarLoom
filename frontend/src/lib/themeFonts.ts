const themeTypographyStyles: Record<string, () => Promise<unknown>> = {
  'ancient-greece': () => import('../themes/ancient-greece-fonts.css'),
}

let loadedThemeTypography = ''

async function waitForGreeKish() {
  if (typeof document === 'undefined' || !document.fonts?.load) {
    return
  }
  try {
    await document.fonts.load("700 1em 'GreeKish'")
    await document.fonts.ready
  } catch {
    // Best-effort; section titles fall back to Georgia if load fails.
  }
}

/** Load bundled unlayered theme typography when a theme defines one. */
export async function syncThemeFontStylesheet(themeName: string) {
  const trimmed = themeName.trim()
  if (trimmed === loadedThemeTypography) {
    return
  }

  const load = trimmed ? themeTypographyStyles[trimmed] : undefined
  if (load) {
    await load()
    if (trimmed === 'ancient-greece') {
      await waitForGreeKish()
    }
  }

  loadedThemeTypography = trimmed
}
