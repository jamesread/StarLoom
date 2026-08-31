const CONTROL_PANEL_PERMISSIONS = [
  'users.view',
  'usergroups.view',
  'rbac.view',
  'system.settings',
]

export const IAM_PERMISSIONS = ['users.view', 'usergroups.view', 'rbac.view']

export type StatusLike = {
  isLoggedIn?: boolean
  rbacIsSuperuser?: boolean
  rbacPermissions?: readonly string[]
}

export function hasPermission(st: StatusLike | null | undefined, name: string): boolean {
  return Boolean(st?.rbacIsSuperuser) || (st?.rbacPermissions || []).includes(name)
}

export function canAccessControlPanelFromStatus(st: StatusLike | null | undefined): boolean {
  if (!st?.isLoggedIn) return false
  if (st.rbacIsSuperuser) return true
  if (
    hasPermission(st, 'stars.view_family') ||
    hasPermission(st, 'family.manage') ||
    hasPermission(st, 'members.manage')
  ) {
    return true
  }
  return CONTROL_PANEL_PERMISSIONS.some((p) => (st.rbacPermissions || []).includes(p))
}

export function canAccessIamFromStatus(st: StatusLike | null | undefined): boolean {
  if (!st?.isLoggedIn) return false
  if (st.rbacIsSuperuser) return true
  return IAM_PERMISSIONS.some((p) => (st.rbacPermissions || []).includes(p))
}

export function canAccessSettingsFromStatus(st: StatusLike | null | undefined): boolean {
  if (!st?.isLoggedIn) return false
  if (st.rbacIsSuperuser) return true
  return (st.rbacPermissions || []).includes('system.settings')
}

export function canAccessWebhooksFromStatus(st: StatusLike | null | undefined): boolean {
  return canAccessSettingsFromStatus(st)
}

export function canViewFamilyHomeFromStatus(st: StatusLike | null | undefined): boolean {
  if (!st?.isLoggedIn) return false
  return (
    hasPermission(st, 'stars.view_family') ||
    hasPermission(st, 'family.manage') ||
    hasPermission(st, 'members.manage')
  )
}

export function canViewChildHomeFromStatus(st: StatusLike | null | undefined): boolean {
  if (!st?.isLoggedIn) return false
  if (canViewFamilyHomeFromStatus(st)) return false
  return hasPermission(st, 'stars.view_own')
}

export function canManageFamilyFromStatus(st: StatusLike | null | undefined): boolean {
  return hasPermission(st, 'members.manage')
}

export function canManageRewardsFromStatus(st: StatusLike | null | undefined): boolean {
  return hasPermission(st, 'rewards.manage')
}

export function canApproveRedemptionsFromStatus(st: StatusLike | null | undefined): boolean {
  return hasPermission(st, 'redemptions.approve')
}

export function canManageChoresFromStatus(st: StatusLike | null | undefined): boolean {
  return hasPermission(st, 'chores.manage')
}

export function canViewChoresFromStatus(st: StatusLike | null | undefined): boolean {
  if (!st?.isLoggedIn) return false
  return hasPermission(st, 'chores.view_family') || canViewFamilyHomeFromStatus(st)
}

export function canCompleteChoresFromStatus(st: StatusLike | null | undefined): boolean {
  return hasPermission(st, 'chores.complete')
}
