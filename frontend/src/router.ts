import { createRouter, createWebHistory } from 'vue-router'
import {
  DashboardSquareSettingIcon,
  GiftIcon,
  Home01Icon,
  Settings01Icon,
  StarIcon,
  Task01Icon,
  UserMultipleIcon,
  UserShield01Icon,
  WebhookIcon,
  Notification03Icon,
} from '@hugeicons/core-free-icons'
import { fetchAppStatus } from './composables/useStatus'
import {
  canAccessControlPanelFromStatus,
  canAccessIamFromStatus,
  canAccessSettingsFromStatus,
  canViewChoresFromStatus,
  canViewFamilyHomeFromStatus,
} from './lib/rbacAccess'
import { applyRouteBreadcrumbs } from './lib/routeBreadcrumbs'
import HomeView from './views/HomeView.vue'
import ControlPanel from './views/ControlPanel.vue'
import SettingsAdmin from './views/SettingsAdmin.vue'
import WebhooksAdmin from './views/WebhooksAdmin.vue'
import WebhookCreate from './views/WebhookCreate.vue'
import WebhookEdit from './views/WebhookEdit.vue'
import IamHub from './views/IamHub.vue'
import UsersAdmin from './views/UsersAdmin.vue'
import UserInfoAdmin from './views/UserInfoAdmin.vue'
import UserGroupsAdmin from './views/UserGroupsAdmin.vue'
import UserGroupEdit from './views/UserGroupEdit.vue'
import RbacRolesAdmin from './views/RbacRolesAdmin.vue'
import RbacPermissionsAdmin from './views/RbacPermissionsAdmin.vue'
import MyPermissions from './views/MyPermissions.vue'
import ApiKeysAdmin from './views/ApiKeysAdmin.vue'
import UserControlPanel from './views/UserControlPanel.vue'
import UserPreferences from './views/UserPreferences.vue'
import ChoreNotificationSubscriptions from './views/ChoreNotificationSubscriptions.vue'
import NotificationHistoryAdmin from './views/NotificationHistoryAdmin.vue'
import ChangePassword from './views/ChangePassword.vue'
import ChildrenAdmin from './views/ChildrenAdmin.vue'
import ChildCreate from './views/ChildCreate.vue'
import ChildDetail from './views/ChildDetail.vue'
import ChildEdit from './views/ChildEdit.vue'
import PersonNotificationSubscriptions from './views/PersonNotificationSubscriptions.vue'
import RewardsAdmin from './views/RewardsAdmin.vue'
import RewardCreate from './views/RewardCreate.vue'
import RewardEdit from './views/RewardEdit.vue'
import ChoresAdmin from './views/ChoresAdmin.vue'
import ChoreEdit from './views/ChoreEdit.vue'
import StarChart from './views/StarChart.vue'
import StarChartRedirect from './views/StarChartRedirect.vue'
import StarChartsAdmin from './views/StarChartsAdmin.vue'
import StarChartCreate from './views/StarChartCreate.vue'
import StarChartEdit from './views/StarChartEdit.vue'

const routes = [
  { path: '/', name: 'home', component: HomeView, meta: { title: 'Home', icon: Home01Icon, requiresAuth: true } },
  {
    path: '/control-panel/people',
    name: 'familyPeople',
    component: ChildrenAdmin,
    meta: { title: 'People', icon: UserMultipleIcon, requiresAuth: true, requiresFamilyAdmin: true },
  },
  {
    path: '/control-panel/people/create',
    name: 'familyPersonCreate',
    component: ChildCreate,
    meta: { title: 'Add person', requiresAuth: true, requiresFamilyAdmin: true },
  },
  {
    path: '/control-panel/people/:id',
    name: 'familyPersonDetail',
    component: ChildDetail,
    meta: { title: 'Person', requiresAuth: true, requiresFamilyAdmin: true },
  },
  {
    path: '/control-panel/people/:id/edit',
    name: 'familyPersonEdit',
    component: ChildEdit,
    meta: { title: 'Edit person', requiresAuth: true, requiresFamilyAdmin: true },
  },
  {
    path: '/control-panel/people/:id/notifications',
    name: 'familyPersonNotifications',
    component: PersonNotificationSubscriptions,
    meta: { title: 'Chore notifications', requiresAuth: true, requiresFamilyAdmin: true },
  },
  { path: '/family/people', redirect: { name: 'familyPeople' } },
  { path: '/family/people/create', redirect: { name: 'familyPersonCreate' } },
  { path: '/family/people/:id', redirect: (to) => ({ name: 'familyPersonDetail', params: { id: to.params.id } }) },
  {
    path: '/family/people/:id/edit',
    redirect: (to) => ({ name: 'familyPersonEdit', params: { id: to.params.id } }),
  },
  { path: '/family/children', redirect: { name: 'familyPeople' } },
  { path: '/family/children/create', redirect: { name: 'familyPersonCreate' } },
  { path: '/family/children/:id', redirect: (to) => ({ name: 'familyPersonDetail', params: { id: to.params.id } }) },
  { path: '/family/children/:id/edit', redirect: (to) => ({ name: 'familyPersonEdit', params: { id: to.params.id } }) },
  {
    path: '/control-panel/rewards',
    name: 'familyRewards',
    component: RewardsAdmin,
    meta: { title: 'Rewards', icon: GiftIcon, requiresAuth: true, requiresFamilyAdmin: true },
  },
  {
    path: '/control-panel/rewards/create',
    name: 'familyRewardCreate',
    component: RewardCreate,
    meta: { title: 'Add reward', requiresAuth: true, requiresFamilyAdmin: true },
  },
  {
    path: '/control-panel/rewards/:id',
    name: 'familyRewardEdit',
    component: RewardEdit,
    meta: { title: 'Edit reward', requiresAuth: true, requiresFamilyAdmin: true },
  },
  { path: '/family/rewards', redirect: { name: 'familyRewards' } },
  { path: '/family/rewards/create', redirect: { name: 'familyRewardCreate' } },
  {
    path: '/family/rewards/:id',
    redirect: (to) => ({ name: 'familyRewardEdit', params: { id: to.params.id } }),
  },
  {
    path: '/family/redemptions',
    redirect: { name: 'familyRewards' },
  },
  {
    path: '/control-panel/chores',
    name: 'familyChores',
    component: ChoresAdmin,
    meta: { title: 'Chores', icon: Task01Icon, requiresAuth: true, requiresFamilyAdmin: true },
  },
  {
    path: '/control-panel/chores/:id',
    name: 'familyChoreEdit',
    component: ChoreEdit,
    meta: { title: 'Edit chore', requiresAuth: true, requiresFamilyAdmin: true },
  },
  { path: '/family/chores', redirect: { name: 'familyChores' } },
  {
    path: '/family/chores/:id',
    redirect: (to) => ({ name: 'familyChoreEdit', params: { id: to.params.id } }),
  },
  {
    path: '/control-panel/star-charts',
    name: 'familyStarCharts',
    component: StarChartsAdmin,
    meta: { title: 'Star Charts', icon: StarIcon, requiresAuth: true, requiresFamilyAdmin: true },
  },
  {
    path: '/control-panel/star-charts/create',
    name: 'familyStarChartCreate',
    component: StarChartCreate,
    meta: { title: 'Add star chart', requiresAuth: true, requiresFamilyAdmin: true },
  },
  {
    path: '/control-panel/star-charts/:id/edit',
    name: 'familyStarChartEdit',
    component: StarChartEdit,
    meta: { title: 'Edit star chart', requiresAuth: true, requiresFamilyAdmin: true },
  },
  { path: '/family/star-charts', redirect: { name: 'familyStarCharts' } },
  { path: '/family/star-charts/create', redirect: { name: 'familyStarChartCreate' } },
  {
    path: '/family/star-charts/:id/edit',
    redirect: (to) => ({ name: 'familyStarChartEdit', params: { id: to.params.id } }),
  },
  {
    path: '/family/star-chart',
    name: 'familyStarChart',
    component: StarChartRedirect,
    meta: { title: 'Star Chart', icon: StarIcon, requiresAuth: true, requiresChoresView: true },
  },
  {
    path: '/family/star-chart/:id',
    name: 'familyStarChartView',
    component: StarChart,
    meta: { title: 'Star Chart', icon: StarIcon, requiresAuth: true, requiresChoresView: true },
  },
  { path: '/user-control-panel', name: 'userControlPanel', component: UserControlPanel, meta: { title: 'User Control Panel', requiresAuth: true } },
  { path: '/user-control-panel/preferences', name: 'userPreferences', component: UserPreferences, meta: { title: 'User preferences', requiresAuth: true } },
  { path: '/user-control-panel/chore-notifications', name: 'choreNotifications', component: ChoreNotificationSubscriptions, meta: { title: 'Chore notifications', requiresAuth: true } },
  { path: '/user-control-panel/permissions', name: 'myPermissions', component: MyPermissions, meta: { title: 'My Permissions', requiresAuth: true } },
  { path: '/change-password', name: 'changePassword', component: ChangePassword, meta: { title: 'Change password', requiresAuth: true } },
  { path: '/api-keys', name: 'apiKeys', component: ApiKeysAdmin, meta: { title: 'API keys', requiresAuth: true } },
  {
    path: '/control-panel',
    name: 'controlPanel',
    component: ControlPanel,
    meta: { title: 'Control Panel', icon: DashboardSquareSettingIcon, requiresAuth: true, requiresControlPanel: true },
  },
  {
    path: '/control-panel/iam',
    name: 'iam',
    component: IamHub,
    meta: { title: 'IAM', icon: UserShield01Icon, requiresAuth: true, requiresIam: true },
  },
  { path: '/users', name: 'users', component: UsersAdmin, meta: { title: 'Users', requiresAuth: true, requiresIam: true } },
  { path: '/users/:id', name: 'userInfo', component: UserInfoAdmin, meta: { title: 'User', requiresAuth: true, requiresIam: true } },
  { path: '/user-groups', name: 'user-groups', component: UserGroupsAdmin, meta: { title: 'User groups', requiresAuth: true, requiresIam: true } },
  { path: '/user-groups/:id', name: 'userGroupEdit', component: UserGroupEdit, meta: { title: 'User group', requiresAuth: true, requiresIam: true } },
  { path: '/settings/rbac', name: 'rbac-roles', component: RbacRolesAdmin, meta: { title: 'Roles', requiresAuth: true, requiresIam: true } },
  {
    path: '/settings/rbac/permissions',
    name: 'rbac-permissions',
    component: RbacPermissionsAdmin,
    meta: { title: 'Permissions', requiresAuth: true, requiresIam: true },
  },
  {
    path: '/control-panel/settings',
    name: 'settings',
    component: SettingsAdmin,
    meta: { title: 'Settings', icon: Settings01Icon, requiresAuth: true, requiresSettings: true },
  },
  {
    path: '/control-panel/webhooks',
    name: 'webhooks',
    component: WebhooksAdmin,
    meta: { title: 'Webhooks', icon: WebhookIcon, requiresAuth: true, requiresSettings: true },
  },
  {
    path: '/control-panel/notifications',
    name: 'notificationHistory',
    component: NotificationHistoryAdmin,
    meta: { title: 'Notification history', icon: Notification03Icon, requiresAuth: true, requiresSettings: true },
  },
  {
    path: '/control-panel/webhooks/create',
    name: 'webhook-create',
    component: WebhookCreate,
    meta: { title: 'Add webhook', requiresAuth: true, requiresSettings: true },
  },
  {
    path: '/control-panel/webhooks/:id',
    name: 'webhook-edit',
    component: WebhookEdit,
    meta: { title: 'Edit webhook', requiresAuth: true, requiresSettings: true },
  },
  { path: '/iam', redirect: { name: 'iam' } },
  { path: '/admin/settings', redirect: { name: 'settings' } },
  { path: '/admin/webhooks', redirect: { name: 'webhooks' } },
  { path: '/admin/webhooks/create', redirect: { name: 'webhook-create' } },
  { path: '/admin/webhooks/:id', redirect: (to) => ({ name: 'webhook-edit', params: { id: to.params.id } }) },
  { path: '/my-permissions', redirect: { name: 'myPermissions' } },
]

routes.forEach(applyRouteBreadcrumbs)

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  let st
  try {
    st = await fetchAppStatus()
  } catch {
    if (to.meta.requiresAuth && to.name !== 'home') {
      return { name: 'home' }
    }
    return
  }
  if (to.meta.requiresAuth && !st?.isLoggedIn) {
    if (to.name !== 'home') {
      return { name: 'home' }
    }
    return
  }
  if (to.meta.requiresControlPanel && !canAccessControlPanelFromStatus(st)) {
    return '/'
  }
  if (to.meta.requiresIam && !canAccessIamFromStatus(st)) {
    return '/'
  }
  if (to.meta.requiresSettings && !canAccessSettingsFromStatus(st)) {
    return '/'
  }
  if (to.meta.requiresFamilyAdmin && !canViewFamilyHomeFromStatus(st)) {
    return '/'
  }
  if (to.meta.requiresChoresView && !canViewChoresFromStatus(st)) {
    return '/'
  }
})

export default router
