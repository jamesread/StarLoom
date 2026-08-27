<script setup lang="ts">
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { fetchAppStatus, invalidateAppStatus, useStatus } from '../composables/useStatus'
import { starapp } from '../api/client'

const router = useRouter()
const status = useStatus()
const navRef = ref<InstanceType<typeof Navigation> | null>(null)
const loggingOut = ref(false)
const error = ref('')

const userData = computed(() => status.status)

async function refreshUserData() {
  error.value = ''
  try {
    await fetchAppStatus({ force: true })
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Refresh failed'
  }
}

function setupQuickActions() {
  const nav = navRef.value
  if (!nav) return
  nav.addCallback('User Preferences', () => router.push({ name: 'userPreferences' }), {
    name: 'user-preferences',
    description: 'Language and personal settings',
  })
  nav.addCallback('Change Password', () => router.push({ name: 'changePassword' }), {
    name: 'change-password',
    description: 'Update your account password',
  })
  nav.addCallback('API Keys', () => router.push({ name: 'apiKeys' }), {
    name: 'api-keys',
    description: 'Manage Bearer tokens for automation',
  })
  nav.addCallback('My Permissions', () => router.push({ name: 'myPermissions' }), {
    name: 'my-permissions',
    description: 'Review groups, roles, and effective access',
  })
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
  row-gap: 0.25em;
  margin: 1em 0 0;
}
.account-info dt {
  font-weight: bold;
}
.account-info dd {
  margin: 0;
}
</style>
