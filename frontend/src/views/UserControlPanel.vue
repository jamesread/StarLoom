<script setup lang="ts">
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import Section from 'picocrank/vue/components/Section.vue'
import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { HugeiconsIcon } from '@hugeicons/vue'
import {
  DashboardSquareSettingIcon,
  Key01Icon,
  Notification03Icon,
  Settings01Icon,
  UserShield01Icon,
} from '@hugeicons/core-free-icons'
import { fetchAppStatus, invalidateAppStatus, useStatus } from '../composables/useStatus'
import { starapp } from '../api/client'
import { canAccessControlPanelFromStatus } from '../lib/rbacAccess'

const iconStrokeWidth = 2.5

const router = useRouter()
const status = useStatus()
const navRef = ref<InstanceType<typeof Navigation> | null>(null)
const loggingOut = ref(false)
const error = ref('')
const testingNotify = ref(false)
const notifyError = ref('')
const notifySuccess = ref('')
const personTag = ref('')

const userData = computed(() => status.status)

async function refreshUserData() {
  error.value = ''
  try {
    await fetchAppStatus({ force: true })
    await nextTick()
    setupQuickActions()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Refresh failed'
  }
}

async function loadPersonTag() {
  try {
    const fam = await starapp.getMyFamily()
    const id = fam.callerMember?.id
    personTag.value = id ? `starloom_uid_${id}` : ''
  } catch {
    personTag.value = ''
  }
}

function setupQuickActions() {
  const nav = navRef.value
  if (!nav) return
  nav.clearNavigationLinks()
  nav.addCallback('User Preferences', () => router.push({ name: 'userPreferences' }), {
    icon: Settings01Icon,
    name: 'user-preferences',
    description: 'Language and personal settings',
  })
  nav.addCallback('Change Password', () => router.push({ name: 'changePassword' }), {
    icon: Key01Icon,
    name: 'change-password',
    description: 'Update your account password',
  })
  nav.addCallback('API Keys', () => router.push({ name: 'apiKeys' }), {
    icon: Key01Icon,
    name: 'api-keys',
    description: 'Manage Bearer tokens for automation',
  })
  nav.addCallback('My Permissions', () => router.push({ name: 'myPermissions' }), {
    icon: UserShield01Icon,
    name: 'my-permissions',
    description: 'Review groups, roles, and effective access',
  })
  if (canAccessControlPanelFromStatus(status.status)) {
    nav.addCallback('Control Panel', () => router.push({ name: 'controlPanel' }), {
      name: 'control-panel',
      icon: DashboardSquareSettingIcon,
      description: 'System administration',
    })
  }
}

async function sendTestNotification() {
  testingNotify.value = true
  notifyError.value = ''
  notifySuccess.value = ''
  try {
    const res = await starapp.testAppriseNotification()
    const tag = res.tag || personTag.value
    if (tag) personTag.value = tag
    notifySuccess.value = res.standardResponse?.message
      ? `${res.standardResponse.message}${tag ? ` (tag ${tag})` : ''}`
      : 'Test notification sent'
  } catch (e) {
    notifyError.value = e instanceof Error ? e.message : String(e)
  } finally {
    testingNotify.value = false
  }
}

async function logout() {
  loggingOut.value = true
  try {
    await starapp.logout()
    invalidateAppStatus()
    await router.push('/')
    window.location.reload()
  } finally {
    loggingOut.value = false
  }
}

function formatDate(value?: string) {
  if (!value) return '—'
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? value : d.toLocaleString()
}

onMounted(async () => {
  await refreshUserData()
  await loadPersonTag()
  await nextTick()
  setupQuickActions()
})
</script>

<template>
  <p v-if="error" class="inline-notification error">{{ error }}</p>

  <Section v-if="userData?.isLoggedIn" title="Identity" :padding="true">
    <template #toolbar>
      <button type="button" class="inline-icon neutral" aria-label="Refresh" @click="refreshUserData">
        Refresh
      </button>
    </template>

    <h3>{{ userData.username }}</h3>
    <dl class="account-info">
      <dt>Username</dt>
      <dd>{{ userData.username }}</dd>
      <dt>Account created</dt>
      <dd>{{ formatDate(userData.accountCreatedAt) }}</dd>
    </dl>
  </Section>

  <Section v-if="userData?.isLoggedIn" title="Notifications" :padding="true">
    <p>
      Send a test Apprise notification to your devices.
      <template v-if="personTag"> Uses tag <code>{{ personTag }}</code>.</template>
    </p>
    <p v-if="notifyError" class="inline-notification error">{{ notifyError }}</p>
    <NotificationBlock
      v-if="notifySuccess"
      type="success"
      :message="notifySuccess"
    />
    <button
      type="button"
      class="inline-icon neutral"
      :disabled="testingNotify"
      @click="sendTestNotification"
    >
      <HugeiconsIcon
        :icon="Notification03Icon"
        width="1em"
        height="1em"
        :strokeWidth="iconStrokeWidth"
        aria-hidden="true"
      />
      <span>{{ testingNotify ? 'Sending…' : 'Send test notification' }}</span>
    </button>
  </Section>

  <Section v-if="userData?.isLoggedIn" title="Quick actions" :padding="true">
    <Navigation ref="navRef">
      <NavigationGrid />
    </Navigation>
  </Section>

  <Section v-if="userData?.isLoggedIn" title="Session" :padding="true">
    <p>End your current session and return to the login page.</p>
    <button type="button" class="bad" :disabled="loggingOut" @click="logout">
      {{ loggingOut ? 'Signing out…' : 'Sign Out' }}
    </button>
  </Section>
</template>

<style scoped>
.account-info {
  display: grid;
  grid-template-columns: 200px 1fr;
  column-gap: 1em;
  margin: 1em 0 0;
}
.account-info dt {
  font-weight: bold;
}
.account-info dd {
  margin: 0;
}
</style>
