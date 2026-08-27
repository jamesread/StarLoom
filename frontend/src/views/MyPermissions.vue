<script setup lang="ts">
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { starapp } from '../api/client'

const router = useRouter()

const audit = ref<Awaited<ReturnType<typeof starapp.getMyPermissionsAudit>> | null>(null)
const error = ref('')

onMounted(async () => {
  try {
    audit.value = await starapp.getMyPermissionsAudit()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Load failed'
  }
})
</script>

<template>
  <Section title="My permissions" :padding="true">
    <template #toolbar>
      <button type="button" class="inline-icon neutral" @click="router.push({ name: 'userControlPanel' })">
        Back
      </button>
    </template>
    <p v-if="error" class="inline-notification error">{{ error }}</p>
    <template v-if="audit">
      <p v-if="audit.isSuperuser" class="inline-notification note">Superuser — all permissions granted.</p>
      <p><strong>Groups:</strong> {{ (audit.groupNames || []).join(', ') || '—' }}</p>
      <p><strong>Roles:</strong> {{ (audit.roleNames || []).join(', ') || '—' }}</p>
      <Table
        :data="(audit.permissions || []).map((p) => ({
          permission: p.permission,
          granted: p.granted ? 'yes' : 'no',
          groups: (p.grantingGroups || []).join(', ') || (audit.isSuperuser ? 'superuser' : '—'),
        }))"
        :headers="[
          { key: 'permission', label: 'Permission' },
          { key: 'granted', label: 'Granted' },
          { key: 'groups', label: 'Granting groups' },
        ]"
      />
    </template>
  </Section>
</template>
