<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { GiftIcon, PlusSignIcon, Refresh01Icon } from '@hugeicons/core-free-icons'
import { starapp, type Reward } from '../api/client'

const iconStrokeWidth = 2.5

const rewards = ref<Reward[]>([])
const error = ref('')
const createError = ref('')
const loading = ref(true)
const creating = ref(false)

const createDialog = ref<HTMLDialogElement | null>(null)
const createTitleInput = ref<HTMLInputElement | null>(null)

const form = reactive({
  title: '',
  description: '',
  costStars: 5,
  approvalRequired: true,
})

const booleanOptions = [
  { label: 'Yes', value: true },
  { label: 'No', value: false },
]

const tableHeaders = [
  { key: 'title', label: 'Title', sortable: true },
  { key: 'costStars', label: 'Cost', sortable: true, width: '6rem' },
  { key: 'status', label: 'Status', sortable: true, width: '6rem' },
  { key: 'approval', label: 'Approval', sortable: true, width: '8rem' },
  { key: 'actions', label: 'Actions', sortable: false, width: '8rem' },
]

const listRows = computed(() =>
  rewards.value.map((r) => ({
    id: r.id,
    title: r.title,
    costStars: r.costStars,
    status: r.active !== false ? 'Active' : 'Inactive',
    approval: r.approvalRequired !== false ? 'Required' : 'Auto',
    active: r.active !== false,
    actions: '',
  })),
)

function resetCreateForm() {
  form.title = ''
  form.description = ''
  form.costStars = 5
  form.approvalRequired = true
}

async function load() {
  loading.value = true
  try {
    const res = await starapp.listRewards({ includeInactive: true })
    rewards.value = res.rewards || []
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
  nextTick(() => createTitleInput.value?.focus())
}

function closeCreateDialog() {
  createDialog.value?.close()
}

async function createReward() {
  creating.value = true
  createError.value = ''
  try {
    await starapp.createReward({
      title: form.title.trim(),
      description: form.description.trim(),
      costStars: form.costStars,
      approvalRequired: form.approvalRequired,
    })
    closeCreateDialog()
    await load()
  } catch (e) {
    createError.value = e instanceof Error ? e.message : String(e)
  } finally {
    creating.value = false
  }
}

async function deactivate(id: number) {
  if (!confirm('Deactivate this reward?')) return
  try {
    await starapp.deleteReward({ id })
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(load)
</script>

<template>
  <dialog ref="createDialog" class="dialog" @close="resetCreateForm">
    <h2>Add reward</h2>
    <p>Define a privilege children can redeem with stars.</p>
    <form class="dialog-form" @submit.prevent="createReward">
      <FormField label="Title" for="reward-title">
        <input
          id="reward-title"
          ref="createTitleInput"
          v-model="form.title"
          type="text"
          required
          placeholder="Extra screen time"
          :disabled="creating"
        />
      </FormField>
      <FormField label="Description" for="reward-description">
        <textarea id="reward-description" v-model="form.description" rows="2" :disabled="creating" />
      </FormField>
      <FormField label="Cost (stars)" for="reward-cost">
        <input
          id="reward-cost"
          v-model.number="form.costStars"
          type="number"
          min="1"
          required
          :disabled="creating"
        />
      </FormField>
      <FormField label="Requires approval" fake>
        <RadioGroup v-model="form.approvalRequired" :options="booleanOptions" name="reward-approval" />
      </FormField>
      <p v-if="createError" class="inline-notification error">{{ createError }}</p>
      <div class="dialog-actions">
        <button type="button" class="neutral" :disabled="creating" @click="closeCreateDialog">Cancel</button>
        <button type="submit" class="good" :disabled="creating || !form.title.trim()">
          {{ creating ? 'Creating…' : 'Create' }}
        </button>
      </div>
    </form>
  </dialog>

  <Section
    subtitle="Privileges children can redeem with stars."
    classes="rewards-list"
    :padding="false"
  >
    <template #title>
      <span class="section-title-with-icon">
        <HugeiconsIcon :icon="GiftIcon" width="22" height="22" aria-hidden="true" />
        Rewards
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
        aria-label="Add reward"
        title="Add reward"
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
    <div v-if="loading && !rewards.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!rewards.length" class="inline-notification note list-banner-pad">
        No rewards yet. Use <strong>+</strong> to add one.
      </p>

      <Table
        v-else
        class="list-table-wrap"
        :data="listRows"
        :headers="tableHeaders"
        :show-pagination="rewards.length > 10"
      >
        <template #cell-title="{ value, row }">
          <RouterLink :to="{ name: 'familyRewardEdit', params: { id: row.id } }" class="title-link">
            <strong>{{ value }}</strong>
          </RouterLink>
        </template>
        <template #cell-actions="{ row }">
          <div class="actions-cell">
            <RouterLink :to="{ name: 'familyRewardEdit', params: { id: row.id } }" class="button neutral small">
              Edit
            </RouterLink>
            <button v-if="row.active" type="button" class="bad small" @click="deactivate(row.id)">
              Deactivate
            </button>
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
