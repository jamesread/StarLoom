<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import CheckGroup from 'picocrank/vue/components/CheckGroup.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import DangerZone from 'picocrank/vue/components/DangerZone.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, Key01Icon, UserGroupIcon, UserIcon } from '@hugeicons/core-free-icons'
import { starapp, type UserAccount, type UserGroup } from '../api/client'
import { useStatus } from '../composables/useStatus'
import { hasPermission } from '../lib/rbacAccess'

const route = useRoute()
const router = useRouter()
const status = useStatus()
const groupId = computed(() => Number(route.params.id))

const group = ref<UserGroup | null>(null)
const users = ref<UserAccount[]>([])
const memberIds = ref<number[]>([])
const roleIds = ref<number[]>([])
const allRoles = ref<{ id: number; name: string }[]>([])
const loading = ref(true)
const error = ref('')
const membersError = ref('')
const rolesError = ref('')
const savingMembers = ref(false)
const savingRoles = ref(false)
const deleting = ref(false)
const deleteError = ref('')

const canManage = computed(() => hasPermission(status.status, 'usergroups.manage'))
const sectionTitle = computed(() => group.value?.name || 'User group')

const memberUsernames = computed(() =>
  memberIds.value
    .map((id) => users.value.find((user) => user.id === id)?.username)
    .filter((username): username is string => Boolean(username)),
)

const assignedRoleNames = computed(() =>
  roleIds.value
    .map((id) => allRoles.value.find((role) => role.id === id)?.name)
    .filter((name): name is string => Boolean(name)),
)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [groupsRes, usersRes, rolesRes] = await Promise.all([
      starapp.listUserGroups(),
      starapp.getUsers(),
      starapp.listRbacRoles(),
    ])
    group.value = (groupsRes.groups || []).find((g) => g.id === groupId.value) || null
    users.value = usersRes.users || []
    allRoles.value = (rolesRes.roles || []).map((role) => ({ id: role.id, name: role.name }))
    if (!group.value) {
      error.value = 'User group not found'
      return
    }
    const [members, roles] = await Promise.all([
      starapp.getUserGroupMembers({ groupId: groupId.value }),
      starapp.getUserGroupRbacRoles({ groupId: groupId.value }),
    ])
    memberIds.value = (members.members || []).map((m) => m.id)
    roleIds.value = roles.roleIds || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Load failed'
    group.value = null
  } finally {
    loading.value = false
  }
}

async function saveMembers() {
  if (!group.value || !canManage.value) return
  savingMembers.value = true
  membersError.value = ''
  try {
    await starapp.setUserGroupMembers({ groupId: group.value.id, userIds: memberIds.value })
    router.push({ name: 'user-groups' })
  } catch (e) {
    membersError.value = e instanceof Error ? e.message : String(e)
  } finally {
    savingMembers.value = false
  }
}

async function saveRoles() {
  if (!group.value || !canManage.value) return
  savingRoles.value = true
  rolesError.value = ''
  try {
    await starapp.setUserGroupRbacRoles({ groupId: group.value.id, roleIds: roleIds.value })
    router.push({ name: 'user-groups' })
  } catch (e) {
    rolesError.value = e instanceof Error ? e.message : String(e)
  } finally {
    savingRoles.value = false
  }
}

async function removeGroup() {
  if (!group.value) return
  if (!confirm(`Delete user group "${group.value.name}"?`)) return
  deleting.value = true
  deleteError.value = ''
  try {
    await starapp.deleteUserGroup({ groupId: group.value.id })
    router.push({ name: 'user-groups' })
  } catch (e) {
    deleteError.value = e instanceof Error ? e.message : String(e)
  } finally {
    deleting.value = false
  }
}

onMounted(load)
</script>

<template>
  <Section :title="sectionTitle" :icon="UserGroupIcon" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'user-groups' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>User groups</span>
      </RouterLink>
      <button type="button" class="inline-icon neutral" :disabled="loading" @click="load">
        Refresh
      </button>
    </template>

    <p v-if="loading" class="subtle">Loading…</p>
    <p v-if="error" class="inline-notification error">{{ error }}</p>

    <dl v-else-if="group" class="group-info">
      <dt>Name</dt>
      <dd>{{ group.name }}</dd>
      <dt>Members</dt>
      <dd>
        {{ memberIds.length }}
        <span v-if="memberUsernames.length" class="subtle">
          —
          {{ memberUsernames.join(', ') }}
        </span>
        <span v-else class="subtle"> — no members yet</span>
      </dd>
      <dt>Roles</dt>
      <dd>
        <span v-if="assignedRoleNames.length">{{ assignedRoleNames.join(', ') }}</span>
        <span v-else class="subtle">No roles assigned</span>
      </dd>
    </dl>
  </Section>

  <Section
    v-if="group && !loading"
    title="Members"
    subtitle="Users who belong to this group."
    :icon="UserIcon"
    :padding="true"
  >
    <FormLayout @submit.prevent="saveMembers">
      <FormField label="Members" component-has-label>
        <CheckGroup
          v-model="memberIds"
          name="group-members"
          :options="users.map((u) => ({ value: u.id, label: u.username }))"
          :disabled="!canManage || savingMembers"
        />
      </FormField>
      <p v-if="membersError" class="inline-notification error">{{ membersError }}</p>
      <p v-if="!canManage" class="subtle">You can view members but not change them.</p>
      <template v-if="canManage" #actions>
        <RouterLink :to="{ name: 'user-groups' }" class="button neutral">Cancel</RouterLink>
        <button type="submit" class="good" :disabled="savingMembers">
          {{ savingMembers ? 'Saving…' : 'Save members' }}
        </button>
      </template>
    </FormLayout>
  </Section>

  <Section
    v-if="group && !loading"
    title="Roles"
    subtitle="Roles assigned to this group grant permissions to its members."
    :icon="Key01Icon"
    :padding="true"
  >
    <FormLayout @submit.prevent="saveRoles">
      <FormField label="Roles" component-has-label>
        <CheckGroup
          v-model="roleIds"
          name="group-roles"
          :options="allRoles.map((r) => ({ value: r.id, label: r.name }))"
          :disabled="!canManage || savingRoles"
        />
      </FormField>
      <p v-if="rolesError" class="inline-notification error">{{ rolesError }}</p>
      <p v-if="!canManage" class="subtle">You can view roles but not change them.</p>
      <template v-if="canManage" #actions>
        <RouterLink :to="{ name: 'user-groups' }" class="button neutral">Cancel</RouterLink>
        <button type="submit" class="good" :disabled="savingRoles">
          {{ savingRoles ? 'Saving…' : 'Save roles' }}
        </button>
      </template>
    </FormLayout>
  </Section>

  <DangerZone
    v-if="group && canManage"
    title="Danger zone"
    subtitle="Delete this user group"
    :warning="`Deleting ${group.name} removes the group and its role assignments. Users are not deleted.`"
  >
    <p v-if="deleteError" class="inline-notification error">{{ deleteError }}</p>
    <div role="toolbar" class="danger-zone-actions">
      <button type="button" class="bad" :disabled="deleting" @click="removeGroup">
        {{ deleting ? 'Deleting…' : 'Delete user group' }}
      </button>
    </div>
  </DangerZone>
</template>

<style scoped>
.group-info {
  display: grid;
  grid-template-columns: 200px 1fr;
  column-gap: 1em;
  margin: 0;
}

.group-info dt {
  font-weight: bold;
}

.group-info dd {
  margin: 0;
}
</style>
