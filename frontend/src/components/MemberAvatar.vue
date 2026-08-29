<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, type RouteLocationRaw } from 'vue-router'
import { memberAvatarUrl } from '../api/client'
import { memberAvatarStyle, type MemberColorSource } from '../lib/memberStarColor'

export type MemberAvatarSource = MemberColorSource & {
  id: number
  displayName: string
  hasAvatar?: boolean
}

const props = withDefaults(
  defineProps<{
    member: MemberAvatarSource
    size?: 'sm' | 'md' | 'lg' | 'xl'
    to?: RouteLocationRaw
    title?: string
  }>(),
  { size: 'md' },
)

const avatarStyle = computed(() => memberAvatarStyle(props.member))
const initial = computed(() => (props.member.displayName || '?').charAt(0))
const linkTitle = computed(() => props.title ?? props.member.displayName)
</script>

<template>
  <component
    :is="to ? RouterLink : 'span'"
    :to="to"
    :class="to ? 'member-avatar-link' : 'member-avatar-wrap'"
    :title="to ? linkTitle : undefined"
  >
    <img
      v-if="member.hasAvatar"
      class="member-avatar"
      :class="`size-${size}`"
      :style="avatarStyle"
      :src="memberAvatarUrl(member.id, true)"
      :alt="member.displayName"
    />
    <span
      v-else
      class="member-avatar member-avatar-placeholder"
      :class="`size-${size}`"
      :style="avatarStyle"
      :aria-label="member.displayName"
    >
      {{ initial }}
    </span>
  </component>
</template>

<style scoped>
.member-avatar-link,
.member-avatar-wrap {
  display: inline-flex;
  line-height: 0;
}

.member-avatar-link {
  text-decoration: none;
  border-radius: 50%;
}

.member-avatar-link:hover {
  opacity: 0.85;
}

.member-avatar {
  border-radius: 50%;
  object-fit: cover;
  box-sizing: border-box;
}

.member-avatar-placeholder {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--pico-muted-border-color);
  font-weight: 600;
}

.size-sm {
  width: 2rem;
  height: 2rem;
  font-size: 0.85rem;
}

.size-md {
  width: 3.5rem;
  height: 3.5rem;
  font-size: 1.25rem;
}

.size-lg {
  width: 4rem;
  height: 4rem;
  font-size: 1.5rem;
}

.size-xl {
  width: 5rem;
  height: 5rem;
  font-size: 2rem;
}
</style>
