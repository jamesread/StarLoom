<script setup lang="ts">
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { computed, nextTick, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, ArrowRight01Icon, Key01Icon, SearchList01Icon, UserGroupIcon, UserIcon } from '@hugeicons/core-free-icons'
import { fetchAppStatus, useStatus } from '../composables/useStatus'
import { canAccessIamFromStatus, hasPermission } from '../lib/rbacAccess'

const router = useRouter()
const status = useStatus()
const navRef = ref<InstanceType<typeof Navigation> | null>(null)
const canAccessIam = computed(() => canAccessIamFromStatus(status.status))
const iconStrokeWidth = 2.5

type IamFlowStep = {
  kind: 'step'
  label: string
  icon: typeof UserIcon
  route: 'users' | 'user-groups' | 'rbac-roles' | 'rbac-permissions'
  permission: string
}

type IamFlowLink = {
  kind: 'link'
  label: string
}

const iamFlowParts: Array<IamFlowStep | IamFlowLink> = [
  { kind: 'step', label: 'Permissions', icon: SearchList01Icon, route: 'rbac-permissions', permission: 'rbac.view' },
  { kind: 'link', label: 'From roles' },
  { kind: 'step', label: 'Roles', icon: Key01Icon, route: 'rbac-roles', permission: 'rbac.view' },
  { kind: 'link', label: 'To groups' },
  { kind: 'step', label: 'User groups', icon: UserGroupIcon, route: 'user-groups', permission: 'usergroups.view' },
  { kind: 'link', label: 'Include users' },
  { kind: 'step', label: 'Users', icon: UserIcon, route: 'users', permission: 'users.view' },
]

function canAccessFlowStep(permission: string) {
  return hasPermission(status.status, permission)
}

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
      icon: SearchList01Icon,
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

    <p class="iam-intro">
      Permissions are granted by roles. Roles are assigned to user groups, and users join those groups to
      receive access in the app.
    </p>

    <div class="iam-flow" aria-label="How users, user groups, roles, and permissions connect">
      <template v-for="(part, index) in iamFlowParts" :key="`${part.kind}-${index}`">
        <div v-if="part.kind === 'step'" class="iam-flow-step">
          <RouterLink
            v-if="canAccessFlowStep(part.permission)"
            :to="{ name: part.route }"
            class="iam-flow-step-link"
            :aria-label="`Open ${part.label}`"
          >
            <span class="iam-flow-icon" aria-hidden="true">
              <HugeiconsIcon
                :icon="part.icon"
                width="1.35em"
                height="1.35em"
                :strokeWidth="iconStrokeWidth"
              />
            </span>
            <span class="iam-flow-label">{{ part.label }}</span>
          </RouterLink>
          <template v-else>
            <span class="iam-flow-icon" aria-hidden="true">
              <HugeiconsIcon
                :icon="part.icon"
                width="1.35em"
                height="1.35em"
                :strokeWidth="iconStrokeWidth"
              />
            </span>
            <span class="iam-flow-label">{{ part.label }}</span>
          </template>
        </div>
        <div v-else class="iam-flow-arrow" :aria-label="part.label">
          <span class="iam-flow-arrow-icon" aria-hidden="true">
            <HugeiconsIcon
              :icon="ArrowRight01Icon"
              width="1.1em"
              height="1.1em"
              :strokeWidth="iconStrokeWidth"
            />
          </span>
          <span class="iam-flow-label">{{ part.label }}</span>
        </div>
      </template>
    </div>

    <Navigation ref="navRef">
      <NavigationGrid />
    </Navigation>
  </Section>
</template>

<style scoped>
.iam-intro {
  margin: 0 0 1rem;
  max-width: 42rem;
}

.iam-flow {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 0.5rem 0.65rem;
  margin: 0 0 1.5rem;
  padding: 0.85rem 1rem;
  border: 1px solid var(--pico-muted-border-color);
  border-radius: var(--pico-border-radius);
  background: var(--pico-card-background-color, var(--pico-background-color));
}

.iam-flow-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.3rem;
  min-width: 4.75rem;
  text-align: center;
}

.iam-flow-step-link {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.3rem;
  color: inherit;
  text-decoration: none;
}

.iam-flow-step-link:hover .iam-flow-icon {
  filter: brightness(1.08);
}

.iam-flow-step-link:hover .iam-flow-label {
  color: var(--pico-primary);
  text-decoration: underline;
}

.iam-flow-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: var(--pico-border-radius);
  background: var(--pico-primary-background);
  color: var(--pico-primary-inverse);
}

.iam-flow-label {
  font-size: 0.82rem;
  line-height: 1.2;
  color: var(--pico-muted-color);
}

.iam-flow-arrow {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.3rem;
  min-width: 4.75rem;
  text-align: center;
  color: var(--pico-muted-color);
}

.iam-flow-arrow-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
}
</style>
