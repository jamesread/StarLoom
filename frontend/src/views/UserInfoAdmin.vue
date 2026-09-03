<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
import DangerZone from 'picocrank/vue/components/DangerZone.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import {
  ArrowLeft01Icon,
  Notification03Icon,
  UserIcon,
} from '@hugeicons/core-free-icons'
import { starapp, type FamilyMember, type UserAccount, type UserGroup } from '../api/client'
import MemberAvatar from '../components/MemberAvatar.vue'
import { useStatus } from '../composables/useStatus'
import { canViewFamilyHomeFromStatus, hasPermission } from '../lib/rbacAccess'

const iconStrokeWidth = 2.5

const route = useRoute()
const router = useRouter()
const status = useStatus()
const userId = computed(() => Number(route.params.id))

const user = ref<UserAccount | null>(null)
const linkedMember = ref<FamilyMember | null>(null)
const userGroups = ref<UserGroup[]>([])
const loading = ref(true)
const error = ref('')
const testingNotify = ref(false)
const notifyError = ref('')
const notifySuccess = ref('')
const personTag = ref('')
const deleteError = ref('')
const deleting = ref(false)

const canDelete = computed(
  () =>
    hasPermission(status.status, 'users.delete') &&
    Boolean(user.value?.username) &&
    user.value?.username !== status.status?.username,
)

const canSendTest = computed(
  () => hasPermission(status.status, 'users.view') && Boolean(linkedMember.value?.id),
)
const showPersonLink = computed(
  () => Boolean(linkedMember.value?.id) && canViewFamilyHomeFromStatus(status.status),
)
const showUserGroupsLink = computed(() => hasPermission(status.status, 'usergroups.view'))

function formatDate(value?: string) {
  if (!value) return '—'
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? value : d.toLocaleString()
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await starapp.getUser({ userId: userId.value })
    user.value = res.user || null
    linkedMember.value = res.linkedMember || null
    userGroups.value = res.userGroups || []
    personTag.value = linkedMember.value?.id ? `starloom_uid_${linkedMember.value.id}` : ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
    user.value = null
    linkedMember.value = null
    userGroups.value = []
    personTag.value = ''
  } finally {
    loading.value = false
  }
}

async function sendTestNotification() {
  if (!user.value?.id || !linkedMember.value?.id) return
  testingNotify.value = true
  notifyError.value = ''
  notifySuccess.value = ''
  try {
    const res = await starapp.sendUserTestNotification({ userId: user.value.id })
    const tag = res.tag || personTag.value
    if (tag) personTag.value = tag
    notifySuccess.value = res.standardResponse?.message
      ? `${res.standardResponse.message}${tag ? ` (tag ${tag})` : ''}`
      : 'Test notification sent'
  } catch (e) {
    notifyError.value = e instanceof Error ? e.message : String(e)
  } finally {
    testingNotify.value = false
  }
}

async function removeUser() {
  if (!user.value?.id) return
  if (!confirm(`Delete user ${user.value.username}?`)) return
  deleting.value = true
  deleteError.value = ''
  try {
    await starapp.deleteUser({ userId: user.value.id })
    router.push({ name: 'users' })
  } catch (e) {
    deleteError.value = e instanceof Error ? e.message : String(e)
  } finally {
    deleting.value = false
  }
}

onMounted(load)
</script>

<template>
  <Section :title="user?.username || 'User'" :icon="UserIcon" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'users' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Users</span>
      </RouterLink>
      <button type="button" class="inline-icon neutral" :disabled="loading" @click="load">
        Refresh
      </button>
    </template>

    <p v-if="loading" class="subtle">Loading…</p>
    <p v-if="error" class="inline-notification error">{{ error }}</p>

    <template v-else-if="user">
      <dl class="account-info">
        <dt>Username</dt>
        <dd>{{ user.username }}</dd>
        <dt>Created by</dt>
        <dd>{{ user.createdBy || '—' }}</dd>
        <dt>Account created</dt>
        <dd>{{ formatDate(user.createdAt) }}</dd>
        <dt>User groups</dt>
        <dd>
          <ul v-if="userGroups.length" class="user-groups">
            <li v-for="group in userGroups" :key="group.id">{{ group.name }}</li>
          </ul>
          <span v-else class="subtle">Not a member of any groups</span>
          <p v-if="showUserGroupsLink" class="subtle user-groups-link">
            <RouterLink :to="{ name: 'user-groups' }">Manage user groups</RouterLink>
          </p>
        </dd>
        <template v-if="linkedMember">
          <dt>Linked person</dt>
          <dd>
            <RouterLink
              v-if="showPersonLink"
              :to="{ name: 'familyPersonDetail', params: { id: linkedMember.id } }"
              class="title-link"
            >
              {{ linkedMember.displayName }}
            </RouterLink>
            <span v-else>{{ linkedMember.displayName }}</span>
          </dd>
          <dt>Person role</dt>
          <dd>{{ linkedMember.role || '—' }}</dd>
        </template>
        <template v-else>
          <dt>Linked person</dt>
          <dd class="subtle">Not linked to a family profile</dd>
        </template>
      </dl>

      <div v-if="linkedMember" class="member-row">
        <MemberAvatar :member="linkedMember" size="lg" />
      </div>
    </template>
  </Section>

  <Section
    v-if="user && canSendTest"
    title="Notifications"
    :padding="true"
  >
    <p>
      Send a test Apprise notification to this user's devices.
      <template v-if="personTag"> Uses tag <code>{{ personTag }}</code>.</template>
    </p>
    <p v-if="notifyError" class="inline-notification error">{{ notifyError }}</p>
    <NotificationBlock
      v-if="notifySuccess"
      type="success"
      :message="notifySuccess"
    />
    <button
      type="button"
      class="inline-icon neutral"
      :disabled="testingNotify"
      @click="sendTestNotification"
    >
      <HugeiconsIcon
        :icon="Notification03Icon"
        width="1em"
        height="1em"
        :strokeWidth="iconStrokeWidth"
        aria-hidden="true"
      />
      <span>{{ testingNotify ? 'Sending…' : 'Send test notification' }}</span>
    </button>
  </Section>

  <DangerZone
    v-if="user && canDelete"
    title="Danger zone"
    subtitle="Delete this sign-in account"
    :warning="`Deleting ${user.username} removes their account and sign-in access. This cannot be undone.`"
  >
    <p v-if="deleteError" class="inline-notification error">{{ deleteError }}</p>
    <div role="toolbar" class="danger-zone-actions">
      <button type="button" class="bad" :disabled="deleting" @click="removeUser">
        {{ deleting ? 'Deleting…' : 'Delete user' }}
      </button>
    </div>
  </DangerZone>
</template>

<style scoped>
.account-info {
  display: grid;
  grid-template-columns: 200px 1fr;
  column-gap: 1em;
  margin: 0;
}
.account-info dt {
  font-weight: bold;
}
.account-info dd {
  margin: 0;
}
.member-row {
  margin-top: 1rem;
}
.user-groups {
  margin: 0;
  padding-left: 1.25rem;
}
.user-groups-link {
  margin: 0.35rem 0 0;
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
