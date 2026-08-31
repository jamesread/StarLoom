<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import DangerZone from 'picocrank/vue/components/DangerZone.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, UserIcon } from '@hugeicons/core-free-icons'
import { starapp, memberAvatarFileUrl, type FamilyMember, type MemberAvatarEntry } from '../api/client'
import MemberAvatar from '../components/MemberAvatar.vue'
import { memberAvatarStyle, memberStarColor } from '../lib/memberStarColor'

const route = useRoute()
const router = useRouter()
const memberId = computed(() => Number(route.params.id))

const member = ref<FamilyMember | null>(null)
const error = ref('')
const loading = ref(true)
const savingProfile = ref(false)
const assigningLogin = ref(false)
const loginError = ref('')
const displayName = ref('')
const starColor = ref('#3498db')
const loginForm = reactive({
  username: '',
  password: '',
})
const avatarHistory = ref<MemberAvatarEntry[]>([])
const avatarLoading = ref(false)
const selectingAvatar = ref('')

const avatarBorderStyle = computed(() => memberAvatarStyle(member.value))

const hasLogin = computed(() => Boolean(member.value?.userAccountId))

const canAssignLogin = computed(() => {
  if (!loginForm.username.trim()) return false
  return loginForm.password.length >= 8
})

const sectionTitle = computed(() =>
  member.value ? `Edit ${member.value.displayName}` : 'Edit person',
)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const members = await starapp.listMembers()
    member.value = (members.members || []).find((m) => m.id === memberId.value) || null
    if (!member.value) {
      error.value = 'Person not found'
      return
    }
    displayName.value = member.value.displayName
    starColor.value = memberStarColor(member.value)
    await loadAvatarHistory()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function loadAvatarHistory() {
  avatarLoading.value = true
  try {
    const res = await starapp.listMemberAvatars({ memberId: memberId.value })
    avatarHistory.value = res.avatars || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    avatarLoading.value = false
  }
}

async function assignLogin() {
  if (!member.value || hasLogin.value) return
  assigningLogin.value = true
  loginError.value = ''
  try {
    const res = await starapp.assignMemberLogin({
      memberId: memberId.value,
      username: loginForm.username.trim(),
      password: loginForm.password,
    })
    member.value = res.member || member.value
    loginForm.username = ''
    loginForm.password = ''
  } catch (e) {
    loginError.value = e instanceof Error ? e.message : String(e)
  } finally {
    assigningLogin.value = false
  }
}

async function saveProfile() {
  if (!member.value) return
  savingProfile.value = true
  error.value = ''
  try {
    const res = await starapp.updateMember({
      memberId: memberId.value,
      displayName: displayName.value.trim(),
      starColor: starColor.value,
    })
    member.value = res.member || member.value
    starColor.value = memberStarColor(member.value)
    router.push({ name: 'familyPersonDetail', params: { id: memberId.value } })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    savingProfile.value = false
  }
}

async function onAvatarChange(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    const res = await starapp.uploadMemberAvatar({ memberId: memberId.value, file })
    member.value = res.member || member.value
    starColor.value = memberStarColor(member.value)
    await loadAvatarHistory()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function selectAvatar(filename: string) {
  selectingAvatar.value = filename
  error.value = ''
  try {
    const res = await starapp.selectMemberAvatar({ memberId: memberId.value, filename })
    member.value = res.member || member.value
    starColor.value = memberStarColor(member.value)
    await loadAvatarHistory()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    selectingAvatar.value = ''
  }
}

async function removeAvatar() {
  try {
    const res = await starapp.deleteMemberAvatar({ memberId: memberId.value })
    member.value = res.member || member.value
    await loadAvatarHistory()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function removePerson() {
  if (!confirm(`Remove ${member.value?.displayName}?`)) return
  try {
    await starapp.deleteMember({ memberId: memberId.value })
    router.push({ name: 'familyPeople' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(load)
</script>

<template>
  <Section :title="sectionTitle" :icon="UserIcon" :padding="true">
    <template #toolbar>
      <RouterLink
        :to="{ name: 'familyPersonDetail', params: { id: memberId } }"
        class="button inline-icon neutral"
      >
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Back</span>
      </RouterLink>
    </template>

    <p v-if="loading" class="subtle">Loading…</p>
    <p v-else-if="error && !member" class="inline-notification error">{{ error }}</p>

    <template v-else-if="member">
      <p v-if="error" class="inline-notification error">{{ error }}</p>

      <div class="profile-row">
        <MemberAvatar :member="member" size="xl" />
        <div>
          <label class="subtle">
            Upload avatar (JPEG, PNG, or WebP)
            <input type="file" accept="image/jpeg,image/png,image/webp,.jpg,.jpeg,.png,.webp" @change="onAvatarChange" />
          </label>
          <button v-if="member.hasAvatar" type="button" class="secondary outline small" @click="removeAvatar">
            Remove avatar
          </button>
        </div>
      </div>

      <FormField
        label="Previous avatars"
        component-has-label
        description="Choose an earlier upload to use it again. Uploading a new image keeps earlier versions here."
      >
        <p v-if="avatarLoading" class="muted">Loading previous avatars…</p>
        <p v-else-if="!avatarHistory.length" class="muted">No previous avatars yet.</p>
        <div v-else class="avatar-history" role="list">
          <button
            v-for="entry in avatarHistory"
            :key="entry.filename"
            type="button"
            class="avatar-history-item"
            :class="{ current: entry.isCurrent }"
            :disabled="Boolean(selectingAvatar) || entry.isCurrent"
            :title="entry.isCurrent ? 'Current avatar' : 'Use this avatar'"
            role="listitem"
            @click="selectAvatar(entry.filename)"
          >
            <img
              :src="memberAvatarFileUrl(memberId, entry.filename)"
              :alt="entry.isCurrent ? 'Current avatar' : 'Previous avatar'"
              :style="avatarBorderStyle"
            />
          </button>
        </div>
      </FormField>

      <FormLayout class="profile-form" @submit.prevent="saveProfile">
        <FormField label="Display name" for="person-edit-display-name">
          <input
            id="person-edit-display-name"
            v-model="displayName"
            type="text"
            required
            :disabled="savingProfile"
          />
        </FormField>
        <FormField label="Star color" for="star-color" description="Used for stars on the chart and avatar border.">
          <input id="star-color" v-model="starColor" type="color" :disabled="savingProfile" />
        </FormField>
        <template #actions>
          <button type="submit" class="good" :disabled="savingProfile || !displayName.trim()">
            {{ savingProfile ? 'Saving…' : 'Save' }}
          </button>
          <RouterLink
            :to="{ name: 'familyPersonDetail', params: { id: memberId } }"
            class="button neutral"
          >
            Cancel
          </RouterLink>
        </template>
      </FormLayout>

      <div class="login-section">
        <h3>Login</h3>
        <template v-if="hasLogin">
          <p class="subtle">
            This person can sign in with username
            <strong>{{ member.username || '—' }}</strong>.
          </p>
        </template>
        <template v-else>
          <p class="subtle">
            This person cannot sign in yet. Create a username and password so they can log in.
          </p>
          <FormLayout class="login-form" @submit.prevent="assignLogin">
            <p v-if="loginError" class="inline-notification error">{{ loginError }}</p>
            <FormField label="Username" for="person-edit-username">
              <input
                id="person-edit-username"
                v-model="loginForm.username"
                type="text"
                autocomplete="off"
                required
                :disabled="assigningLogin"
              />
            </FormField>
            <FormField label="Password" for="person-edit-password" description="At least 8 characters.">
              <input
                id="person-edit-password"
                v-model="loginForm.password"
                type="password"
                minlength="8"
                autocomplete="new-password"
                required
                :disabled="assigningLogin"
              />
            </FormField>
            <template #actions>
              <button type="submit" class="good" :disabled="assigningLogin || !canAssignLogin">
                {{ assigningLogin ? 'Creating…' : 'Create login' }}
              </button>
            </template>
          </FormLayout>
        </template>
      </div>
    </template>
  </Section>

  <DangerZone
    v-if="member?.role === 'child'"
    title="Danger zone"
    subtitle="Remove this person from the family"
    :warning="`Removing ${member.displayName} deletes their star balance and history. This cannot be undone.`"
  >
    <div role="toolbar" class="danger-zone-actions">
      <button type="button" class="bad" @click="removePerson">Remove person</button>
    </div>
  </DangerZone>
</template>

<style scoped>
.profile-row {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
  margin-bottom: 1.5rem;
}
.avatar-history {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}
.avatar-history-item {
  display: inline-flex;
  padding: 0;
  border: 2px solid transparent;
  border-radius: 50%;
  background: none;
  cursor: pointer;
  line-height: 0;
}
.avatar-history-item:hover:not(:disabled) {
  border-color: var(--pico-primary);
}
.avatar-history-item.current {
  border-color: var(--pico-primary);
  cursor: default;
}
.avatar-history-item:disabled:not(.current) {
  opacity: 0.6;
  cursor: wait;
}
.avatar-history-item img {
  width: 4rem;
  height: 4rem;
  border-radius: 50%;
  object-fit: cover;
  box-sizing: border-box;
}
.profile-form {
  max-width: 24rem;
  margin-bottom: 1.5rem;
}
.login-section {
  max-width: 24rem;
  margin-bottom: 1.5rem;
}
.login-section h3 {
  margin-bottom: 0.5rem;
}
.login-form {
  margin-top: 1rem;
}
</style>
