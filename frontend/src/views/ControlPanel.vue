<template>
  <Section v-if="canAccessControlPanel" title="Control Panel" subtitle="System functionality">
    <template #toolbar>
      <button type="button" class="neutral" :disabled="!clientReady" @click="refresh">Refresh</button>
    </template>
    <Navigation ref="localNavigation">
      <NavigationGrid />
    </Navigation>
  </Section>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import { Settings01Icon, UserShield01Icon, WebhookIcon } from '@hugeicons/core-free-icons'
import { fetchAppStatus, useStatus } from '../composables/useStatus'
import {
  canAccessControlPanelFromStatus,
  canAccessIamFromStatus,
  canAccessSettingsFromStatus,
  canAccessWebhooksFromStatus,
} from '../lib/rbacAccess'

const router = useRouter()
const status = useStatus()
const clientReady = ref(false)
const localNavigation = ref<InstanceType<typeof Navigation> | null>(null)
const canAccessControlPanel = computed(() => canAccessControlPanelFromStatus(status.status))

async function refresh() {
  const st = await fetchAppStatus({ force: true })
  if (!canAccessControlPanelFromStatus(st)) {
    router.push('/')
    return
  }
  clientReady.value = true
  await nextTick()
  populateHubTiles()
}

function populateHubTiles() {
  const nav = localNavigation.value
  if (!nav) return
  nav.clearNavigationLinks()
  const st = status.status
  let tileCount = 0
  if (canAccessIamFromStatus(st)) {
    tileCount += 1
    nav.addCallback('IAM', () => router.push({ name: 'iam' }), {
      icon: UserShield01Icon,
      name: 'iam',
      description: 'Users, groups, and roles',
    })
  }
  if (canAccessSettingsFromStatus(st)) {
    tileCount += 1
    nav.addCallback('Settings', () => router.push({ name: 'settings' }), {
      icon: Settings01Icon,
      name: 'settings',
      description: 'Configure system settings',
    })
  }
  if (canAccessWebhooksFromStatus(st)) {
    tileCount += 1
    nav.addCallback('Webhooks', () => router.push({ name: 'webhooks' }), {
      icon: WebhookIcon,
      name: 'webhooks',
      description: 'HTTP callbacks for StarLoom events',
    })
  }
  if (tileCount === 0) {
    router.push('/')
  }
}

onMounted(refresh)
</script>
