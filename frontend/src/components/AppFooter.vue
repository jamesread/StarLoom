<script setup lang="ts">
import { computed } from 'vue'
import { useStatus } from '../composables/useStatus'

const props = defineProps<{
  appName?: string
  logoUrl?: string
  docsUrl?: string
  version?: string
}>()

const status = useStatus()

const showFooter = computed(() => status.status?.showFooter !== false)
const showVersionNumber = computed(() => status.status?.showVersionNumber !== false)
const currentVersion = computed(() => props.version || status.status?.currentVersion || '')
const showNewVersions = computed(() => status.status?.showNewVersions === true)
const availableVersion = computed(
  () => status.status?.availableVersion?.trim() || '',
)

const showUpdateLink = computed(
  () => showNewVersions.value && availableVersion.value !== '',
)
</script>

<template>
  <footer v-if="showFooter" title="footer">
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
      <span v-if="showVersionNumber && currentVersion">{{ currentVersion }}</span>
    </p>
    <p>
      <span v-if="docsUrl">
        <a :href="docsUrl" target="_blank" rel="noopener noreferrer">Docs</a>
      </span>
    </p>
    <p v-if="showUpdateLink">
      <a
        id="available-version"
        href="#"
        target="_blank"
        rel="noopener noreferrer"
      >
        {{ availableVersion }}
      </a>
    </p>
  </footer>
</template>
