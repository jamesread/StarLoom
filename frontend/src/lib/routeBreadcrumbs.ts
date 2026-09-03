import type { RouteRecordRaw } from 'vue-router'

export type BreadcrumbItem = {
  name: string
  href?: string
  to?: { name: string; params?: Record<string, string> }
}

export const crumb = {
  controlPanel: { name: 'Control Panel', href: '/control-panel', to: { name: 'controlPanel' } },
  iam: { name: 'IAM', href: '/control-panel/iam', to: { name: 'iam' } },
  users: { name: 'Users', href: '/users', to: { name: 'users' } },
  userGroups: { name: 'User groups', href: '/user-groups', to: { name: 'user-groups' } },
  rbac: { name: 'Roles', href: '/settings/rbac', to: { name: 'rbac-roles' } },
  permissions: { name: 'Permissions', href: '/settings/rbac/permissions', to: { name: 'rbac-permissions' } },
  settings: { name: 'Settings', href: '/control-panel/settings', to: { name: 'settings' } },
  webhooks: { name: 'Webhooks', href: '/control-panel/webhooks', to: { name: 'webhooks' } },
  notifications: { name: 'Notifications', href: '/control-panel/notifications', to: { name: 'notificationHistory' } },
  webhookCreate: { name: 'Add webhook', href: '/control-panel/webhooks/create', to: { name: 'webhook-create' } },
  people: { name: 'People', href: '/control-panel/people', to: { name: 'familyPeople' } },
  personCreate: { name: 'Add person', href: '/control-panel/people/create', to: { name: 'familyPersonCreate' } },
  chores: { name: 'Chores', href: '/control-panel/chores', to: { name: 'familyChores' } },
  starCharts: { name: 'Star Charts', href: '/control-panel/star-charts', to: { name: 'familyStarCharts' } },
  starChartCreate: { name: 'Add star chart', href: '/control-panel/star-charts/create', to: { name: 'familyStarChartCreate' } },
  rewards: { name: 'Rewards', href: '/control-panel/rewards', to: { name: 'familyRewards' } },
  rewardCreate: { name: 'Add reward', href: '/control-panel/rewards/create', to: { name: 'familyRewardCreate' } },
}

function trail(...items: BreadcrumbItem[]): BreadcrumbItem[] {
  return items
}

export function iamTrail(...tail: BreadcrumbItem[]): BreadcrumbItem[] {
  return trail(crumb.controlPanel, crumb.iam, ...tail)
}

export const breadcrumbsByRouteName: Record<string, () => BreadcrumbItem[]> = {
  controlPanel: () => trail(crumb.controlPanel),
  iam: () => iamTrail(),
  users: () => iamTrail(crumb.users),
  userInfo: () => iamTrail(crumb.users),
  'user-groups': () => iamTrail(crumb.userGroups),
  'rbac-roles': () => iamTrail(crumb.rbac),
  'rbac-permissions': () => iamTrail(crumb.permissions),
  settings: () => trail(crumb.controlPanel, crumb.settings),
  webhooks: () => trail(crumb.controlPanel, crumb.webhooks),
  notificationHistory: () => trail(crumb.controlPanel, crumb.notifications),
  'webhook-create': () => trail(crumb.controlPanel, crumb.webhooks, crumb.webhookCreate),
  familyPeople: () => trail(crumb.controlPanel, crumb.people),
  familyPersonCreate: () => trail(crumb.controlPanel, crumb.people, crumb.personCreate),
  familyPersonDetail: () => trail(crumb.controlPanel, crumb.people),
  familyPersonEdit: () => trail(crumb.controlPanel, crumb.people),
  familyChores: () => trail(crumb.controlPanel, crumb.chores),
  familyChoreEdit: () => trail(crumb.controlPanel, crumb.chores),
  familyStarCharts: () => trail(crumb.controlPanel, crumb.starCharts),
  familyStarChartCreate: () => trail(crumb.controlPanel, crumb.starCharts, crumb.starChartCreate),
  familyStarChartEdit: () => trail(crumb.controlPanel, crumb.starCharts),
  familyRewards: () => trail(crumb.controlPanel, crumb.rewards),
  familyRewardCreate: () => trail(crumb.controlPanel, crumb.rewards, crumb.rewardCreate),
  familyRewardEdit: () => trail(crumb.controlPanel, crumb.rewards),
}

export function applyRouteBreadcrumbs(route: RouteRecordRaw) {
  if (!route.meta) {
    route.meta = {}
  }
  const name = typeof route.name === 'string' ? route.name : ''
  const builder = breadcrumbsByRouteName[name]
  if (builder) {
    route.meta.breadcrumbs = builder
  }
}
