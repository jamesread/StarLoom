<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { GiftIcon, PlusSignIcon, Refresh01Icon } from '@hugeicons/core-free-icons'
import { starapp, type Reward } from '../api/client'

const iconStrokeWidth = 2.5

const rewards = ref<Reward[]>([])
const error = ref('')
const loading = ref(true)

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
      <RouterLink
        :to="{ name: 'familyRewardCreate' }"
        class="button inline-icon good"
        aria-label="Add reward"
        title="Add reward"
      >
        <HugeiconsIcon
          :icon="PlusSignIcon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
      </RouterLink>
    </template>

    <p v-if="error" class="inline-notification error list-banner-pad">{{ error }}</p>
    <div v-if="loading && !rewards.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!rewards.length" class="inline-notification note list-banner-pad">
        No rewards yet.
        <RouterLink :to="{ name: 'familyRewardCreate' }">Add a reward</RouterLink>.
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
</style>
