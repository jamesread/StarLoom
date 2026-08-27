import { createRouter, createWebHistory } from 'vue-router'
import {
  DashboardSquareSettingIcon,
  GiftIcon,
  Home01Icon,
  Settings01Icon,
  TaskDone01Icon,
  UserMultipleIcon,
  UserShield01Icon,
  WebhookIcon,
} from '@hugeicons/core-free-icons'
import { fetchAppStatus } from './composables/useStatus'
import {
  canAccessControlPanelFromStatus,
  canAccessIamFromStatus,
  canAccessSettingsFromStatus,
  canViewFamilyHomeFromStatus,
} from './lib/rbacAccess'
import { applyRouteBreadcrumbs } from './lib/routeBreadcrumbs'
import HomeView from './views/HomeView.vue'
import ControlPanel from './views/ControlPanel.vue'
import SettingsAdmin from './views/SettingsAdmin.vue'
import WebhooksAdmin from './views/WebhooksAdmin.vue'
import WebhookCreate from './views/WebhookCreate.vue'
import IamHub from './views/IamHub.vue'
import UsersAdmin from './views/UsersAdmin.vue'
import UserGroupsAdmin from './views/UserGroupsAdmin.vue'
import RbacRolesAdmin from './views/RbacRolesAdmin.vue'
import RbacPermissionsAdmin from './views/RbacPermissionsAdmin.vue'
import MyPermissions from './views/MyPermissions.vue'
import ApiKeysAdmin from './views/ApiKeysAdmin.vue'
import UserControlPanel from './views/UserControlPanel.vue'
import UserPreferences from './views/UserPreferences.vue'
import ChangePassword from './views/ChangePassword.vue'
import ChildrenAdmin from './views/ChildrenAdmin.vue'
import ChildDetail from './views/ChildDetail.vue'
import RewardsAdmin from './views/RewardsAdmin.vue'
import RedemptionsAdmin from './views/RedemptionsAdmin.vue'

const routes = [
  { path: '/', name: 'home', component: HomeView, meta: { title: 'Home', icon: Home01Icon, requiresAuth: true } },
  {
    path: '/family/children',
    name: 'familyChildren',
    component: ChildrenAdmin,
    meta: { title: 'Children', icon: UserMultipleIcon, requiresAuth: true, requiresFamilyAdmin: true },
  },
  {
    path: '/family/children/:id',
    name: 'familyChildDetail',
    component: ChildDetail,
    meta: { title: 'Child', requiresAuth: true, requiresFamilyAdmin: true },
  },
  {
    path: '/family/rewards',
    name: 'familyRewards',
    component: RewardsAdmin,
    meta: { title: 'Rewards', icon: GiftIcon, requiresAuth: true, requiresFamilyAdmin: true },
  },
  {
    path: '/family/redemptions',
    name: 'familyRedemptions',
    component: RedemptionsAdmin,
    meta: { title: 'Redemptions', icon: TaskDone01Icon, requiresAuth: true, requiresFamilyAdmin: true },
  },
  { path: '/user-control-panel', name: 'userControlPanel', component: UserControlPanel, meta: { title: 'User Control Panel', requiresAuth: true } },
  { path: '/user-control-panel/preferences', name: 'userPreferences', component: UserPreferences, meta: { title: 'User preferences', requiresAuth: true } },
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
  { path: '/user-groups', name: 'user-groups', component: UserGroupsAdmin, meta: { title: 'User groups', requiresAuth: true, requiresIam: true } },
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
    path: '/control-panel/webhooks/create',
    name: 'webhook-create',
    component: WebhookCreate,
    meta: { title: 'Add webhook', requiresAuth: true, requiresSettings: true },
  },
  { path: '/iam', redirect: { name: 'iam' } },
  { path: '/admin/settings', redirect: { name: 'settings' } },
  { path: '/admin/webhooks', redirect: { name: 'webhooks' } },
  { path: '/admin/webhooks/create', redirect: { name: 'webhook-create' } },
  { path: '/my-permissions', redirect: { name: 'myPermissions' } },
]

routes.forEach(applyRouteBreadcrumbs)

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to, _from, next) => {
  let st
  try {
    st = await fetchAppStatus()
  } catch {
    if (to.meta.requiresAuth) {
      next('/')
    } else {
      next()
    }
    return
  }
  if (to.meta.requiresAuth && !st?.isLoggedIn) {
    next('/')
    return
  }
  if (to.meta.requiresControlPanel && !canAccessControlPanelFromStatus(st)) {
    next('/')
    return
  }
  if (to.meta.requiresIam && !canAccessIamFromStatus(st)) {
    next('/')
    return
  }
  if (to.meta.requiresSettings && !canAccessSettingsFromStatus(st)) {
    next('/')
    return
  }
  if (to.meta.requiresFamilyAdmin && !canViewFamilyHomeFromStatus(st)) {
    next('/')
    return
  }
  next()
})

export default router
