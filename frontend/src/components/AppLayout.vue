<script setup lang="ts">
import Header from 'picocrank/vue/components/Header.vue'
import Navigation from 'picocrank/vue/components/Navigation.vue'
import Sidebar from 'picocrank/vue/components/Sidebar.vue'
import { computed, nextTick, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import AppFooter from './AppFooter.vue'
import { useInit } from '../composables/useInit'
import { useStatus } from '../composables/useStatus'
import { canAccessControlPanelFromStatus, canViewFamilyHomeFromStatus } from '../lib/rbacAccess'
import { setupSidebarNavigation } from '../lib/sidebarNavigation'

defineProps<{
  sidebarPreferenceEnabled?: boolean
  themeTogglePreferenceEnabled?: boolean
}>()

const router = useRouter()
const status = useStatus()
const init = useInit()
const navigation = ref<InstanceType<typeof Navigation> | null>(null)
const topBarNavigation = ref<InstanceType<typeof Navigation> | null>(null)
const sidebar = ref<InstanceType<typeof Sidebar> | null>(null)

const siteTitle = computed(() => status.status?.siteTitle || 'StarApp')
const showFooter = computed(() => init.init?.showFooter !== false)
const username = computed(() => (status.status?.isLoggedIn ? status.status.username || '' : ''))
const isLoggedIn = computed(() => Boolean(status.status?.isLoggedIn))

watch(
  siteTitle,
  (title) => {
    document.title = title
  },
  { immediate: true },
)

watch(
  [navigation, topBarNavigation, () => status.status],
  async () => {
    await nextTick()
    const showControlPanel = canAccessControlPanelFromStatus(status.status)
    const showFamilyNav = canViewFamilyHomeFromStatus(status.status)
    if (navigation.value) {
      setupSidebarNavigation(navigation.value, { showControlPanel, showFamilyNav })
    }
    if (topBarNavigation.value) {
      setupSidebarNavigation(topBarNavigation.value, {
        showControlPanel,
        showFamilyNav,
        excludeRoutes: ['familyChores', 'controlPanel'],
      })
    }
  },
  { immediate: true },
)

function toggleSidebar() {
  sidebar.value?.toggle()
}

function goHome() {
  router.push({ name: 'home' })
}

function goToUserControlPanel() {
  router.push({ name: 'userControlPanel' })
}
</script>

<template>
  <Navigation ref="navigation">
    <Navigation ref="topBarNavigation">
      <Header
        :title="siteTitle"
        logo-url="/favicon.svg"
        :username="username"
        :sidebar-enabled="isLoggedIn && (sidebarPreferenceEnabled ?? true)"
        :theme-toggle-enabled="isLoggedIn && (themeTogglePreferenceEnabled ?? false)"
        :breadcrumbs="false"
        :top-bar-enabled="true"
        :navigation="navigation"
        :top-bar-navigation="topBarNavigation"
        @logo-click="goHome"
        @toggle-sidebar="toggleSidebar"
        @user-click="goToUserControlPanel"
      />
    </Navigation>

    <div id="layout">
      <Sidebar
        v-if="isLoggedIn && (sidebarPreferenceEnabled ?? true)"
        ref="sidebar"
        :navigation="navigation"
      />

      <div id="content">
        <main title="Main content">
          <slot />
        </main>

        <AppFooter v-if="showFooter" logo-url="/favicon.svg" />
      </div>
    </div>
  </Navigation>
</template>
