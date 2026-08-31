const base = '/api'

export type Features = {
  redemptionApprovalDefault?: boolean
  themeColorSchemeSwitcherEnabled?: boolean
  themeName?: string
  themeControl?: string
  availableThemes?: string[]
}

export type StandardResponse = {
  success?: boolean
  message?: string
}

export type GetStatusResponse = {
  username?: string
  isLoggedIn?: boolean
  rbacPermissions?: string[]
  rbacIsSuperuser?: boolean
  isImpersonating?: boolean
  impersonatorUsername?: string
  showFooter?: boolean
  showNewVersions?: boolean
  availableVersion?: string
  currentVersion?: string
  pageTitle?: string
  showVersionNumber?: boolean
  siteTitle?: string
  features?: Features
  webhookEvents?: string[]
  usesSecureCookies?: boolean
  accountCreatedAt?: string
}

export type InitResponse = GetStatusResponse

export type UserAccount = {
  id: number
  username: string
  createdAt?: string
  createdBy?: string
}

export type RbacPermission = {
  id: number
  name: string
  description?: string
}

export type RbacRole = {
  id: number
  name: string
  description?: string
  permissionIds?: number[]
  groupCount?: number
  userCount?: number
}

export type UserGroup = {
  id: number
  name: string
  memberCount?: number
}

export type Webhook = {
  id: number
  url: string
  events?: string[]
  enabled?: boolean
  created?: string
  updated?: string
}

export type WebhookDelivery = {
  id: number
  webhookTargetId?: number
  event: string
  url: string
  success?: boolean
  httpStatus?: number
  errorMessage?: string
  firedAt?: string
}

export type Cvar = {
  key: string
  mainType: string
  valueInt?: number
  valueString?: string
  title?: string
  description?: string
  category?: string
  ordinal?: number
}

export type UserPreferences = {
  language: string
  availableLanguages: string[]
  sidebarEnabled: boolean
}

function normalizeUserPreferences(res: {
  language?: string
  availableLanguages?: string[]
  sidebarEnabled?: boolean
}): UserPreferences {
  return {
    language: res.language ?? '',
    availableLanguages: res.availableLanguages ?? [],
    sidebarEnabled: typeof res.sidebarEnabled === 'boolean' ? res.sidebarEnabled : true,
  }
}

export type ApiKey = {
  id: number
  name: string
  createdAt?: string
  lastUsedAt?: string
  readOnly?: boolean
}

export type MyPermissionAuditRow = {
  permission: string
  granted?: boolean
  grantingGroups?: string[]
}

async function connectFetch<T>(
  procedure: string,
  request: Record<string, unknown> = {},
): Promise<T> {
  const res = await fetch(base + procedure, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(request),
  })
  if (!res.ok) {
    throw new Error(res.statusText || `Request failed: ${res.status}`)
  }
  return res.json() as Promise<T>
}

export const starapp = {
  getStatus() {
    return connectFetch<GetStatusResponse>('/starapp.api.v1.StarAppService/GetStatus', {})
  },
  init() {
    return connectFetch<InitResponse>('/starapp.api.v1.StarAppService/Init', {})
  },
  loginWithUsernameAndPassword(body: { username: string; password: string }) {
    return connectFetch<{ standardResponse?: StandardResponse; username?: string }>(
      '/starapp.api.v1.StarAppService/LoginWithUsernameAndPassword',
      body,
    )
  },
  logout() {
    return connectFetch<{ standardResponse?: StandardResponse }>(
      '/starapp.api.v1.StarAppService/Logout',
      {},
    )
  },
  changePassword(body: { currentPassword: string; newPassword: string }) {
    return connectFetch<{ standardResponse?: StandardResponse }>(
      '/starapp.api.v1.StarAppService/ChangePassword',
      body,
    )
  },
  getUserPreferences() {
    return connectFetch<{
      language?: string
      availableLanguages?: string[]
      sidebarEnabled?: boolean
    }>('/starapp.api.v1.StarAppService/GetUserPreferences', {}).then(normalizeUserPreferences)
  },
  saveUserPreferences(body: {
    language: string
    sidebarEnabled: boolean
  }) {
    return connectFetch<{ standardResponse?: StandardResponse; username?: string }>(
      '/starapp.api.v1.StarAppService/SaveUserPreferences',
      body,
    )
  },
  getUsers() {
    return connectFetch<{ users: UserAccount[] }>('/starapp.api.v1.StarAppService/GetUsers', {})
  },
  createUser(body: { username: string; password?: string }) {
    return connectFetch<{ standardResponse?: StandardResponse; user?: UserAccount }>(
      '/starapp.api.v1.StarAppService/CreateUser',
      body,
    )
  },
  deleteUser(body: { userId: number }) {
    return connectFetch<{ standardResponse?: StandardResponse }>(
      '/starapp.api.v1.StarAppService/DeleteUser',
      body,
    )
  },
  listUserGroups() {
    return connectFetch<{ groups: UserGroup[] }>(
      '/starapp.api.v1.StarAppService/ListUserGroups',
      {},
    )
  },
  getUserGroupMembers(body: { groupId: number }) {
    return connectFetch<{ members: UserAccount[] }>(
      '/starapp.api.v1.StarAppService/GetUserGroupMembers',
      body,
    )
  },
  setUserGroupMembers(body: { groupId: number; userIds: number[] }) {
    return connectFetch<{ standardResponse?: StandardResponse }>(
      '/starapp.api.v1.StarAppService/SetUserGroupMembers',
      body,
    )
  },
  getUserGroupRbacRoles(body: { groupId: number }) {
    return connectFetch<{ roleIds: number[] }>(
      '/starapp.api.v1.StarAppService/GetUserGroupRbacRoles',
      body,
    )
  },
  setUserGroupRbacRoles(body: { groupId: number; roleIds: number[] }) {
    return connectFetch<{ standardResponse?: StandardResponse }>(
      '/starapp.api.v1.StarAppService/SetUserGroupRbacRoles',
      body,
    )
  },
  listRbacPermissions() {
    return connectFetch<{ permissions: RbacPermission[] }>(
      '/starapp.api.v1.StarAppService/ListRbacPermissions',
      {},
    )
  },
  listRbacRoles() {
    return connectFetch<{ roles: RbacRole[] }>(
      '/starapp.api.v1.StarAppService/ListRbacRoles',
      {},
    )
  },
  createRbacRole(body: { name: string; description?: string; permissionIds?: number[] }) {
    return connectFetch<{ role: RbacRole }>(
      '/starapp.api.v1.StarAppService/CreateRbacRole',
      body,
    )
  },
  updateRbacRole(body: {
    id: number
    name: string
    description?: string
    permissionIds?: number[]
  }) {
    return connectFetch<RbacRole>('/starapp.api.v1.StarAppService/UpdateRbacRole', body)
  },
  deleteRbacRole(body: { id: number }) {
    return connectFetch<object>('/starapp.api.v1.StarAppService/DeleteRbacRole', body)
  },
  getMyPermissionsAudit() {
    return connectFetch<{
      groupNames?: string[]
      roleNames?: string[]
      isSuperuser?: boolean
      permissions?: MyPermissionAuditRow[]
    }>('/starapp.api.v1.StarAppService/GetMyPermissionsAudit', {})
  },
  listApiKeys() {
    return connectFetch<{ keys: ApiKey[] }>('/starapp.api.v1.StarAppService/ListApiKeys', {})
  },
  createApiKey(body: { name: string; readOnly?: boolean }) {
    return connectFetch<{ key: ApiKey; secret?: string }>(
      '/starapp.api.v1.StarAppService/CreateApiKey',
      body,
    )
  },
  deleteApiKey(body: { id: number }) {
    return connectFetch<object>('/starapp.api.v1.StarAppService/DeleteApiKey', body)
  },
  listCvars() {
    return connectFetch<{ cvars: Cvar[] }>('/starapp.api.v1.StarAppService/ListCvars', {})
  },
  updateCvar(body: { key: string; valueInt?: number; valueString?: string }) {
    return connectFetch<Cvar>('/starapp.api.v1.StarAppService/UpdateCvar', body)
  },
  listWebhooks() {
    return connectFetch<{ webhooks: Webhook[]; events: string[] }>(
      '/starapp.api.v1.StarAppService/ListWebhooks',
      {},
    )
  },
  createWebhook(body: {
    url: string
    secret: string
    events: string[]
    enabled: boolean
  }) {
    return connectFetch<{ webhook: Webhook }>(
      '/starapp.api.v1.StarAppService/CreateWebhook',
      body,
    )
  },
  updateWebhook(body: {
    id: number
    url: string
    secret?: string
    events: string[]
    enabled: boolean
  }) {
    return connectFetch<Webhook>('/starapp.api.v1.StarAppService/UpdateWebhook', body)
  },
  deleteWebhook(body: { id: number }) {
    return connectFetch<object>('/starapp.api.v1.StarAppService/DeleteWebhook', body)
  },
  listWebhookDeliveries(body: { limit?: number } = {}) {
    return connectFetch<{ deliveries: WebhookDelivery[] }>(
      '/starapp.api.v1.StarAppService/ListWebhookDeliveries',
      body,
    )
  },
  fireTestWebhooks() {
    return connectFetch<{ standardResponse?: StandardResponse; targetsFired?: number }>(
      '/starapp.api.v1.StarAppService/FireTestWebhooks',
      {},
    )
  },
  getMyFamily() {
    return connectFetch<{ family?: Family; callerMember?: FamilyMember }>(
      '/starapp.api.v1.StarAppService/GetMyFamily',
      {},
    )
  },
  createFamily(body: { name: string }) {
    return connectFetch<{ standardResponse?: StandardResponse; family?: Family; callerMember?: FamilyMember }>(
      '/starapp.api.v1.StarAppService/CreateFamily',
      body,
    )
  },
  listMembers() {
    return connectFetch<{ members: FamilyMember[] }>(
      '/starapp.api.v1.StarAppService/ListMembers',
      {},
    )
  },
  createChildMember(body: { displayName: string; username?: string; password?: string }) {
    return connectFetch<{ standardResponse?: StandardResponse; member?: FamilyMember }>(
      '/starapp.api.v1.StarAppService/CreateChildMember',
      body,
    )
  },
  updateMember(body: { memberId: number; displayName: string; starColor?: string }) {
    return connectFetch<{ standardResponse?: StandardResponse; member?: FamilyMember }>(
      '/starapp.api.v1.StarAppService/UpdateMember',
      body,
    )
  },
  assignMemberLogin(body: { memberId: number; username: string; password: string }) {
    return connectFetch<{ standardResponse?: StandardResponse; member?: FamilyMember }>(
      '/starapp.api.v1.StarAppService/AssignMemberLogin',
      body,
    )
  },
  deleteMember(body: { memberId: number }) {
    return connectFetch<{ standardResponse?: StandardResponse }>(
      '/starapp.api.v1.StarAppService/DeleteMember',
      body,
    )
  },
  async uploadMemberAvatar(body: { memberId: number; file: File }) {
    const data = await fileToBase64(body.file)
    const contentType = body.file.type || inferImageContentType(body.file.name)
    return connectFetch<{ standardResponse?: StandardResponse; member?: FamilyMember }>(
      '/starapp.api.v1.StarAppService/UploadMemberAvatar',
      { memberId: body.memberId, data, contentType },
    )
  },
  deleteMemberAvatar(body: { memberId: number }) {
    return connectFetch<{ standardResponse?: StandardResponse; member?: FamilyMember }>(
      '/starapp.api.v1.StarAppService/DeleteMemberAvatar',
      body,
    )
  },
  listMemberAvatars(body: { memberId: number }) {
    return connectFetch<{ avatars: MemberAvatarEntry[] }>(
      '/starapp.api.v1.StarAppService/ListMemberAvatars',
      body,
    )
  },
  selectMemberAvatar(body: { memberId: number; filename: string }) {
    return connectFetch<{ standardResponse?: StandardResponse; member?: FamilyMember }>(
      '/starapp.api.v1.StarAppService/SelectMemberAvatar',
      body,
    )
  },
  awardStars(body: { childMemberId: number; amount?: number; note?: string }) {
    return connectFetch<{ standardResponse?: StandardResponse; entry?: StarLedgerEntry; newBalance?: number }>(
      '/starapp.api.v1.StarAppService/AwardStars',
      body,
    )
  },
  revokeStars(body: { childMemberId: number; amount: number; note?: string }) {
    return connectFetch<{ standardResponse?: StandardResponse; entry?: StarLedgerEntry; newBalance?: number }>(
      '/starapp.api.v1.StarAppService/RevokeStars',
      body,
    )
  },
  getMemberBalance(body: { memberId: number }) {
    return connectFetch<{ balance: number }>(
      '/starapp.api.v1.StarAppService/GetMemberBalance',
      body,
    )
  },
  listLedger(body: { memberId: number; limit?: number }) {
    return connectFetch<{ entries: StarLedgerEntry[] }>(
      '/starapp.api.v1.StarAppService/ListLedger',
      body,
    )
  },
  listRewards(body: { includeInactive?: boolean } = {}) {
    return connectFetch<{ rewards: Reward[] }>(
      '/starapp.api.v1.StarAppService/ListRewards',
      body,
    )
  },
  createReward(body: {
    title: string
    description?: string
    costStars: number
    approvalRequired?: boolean
    availabilityExpression?: string
  }) {
    return connectFetch<{ standardResponse?: StandardResponse; reward?: Reward }>(
      '/starapp.api.v1.StarAppService/CreateReward',
      body,
    )
  },
  updateReward(body: {
    id: number
    title: string
    description?: string
    costStars: number
    active: boolean
    approvalRequired: boolean
    availabilityExpression?: string
  }) {
    return connectFetch<{ standardResponse?: StandardResponse; reward?: Reward }>(
      '/starapp.api.v1.StarAppService/UpdateReward',
      body,
    )
  },
  deleteReward(body: { id: number }) {
    return connectFetch<{ standardResponse?: StandardResponse }>(
      '/starapp.api.v1.StarAppService/DeleteReward',
      body,
    )
  },
  requestRedemption(body: { rewardId: number; childMemberId?: number }) {
    return connectFetch<{ standardResponse?: StandardResponse; redemption?: Redemption }>(
      '/starapp.api.v1.StarAppService/RequestRedemption',
      body,
    )
  },
  approveRedemption(body: { redemptionId: number }) {
    return connectFetch<{ standardResponse?: StandardResponse; redemption?: Redemption }>(
      '/starapp.api.v1.StarAppService/ApproveRedemption',
      body,
    )
  },
  rejectRedemption(body: { redemptionId: number }) {
    return connectFetch<{ standardResponse?: StandardResponse; redemption?: Redemption }>(
      '/starapp.api.v1.StarAppService/RejectRedemption',
      body,
    )
  },
  listRedemptions(body: { status?: string } = {}) {
    return connectFetch<{ redemptions: Redemption[] }>(
      '/starapp.api.v1.StarAppService/ListRedemptions',
      body,
    )
  },
  getParentHomeSummary() {
    return connectFetch<ParentHomeSummary>(
      '/starapp.api.v1.StarAppService/GetParentHomeSummary',
      {},
    )
  },
  getChildHomeSummary() {
    return connectFetch<ChildHomeSummary>(
      '/starapp.api.v1.StarAppService/GetChildHomeSummary',
      {},
    )
  },
  listChores(body: { includeInactive?: boolean; starChartId?: number } = {}) {
    return connectFetch<{ chores: Chore[] }>(
      '/starapp.api.v1.StarAppService/ListChores',
      body,
    )
  },
  createChore(body: {
    title: string
    starReward: number
    weekdays: number[]
    childMemberIds: number[]
    starChartId?: number
  }) {
    return connectFetch<{ standardResponse?: StandardResponse; chore?: Chore }>(
      '/starapp.api.v1.StarAppService/CreateChore',
      body,
    )
  },
  updateChore(body: {
    id: number
    title: string
    starReward: number
    weekdays: number[]
    childMemberIds: number[]
    active: boolean
    starChartId?: number
  }) {
    return connectFetch<{ standardResponse?: StandardResponse; chore?: Chore }>(
      '/starapp.api.v1.StarAppService/UpdateChore',
      body,
    )
  },
  deleteChore(body: { id: number }) {
    return connectFetch<{ standardResponse?: StandardResponse }>(
      '/starapp.api.v1.StarAppService/DeleteChore',
      body,
    )
  },
  listChorePauses() {
    return connectFetch<{ pauses: ChorePause[] }>(
      '/starapp.api.v1.StarAppService/ListChorePauses',
      {},
    )
  },
  createChorePause(body: { startDate: string; endDate: string; reason?: string }) {
    return connectFetch<{ standardResponse?: StandardResponse; pause?: ChorePause }>(
      '/starapp.api.v1.StarAppService/CreateChorePause',
      body,
    )
  },
  deleteChorePause(body: { id: number }) {
    return connectFetch<{ standardResponse?: StandardResponse }>(
      '/starapp.api.v1.StarAppService/DeleteChorePause',
      body,
    )
  },
  getWeeklyStarChart(body: { weekStart?: string; starChartId?: number } = {}) {
    return connectFetch<WeeklyStarChart>(
      '/starapp.api.v1.StarAppService/GetWeeklyStarChart',
      body,
    )
  },
  listStarCharts(body: { includeInactive?: boolean } = {}) {
    return connectFetch<{ starCharts: StarChart[] }>(
      '/starapp.api.v1.StarAppService/ListStarCharts',
      body,
    )
  },
  createStarChart(body: { name: string; sortOrder?: number }) {
    return connectFetch<{ standardResponse?: StandardResponse; starChart?: StarChart }>(
      '/starapp.api.v1.StarAppService/CreateStarChart',
      body,
    )
  },
  updateStarChart(body: { id: number; name: string; sortOrder?: number; active: boolean }) {
    return connectFetch<{ standardResponse?: StandardResponse; starChart?: StarChart }>(
      '/starapp.api.v1.StarAppService/UpdateStarChart',
      body,
    )
  },
  deleteStarChart(body: { id: number }) {
    return connectFetch<{ standardResponse?: StandardResponse }>(
      '/starapp.api.v1.StarAppService/DeleteStarChart',
      body,
    )
  },
  completeChore(body: { choreId: number; childMemberId: number; date: string }) {
    return connectFetch<{ standardResponse?: StandardResponse; newBalance?: number }>(
      '/starapp.api.v1.StarAppService/CompleteChore',
      body,
    )
  },
  uncompleteChore(body: { choreId: number; childMemberId: number; date: string }) {
    return connectFetch<{ standardResponse?: StandardResponse; newBalance?: number }>(
      '/starapp.api.v1.StarAppService/UncompleteChore',
      body,
    )
  },
}

export type Family = {
  id: number
  name: string
  createdAt?: string
}

export type FamilyMember = {
  id: number
  familyId: number
  userAccountId?: number
  displayName: string
  role: string
  hasAvatar?: boolean
  createdAt?: string
  username?: string
  starColor?: string
}

export type MemberAvatarEntry = {
  filename: string
  isCurrent?: boolean
}

export type StarLedgerEntry = {
  id: number
  familyId: number
  childMemberId: number
  amount: number
  entryType: string
  note?: string
  relatedRewardId?: number
  createdByMemberId?: number
  createdAt?: string
}

export type Reward = {
  id: number
  familyId: number
  title: string
  description?: string
  costStars: number
  active?: boolean
  approvalRequired?: boolean
  availabilityExpression?: string
}

export type Redemption = {
  id: number
  familyId: number
  childMemberId: number
  rewardId: number
  starsSpent: number
  status: string
  createdAt?: string
  resolvedAt?: string
  rewardTitle?: string
  childDisplayName?: string
}

export type ChildHomeSummary = {
  member?: FamilyMember
  balance?: number
  recentAwards?: StarLedgerEntry[]
  rewards?: Reward[]
  pendingRewardIds?: number[]
  unavailableRewardIds?: number[]
}

export type ParentHomeSummary = {
  family?: Family
  children?: Array<{
    member?: FamilyMember
    balance?: number
    lastAward?: StarLedgerEntry
  }>
  pendingRedemptions?: number
}

export type Chore = {
  id: number
  familyId: number
  title: string
  starReward: number
  weekdays?: number[]
  active?: boolean
  childMemberIds?: number[]
  createdAt?: string
  starChartId?: number
}

export type StarChart = {
  id: number
  familyId: number
  name: string
  sortOrder?: number
  active?: boolean
  createdAt?: string
  choreCount?: number
}

export type ChorePause = {
  id: number
  familyId: number
  startDate: string
  endDate: string
  reason?: string
  createdAt?: string
}

export type WeeklyStarChartDay = {
  date?: string
  scheduled?: boolean
  completed?: boolean
  starsEarned?: number
  paused?: boolean
}

export type WeeklyStarChartChild = {
  assignmentId?: number
  child?: FamilyMember
  days?: WeeklyStarChartDay[]
}

export type WeeklyStarChartRow = {
  choreId?: number
  title?: string
  starReward?: number
  weekdays?: number[]
  children?: WeeklyStarChartChild[]
}

export type WeeklyStarChartBonusChild = {
  child?: FamilyMember
  days?: WeeklyStarChartDay[]
}

export type WeeklyStarChart = {
  weekStart?: string
  weekEnd?: string
  rows?: WeeklyStarChartRow[]
  bonusChildren?: WeeklyStarChartBonusChild[]
  starChartId?: number
  starChartName?: string
}

export function memberAvatarUrl(memberId: number, hasAvatar?: boolean): string {
  if (!hasAvatar) return ''
  return `/avatars/${memberId}`
}

export function memberAvatarFileUrl(memberId: number, filename: string): string {
  return `/avatars/${memberId}/${encodeURIComponent(filename)}`
}

function inferImageContentType(filename: string): string {
  const lower = filename.toLowerCase()
  if (lower.endsWith('.jpg') || lower.endsWith('.jpeg')) return 'image/jpeg'
  if (lower.endsWith('.png')) return 'image/png'
  if (lower.endsWith('.webp')) return 'image/webp'
  return ''
}

async function fileToBase64(file: File): Promise<string> {
  const dataUrl = await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(reader.error ?? new Error('Could not read image file'))
    reader.readAsDataURL(file)
  })
  const comma = dataUrl.indexOf(',')
  if (comma < 0) {
    throw new Error('Could not read image file')
  }
  return dataUrl.slice(comma + 1)
}
