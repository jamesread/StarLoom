<script setup lang="ts">
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import Section from 'picocrank/vue/components/Section.vue'
import { useTheme } from 'picocrank/vue/composables/useTheme.js'
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { starapp } from '../api/client'
import {
  applyUserLanguage,
  applyUserSidebar,
  applyUserThemeToggle,
  languageLabel,
} from '../lib/userPreferences'

const router = useRouter()
const { theme } = useTheme()

const language = ref('')
const sidebarEnabled = ref(true)
const themeToggleEnabled = ref(false)
const availableLanguages = ref<string[]>([])
const savedLanguage = ref('')
const savedSidebar = ref(true)
const savedThemeToggle = ref(false)
const saving = ref(false)
const error = ref('')
const success = ref('')

const dirty = computed(
  () =>
    language.value !== savedLanguage.value ||
    sidebarEnabled.value !== savedSidebar.value ||
    themeToggleEnabled.value !== savedThemeToggle.value,
)

async function load() {
  error.value = ''
  const res = await starapp.getUserPreferences()
  language.value = res.language
  sidebarEnabled.value = res.sidebarEnabled
  themeToggleEnabled.value = res.themeToggleEnabled
  availableLanguages.value = res.availableLanguages
  savedLanguage.value = language.value
  savedSidebar.value = sidebarEnabled.value
  savedThemeToggle.value = themeToggleEnabled.value
}

async function save() {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    const res = await starapp.saveUserPreferences({
      language: language.value,
      sidebarEnabled: sidebarEnabled.value,
      themeToggleEnabled: themeToggleEnabled.value,
    })
    if (!res.standardResponse?.success) {
      error.value = res.standardResponse?.message || 'Save failed'
      return
    }
    savedLanguage.value = language.value
    savedSidebar.value = sidebarEnabled.value
    savedThemeToggle.value = themeToggleEnabled.value
    applyUserLanguage(language.value)
    applyUserSidebar(sidebarEnabled.value)
    applyUserThemeToggle(themeToggleEnabled.value)
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

      <FormField label="Sidebar" fake :disabled="saving" description="Show navigation sidebar on wide screens.">
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

      <FormField label="Theme switcher" fake :disabled="saving" description="Show the theme button in the header.">
        <RadioGroup
          v-model="themeToggleEnabled"
          name="user-preferences-theme-toggle"
          variant="boolean"
          :options="[
            { label: 'Enabled', value: true },
            { label: 'Disabled', value: false },
          ]"
          :disabled="saving"
        />
      </FormField>

      <FormField label="Theme" fake description="Saved in this browser only.">
        <RadioGroup
          v-model="theme"
          name="user-preferences-theme"
          variant="list"
          :options="[
            { label: 'Auto (system)', value: 'auto' },
            { label: 'Light', value: 'light' },
            { label: 'Dark', value: 'dark' },
          ]"
        />
      </FormField>

      <template #actions>
        <button type="submit" class="good" :disabled="saving || !dirty">
          {{ saving ? 'Saving…' : 'Save preferences' }}
        </button>
      </template>
    </FormLayout>
  </Section>
</template>
