<script setup lang="ts">
import Section from 'picocrank/vue/components/Section.vue'
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { starapp, type ApiKey } from '../api/client'

const router = useRouter()

const keys = ref<ApiKey[]>([])
const name = ref('')
const readOnly = ref(false)
const newSecret = ref('')
const error = ref('')

async function load() {
  const res = await starapp.listApiKeys()
  keys.value = res.keys || []
}

async function createKey() {
  newSecret.value = ''
  const res = await starapp.createApiKey({ name: name.value || 'API key', readOnly: readOnly.value })
  newSecret.value = res.secret || ''
  name.value = ''
  await load()
}

async function removeKey(id: number) {
  await starapp.deleteApiKey({ id })
  await load()
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
  <Section title="API keys" subtitle="Bearer tokens for automation and MCP" :padding="true">
    <template #toolbar>
      <button type="button" class="inline-icon neutral" @click="router.push({ name: 'userControlPanel' })">
        Back
      </button>
    </template>
    <p v-if="error" class="inline-notification error">{{ error }}</p>
    <div class="toolbar-pad">
      <input v-model="name" placeholder="Key name" />
      <label><input v-model="readOnly" type="checkbox" /> Read-only</label>
      <button type="button" class="good" @click="createKey">Create key</button>
    </div>
    <p v-if="newSecret" class="inline-notification note">Copy this secret now — it will not be shown again: <code>{{ newSecret }}</code></p>
    <ul>
      <li v-for="k in keys" :key="k.id">
        {{ k.name }} ({{ k.readOnly ? 'read-only' : 'read/write' }})
        <button type="button" class="bad small" @click="removeKey(k.id)">Delete</button>
      </li>
    </ul>
  </Section>
</template>

<style scoped>
.toolbar-pad {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  align-items: center;
  margin-bottom: 1rem;
}
</style>
