<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, PlusSignIcon, Refresh01Icon, TestTubeIcon, WebhookIcon } from '@hugeicons/core-free-icons'
import { starapp, type Webhook, type WebhookDelivery } from '../api/client'

const iconStrokeWidth = 2.5

const webhooks = ref<Webhook[]>([])
const deliveries = ref<WebhookDelivery[]>([])
const error = ref('')
const historyError = ref('')
const loading = ref(true)
const historyLoading = ref(true)
const testLoading = ref(false)
const testMessage = ref('')
const testError = ref('')

const definitionHeaders = [
  { key: 'id', label: 'ID', sortable: true, width: '4rem' },
  { key: 'url', label: 'URL', sortable: true },
  { key: 'events', label: 'Events', sortable: false },
  { key: 'status', label: 'Status', sortable: true, width: '6rem' },
  { key: 'actions', label: 'Actions', sortable: false, width: '6rem' },
]

const historyHeaders = [
  { key: 'firedAt', label: 'Fired at', sortable: true, width: '11rem' },
  { key: 'event', label: 'Event', sortable: true, width: '10rem' },
  { key: 'url', label: 'URL', sortable: true },
  { key: 'result', label: 'Result', sortable: true, width: '7rem' },
  { key: 'httpStatus', label: 'HTTP', sortable: true, width: '5rem' },
  { key: 'errorMessage', label: 'Details', sortable: false },
]

const definitionRows = computed(() =>
  webhooks.value.map((wh) => ({
    id: wh.id,
    url: wh.url,
    events: (wh.events || []).join(', '),
    status: wh.enabled ? 'Enabled' : 'Disabled',
    actions: '',
  })),
)

const historyRows = computed(() =>
  deliveries.value.map((entry) => ({
    id: entry.id,
    firedAt: entry.firedAt || '—',
    event: entry.event,
    url: entry.url,
    result: entry.success ? 'Success' : 'Failed',
    httpStatus: entry.httpStatus ? String(entry.httpStatus) : '—',
    errorMessage: entry.errorMessage || '—',
    success: entry.success === true,
  })),
)

async function loadDefinitions() {
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

async function loadHistory() {
  historyLoading.value = true
  try {
    const res = await starapp.listWebhookDeliveries({ limit: 100 })
    deliveries.value = res.deliveries || []
    historyError.value = ''
  } catch (e) {
    historyError.value = e instanceof Error ? e.message : String(e)
  } finally {
    historyLoading.value = false
  }
}

async function load() {
  await Promise.all([loadDefinitions(), loadHistory()])
}

async function fireTest() {
  testLoading.value = true
  testMessage.value = ''
  testError.value = ''
  try {
    const res = await starapp.fireTestWebhooks()
    testMessage.value = res.standardResponse?.message || 'Test webhook sent'
    await loadHistory()
  } catch (e) {
    testError.value = e instanceof Error ? e.message : String(e)
  } finally {
    testLoading.value = false
  }
}

onMounted(load)
</script>

<template>
  <Section title="Webhook definitions" subtitle="HTTP callbacks for StarLoom events" :padding="false">
    <template #toolbar>
      <router-link :to="{ name: 'controlPanel' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Control Panel</span>
      </router-link>
      <button
        type="button"
        class="inline-icon neutral"
        aria-label="Refresh"
        title="Refresh"
        :disabled="loading || historyLoading"
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
        class="inline-icon neutral"
        :disabled="testLoading"
        @click="fireTest"
      >
        <HugeiconsIcon
          :icon="TestTubeIcon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>Send test</span>
      </button>
      <RouterLink
        :to="{ name: 'webhook-create' }"
        class="button inline-icon good"
        title="Add webhook"
        aria-label="Add webhook"
      >
        <HugeiconsIcon :icon="PlusSignIcon" width="1em" height="1em" :strokeWidth="iconStrokeWidth" aria-hidden="true" />
      </RouterLink>
    </template>

    <p v-if="testError" class="inline-notification error list-banner-pad">{{ testError }}</p>
    <p v-else-if="testMessage" class="inline-notification note list-banner-pad">{{ testMessage }}</p>
    <p v-if="error" class="inline-notification error list-banner-pad">{{ error }}</p>
    <div v-if="loading && !webhooks.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!webhooks.length" class="inline-notification note list-banner-pad">
        No webhooks configured yet.
      </p>

      <Table
        v-else
        class="list-table-wrap"
        :data="definitionRows"
        :headers="definitionHeaders"
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

  <Section
    title="Webhook history"
    subtitle="Recent outbound delivery attempts."
    :icon="WebhookIcon"
    :padding="false"
  >
    <template #toolbar>
      <button
        type="button"
        class="inline-icon neutral"
        aria-label="Refresh history"
        title="Refresh history"
        :disabled="historyLoading"
        @click="loadHistory"
      >
        <HugeiconsIcon
          :icon="Refresh01Icon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
      </button>
    </template>

    <p v-if="historyError" class="inline-notification error list-banner-pad">{{ historyError }}</p>
    <div v-if="historyLoading && !deliveries.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!deliveries.length" class="inline-notification note list-banner-pad">
        No webhook deliveries recorded yet.
      </p>

      <Table
        v-else
        class="list-table-wrap"
        :data="historyRows"
        :headers="historyHeaders"
        :show-pagination="deliveries.length > 10"
      >
        <template #cell-url="{ value }">
          <code class="url-cell">{{ value }}</code>
        </template>
        <template #cell-result="{ value, row }">
          <span :class="row.success ? 'result-good' : 'result-bad'">{{ value }}</span>
        </template>
        <template #cell-errorMessage="{ value }">
          <span class="details-cell" :title="value">{{ value }}</span>
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

.url-cell,
.details-cell {
  word-break: break-all;
}

.details-cell {
  display: inline-block;
  max-width: 18rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-good {
  color: var(--karma-good, #2e7d32);
  font-weight: 600;
}

.result-bad {
  color: var(--karma-bad, #c62828);
  font-weight: 600;
}
</style>
