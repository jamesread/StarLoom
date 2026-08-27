<script setup lang="ts">
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, Key01Icon, UserGroupIcon, UserIcon } from '@hugeicons/core-free-icons'
import { fetchAppStatus, useStatus } from '../composables/useStatus'
import { canAccessIamFromStatus, hasPermission } from '../lib/rbacAccess'

const router = useRouter()
const status = useStatus()
const navRef = ref<InstanceType<typeof Navigation> | null>(null)
const canAccessIam = computed(() => canAccessIamFromStatus(status.status))

onMounted(async () => {
  const st = await fetchAppStatus()
  if (!canAccessIamFromStatus(st)) {
    router.push('/')
    return
  }
  await nextTick()
  const nav = navRef.value
  if (!nav) return
  if (hasPermission(st, 'users.view')) {
    nav.addCallback('Users', () => router.push({ name: 'users' }), {
      icon: UserIcon,
      name: 'users',
      description: 'Manage accounts',
    })
  }
  if (hasPermission(st, 'usergroups.view')) {
    nav.addCallback('User groups', () => router.push({ name: 'user-groups' }), {
      icon: UserGroupIcon,
      name: 'groups',
      description: 'Membership and roles',
    })
  }
  if (hasPermission(st, 'rbac.view')) {
    nav.addCallback('Roles', () => router.push({ name: 'rbac-roles' }), {
      icon: Key01Icon,
      name: 'roles',
      description: 'Permissions via groups',
    })
    nav.addCallback('Permission catalog', () => router.push({ name: 'rbac-permissions' }), {
      name: 'perms',
      description: 'Read-only reference',
    })
  }
})
</script>

<template>
  <Section v-if="canAccessIam" title="IAM" subtitle="Users, groups, and roles" :padding="true">
    <template #toolbar>
      <router-link :to="{ name: 'controlPanel' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Control Panel</span>
      </router-link>
    </template>
    <Navigation ref="navRef">
      <NavigationGrid />
    </Navigation>
  </Section>
</template>
