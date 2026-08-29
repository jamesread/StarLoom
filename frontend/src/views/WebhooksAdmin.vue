<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, PlusSignIcon } from '@hugeicons/core-free-icons'
import { starapp, type Webhook } from '../api/client'

const webhooks = ref<Webhook[]>([])
const error = ref('')
const loading = ref(true)

const listHeaders = [
  { key: 'id', label: 'ID', sortable: true, width: '4rem' },
  { key: 'url', label: 'URL', sortable: true },
  { key: 'events', label: 'Events', sortable: false },
  { key: 'status', label: 'Status', sortable: true, width: '6rem' },
  { key: 'actions', label: 'Actions', sortable: false, width: '6rem' },
]

const listRows = computed(() =>
  webhooks.value.map((wh) => ({
    id: wh.id,
    url: wh.url,
    events: (wh.events || []).join(', '),
    status: wh.enabled ? 'Enabled' : 'Disabled',
    actions: '',
  })),
)

async function load() {
  loading.value = true
  try {
    const res = await starapp.listWebhooks()
    webhooks.value = res.webhooks || []
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
  <Section title="Webhooks" subtitle="HTTP callbacks for StarLoom events" :padding="false">
    <template #toolbar>
      <router-link :to="{ name: 'controlPanel' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Control Panel</span>
      </router-link>
      <RouterLink
        :to="{ name: 'webhook-create' }"
        class="button"
        title="Add webhook"
        aria-label="Add webhook"
      >
        <HugeiconsIcon :icon="PlusSignIcon" width="1em" height="1em" />
      </RouterLink>
    </template>

    <p v-if="error" class="inline-notification error list-banner-pad">{{ error }}</p>
    <div v-if="loading && !webhooks.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!webhooks.length" class="inline-notification note list-banner-pad">
        No webhooks configured yet.
      </p>

      <Table
        v-else
        class="list-table-wrap"
        :data="listRows"
        :headers="listHeaders"
        :show-pagination="webhooks.length > 10"
      >
        <template #cell-url="{ value, row }">
          <RouterLink :to="{ name: 'webhook-edit', params: { id: row.id } }" class="title-link">
            <code>{{ value }}</code>
          </RouterLink>
        </template>
        <template #cell-actions="{ row }">
          <div class="actions-cell">
            <RouterLink :to="{ name: 'webhook-edit', params: { id: row.id } }" class="button neutral small">
              Edit
            </RouterLink>
          </div>
        </template>
      </Table>
    </template>
  </Section>
</template>

<style scoped>
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
  color: var(--pico-primary);
  text-decoration: none;
}

.title-link:hover {
  text-decoration: underline;
}
</style>
