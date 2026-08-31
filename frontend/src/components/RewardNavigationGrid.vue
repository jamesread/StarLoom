<script setup lang="ts">
import { computed, inject, ref, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import { HugeiconsIcon } from '@hugeicons/vue'

type NavLink = {
  name: string
  type: string
  title: string
  description?: string | null
  disabled?: boolean
  rewardBlocked?: boolean
  icon?: unknown
  callback?: () => void
  to?: unknown
  path?: string
  iconColor?: string | null
  indicator?: boolean
  count?: number | null
}

defineProps({
  iconSize: {
    type: String,
    default: '1.5em',
  },
})

const router = useRouter()
const navigation = inject<{ navigationLinks: Ref<NavLink[]>; isActive: (link: NavLink) => boolean } | null>(
  'navigation',
  null,
)
const navigationLinks = navigation ? navigation.navigationLinks : ref<NavLink[]>([])
const isActive = navigation ? navigation.isActive : () => false

const filteredLinks = computed(() =>
  (navigationLinks.value || []).filter(
    (link) => link.type !== 'separator' && link.type !== 'html' && link.type !== 'section',
  ),
)

function handleLinkClick(link: NavLink) {
  if (link.disabled) return
  if (link.type === 'route') {
    const to = link.to || link.path
    if (to) {
      void router.push(to as string)
    }
  } else if (link.type === 'callback' && link.callback) {
    link.callback()
  }
}

function linkClasses(link: NavLink) {
  const blocked = Boolean(link.rewardBlocked)
  return {
    active: !link.disabled && isActive(link),
    disabled: link.disabled && !blocked,
    'reward-blocked': blocked,
  }
}
</script>

<template>
  <div class="navigation-grid-container">
    <div class="navigation-grid">
      <button
        v-for="link in filteredLinks"
        :key="link.name"
        :class="['nav-button', linkClasses(link)]"
        :disabled="link.disabled"
        :title="link.title"
        :aria-disabled="link.disabled ? 'true' : undefined"
        @click="handleLinkClick(link)"
      >
        <div class="nav-button-heading">
          <div class="nav-button-icon">
            <HugeiconsIcon
              v-if="link.icon"
              :icon="link.icon as never"
              :width="iconSize"
              :height="iconSize"
            />
          </div>
          <div class="nav-button-label">{{ link.title }}</div>
        </div>
        <div v-if="link.description" class="nav-button-description">
          {{ link.description }}
        </div>
      </button>
    </div>
    <div v-if="!filteredLinks.length" class="navigation-grid-empty">
      <p>No navigation links available</p>
    </div>
  </div>
</template>
