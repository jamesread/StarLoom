<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import CheckGroup from 'picocrank/vue/components/CheckGroup.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
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
    <p v-if="error" class="inline-notification error">{{ error }}</p>

    <FormLayout @submit.prevent="submit">
      <FormField label="URL" for="webhook-create-url">
        <input
          id="webhook-create-url"
          v-model="url"
          type="url"
          required
          placeholder="https://example.com/hooks/starapp"
        >
      </FormField>
      <FormField label="Secret" for="webhook-create-secret">
        <input
          id="webhook-create-secret"
          v-model="secret"
          type="text"
          required
          placeholder="Shared signing secret"
        >
      </FormField>
      <FormField label="Events" component-has-label description="Select one or more events that should POST to this URL.">
        <CheckGroup
          v-model="selectedEvents"
          :options="eventOptions"
          name="webhook-events-create"
          aria-label="Webhook events"
        />
      </FormField>
      <FormField label="Enabled" component-has-label>
        <RadioGroup
          v-model="enabled"
          name="webhook-enabled-create"
          variant="boolean"
          :options="booleanOptions"
          aria-label="Webhook enabled"
        />
      </FormField>
      <template #actions>
        <button type="submit" class="good">Create</button>
        <RouterLink :to="{ name: 'webhooks' }" class="button neutral">Cancel</RouterLink>
      </template>
    </FormLayout>
  </Section>
</template>
