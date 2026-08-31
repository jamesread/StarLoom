export type SidebarNavigation = {
  clearNavigationLinks: () => void
  addSection: (title: string, options?: { name?: string }) => void
  addRouterLink: (link: string, altTitle?: string | null, options?: Record<string, unknown>) => void
}

export function setupSidebarNavigation(
  navigation: SidebarNavigation,
  {
    showControlPanel = false,
    showFamilyNav = false,
    excludeRoutes = [],
    flat = false,
  }: {
    showControlPanel?: boolean
    showFamilyNav?: boolean
    excludeRoutes?: string[]
    /** Omit section headers — required for PicoCrank TopBar, which cannot skip section rows. */
    flat?: boolean
  } = {},
) {
  const excluded = new Set(excludeRoutes)

  navigation.clearNavigationLinks()

  navigation.addRouterLink('home')

  if (showFamilyNav) {
    if (!flat) {
      navigation.addSection('Family', { name: 'nav-family' })
    }
    if (!excluded.has('familyStarCharts')) {
      navigation.addRouterLink('familyStarCharts', null, { description: 'Manage star charts' })
    }
    if (!excluded.has('familyRewards')) {
      navigation.addRouterLink('familyRewards', null, { description: 'Reward catalog' })
    }
  }

  if (showControlPanel && !excluded.has('controlPanel')) {
    if (!flat) {
      navigation.addSection('Control Panel', { name: 'nav-control-panel' })
    }
    navigation.addRouterLink('controlPanel', null, {
      description: 'System administration',
    })
  }
}
