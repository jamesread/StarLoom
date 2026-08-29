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
  }: {
    showControlPanel?: boolean
    showFamilyNav?: boolean
    excludeRoutes?: string[]
  } = {},
) {
  const excluded = new Set(excludeRoutes)

  navigation.clearNavigationLinks()

  navigation.addRouterLink('home')

  if (showFamilyNav) {
    navigation.addSection('Family', { name: 'nav-family' })
    if (!excluded.has('familyPeople')) {
      navigation.addRouterLink('familyPeople', null, { description: 'Manage people' })
    }
    if (!excluded.has('familyStarCharts')) {
      navigation.addRouterLink('familyStarCharts', null, { description: 'Manage star charts' })
    }
    if (!excluded.has('familyChores')) {
      navigation.addRouterLink('familyChores', null, { description: 'Chore definitions' })
    }
    if (!excluded.has('familyRewards')) {
      navigation.addRouterLink('familyRewards', null, { description: 'Reward catalog' })
    }
    if (!excluded.has('familyRedemptions')) {
      navigation.addRouterLink('familyRedemptions', null, { description: 'Approval queue' })
    }
  }

  if (showControlPanel && !excluded.has('controlPanel')) {
    navigation.addSection('Control Panel', { name: 'nav-control-panel' })
    navigation.addRouterLink('controlPanel', null, {
      description: 'System administration',
    })
  }
}
