<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { GiftIcon } from '@hugeicons/core-free-icons'
import { starapp, type Reward } from '../api/client'

const rewards = ref<Reward[]>([])
const error = ref('')
const showCreate = ref(false)
const creating = ref(false)
const form = reactive({
  title: '',
  description: '',
  costStars: 5,
  approvalRequired: true,
})

const booleanOptions = [
  { label: 'Yes', value: true },
  { label: 'No', value: false },
]

async function load() {
  try {
    const res = await starapp.listRewards({ includeInactive: true })
    rewards.value = res.rewards || []
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function createReward() {
  creating.value = true
  try {
    await starapp.createReward({
      title: form.title.trim(),
      description: form.description.trim(),
      costStars: form.costStars,
      approvalRequired: form.approvalRequired,
    })
    form.title = ''
    form.description = ''
    form.costStars = 5
    showCreate.value = false
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    creating.value = false
  }
}

async function deactivate(id: number) {
  if (!confirm('Deactivate this reward?')) return
  try {
    await starapp.deleteReward({ id })
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(load)
</script>

<template>
  <Section title="Rewards" :icon="GiftIcon">
    <p v-if="error" class="error">{{ error }}</p>
    <div class="toolbar">
      <button type="button" @click="load">Refresh</button>
      <button type="button" @click="showCreate = true">Add reward</button>
    </div>

    <table v-if="rewards.length">
      <thead>
        <tr>
          <th>Title</th>
          <th>Cost</th>
          <th>Status</th>
          <th>Approval</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rewards" :key="r.id">
          <td>{{ r.title }}</td>
          <td>{{ r.costStars }}</td>
          <td>{{ r.active ? 'Active' : 'Inactive' }}</td>
          <td>{{ r.approvalRequired ? 'Required' : 'Auto' }}</td>
          <td>
            <button v-if="r.active" type="button" class="secondary outline" @click="deactivate(r.id)">
              Deactivate
            </button>
          </td>
        </tr>
      </tbody>
    </table>
    <p v-else class="subtle">No rewards yet.</p>

    <div v-if="showCreate" class="create-form">
      <FormField label="Title">
        <input v-model="form.title" type="text" required />
      </FormField>
      <FormField label="Description">
        <textarea v-model="form.description" rows="2" />
      </FormField>
      <FormField label="Cost (stars)">
        <input v-model.number="form.costStars" type="number" min="1" required />
      </FormField>
      <FormField label="Requires approval">
        <RadioGroup v-model="form.approvalRequired" :options="booleanOptions" name="approval" />
      </FormField>
      <button type="button" class="secondary" @click="showCreate = false">Cancel</button>
      <button type="button" :disabled="creating" @click="createReward">Create</button>
    </div>
  </Section>
</template>

<style scoped>
.toolbar {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
}
.create-form {
  margin-top: 1rem;
  padding: 1rem;
  border: 1px solid var(--pico-muted-border-color);
  border-radius: var(--pico-border-radius);
}
.error {
  color: var(--pico-del-color);
}
</style>
