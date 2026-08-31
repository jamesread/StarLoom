<script setup lang="ts">
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { useTheme } from 'picocrank/vue/composables/useTheme.js'
import { useCustomTheme } from 'picocrank/vue/composables/useCustomTheme.js'
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { starapp } from '../api/client'
import { useInit } from '../composables/useInit'
import { applyAppTheming, themeControlFromFeatures } from '../lib/applyAppTheming'
import { applyUserLanguage, applyUserSidebar, languageLabel } from '../lib/userPreferences'

const router = useRouter()
const init = useInit()
const { theme } = useTheme()
const { availableThemes, themeLabels, themePreference, setTheme } = useCustomTheme()

const language = ref('')
const sidebarEnabled = ref(true)
const availableLanguages = ref<string[]>([])
const savedLanguage = ref('')
const savedSidebar = ref(true)
const saving = ref(false)
const error = ref('')
const success = ref('')

const initFeatures = computed(() => init.init?.features ?? {})
const themeControl = computed(() => themeControlFromFeatures(initFeatures.value))
const userThemeControl = computed(() => themeControl.value === 'user')

const colorSchemeOptions = [
  { label: 'Auto (system)', value: 'auto' },
  { label: 'Light', value: 'light' },
  { label: 'Dark', value: 'dark' },
]

const dirty = computed(
  () => language.value !== savedLanguage.value || sidebarEnabled.value !== savedSidebar.value,
)

function onThemeNameChange(event: Event) {
  const target = event.target as HTMLSelectElement
  setTheme(target.value)
}

async function load() {
  error.value = ''
  const res = await starapp.getUserPreferences()
  language.value = res.language
  sidebarEnabled.value = res.sidebarEnabled
  availableLanguages.value = res.availableLanguages
  savedLanguage.value = language.value
  savedSidebar.value = sidebarEnabled.value
}

async function save() {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    const res = await starapp.saveUserPreferences({
      language: language.value,
      sidebarEnabled: sidebarEnabled.value,
    })
    if (!res.standardResponse?.success) {
      error.value = res.standardResponse?.message || 'Save failed'
      return
    }
    savedLanguage.value = language.value
    savedSidebar.value = sidebarEnabled.value
    applyUserLanguage(language.value)
    applyUserSidebar(sidebarEnabled.value)
    success.value = 'Preferences saved'
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Save failed'
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Load failed'
  }
})
</script>

<template>
  <Section title="User preferences" :padding="true">
    <template #toolbar>
      <button type="button" class="inline-icon neutral" @click="router.push({ name: 'userControlPanel' })">
        Back
      </button>
    </template>

    <p v-if="error" class="inline-notification error">{{ error }}</p>
    <p v-if="success" class="inline-notification note">{{ success }}</p>

    <FormLayout @submit.prevent="save">
      <FormField
        label="Language"
        for="user-preferences-language"
        :disabled="saving"
        description="Empty (browser default) uses the browser language."
      >
        <select id="user-preferences-language" v-model="language" :disabled="saving">
          <option value="">Browser default</option>
          <option v-for="code in availableLanguages" :key="code" :value="code">
            {{ languageLabel(code) }}
          </option>
        </select>
      </FormField>

      <FormField label="Sidebar" component-has-label :disabled="saving" description="Show navigation sidebar on wide screens.">
        <RadioGroup
          v-model="sidebarEnabled"
          name="user-preferences-sidebar"
          variant="boolean"
          :options="[
            { label: 'Enabled', value: true },
            { label: 'Disabled', value: false },
          ]"
          :disabled="saving"
        />
      </FormField>

      <FormField
        label="Color scheme preference"
        component-has-label
        description="Choose light, dark, or match your operating system. Saved in this browser."
      >
        <RadioGroup
          v-model="theme"
          name="user-preferences-color-scheme"
          variant="list"
          :options="colorSchemeOptions"
          aria-label="Color scheme preference"
        />
      </FormField>

      <FormField
        v-if="userThemeControl"
        label="Theme name"
        for="user-preferences-theme-name"
        description="Override the administrator default theme. Saved in this browser."
      >
        <select
          id="user-preferences-theme-name"
          :value="themePreference"
          @change="onThemeNameChange"
        >
          <option value="">Default (Femtocrank only)</option>
          <option v-for="name in availableThemes" :key="name" :value="name">
            {{ themeLabels[name] || name }}
          </option>
        </select>
      </FormField>
      <FormField
        v-else
        label="Theme name"
        component-has-label
        description="Set by administrator in Settings."
      >
        <p class="subtle">{{ initFeatures.themeName || 'Default (Femtocrank only)' }}</p>
      </FormField>

      <template #actions>
        <button type="submit" class="good" :disabled="saving || !dirty">
          {{ saving ? 'Saving…' : 'Save preferences' }}
        </button>
      </template>
    </FormLayout>
  </Section>
</template>
