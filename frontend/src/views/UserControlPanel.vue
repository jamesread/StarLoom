<script setup lang="ts">
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { computed, nextTick, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
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
import { canAccessControlPanelFromStatus, canViewFamilyHomeFromStatus } from '../lib/rbacAccess'

const router = useRouter()
const status = useStatus()
const navRef = ref<InstanceType<typeof Navigation> | null>(null)
const loggingOut = ref(false)
const error = ref('')
const personId = ref<number | null>(null)
const personName = ref('')

const userData = computed(() => status.status)
const showPersonLink = computed(
  () => Boolean(personId.value) && canViewFamilyHomeFromStatus(status.status),
)

async function refreshUserData() {
  error.value = ''
  try {
    await fetchAppStatus({ force: true })
    await loadPerson()
    await nextTick()
    setupQuickActions()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Refresh failed'
  }
}

async function loadPerson() {
  try {
    const fam = await starapp.getMyFamily()
    const id = fam.callerMember?.id
    personId.value = id || null
    personName.value = fam.callerMember?.displayName?.trim() || ''
  } catch {
    personId.value = null
    personName.value = ''
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
  nav.addCallback('Chore notifications', () => router.push({ name: 'choreNotifications' }), {
    icon: Notification03Icon,
    name: 'chore-notifications',
    description: 'Apprise alerts when family members complete chores',
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

onMounted(() => {
  void refreshUserData()
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

    <h3>{{ userData.displayName?.trim() || userData.username }}</h3>
    <dl class="account-info">
      <dt>Username</dt>
      <dd>{{ userData.username }}</dd>
      <dt>Account created</dt>
      <dd>{{ formatDate(userData.accountCreatedAt) }}</dd>
      <template v-if="showPersonLink">
        <dt>Person</dt>
        <dd>
          <RouterLink
            :to="{ name: 'familyPersonDetail', params: { id: personId } }"
            class="title-link"
          >
            {{ personName || 'View person' }}
          </RouterLink>
        </dd>
      </template>
    </dl>
  </Section>

  <Section v-if="userData?.isLoggedIn" title="Notifications" :padding="true">
    <p>
      Choose which chore completions should notify you, or manage subscriptions from your
      <RouterLink v-if="showPersonLink" :to="{ name: 'familyPersonDetail', params: { id: personId } }">
        person profile
      </RouterLink>
      <span v-else>person profile</span>.
    </p>
    <p>
      <RouterLink :to="{ name: 'choreNotifications' }">Manage chore completion subscriptions</RouterLink>
    </p>
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

.title-link {
  font: inherit;
  color: var(--pico-primary);
  text-decoration: none;
}

.title-link:hover {
  text-decoration: underline;
}
</style>
