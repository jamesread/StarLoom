<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { PlusSignIcon, Refresh01Icon, UserMultipleIcon } from '@hugeicons/core-free-icons'
import { starapp, memberAvatarUrl, type FamilyMember } from '../api/client'
import { memberAvatarStyle, memberStarColor } from '../lib/memberStarColor'

const iconStrokeWidth = 2.5

const members = ref<FamilyMember[]>([])
const error = ref('')
const createError = ref('')
const loading = ref(true)
const creating = ref(false)

const createDialog = ref<HTMLDialogElement | null>(null)
const createNameInput = ref<HTMLInputElement | null>(null)

const form = reactive({
  displayName: '',
  username: '',
  password: '',
})

const tableHeaders = [
  { key: 'displayName', label: 'Name', sortable: true },
  { key: 'username', label: 'Username', sortable: true },
  { key: 'starColor', label: 'Star color', sortable: false, width: '7rem' },
  { key: 'actions', label: 'Actions', sortable: false, width: '8rem' },
]

const listRows = computed(() =>
  members.value.map((m) => ({
    id: m.id,
    displayName: m.displayName,
    username: m.username || '—',
    starColor: memberStarColor(m),
    member: m,
    actions: '',
  })),
)

function resetCreateForm() {
  form.displayName = ''
  form.username = ''
  form.password = ''
}

async function load() {
  loading.value = true
  try {
    const res = await starapp.listMembers()
    members.value = (res.members || []).filter((m) => m.role === 'child')
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  createError.value = ''
  resetCreateForm()
  createDialog.value?.showModal()
  nextTick(() => createNameInput.value?.focus())
}

function closeCreateDialog() {
  createDialog.value?.close()
}

async function createChild() {
  creating.value = true
  createError.value = ''
  try {
    await starapp.createChildMember({
      displayName: form.displayName.trim(),
      username: form.username.trim(),
      password: form.password,
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
    <h2>Add child</h2>
    <p>Create a login for a child in your family. A star color is assigned automatically.</p>
    <form class="dialog-form" @submit.prevent="createChild">
      <FormField label="Display name" for="child-display-name">
        <input
          id="child-display-name"
          ref="createNameInput"
          v-model="form.displayName"
          type="text"
          required
          placeholder="Alex"
          :disabled="creating"
        />
      </FormField>
      <FormField label="Username" for="child-username">
        <input
          id="child-username"
          v-model="form.username"
          type="text"
          required
          autocomplete="off"
          :disabled="creating"
        />
      </FormField>
      <FormField label="Password" for="child-password" description="At least 8 characters.">
        <input
          id="child-password"
          v-model="form.password"
          type="password"
          required
          minlength="8"
          autocomplete="new-password"
          :disabled="creating"
        />
      </FormField>
      <p v-if="createError" class="inline-notification error">{{ createError }}</p>
      <div class="dialog-actions">
        <button type="button" class="neutral" :disabled="creating" @click="closeCreateDialog">Cancel</button>
        <button type="submit" class="good" :disabled="creating || !form.displayName.trim() || !form.username.trim()">
          {{ creating ? 'Creating…' : 'Create' }}
        </button>
      </div>
    </form>
  </dialog>

  <Section
    subtitle="Add children with their own login accounts."
    classes="children-list"
    :padding="false"
  >
    <template #title>
      <span class="section-title-with-icon">
        <HugeiconsIcon :icon="UserMultipleIcon" width="22" height="22" aria-hidden="true" />
        Children
      </span>
    </template>

    <template #toolbar>
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
        type="button"
        class="inline-icon good"
        aria-label="Add child"
        title="Add child"
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
    <div v-if="loading && !members.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!members.length" class="inline-notification note list-banner-pad">
        No children yet. Use <strong>+</strong> to add one.
      </p>

      <Table
        v-else
        class="list-table-wrap"
        :data="listRows"
        :headers="tableHeaders"
        :show-pagination="members.length > 10"
      >
        <template #cell-displayName="{ value, row }">
          <RouterLink :to="{ name: 'familyChildDetail', params: { id: row.id } }" class="title-link name-cell">
            <img
              v-if="row.member?.hasAvatar"
              class="row-avatar"
              :style="memberAvatarStyle(row.member)"
              :src="memberAvatarUrl(row.member.id, true)"
              :alt="value"
            />
            <span v-else class="row-avatar row-avatar-placeholder" :style="memberAvatarStyle(row.member)">
              {{ String(value).charAt(0) }}
            </span>
            <strong>{{ value }}</strong>
          </RouterLink>
        </template>
        <template #cell-starColor="{ value }">
          <span class="color-swatch" :style="{ backgroundColor: value }" :title="value" />
        </template>
        <template #cell-actions="{ row }">
          <div class="actions-cell">
            <RouterLink :to="{ name: 'familyChildDetail', params: { id: row.id } }" class="button neutral small">
              Manage
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

.actions-cell {
  text-align: right;
}

.title-link {
  font: inherit;
  color: var(--pico-primary);
  text-decoration: none;
}

.title-link:hover {
  text-decoration: underline;
}

.name-cell {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.row-avatar {
  width: 2rem;
  height: 2rem;
  border-radius: 50%;
  object-fit: cover;
  box-sizing: border-box;
  flex-shrink: 0;
}

.row-avatar-placeholder {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--pico-muted-border-color);
  font-size: 0.85rem;
  font-weight: 600;
}

.color-swatch {
  display: inline-block;
  width: 1.25rem;
  height: 1.25rem;
  border-radius: 50%;
  vertical-align: middle;
  border: 1px solid color-mix(in srgb, var(--pico-color) 20%, transparent);
}

.dialog-form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
