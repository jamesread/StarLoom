<script setup lang="ts">
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { onMounted, ref, computed } from 'vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon } from '@hugeicons/core-free-icons'
import { starapp, type UserAccount } from '../api/client'
import { useStatus } from '../composables/useStatus'
import { hasPermission } from '../lib/rbacAccess'

const users = ref<UserAccount[]>([])
const loading = ref(true)
const error = ref('')
const createUsername = ref('')
const createPassword = ref('')
const status = useStatus()

const canCreate = computed(
  () => hasPermission(status.status, 'users.create'),
)
const canDelete = computed(
  () => hasPermission(status.status, 'users.delete'),
)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await starapp.getUsers()
    users.value = res.users || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Load failed'
  } finally {
    loading.value = false
  }
}

async function createUser() {
  if (!createUsername.value.trim()) return
  await starapp.createUser({
    username: createUsername.value.trim(),
    password: createPassword.value || undefined,
  })
  createUsername.value = ''
  createPassword.value = ''
  await load()
}

async function removeUser(u: UserAccount) {
  if (!confirm(`Delete user ${u.username}?`)) return
  await starapp.deleteUser({ userId: u.id })
  await load()
}

onMounted(load)

const rows = computed(() =>
  users.value.map((u) => ({
    id: u.id,
    username: u.username,
    createdBy: u.createdBy || '—',
    createdAt: u.createdAt || '—',
    raw: u,
  })),
)
</script>

<template>
  <Section title="Users" :padding="false">
    <template #toolbar>
      <router-link :to="{ name: 'iam' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>IAM</span>
      </router-link>
      <button type="button" class="inline-icon neutral" :disabled="loading" @click="load">Refresh</button>
    </template>
    <p v-if="error" class="inline-notification error">{{ error }}</p>
    <FormLayout v-if="canCreate" class="toolbar-pad" @submit.prevent="createUser">
      <FormField label="Username" for="create-username">
        <input id="create-username" v-model="createUsername" type="text" required placeholder="Username" />
      </FormField>
      <FormField label="Password" for="create-password" description="Optional; at least 8 characters if set.">
        <input id="create-password" v-model="createPassword" type="password" placeholder="Password (optional, min 8)" autocomplete="new-password" />
      </FormField>
      <template #actions>
        <button type="submit" class="good">Create user</button>
      </template>
    </FormLayout>
    <Table
      :data="rows"
      :headers="[
        { key: 'username', label: 'Username' },
        { key: 'createdBy', label: 'Created by' },
        { key: 'createdAt', label: 'Created' },
      ]"
    />
    <div v-if="canDelete" class="toolbar-pad">
      <button
        v-for="u in users"
        :key="u.id"
        type="button"
        class="bad small"
        @click="removeUser(u)"
      >
        Delete {{ u.username }}
      </button>
    </div>
  </Section>
</template>

<style scoped>
.toolbar-pad {
  padding: 1rem;
}
</style>
