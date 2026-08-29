<script setup lang="ts">
import { onMounted, computed, ref } from 'vue'
import AppLayout from './components/AppLayout.vue'
import LoginForm from './components/LoginForm.vue'
import { fetchAppStatus, useStatus } from './composables/useStatus'
import { loadInit } from './composables/useInit'
import { loadAndApplyUserPreferences, registerSidebarApplier, registerThemeToggleApplier } from './lib/userPreferences'

const status = useStatus()
const sidebarPreferenceEnabled = ref(true)
const themeTogglePreferenceEnabled = ref(false)

const isLoggedIn = computed(() => Boolean(status.status?.isLoggedIn))

registerSidebarApplier((enabled) => {
  sidebarPreferenceEnabled.value = enabled
})
registerThemeToggleApplier((enabled) => {
  themeTogglePreferenceEnabled.value = enabled
})

onMounted(async () => {
  try {
    await Promise.all([loadInit(), fetchAppStatus()])
    if (status.status?.isLoggedIn) {
      await loadAndApplyUserPreferences()
    }
  } catch {
    // login form shown
  }
})

async function onLogin() {
  await fetchAppStatus({ force: true })
  await loadAndApplyUserPreferences()
}
</script>

<template>
  <AppLayout
    v-if="isLoggedIn"
    :sidebar-preference-enabled="sidebarPreferenceEnabled"
    :theme-toggle-preference-enabled="themeTogglePreferenceEnabled"
  >
    <p v-if="status.loading && !status.loaded">Loading…</p>
    <p v-else-if="status.error">{{ status.error }}</p>
    <RouterView v-else />
  </AppLayout>
  <main v-else class="login-shell">
    <p v-if="status.loading">Loading…</p>
    <LoginForm v-else @login-success="onLogin" />
  </main>
</template>

<style scoped>
.login-shell {
  max-width: 28rem;
  margin: 4rem auto;
  padding: 0 1rem;
}
</style>
