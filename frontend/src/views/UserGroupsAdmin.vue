<script setup lang="ts">
import Section from 'picocrank/vue/components/Section.vue'
import CheckGroup from 'picocrank/vue/components/CheckGroup.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { onMounted, ref } from 'vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon } from '@hugeicons/core-free-icons'
import { starapp, type UserAccount, type UserGroup } from '../api/client'

const groups = ref<UserGroup[]>([])
const users = ref<UserAccount[]>([])
const selectedGroupId = ref<number | null>(null)
const memberIds = ref<number[]>([])
const roleIds = ref<number[]>([])
const allRoles = ref<{ id: number; name: string }[]>([])
const error = ref('')

async function load() {
  const [g, u, r] = await Promise.all([
    starapp.listUserGroups(),
    starapp.getUsers(),
    starapp.listRbacRoles(),
  ])
  groups.value = g.groups || []
  users.value = u.users || []
  allRoles.value = (r.roles || []).map((role) => ({ id: role.id, name: role.name }))
  if (!selectedGroupId.value && groups.value.length) {
    selectedGroupId.value = groups.value[0].id
    await loadGroupDetails()
  }
}

async function loadGroupDetails() {
  if (!selectedGroupId.value) return
  const [members, roles] = await Promise.all([
    starapp.getUserGroupMembers({ groupId: selectedGroupId.value }),
    starapp.getUserGroupRbacRoles({ groupId: selectedGroupId.value }),
  ])
  memberIds.value = (members.members || []).map((m) => m.id)
  roleIds.value = roles.roleIds || []
}

async function saveMembers() {
  if (!selectedGroupId.value) return
  await starapp.setUserGroupMembers({ groupId: selectedGroupId.value, userIds: memberIds.value })
}

async function saveRoles() {
  if (!selectedGroupId.value) return
  await starapp.setUserGroupRbacRoles({ groupId: selectedGroupId.value, roleIds: roleIds.value })
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
  <Section title="User groups" :padding="true">
    <template #toolbar>
      <router-link :to="{ name: 'iam' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>IAM</span>
      </router-link>
    </template>
    <p v-if="error" class="inline-notification error">{{ error }}</p>
    <FormLayout @submit.prevent="saveMembers">
      <FormField label="Group" for="user-group-select">
        <select id="user-group-select" v-model="selectedGroupId" @change="loadGroupDetails">
          <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }} ({{ g.memberCount }} members)</option>
        </select>
      </FormField>
      <FormField label="Members" fake>
        <CheckGroup
          v-model="memberIds"
          name="group-members"
          :options="users.map((u) => ({ value: u.id, label: u.username }))"
        />
      </FormField>
      <template #actions>
        <button type="submit" class="good">Save members</button>
      </template>
    </FormLayout>
    <FormLayout @submit.prevent="saveRoles">
      <FormField label="Roles for this group" fake>
        <CheckGroup
          v-model="roleIds"
          name="group-roles"
          :options="allRoles.map((r) => ({ value: r.id, label: r.name }))"
        />
      </FormField>
      <template #actions>
        <button type="submit" class="good">Save roles</button>
      </template>
    </FormLayout>
  </Section>
</template>
