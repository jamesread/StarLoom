<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import {
  ArrowLeft01Icon,
  PlusSignIcon,
  Refresh01Icon,
  UserIcon,
} from '@hugeicons/core-free-icons'
import { starapp, type UserAccount, type UserGroup } from '../api/client'
import { useStatus } from '../composables/useStatus'
import { canViewFamilyHomeFromStatus, hasPermission } from '../lib/rbacAccess'

const iconStrokeWidth = 2.5

const users = ref<UserAccount[]>([])
const loading = ref(true)
const error = ref('')
const createUsername = ref('')
const createPassword = ref('')
const createError = ref('')
const creating = ref(false)
const createDialog = ref<HTMLDialogElement | null>(null)
const createUsernameInput = ref<HTMLInputElement | null>(null)
const status = useStatus()

const canCreate = computed(() => hasPermission(status.status, 'users.create'))
const canLinkUserGroups = computed(() => hasPermission(status.status, 'usergroups.view'))
const canLinkPerson = computed(() => canViewFamilyHomeFromStatus(status.status))

const tableHeaders = [
  { key: 'username', label: 'Username', sortable: true },
  { key: 'linkedPerson', label: 'Person', sortable: true },
  { key: 'userGroups', label: 'Usergroups', sortable: false },
  { key: 'createdBy', label: 'Created by', sortable: true, width: '10rem' },
  { key: 'createdAt', label: 'Created', sortable: true, width: '11rem' },
]

function visibleUserGroups(groups?: UserGroup[]) {
  return (groups || []).slice(0, 3)
}

const rows = computed(() =>
  users.value.map((u) => ({
    id: u.id,
    username: u.username,
    linkedPerson: u.linkedMember?.displayName || '',
    linkedMemberId: u.linkedMember?.id || 0,
    userGroups: u.userGroups || [],
    createdBy: u.createdBy || '—',
    createdAt: u.createdAt || '',
  })),
)

function formatDate(value?: string) {
  if (!value) return '—'
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? value : d.toLocaleString()
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await starapp.getUsers()
    users.value = res.users || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Load failed'
  } finally {
    loading.value = false
  }
}

function resetCreateForm() {
  createUsername.value = ''
  createPassword.value = ''
  createError.value = ''
  creating.value = false
}

function openCreateDialog() {
  resetCreateForm()
  createDialog.value?.showModal()
  nextTick(() => createUsernameInput.value?.focus())
}

function closeCreateDialog() {
  createDialog.value?.close()
}

async function submitCreate() {
  if (!createUsername.value.trim()) {
    createError.value = 'Username is required.'
    return
  }
  creating.value = true
  createError.value = ''
  try {
    await starapp.createUser({
      username: createUsername.value.trim(),
      password: createPassword.value || undefined,
    })
    closeCreateDialog()
    await load()
  } catch (e) {
    createError.value = e instanceof Error ? e.message : String(e)
  } finally {
    creating.value = false
  }
}

onMounted(load)
</script>

<template>
  <dialog ref="createDialog" class="dialog" @close="resetCreateForm">
    <h2>Create user</h2>
    <p>Add a sign-in account. Password is optional; set one with at least 8 characters if provided.</p>
    <FormLayout @submit.prevent="submitCreate">
      <FormField label="Username" for="create-username">
        <input
          id="create-username"
          ref="createUsernameInput"
          v-model="createUsername"
          type="text"
          autocomplete="off"
          required
          :disabled="creating"
        />
      </FormField>
      <FormField label="Password" for="create-password" description="Optional; at least 8 characters if set.">
        <input
          id="create-password"
          v-model="createPassword"
          type="password"
          autocomplete="new-password"
          minlength="8"
          :disabled="creating"
        />
      </FormField>
      <p v-if="createError" class="inline-notification error">{{ createError }}</p>
      <template #actions>
        <button type="button" class="neutral" :disabled="creating" @click="closeCreateDialog">Cancel</button>
        <button type="submit" class="good" :disabled="creating || !createUsername.trim()">
          {{ creating ? 'Creating…' : 'Create' }}
        </button>
      </template>
    </FormLayout>
  </dialog>

  <Section
    subtitle="Sign-in accounts for parents, children, and administrators."
    classes="users-list"
    :padding="false"
  >
    <template #title>
      <span class="section-title-with-icon">
        <HugeiconsIcon :icon="UserIcon" width="22" height="22" aria-hidden="true" />
        Users
      </span>
    </template>

    <template #toolbar>
      <RouterLink :to="{ name: 'iam' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>IAM</span>
      </RouterLink>
      <button
        type="button"
        class="inline-icon neutral"
        aria-label="Refresh"
        title="Refresh"
        :disabled="loading"
        @click="load"
      >
        <HugeiconsIcon
          :icon="Refresh01Icon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
      </button>
      <button
        v-if="canCreate"
        type="button"
        class="inline-icon good"
        aria-label="Create user"
        title="Create user"
        :disabled="loading"
        @click="openCreateDialog"
      >
        <HugeiconsIcon
          :icon="PlusSignIcon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
      </button>
    </template>

    <p v-if="error" class="inline-notification error list-banner-pad">{{ error }}</p>
    <div v-if="loading && !users.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!users.length" class="inline-notification note list-banner-pad">No users yet.</p>

      <Table
        v-else
        class="list-table-wrap"
        :data="rows"
        :headers="tableHeaders"
        :show-pagination="users.length > 10"
      >
        <template #cell-username="{ value, row }">
          <RouterLink :to="{ name: 'userInfo', params: { id: row.id } }" class="title-link">
            <strong>{{ value }}</strong>
          </RouterLink>
        </template>
        <template #cell-linkedPerson="{ row }">
          <span v-if="!row.linkedPerson" class="subtle">—</span>
          <RouterLink
            v-else-if="canLinkPerson && row.linkedMemberId"
            :to="{ name: 'familyPersonDetail', params: { id: row.linkedMemberId } }"
            class="title-link"
          >
            {{ row.linkedPerson }}
          </RouterLink>
          <span v-else>{{ row.linkedPerson }}</span>
        </template>
        <template #cell-userGroups="{ row }">
          <span v-if="!row.userGroups.length" class="subtle">—</span>
          <span v-else class="user-group-links">
            <template v-for="(group, index) in visibleUserGroups(row.userGroups)" :key="group.id">
              <span v-if="index > 0" class="user-group-sep">, </span>
              <RouterLink
                v-if="canLinkUserGroups"
                :to="{ name: 'userGroupEdit', params: { id: group.id } }"
                class="title-link"
              >
                {{ group.name }}
              </RouterLink>
              <span v-else>{{ group.name }}</span>
            </template>
          </span>
        </template>
        <template #cell-createdAt="{ value }">
          {{ formatDate(value) }}
        </template>
      </Table>
    </template>
  </Section>
</template>

<style scoped>
.section-title-with-icon {
  display: inline-flex;
  align-items: center;
  gap: 0.45em;
  vertical-align: middle;
}

.list-banner-pad {
  padding-left: 1em;
  padding-right: 1em;
}

.list-table-wrap {
  margin-top: 0.5rem;
  margin-bottom: 1.5rem;
}

.title-link {
  color: var(--pico-primary);
  text-decoration: none;
}

.title-link:hover {
  text-decoration: underline;
}

.user-group-links {
  display: inline;
}

.user-group-sep {
  color: inherit;
}
</style>
