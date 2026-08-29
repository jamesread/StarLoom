<script setup lang="ts">
import { computed } from 'vue'
import { useInit } from '../composables/useInit'

const DOCS_URL = 'https://jamesread.github.io/StarLoom/'
const ISSUES_URL = 'https://github.com/jamesread/StarLoom/issues'
const RELEASES_URL = 'https://github.com/jamesread/StarLoom/releases'

defineProps<{
  logoUrl?: string
}>()

const init = useInit()

const showVersionNumber = computed(() => init.init?.showVersionNumber !== false)
const appName = computed(() => init.init?.siteTitle?.trim() || 'StarApp')
const currentVersion = computed(() => init.init?.currentVersion?.trim() || '')
const showNewVersions = computed(() => init.init?.showNewVersions === true)
const availableVersion = computed(() => init.init?.availableVersion?.trim() || '')

const showUpdateLink = computed(() => {
  if (!showVersionNumber.value || !showNewVersions.value) return false
  const v = availableVersion.value
  if (!v || v === 'none' || v === '?' || v.toLowerCase().startsWith('you-are-using')) {
    return false
  }
  return true
})
</script>

<template>
  <footer title="footer">
    <p>
      <img
        v-if="logoUrl"
        class="logo"
        :src="logoUrl"
        :alt="appName + ' logo'"
        title="application icon"
        style="height: 1em;"
      >
      {{ appName }}
    </p>
    <p>
      <span v-if="showVersionNumber && currentVersion">{{ currentVersion }}</span>
      <span>
        <a :href="DOCS_URL" target="_blank" rel="noopener noreferrer">Docs</a>
      </span>
      <span>
        <a :href="ISSUES_URL" target="_blank" rel="noopener noreferrer">Raise an issue</a>
      </span>
    </p>
    <p v-if="showUpdateLink">
      <a
        id="available-version"
        :href="RELEASES_URL"
        target="_blank"
        rel="noopener noreferrer"
      >
        {{ availableVersion }}
      </a>
    </p>
  </footer>
</template>
