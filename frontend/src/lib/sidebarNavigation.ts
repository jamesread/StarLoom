export type SidebarNavigation = {
  clearNavigationLinks: () => void
  addSection: (title: string, options?: { name?: string }) => void
  addRouterLink: (link: string, altTitle?: string | null, options?: Record<string, unknown>) => void
}

export function setupSidebarNavigation(
  navigation: SidebarNavigation,
  { showControlPanel = false, showFamilyNav = false, showChoreChart = false } = {},
) {
  navigation.clearNavigationLinks()

  navigation.addRouterLink('home')

  if (showChoreChart && !showFamilyNav) {
    navigation.addRouterLink('familyStarChart', null, { description: 'Weekly star chart' })
  }

  if (showFamilyNav) {
    navigation.addSection('Family', { name: 'nav-family' })
    navigation.addRouterLink('familyPeople', null, { description: 'Manage people' })
    navigation.addRouterLink('familyStarChart', null, { description: 'Weekly star chart' })
    navigation.addRouterLink('familyStarCharts', null, { description: 'Manage star charts' })
    navigation.addRouterLink('familyChores', null, { description: 'Chore definitions' })
    navigation.addRouterLink('familyRewards', null, { description: 'Reward catalog' })
    navigation.addRouterLink('familyRedemptions', null, { description: 'Approval queue' })
  }

  if (showControlPanel) {
    navigation.addSection('Control Panel', { name: 'nav-control-panel' })
    navigation.addRouterLink('controlPanel', null, {
      description: 'System administration',
    })
  }
}
