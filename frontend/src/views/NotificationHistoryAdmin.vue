<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, Notification03Icon, Refresh01Icon } from '@hugeicons/core-free-icons'
import { starapp, type NotificationDelivery } from '../api/client'

const iconStrokeWidth = 2.5

const deliveries = ref<NotificationDelivery[]>([])
const error = ref('')
const loading = ref(true)

const headers = [
  { key: 'sentAt', label: 'Sent at', sortable: true, width: '11rem' },
  { key: 'recipient', label: 'Recipient', sortable: true, width: '10rem' },
  { key: 'notificationType', label: 'Type', sortable: true, width: '10rem' },
  { key: 'title', label: 'Title', sortable: true },
  { key: 'status', label: 'Status', sortable: true, width: '7rem' },
  { key: 'errorMessage', label: 'Details', sortable: false },
]

function notificationTypeLabel(value?: string) {
  switch (value) {
    case 'test':
      return 'Test'
    case 'chore_completed':
      return 'Chore completed'
    case 'redemption_requested':
      return 'Redemption requested'
    default:
      return value || '—'
  }
}

const rows = computed(() =>
  deliveries.value.map((entry) => ({
    id: entry.id,
    sentAt: entry.sentAt || '—',
    recipient: entry.recipientDisplayName || (entry.recipientMemberId ? `Person #${entry.recipientMemberId}` : '—'),
    recipientMemberId: entry.recipientMemberId,
    notificationType: notificationTypeLabel(entry.notificationType),
    title: entry.title || '—',
    status: entry.success ? 'Sent' : 'Failed',
    errorMessage: entry.errorMessage || '—',
    success: entry.success === true,
  })),
)

async function load() {
  loading.value = true
  try {
    const res = await starapp.listNotificationDeliveries({ limit: 30 })
    deliveries.value = res.deliveries || []
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
    title="Notification history"
    subtitle="The 30 most recent Apprise notifications sent by StarApp."
    :icon="Notification03Icon"
    :padding="false"
  >
    <template #toolbar>
      <RouterLink :to="{ name: 'controlPanel' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Control Panel</span>
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
    </template>

    <p v-if="error" class="inline-notification error list-banner-pad">{{ error }}</p>
    <div v-if="loading && !deliveries.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!deliveries.length" class="inline-notification note list-banner-pad">
        No notifications recorded yet.
      </p>

      <Table
        v-else
        class="list-table-wrap"
        :data="rows"
        :headers="headers"
        :show-pagination="false"
      >
        <template #cell-recipient="{ value, row }">
          <RouterLink
            v-if="row.recipientMemberId"
            :to="{ name: 'familyPersonDetail', params: { id: row.recipientMemberId } }"
            class="title-link"
          >
            <strong>{{ value }}</strong>
          </RouterLink>
          <span v-else>{{ value }}</span>
        </template>
        <template #cell-status="{ value, row }">
          <span class="tag" :class="row.success ? 'good' : 'bad'">{{ value }}</span>
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

.details-cell {
  display: inline-block;
  max-width: 18rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  word-break: break-all;
}

.title-link {
  color: var(--pico-primary);
  text-decoration: none;
}
.title-link:hover {
  text-decoration: underline;
}
</style>
