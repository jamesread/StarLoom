<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, GiftIcon } from '@hugeicons/core-free-icons'
import { starapp, type Reward } from '../api/client'

const route = useRoute()
const router = useRouter()
const rewardId = computed(() => Number(route.params.id))

const reward = ref<Reward | null>(null)
const error = ref('')
const saving = ref(false)
const loading = ref(true)

const form = reactive({
  title: '',
  description: '',
  costStars: 5,
  active: true,
  approvalRequired: true,
  availabilityExpression: '',
})

const booleanOptions = [
  { label: 'Yes', value: true },
  { label: 'No', value: false },
]

const statusOptions = [
  { label: 'Active', value: true },
  { label: 'Inactive', value: false },
]

const sectionTitle = computed(() => reward.value?.title || 'Edit reward')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await starapp.listRewards({ includeInactive: true })
    reward.value = (res.rewards || []).find((r) => r.id === rewardId.value) || null
    if (!reward.value) {
      error.value = 'Reward not found'
      return
    }
    form.title = reward.value.title
    form.description = reward.value.description || ''
    form.costStars = reward.value.costStars
    form.active = reward.value.active !== false
    form.approvalRequired = reward.value.approvalRequired === true
    form.availabilityExpression = reward.value.availabilityExpression || ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!reward.value) return
  saving.value = true
  error.value = ''
  try {
    await starapp.updateReward({
      id: reward.value.id,
      title: form.title.trim(),
      description: form.description.trim(),
      costStars: form.costStars,
      active: form.active,
      approvalRequired: form.approvalRequired,
      availabilityExpression: form.availabilityExpression.trim(),
    })
    router.push({ name: 'familyRewards' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

async function deactivate() {
  if (!reward.value || !confirm('Deactivate this reward?')) return
  try {
    await starapp.deleteReward({ id: reward.value.id })
    router.push({ name: 'familyRewards' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(load)
</script>

<template>
  <Section :title="sectionTitle" :icon="GiftIcon" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'familyRewards' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Rewards</span>
      </RouterLink>
    </template>

    <p v-if="loading" class="subtle">Loading…</p>
    <p v-else-if="error && !reward" class="inline-notification error">{{ error }}</p>

    <FormLayout v-else @submit.prevent="save">
      <p v-if="error" class="inline-notification error">{{ error }}</p>
      <FormField label="Title" for="reward-edit-title">
        <input id="reward-edit-title" v-model="form.title" type="text" required />
      </FormField>
      <FormField label="Description" for="reward-edit-description">
        <textarea id="reward-edit-description" v-model="form.description" rows="2" />
      </FormField>
      <FormField label="Cost (stars)" for="reward-edit-cost">
        <input id="reward-edit-cost" v-model.number="form.costStars" type="number" min="1" required />
      </FormField>
      <FormField label="Requires approval" component-has-label>
        <RadioGroup
          v-model="form.approvalRequired"
          name="reward-edit-approval"
          variant="boolean"
          :options="booleanOptions"
        />
      </FormField>
      <FormField label="Status" component-has-label>
        <RadioGroup
          v-model="form.active"
          name="reward-edit-active"
          variant="boolean"
          :options="statusOptions"
        />
      </FormField>
      <FormField
        label="Availability expression"
        for="reward-edit-availability"
        description="Optional. expr language; must evaluate to true when the reward can be redeemed. Leave blank for always available. Variables: hour, minute, dayName (Mon–Sun), day, month, year, balance, costStars, countPerDay, countPerWeek. Times use server local timezone."
      >
        <textarea
          id="reward-edit-availability"
          v-model="form.availabilityExpression"
          rows="3"
          placeholder='(hour > 9 && hour < 18) && (dayName == "Sat" || dayName == "Sun")'
        />
      </FormField>
      <template #actions>
        <button type="submit" class="good" :disabled="saving">{{ saving ? 'Saving…' : 'Save' }}</button>
        <RouterLink :to="{ name: 'familyRewards' }" class="button neutral">Cancel</RouterLink>
        <button v-if="form.active" type="button" class="bad" @click="deactivate">Deactivate</button>
      </template>
    </FormLayout>
  </Section>
</template>
