<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import AppLayout from './components/AppLayout.vue'
import LoginForm from './components/LoginForm.vue'
import { fetchAppStatus, useStatus } from './composables/useStatus'
import { loadInit, useInit } from './composables/useInit'
import { loadAndApplyUserPreferences, registerSidebarApplier } from './lib/userPreferences'
import {
  applyAppTheming,
  featuresFromShell,
  themeColorSchemeSwitcherEnabledFromFeatures,
} from './lib/applyAppTheming'
import { syncThemeFontStylesheet } from './lib/themeFonts'
import { useCustomTheme } from 'picocrank/vue/composables/useCustomTheme.js'

const status = useStatus()
const init = useInit()
const sidebarPreferenceEnabled = ref(true)
const { discoverThemes, themePreference } = useCustomTheme()

const isLoggedIn = computed(() => Boolean(status.status?.isLoggedIn))
const initFeatures = computed(() => featuresFromShell(init.init, status.status))
const themeColorSchemeSwitcherEnabled = computed(() =>
  themeColorSchemeSwitcherEnabledFromFeatures(initFeatures.value),
)

registerSidebarApplier((enabled) => {
  sidebarPreferenceEnabled.value = enabled
})

watch(
  themePreference,
  (name) => {
    void syncThemeFontStylesheet(name)
  },
  { immediate: true },
)

async function applyShellTheming() {
  await discoverThemes()
  applyAppTheming(initFeatures.value)
}

onMounted(async () => {
  try {
    await Promise.all([loadInit(), fetchAppStatus()])
    if (status.status?.isLoggedIn) {
      await loadAndApplyUserPreferences()
    }
  } catch {
    // login form shown
  }
  await applyShellTheming()
})

async function onLogin() {
  await fetchAppStatus({ force: true })
  await loadInit({ force: true })
  await loadAndApplyUserPreferences()
  await applyShellTheming()
}
</script>

<template>
  <AppLayout
    v-if="isLoggedIn"
    :sidebar-preference-enabled="sidebarPreferenceEnabled"
    :theme-color-scheme-switcher-enabled="themeColorSchemeSwitcherEnabled"
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
