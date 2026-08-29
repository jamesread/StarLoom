<script setup lang="ts">
import { reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, StarIcon } from '@hugeicons/core-free-icons'
import { starapp } from '../api/client'

const router = useRouter()
const error = ref('')
const creating = ref(false)

const form = reactive({
  name: '',
  sortOrder: 0,
})

async function createChart() {
  creating.value = true
  error.value = ''
  try {
    const res = await starapp.createStarChart({
      name: form.name.trim(),
      sortOrder: form.sortOrder,
    })
    const id = res.starChart?.id
    if (id) {
      router.push({ name: 'familyStarChartEdit', params: { id } })
    } else {
      router.push({ name: 'familyStarCharts' })
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <Section title="Add star chart" :icon="StarIcon" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'familyStarCharts' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Star Charts</span>
      </RouterLink>
    </template>

    <p class="subtle">Create a chart, then assign chores to it from the Chores page.</p>

    <FormLayout @submit.prevent="createChart">
      <p v-if="error" class="inline-notification error">{{ error }}</p>
      <FormField label="Name" for="star-chart-create-name">
        <input
          id="star-chart-create-name"
          v-model="form.name"
          type="text"
          required
          placeholder="Morning routine"
          :disabled="creating"
        />
      </FormField>
      <FormField label="Sort order" for="star-chart-create-sort">
        <input
          id="star-chart-create-sort"
          v-model.number="form.sortOrder"
          type="number"
          min="0"
          :disabled="creating"
        />
      </FormField>
      <template #actions>
        <button type="submit" class="good" :disabled="creating || !form.name.trim()">
          {{ creating ? 'Creating…' : 'Create' }}
        </button>
        <RouterLink :to="{ name: 'familyStarCharts' }" class="button neutral">Cancel</RouterLink>
      </template>
    </FormLayout>
  </Section>
</template>
