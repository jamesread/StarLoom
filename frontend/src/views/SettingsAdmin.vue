<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon } from '@hugeicons/core-free-icons'
import { useCustomTheme } from 'picocrank/vue/composables/useCustomTheme.js'
import { starapp, type Cvar } from '../api/client'
import { fetchAppStatus } from '../composables/useStatus'
import { loadInit } from '../composables/useInit'
import { applyAppTheming } from '../lib/applyAppTheming'

const THEME_CVAR_KEYS = new Set([
  'theme_color_scheme_switcher_enabled',
  'theme_name',
  'theme_control',
])

const cvars = ref<Cvar[]>([])
const edits = reactive<Record<string, {
  valueString: string
  valueInt: number
  boolValue: boolean
}>>({})
const dirtySections = reactive<Record<string, boolean>>({})
const themeEdits = reactive({
  colorSchemeSwitcher: false,
  themeName: '',
  themeControl: 'user',
})
const themeSectionDirty = ref(false)
const error = ref('')
const success = ref('')
const savingSection = ref('')
const themeSaving = ref(false)
const { availableThemes, themeLabels } = useCustomTheme()

const booleanOptions = [
  { label: 'On', value: true },
  { label: 'Off', value: false },
]

const themeControlOptions = [
  { label: 'System preference', value: 'system' },
  { label: 'User preference', value: 'user' },
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
    if (THEME_CVAR_KEYS.has(c.key)) {
      continue
    }
    const name = c.category || 'Other'
    if (indexByName[name] === undefined) {
      indexByName[name] = groups.length
      groups.push({ name, cvars: [] })
    }
    groups[indexByName[name]].cvars.push(c)
  }
  return groups
})

function syncThemeEdits() {
  themeEdits.colorSchemeSwitcher = !!edits.theme_color_scheme_switcher_enabled?.boolValue
  themeEdits.themeName = edits.theme_name?.valueString ?? ''
  themeEdits.themeControl = edits.theme_control?.valueString || 'user'
  themeSectionDirty.value = false
}

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
  syncThemeEdits()
}

function valuesFor(cvar: Cvar) {
  const edit = edits[cvar.key]
  if (cvar.mainType === 'bool') {
    return { valueInt: edit.boolValue ? 1 : 0, valueString: '' }
  }
  if (cvar.mainType === 'int') {
    return { valueInt: Number(edit.valueInt) || 0, valueString: '' }
  }
  // string + textarea
  return { valueInt: 0, valueString: edit.valueString || '' }
}

async function reloadShell() {
  const init = await loadInit({ force: true })
  await fetchAppStatus({ force: true })
  applyAppTheming(init.features)
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
    await reloadShell()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    savingSection.value = ''
  }
}

async function saveThemeSection() {
  themeSaving.value = true
  success.value = ''
  error.value = ''
  try {
    await starapp.updateCvar({
      key: 'theme_color_scheme_switcher_enabled',
      valueInt: themeEdits.colorSchemeSwitcher ? 1 : 0,
    })
    await starapp.updateCvar({
      key: 'theme_name',
      valueString: themeEdits.themeName,
    })
    await starapp.updateCvar({
      key: 'theme_control',
      valueString: themeEdits.themeControl,
    })
    success.value = 'Theme settings saved.'
    themeSectionDirty.value = false
    await load()
    await reloadShell()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    themeSaving.value = false
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

  <Section title="Theme" subtitle="Appearance and theme policy" :padding="true">
    <FormLayout @submit.prevent="saveThemeSection">
      <FormField label="Color scheme switcher" component-has-label>
        <RadioGroup
          v-model="themeEdits.colorSchemeSwitcher"
          name="settings-theme-color-scheme-switcher"
          variant="boolean"
          :options="booleanOptions"
          aria-label="Color scheme switcher"
          @change="themeSectionDirty = true"
        />
        <p class="subtle">Show the auto/light/dark control in the header.</p>
      </FormField>

      <FormField label="Theme name" for="settings-theme-name">
        <select
          id="settings-theme-name"
          v-model="themeEdits.themeName"
          @change="themeSectionDirty = true"
        >
          <option value="">Default (Femtocrank only)</option>
          <option v-for="name in availableThemes" :key="name" :value="name">
            {{ themeLabels[name] || name }}
          </option>
        </select>
        <p class="subtle">Drop-in CSS theme for the app.</p>
      </FormField>

      <FormField label="Theme control" component-has-label>
        <RadioGroup
          v-model="themeEdits.themeControl"
          name="settings-theme-control"
          variant="list"
          :options="themeControlOptions"
          aria-label="Theme control"
          @change="themeSectionDirty = true"
        />
        <p class="subtle">
          System preference forces the theme name for all users.
          User preference uses this theme as the default; users may override on User Preferences.
        </p>
      </FormField>

      <template #actions>
        <button type="submit" class="good" :disabled="!themeSectionDirty || themeSaving">
          {{ themeSaving ? 'Saving…' : 'Save' }}
        </button>
      </template>
    </FormLayout>
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
              :required="cvar.key !== 'apprise_url' && cvar.key !== 'external_base_url'"
              maxlength="255"
              @change="markDirty(group.name)"
            >
            <p v-if="cvar.description" class="subtle">{{ cvar.description }}</p>
          </div>
        </FormField>

        <FormField
          v-else-if="cvar.mainType === 'textarea'"
          :label="labelFor(cvar)"
          :for="fieldId(cvar)"
        >
          <div>
            <textarea
              :id="fieldId(cvar)"
              v-model="edits[cvar.key].valueString"
              rows="6"
              maxlength="4000"
              @change="markDirty(group.name)"
            />
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
          component-has-label
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
