<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { PlusSignIcon, Refresh01Icon, StarIcon } from '@hugeicons/core-free-icons'
import { starapp, type StarChart } from '../api/client'

const iconStrokeWidth = 2.5

const charts = ref<StarChart[]>([])
const error = ref('')
const loading = ref(true)

const tableHeaders = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'choreCount', label: 'Chores', sortable: true, width: '7rem' },
  { key: 'status', label: 'Status', sortable: true, width: '7rem' },
  { key: 'actions', label: 'Actions', sortable: false, width: '10rem' },
]

const listRows = computed(() =>
  charts.value.map((c) => ({
    id: c.id,
    name: c.name,
    choreCount: c.choreCount ?? 0,
    status: c.active !== false ? 'Active' : 'Inactive',
    active: c.active !== false,
    actions: '',
  })),
)

async function load() {
  loading.value = true
  try {
    const res = await starapp.listStarCharts({ includeInactive: true })
    charts.value = res.starCharts || []
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <Section
    subtitle="Each chart shows a weekly grid for its assigned chores."
    classes="star-charts-list"
    :padding="false"
  >
    <template #title>
      <span class="section-title-with-icon">
        <HugeiconsIcon :icon="StarIcon" width="22" height="22" aria-hidden="true" />
        Star Charts
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
        :to="{ name: 'familyStarChartCreate' }"
        class="button inline-icon good"
        aria-label="Add star chart"
        title="Add star chart"
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
    <div v-if="loading && !charts.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!charts.length" class="inline-notification note list-banner-pad">
        No star charts yet.
        <RouterLink :to="{ name: 'familyStarChartCreate' }">Add a star chart</RouterLink>.
      </p>

      <Table
        v-else
        class="list-table-wrap"
        :data="listRows"
        :headers="tableHeaders"
        row-hover
        default-sort-key="name"
      >
        <template #cell-name="{ row }">
          <RouterLink :to="{ name: 'familyStarChartView', params: { id: row.id } }" class="title-link">
            <strong>{{ row.name }}</strong>
          </RouterLink>
        </template>
        <template #cell-actions="{ row }">
          <RouterLink :to="{ name: 'familyStarChartEdit', params: { id: row.id } }">Edit</RouterLink>
          <RouterLink :to="{ name: 'familyStarChartView', params: { id: row.id } }">View</RouterLink>
        </template>
      </Table>
    </template>
  </Section>
</template>
