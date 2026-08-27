<script setup lang="ts">
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import { onMounted, ref } from 'vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon } from '@hugeicons/core-free-icons'
import { starapp, type RbacPermission } from '../api/client'

const permissions = ref<RbacPermission[]>([])
const error = ref('')

onMounted(async () => {
  try {
    const res = await starapp.listRbacPermissions()
    permissions.value = res.permissions || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Load failed'
  }
})
</script>

<template>
  <Section title="Permission catalog" :padding="false">
    <template #toolbar>
      <router-link :to="{ name: 'iam' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>IAM</span>
      </router-link>
    </template>
    <p v-if="error" class="inline-notification error">{{ error }}</p>
    <Table
      :data="permissions.map((p) => ({ name: p.name, description: p.description || '—' }))"
      :headers="[
        { key: 'name', label: 'Name' },
        { key: 'description', label: 'Description' },
      ]"
    />
  </Section>
</template>
