<script setup lang="ts">
import Section from 'picocrank/vue/components/Section.vue'
import { onMounted, ref } from 'vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon } from '@hugeicons/core-free-icons'
import { starapp, type RbacRole } from '../api/client'

const roles = ref<RbacRole[]>([])
const permissions = ref<{ id: number; name: string }[]>([])
const error = ref('')

async function load() {
  const [r, p] = await Promise.all([starapp.listRbacRoles(), starapp.listRbacPermissions()])
  roles.value = r.roles || []
  permissions.value = (p.permissions || []).map((x) => ({ id: x.id, name: x.name }))
}

async function saveRole(role: RbacRole, permissionIds: number[]) {
  await starapp.updateRbacRole({
    id: role.id,
    name: role.name,
    description: role.description || '',
    permissionIds,
  })
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
  <Section title="Roles" subtitle="Assign permissions to roles, then attach roles to groups" :padding="true">
    <template #toolbar>
      <router-link :to="{ name: 'iam' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>IAM</span>
      </router-link>
    </template>
    <p v-if="error" class="inline-notification error">{{ error }}</p>
    <article v-for="role in roles" :key="role.id" class="role-card">
      <h3>
        {{ role.name }}
        <span v-if="role.name === 'superuser'" class="tag">all access</span>
        <span v-if="role.name === 'superuser' || role.name === 'member'" class="tag">system role</span>
      </h3>
      <p>{{ role.description }}</p>
      <table v-if="role.name !== 'superuser'" class="perm-table data-table">
        <thead>
          <tr>
            <th>Grant</th>
            <th>Permission</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in permissions" :key="p.id">
            <td>
              <input
                type="checkbox"
                :checked="(role.permissionIds || []).includes(p.id)"
                @change="(e) => {
                  const checked = (e.target as HTMLInputElement).checked
                  const ids = new Set(role.permissionIds || [])
                  checked ? ids.add(p.id) : ids.delete(p.id)
                  saveRole(role, [...ids])
                }"
              />
            </td>
            <td><code>{{ p.name }}</code></td>
          </tr>
        </tbody>
      </table>
    </article>
  </Section>
</template>

<style scoped>
.role-card {
  margin-bottom: 2rem;
}
.tag {
  font-size: 0.75rem;
  margin-left: 0.5rem;
}
</style>
