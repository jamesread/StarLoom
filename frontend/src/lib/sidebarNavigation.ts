export type SidebarNavigation = {
  clearNavigationLinks: () => void
  addSection: (title: string, options?: { name?: string }) => void
  addRouterLink: (link: string, altTitle?: string | null, options?: Record<string, unknown>) => void
}

export type TopBarNavigation = SidebarNavigation & {
  addCallback: (title: string, callback: () => void, options?: Record<string, unknown>) => void
}

export type TopBarStarChart = {
  id: number
  name: string
  choreCount?: number
}

export function appendTopBarStarChartLinks(
  navigation: TopBarNavigation,
  charts: TopBarStarChart[],
  onOpen: (chartId: number) => void,
  icon: unknown,
) {
  for (const chart of charts) {
    const count = chart.choreCount ?? 0
    navigation.addCallback(chart.name, () => onOpen(chart.id), {
      name: `topbar-star-chart-${chart.id}`,
      icon,
      description: count === 1 ? '1 chore for you' : `${count} chores for you`,
    })
  }
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
