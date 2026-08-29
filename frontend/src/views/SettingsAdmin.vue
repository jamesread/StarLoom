<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon } from '@hugeicons/core-free-icons'
import { starapp, type Cvar } from '../api/client'
import { fetchAppStatus } from '../composables/useStatus'
import { loadInit } from '../composables/useInit'

const cvars = ref<Cvar[]>([])
const edits = reactive<Record<string, {
  valueString: string
  valueInt: number
  boolValue: boolean
}>>({})
const dirtySections = reactive<Record<string, boolean>>({})
const error = ref('')
const success = ref('')
const savingSection = ref('')

const booleanOptions = [
  { label: 'On', value: true },
  { label: 'Off', value: false },
]

function labelFor(cvar: Cvar) {
  return cvar.title || cvar.key.replace(/_/g, ' ')
}

function fieldId(cvar: Cvar) {
  return `cvar-${cvar.key}`
}

function markDirty(sectionName: string) {
  dirtySections[sectionName] = true
}

function clearDirty() {
  for (const key of Object.keys(dirtySections)) {
    delete dirtySections[key]
  }
}

const categories = computed(() => {
  const groups: { name: string; cvars: Cvar[] }[] = []
  const indexByName: Record<string, number> = {}
  for (const c of cvars.value) {
    const name = c.category || 'Other'
    if (indexByName[name] === undefined) {
      indexByName[name] = groups.length
      groups.push({ name, cvars: [] })
    }
    groups[indexByName[name]].cvars.push(c)
  }
  return groups
})

function syncEdits() {
  for (const key of Object.keys(edits)) delete edits[key]
  for (const c of cvars.value) {
    edits[c.key] = {
      valueString: c.valueString || '',
      valueInt: c.valueInt || 0,
      boolValue: !!c.valueInt,
    }
  }
  clearDirty()
}

function valuesFor(cvar: Cvar) {
  const edit = edits[cvar.key]
  if (cvar.mainType === 'bool') {
    return { valueInt: edit.boolValue ? 1 : 0, valueString: '' }
  }
  if (cvar.mainType === 'int') {
    return { valueInt: Number(edit.valueInt) || 0, valueString: '' }
  }
  return { valueInt: 0, valueString: edit.valueString || '' }
}

async function load() {
  try {
    const res = await starapp.listCvars()
    cvars.value = res.cvars || []
    syncEdits()
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function saveSection(group: { name: string; cvars: Cvar[] }) {
  savingSection.value = group.name
  success.value = ''
  error.value = ''
  try {
    for (const cvar of group.cvars) {
      const { valueInt, valueString } = valuesFor(cvar)
      await starapp.updateCvar({ key: cvar.key, valueInt, valueString })
    }
    success.value = `${group.name} settings saved.`
    await load()
    await Promise.all([loadInit({ force: true }), fetchAppStatus({ force: true })])
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    savingSection.value = ''
  }
}

onMounted(load)
</script>

<template>
  <Section title="Settings" subtitle="Configuration variables" :padding="true">
    <template #toolbar>
      <router-link :to="{ name: 'controlPanel' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Control Panel</span>
      </router-link>
    </template>
    <p>Family-wide options stored in the database. Edits apply after you save.</p>
    <p v-if="error" class="form-error">{{ error }}</p>
    <NotificationBlock
      v-if="success"
      type="success"
      :message="success"
    />
    <p v-if="cvars.length === 0 && !error" class="subtle">No configuration variables found.</p>
  </Section>

  <Section
    v-for="group in categories"
    :key="group.name"
    :title="group.name"
    :padding="true"
  >
    <FormLayout @submit.prevent="saveSection(group)">
      <template v-for="cvar in group.cvars" :key="cvar.key">
        <FormField
          v-if="cvar.mainType === 'string'"
          :label="labelFor(cvar)"
          :for="fieldId(cvar)"
        >
          <div>
            <input
              :id="fieldId(cvar)"
              v-model="edits[cvar.key].valueString"
              type="text"
              required
              maxlength="255"
              @change="markDirty(group.name)"
            >
            <p v-if="cvar.description" class="subtle">{{ cvar.description }}</p>
          </div>
        </FormField>

        <FormField
          v-else-if="cvar.mainType === 'int'"
          :label="labelFor(cvar)"
          :for="fieldId(cvar)"
        >
          <div>
            <input
              :id="fieldId(cvar)"
              v-model.number="edits[cvar.key].valueInt"
              type="number"
              required
              :min="cvar.key === 'default_award_stars' ? 1 : undefined"
              :max="cvar.key === 'default_award_stars' ? 100 : undefined"
              @change="markDirty(group.name)"
            >
            <p v-if="cvar.description" class="subtle">{{ cvar.description }}</p>
          </div>
        </FormField>

        <FormField
          v-else-if="cvar.mainType === 'bool'"
          :label="labelFor(cvar)"
          fake
        >
          <div>
            <RadioGroup
              v-model="edits[cvar.key].boolValue"
              :name="fieldId(cvar)"
              variant="boolean"
              :options="booleanOptions"
              :aria-label="labelFor(cvar)"
              @change="markDirty(group.name)"
            />
            <p v-if="cvar.description" class="subtle">{{ cvar.description }}</p>
          </div>
        </FormField>

        <p v-else class="form-error">Unsupported type: {{ cvar.mainType }}</p>
      </template>

      <template #actions>
        <button
          type="submit"
          class="button"
          :disabled="!dirtySections[group.name] || savingSection === group.name"
        >
          {{ savingSection === group.name ? 'Saving…' : 'Save' }}
        </button>
      </template>
    </FormLayout>
  </Section>
</template>
