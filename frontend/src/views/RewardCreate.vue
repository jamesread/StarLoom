<script setup lang="ts">
import { reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, GiftIcon } from '@hugeicons/core-free-icons'
import { starapp } from '../api/client'

const router = useRouter()
const error = ref('')
const creating = ref(false)

const form = reactive({
  title: '',
  description: '',
  costStars: 5,
  approvalRequired: true,
  active: true,
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

async function createReward() {
  creating.value = true
  error.value = ''
  try {
    const res = await starapp.createReward({
      title: form.title.trim(),
      description: form.description.trim(),
      costStars: form.costStars,
      approvalRequired: form.approvalRequired,
      availabilityExpression: form.availabilityExpression.trim(),
    })
    const created = res.reward
    if (created?.id && !form.active) {
      await starapp.updateReward({
        id: created.id,
        title: form.title.trim(),
        description: form.description.trim(),
        costStars: form.costStars,
        active: false,
        approvalRequired: form.approvalRequired,
        availabilityExpression: form.availabilityExpression.trim(),
      })
    }
    router.push({ name: 'familyRewards' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <Section title="Add reward" :icon="GiftIcon" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'familyRewards' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Rewards</span>
      </RouterLink>
    </template>

    <p class="subtle">Define a privilege children can redeem with stars.</p>

    <FormLayout @submit.prevent="createReward">
      <p v-if="error" class="inline-notification error">{{ error }}</p>
      <FormField label="Title" for="reward-create-title">
        <input
          id="reward-create-title"
          v-model="form.title"
          type="text"
          required
          placeholder="Extra screen time"
          :disabled="creating"
        />
      </FormField>
      <FormField label="Description" for="reward-create-description">
        <textarea id="reward-create-description" v-model="form.description" rows="2" :disabled="creating" />
      </FormField>
      <FormField label="Cost (stars)" for="reward-create-cost">
        <input
          id="reward-create-cost"
          v-model.number="form.costStars"
          type="number"
          min="1"
          required
          :disabled="creating"
        />
      </FormField>
      <FormField label="Requires approval" component-has-label>
        <RadioGroup
          v-model="form.approvalRequired"
          name="reward-create-approval"
          variant="boolean"
          :options="booleanOptions"
        />
      </FormField>
      <FormField label="Status" component-has-label>
        <RadioGroup
          v-model="form.active"
          name="reward-create-active"
          variant="boolean"
          :options="statusOptions"
        />
      </FormField>
      <FormField
        label="Availability expression"
        for="reward-create-availability"
        description="Optional. expr language; must evaluate to true when the reward can be redeemed. Leave blank for always available. Variables: hour, minute, dayName (Mon–Sun), day, month, year, balance, costStars, countPerDay, countPerWeek. Times use server local timezone."
      >
        <textarea
          id="reward-create-availability"
          v-model="form.availabilityExpression"
          rows="3"
          placeholder='(hour > 9 && hour < 18) && (dayName == "Sat" || dayName == "Sun")'
          :disabled="creating"
        />
      </FormField>
      <template #actions>
        <button type="submit" class="good" :disabled="creating || !form.title.trim()">
          {{ creating ? 'Creating…' : 'Create' }}
        </button>
        <RouterLink :to="{ name: 'familyRewards' }" class="button neutral">Cancel</RouterLink>
      </template>
    </FormLayout>
  </Section>
</template>
