<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import CheckGroup from 'picocrank/vue/components/CheckGroup.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, PlusSignIcon } from '@hugeicons/core-free-icons'
import { starapp, type Webhook } from '../api/client'

const webhooks = ref<Webhook[]>([])
const catalog = ref<string[]>([])
const edits = reactive<Record<number, {
  url: string
  secret: string
  events: string[]
  enabled: boolean
}>>({})
const editingId = ref<number | null>(null)
const error = ref('')

const booleanOptions = [
  { label: 'On', value: true },
  { label: 'Off', value: false },
]

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

const eventOptions = computed(() =>
  catalog.value.map((e) => ({ label: e, value: e })),
)

function syncEdits() {
  for (const key of Object.keys(edits)) delete edits[Number(key)]
  for (const wh of webhooks.value) {
    edits[wh.id] = {
      url: wh.url,
      secret: '',
      events: [...(wh.events || [])],
      enabled: !!wh.enabled,
    }
  }
  if (editingId.value && !edits[editingId.value]) editingId.value = null
}

async function load() {
  try {
    const res = await starapp.listWebhooks()
    webhooks.value = res.webhooks || []
    catalog.value = res.events?.length ? res.events : []
    syncEdits()
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function update(id: number) {
  const e = edits[id]
  await starapp.updateWebhook({
    id,
    url: e.url,
    secret: e.secret || '',
    events: e.events,
    enabled: e.enabled,
  })
  editingId.value = null
  await load()
}

async function destroy(id: number) {
  if (!confirm('Delete this webhook?')) return
  await starapp.deleteWebhook({ id })
  editingId.value = null
  await load()
}

onMounted(load)
watch(webhooks, syncEdits, { deep: true })
</script>

<template>
  <Section title="Webhooks" subtitle="HTTP callbacks for StarApp events" :padding="false">
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

    <p v-if="error" class="form-error padding">{{ error }}</p>

    <Table
      v-if="webhooks.length > 0"
      :data="listRows"
      :headers="listHeaders"
      :show-pagination="webhooks.length > 10"
    >
      <template #cell-url="{ value }">
        <code>{{ value }}</code>
      </template>
      <template #cell-actions="{ row }">
        <button type="button" class="button" @click="editingId = row.id">Edit</button>
      </template>
    </Table>
    <p v-else class="padding subtle">No webhooks configured yet.</p>
  </Section>

  <Section
    v-if="editingId && edits[editingId]"
    :title="`Edit webhook #${editingId}`"
    :padding="true"
  >
    <form class="form-stack" @submit.prevent="update(editingId)">
      <label>
        URL
        <input v-model="edits[editingId].url" type="url" required>
      </label>
      <label>
        Secret
        <input v-model="edits[editingId].secret" type="text" placeholder="Leave blank to keep current secret">
      </label>
      <FormField label="Events" fake>
        <CheckGroup
          v-model="edits[editingId].events"
          :options="eventOptions"
          :name="`webhook-events-${editingId}`"
          aria-label="Webhook events"
        />
      </FormField>
      <FormField label="Enabled" fake>
        <RadioGroup
          v-model="edits[editingId].enabled"
          :name="`webhook-enabled-${editingId}`"
          variant="boolean"
          :options="booleanOptions"
          aria-label="Webhook enabled"
        />
      </FormField>
      <div class="form-actions">
        <button type="submit" class="button">Save</button>
        <button type="button" class="button" @click="editingId = null">Cancel</button>
        <button type="button" class="button danger" @click="destroy(editingId)">Delete</button>
      </div>
    </form>
  </Section>
</template>
