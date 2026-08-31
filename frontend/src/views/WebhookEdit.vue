<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import CheckGroup from 'picocrank/vue/components/CheckGroup.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, WebhookIcon } from '@hugeicons/core-free-icons'
import { starapp, type Webhook } from '../api/client'
import { useInit } from '../composables/useInit'

const route = useRoute()
const router = useRouter()
const init = useInit()
const webhookId = computed(() => Number(route.params.id))

const webhook = ref<Webhook | null>(null)
const error = ref('')
const saving = ref(false)
const loading = ref(true)

const form = reactive({
  url: '',
  secret: '',
  events: [] as string[],
  enabled: true,
})

const booleanOptions = [
  { label: 'On', value: true },
  { label: 'Off', value: false },
]

const eventOptions = computed(() =>
  (init.init?.webhookEvents?.length ? init.init.webhookEvents : ['stars.awarded']).map((e) => ({
    label: e,
    value: e,
  })),
)

const sectionTitle = computed(() =>
  webhook.value ? `Edit webhook #${webhook.value.id}` : 'Edit webhook',
)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await starapp.listWebhooks()
    webhook.value = (res.webhooks || []).find((wh) => wh.id === webhookId.value) || null
    if (!webhook.value) {
      error.value = 'Webhook not found'
      return
    }
    form.url = webhook.value.url
    form.secret = ''
    form.events = [...(webhook.value.events || [])]
    form.enabled = !!webhook.value.enabled
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!webhook.value) return
  saving.value = true
  error.value = ''
  try {
    await starapp.updateWebhook({
      id: webhook.value.id,
      url: form.url,
      secret: form.secret || '',
      events: form.events,
      enabled: form.enabled,
    })
    router.push({ name: 'webhooks' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

async function destroy() {
  if (!webhook.value || !confirm('Delete this webhook?')) return
  try {
    await starapp.deleteWebhook({ id: webhook.value.id })
    router.push({ name: 'webhooks' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(load)
</script>

<template>
  <Section :title="sectionTitle" :icon="WebhookIcon" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'webhooks' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Webhooks</span>
      </RouterLink>
    </template>

    <p v-if="loading" class="subtle">Loading…</p>
    <p v-else-if="error && !webhook" class="inline-notification error">{{ error }}</p>

    <FormLayout v-else @submit.prevent="save">
      <p v-if="error" class="inline-notification error">{{ error }}</p>
      <FormField label="URL" for="webhook-edit-url">
        <input id="webhook-edit-url" v-model="form.url" type="url" required />
      </FormField>
      <FormField label="Secret" for="webhook-edit-secret" description="Leave blank to keep the current secret.">
        <input id="webhook-edit-secret" v-model="form.secret" type="text" />
      </FormField>
      <FormField label="Events" component-has-label>
        <CheckGroup
          v-model="form.events"
          :options="eventOptions"
          name="webhook-events-edit"
          aria-label="Webhook events"
        />
      </FormField>
      <FormField label="Enabled" component-has-label>
        <RadioGroup
          v-model="form.enabled"
          name="webhook-enabled-edit"
          variant="boolean"
          :options="booleanOptions"
          aria-label="Webhook enabled"
        />
      </FormField>
      <template #actions>
        <button type="submit" class="good" :disabled="saving">{{ saving ? 'Saving…' : 'Save' }}</button>
        <RouterLink :to="{ name: 'webhooks' }" class="button neutral">Cancel</RouterLink>
        <button type="button" class="bad" @click="destroy">Delete</button>
      </template>
    </FormLayout>
  </Section>
</template>
