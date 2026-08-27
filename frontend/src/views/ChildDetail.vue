<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import { UserIcon } from '@hugeicons/core-free-icons'
import {
  starapp,
  memberAvatarUrl,
  type FamilyMember,
  type StarLedgerEntry,
} from '../api/client'

const route = useRoute()
const router = useRouter()
const memberId = computed(() => Number(route.params.id))

const member = ref<FamilyMember | null>(null)
const balance = ref(0)
const ledger = ref<StarLedgerEntry[]>([])
const error = ref('')
const awarding = ref(false)
const awardForm = reactive({ amount: 1, note: '' })

async function load() {
  try {
    const members = await starapp.listMembers()
    member.value = (members.members || []).find((m) => m.id === memberId.value) || null
    if (!member.value) {
      error.value = 'Child not found'
      return
    }
    const bal = await starapp.getMemberBalance({ memberId: memberId.value })
    balance.value = bal.balance ?? 0
    ledger.value = (await starapp.listLedger({ memberId: memberId.value, limit: 30 })).entries || []
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function award() {
  awarding.value = true
  try {
    const res = await starapp.awardStars({
      childMemberId: memberId.value,
      amount: awardForm.amount,
      note: awardForm.note,
    })
    balance.value = res.newBalance ?? balance.value
    awardForm.note = ''
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    awarding.value = false
  }
}

async function onAvatarChange(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    const res = await starapp.uploadMemberAvatar({ memberId: memberId.value, file })
    member.value = res.member || member.value
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function removeAvatar() {
  try {
    const res = await starapp.deleteMemberAvatar({ memberId: memberId.value })
    member.value = res.member || member.value
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function removeChild() {
  if (!confirm(`Remove ${member.value?.displayName}?`)) return
  try {
    await starapp.deleteMember({ memberId: memberId.value })
    router.push({ name: 'familyChildren' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(load)
</script>

<template>
  <Section :title="member?.displayName || 'Child'" :icon="UserIcon">
    <p v-if="error" class="error">{{ error }}</p>
    <template v-if="member">
      <div class="profile-row">
        <img
          v-if="member.hasAvatar"
          class="avatar"
          :src="memberAvatarUrl(member.id, true)"
          :alt="member.displayName"
        />
        <div v-else class="avatar avatar-placeholder">{{ member.displayName.charAt(0) }}</div>
        <div>
          <div class="balance">★ {{ balance }}</div>
          <label class="subtle">
            Upload avatar (JPEG, PNG, or WebP)
            <input type="file" accept="image/jpeg,image/png,image/webp,.jpg,.jpeg,.png,.webp" @change="onAvatarChange" />
          </label>
          <button v-if="member.hasAvatar" type="button" class="secondary outline small" @click="removeAvatar">
            Remove avatar
          </button>
        </div>
      </div>

      <h3>Award stars</h3>
      <div class="award-form">
        <FormField label="Amount">
          <input v-model.number="awardForm.amount" type="number" min="1" max="100" />
        </FormField>
        <FormField label="Note">
          <input v-model="awardForm.note" type="text" placeholder="Optional" />
        </FormField>
        <button type="button" :disabled="awarding" @click="award">Award</button>
      </div>

      <h3>History</h3>
      <ul v-if="ledger.length" class="ledger-list">
        <li v-for="entry in ledger" :key="entry.id">
          <span :class="{ positive: entry.amount > 0, negative: entry.amount < 0 }">
            {{ entry.amount > 0 ? '+' : '' }}{{ entry.amount }}
          </span>
          <span v-if="entry.note"> {{ entry.note }}</span>
          <span class="subtle"> — {{ entry.createdAt }}</span>
        </li>
      </ul>
      <p v-else class="subtle">No ledger entries yet.</p>

      <button type="button" class="outline contrast" @click="removeChild">Remove child</button>
    </template>
  </Section>
</template>

<style scoped>
.profile-row {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
  margin-bottom: 1.5rem;
}
.avatar {
  width: 5rem;
  height: 5rem;
  border-radius: 50%;
  object-fit: cover;
}
.avatar-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--pico-muted-border-color);
  font-size: 2rem;
  font-weight: bold;
}
.balance {
  font-size: 1.75rem;
  margin-bottom: 0.5rem;
}
.award-form {
  display: grid;
  gap: 0.5rem;
  max-width: 24rem;
  margin-bottom: 1.5rem;
}
.ledger-list {
  padding-left: 1.25rem;
}
.positive {
  color: var(--pico-ins-color);
}
.negative {
  color: var(--pico-del-color);
}
.error {
  color: var(--pico-del-color);
}
</style>
