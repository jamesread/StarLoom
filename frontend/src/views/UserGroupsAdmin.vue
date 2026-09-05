<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import {
  ArrowLeft01Icon,
  PlusSignIcon,
  Refresh01Icon,
  UserGroupIcon,
} from '@hugeicons/core-free-icons'
import { starapp, type UserGroup } from '../api/client'
import { useStatus } from '../composables/useStatus'
import { hasPermission } from '../lib/rbacAccess'

const iconStrokeWidth = 2.5

const route = useRoute()
const router = useRouter()
const status = useStatus()

const groups = ref<UserGroup[]>([])
const loading = ref(true)
const error = ref('')
const createName = ref('')
const createError = ref('')
const creating = ref(false)
const createDialog = ref<HTMLDialogElement | null>(null)
const createNameInput = ref<HTMLInputElement | null>(null)

const canCreate = computed(() => hasPermission(status.status, 'usergroups.manage'))
const canManage = computed(() => hasPermission(status.status, 'usergroups.manage'))

const tableHeaders = computed(() => {
  const headers = [
    { key: 'name', label: 'Name', sortable: true },
    { key: 'memberCount', label: 'Members', sortable: true, width: '8rem' },
  ]
  if (canManage.value) {
    headers.push({ key: 'actions', label: 'Actions', sortable: false, width: '8rem' })
  }
  return headers
})

const rows = computed(() =>
  groups.value.map((g) => ({
    id: g.id,
    name: g.name,
    memberCount: g.memberCount ?? 0,
    actions: '',
  })),
)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await starapp.listUserGroups()
    groups.value = res.groups || []
    const queryGroupId = Number(route.query.groupId)
    if (queryGroupId && groups.value.some((group) => group.id === queryGroupId)) {
      await router.replace({ name: 'userGroupEdit', params: { id: queryGroupId } })
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Load failed'
  } finally {
    loading.value = false
  }
}

function resetCreateForm() {
  createName.value = ''
  createError.value = ''
  creating.value = false
}

function openCreateDialog() {
  resetCreateForm()
  createDialog.value?.showModal()
  nextTick(() => createNameInput.value?.focus())
}

function closeCreateDialog() {
  createDialog.value?.close()
}

async function submitCreate() {
  const name = createName.value.trim()
  if (!name) {
    createError.value = 'Name is required.'
    return
  }
  creating.value = true
  createError.value = ''
  try {
    const res = await starapp.createUserGroup({ name })
    closeCreateDialog()
    if (res.group?.id) {
      router.push({ name: 'userGroupEdit', params: { id: res.group.id } })
      return
    }
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
    <h2>Create user group</h2>
    <p>Add a group for organizing users and assigning roles.</p>
    <FormLayout @submit.prevent="submitCreate">
      <FormField label="Name" for="create-user-group-name">
        <input
          id="create-user-group-name"
          ref="createNameInput"
          v-model="createName"
          type="text"
          autocomplete="off"
          required
          :disabled="creating"
        />
      </FormField>
      <p v-if="createError" class="inline-notification error">{{ createError }}</p>
      <template #actions>
        <button type="button" class="neutral" :disabled="creating" @click="closeCreateDialog">Cancel</button>
        <button type="submit" class="good" :disabled="creating || !createName.trim()">
          {{ creating ? 'Creating…' : 'Create' }}
        </button>
      </template>
    </FormLayout>
  </dialog>

  <Section
    subtitle="Groups organize users and receive role assignments that grant permissions."
    classes="user-groups-list"
    :padding="false"
  >
    <template #title>
      <span class="section-title-with-icon">
        <HugeiconsIcon :icon="UserGroupIcon" width="22" height="22" aria-hidden="true" />
        User groups
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
        aria-label="Create user group"
        title="Create user group"
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
    <div v-if="loading && !groups.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!groups.length" class="inline-notification note list-banner-pad">No user groups yet.</p>

      <Table
        v-else
        class="list-table-wrap"
        :data="rows"
        :headers="tableHeaders"
        :show-pagination="groups.length > 10"
      >
        <template #cell-name="{ value, row }">
          <RouterLink :to="{ name: 'userGroupEdit', params: { id: row.id } }" class="title-link">
            <strong>{{ value }}</strong>
          </RouterLink>
        </template>
        <template #cell-actions="{ row }">
          <div class="actions-cell">
            <RouterLink
              :to="{ name: 'userGroupEdit', params: { id: row.id } }"
              class="button neutral small"
            >
              Edit
            </RouterLink>
          </div>
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

.actions-cell {
  text-align: right;
}
</style>
