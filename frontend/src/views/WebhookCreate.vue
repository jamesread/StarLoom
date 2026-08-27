<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import CheckGroup from 'picocrank/vue/components/CheckGroup.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import { starapp } from '../api/client'
import { useInit } from '../composables/useInit'

const router = useRouter()
const init = useInit()

const url = ref('')
const secret = ref('')
const selectedEvents = ref<string[]>(['stars.awarded'])
const enabled = ref(true)
const error = ref('')

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

async function submit() {
  error.value = ''
  try {
    await starapp.createWebhook({
      url: url.value,
      secret: secret.value,
      events: selectedEvents.value,
      enabled: enabled.value,
    })
    router.push({ name: 'webhooks' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}
</script>

<template>
  <Section title="Add webhook" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'webhooks' }" class="button">Back</RouterLink>
    </template>
    <form class="form-stack" @submit.prevent="submit">
      <label>
        URL
        <input v-model="url" type="url" required placeholder="https://example.com/hooks/starapp">
      </label>
      <label>
        Secret
        <input v-model="secret" type="text" required placeholder="Shared signing secret">
      </label>
      <FormField label="Events" fake>
        <div>
          <CheckGroup
            v-model="selectedEvents"
            :options="eventOptions"
            name="webhook-events-create"
            aria-label="Webhook events"
          />
          <p class="subtle">Select one or more events that should POST to this URL.</p>
        </div>
      </FormField>
      <FormField label="Enabled" fake>
        <RadioGroup
          v-model="enabled"
          name="webhook-enabled-create"
          variant="boolean"
          :options="booleanOptions"
          aria-label="Webhook enabled"
        />
      </FormField>
      <p v-if="error" class="form-error">{{ error }}</p>
      <div class="form-actions">
        <button type="submit" class="button">Create</button>
        <RouterLink :to="{ name: 'webhooks' }" class="button">Cancel</RouterLink>
      </div>
    </form>
  </Section>
</template>
