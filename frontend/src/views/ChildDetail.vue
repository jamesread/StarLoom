<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import {
  ArrowLeft01Icon,
  Clock01Icon,
  PencilEdit01Icon,
  Refresh01Icon,
  StarIcon,
  UserIcon,
} from '@hugeicons/core-free-icons'
import {
  starapp,
  type FamilyMember,
  type StarLedgerEntry,
} from '../api/client'
import MemberAvatar from '../components/MemberAvatar.vue'
import { useStatus } from '../composables/useStatus'
import { hasPermission } from '../lib/rbacAccess'
import { memberStarStyle } from '../lib/memberStarColor'

const iconStrokeWidth = 2.5

const route = useRoute()
const statusState = useStatus()
const memberId = computed(() => Number(route.params.id))

const member = ref<FamilyMember | null>(null)
const balance = ref(0)
const ledger = ref<StarLedgerEntry[]>([])
const error = ref('')
const awardError = ref('')
const ledgerError = ref('')
const loading = ref(true)
const awarding = ref(false)
const revokingId = ref<number | null>(null)
const awardForm = reactive({ amount: 1, note: '' })

const starStyle = computed(() => memberStarStyle(member.value))
const canRevoke = computed(() => hasPermission(statusState.status, 'stars.revoke'))
const myMemberId = ref(0)
const isOwnProfile = computed(
  () => Boolean(myMemberId.value) && memberId.value === myMemberId.value,
)
const canAwardBonus = computed(
  () => hasPermission(statusState.status, 'stars.award') && !isOwnProfile.value,
)

const ledgerHeaders = [
  { key: 'createdAt', label: 'Date', sortable: true, width: '11rem' },
  { key: 'entryType', label: 'Type', sortable: true, width: '7rem' },
  { key: 'amount', label: 'Stars', sortable: true, width: '6rem' },
  { key: 'note', label: 'Note', sortable: false },
  { key: 'actions', label: 'Actions', sortable: false, width: '7rem' },
]

const ledgerRows = computed(() =>
  ledger.value.map((entry) => ({
    id: entry.id,
    createdAt: entry.createdAt || '—',
    entryType: formatEntryType(entry.entryType),
    amount: entry.amount,
    amountLabel: entry.amount > 0 ? `+${entry.amount}` : String(entry.amount),
    note: entry.note || '—',
    raw: entry,
    actions: '',
  })),
)

function formatEntryType(entryType?: string) {
  switch (entryType) {
    case 'award':
      return 'Award'
    case 'revoke':
      return 'Revoke'
    case 'redeem':
      return 'Redeem'
    default:
      return entryType || '—'
  }
}

async function load() {
  loading.value = true
  error.value = ''
  ledgerError.value = ''
  try {
    const [membersRes, familyRes] = await Promise.all([
      starapp.listMembers(),
      starapp.getMyFamily().catch(() => null),
    ])
    myMemberId.value = familyRes?.callerMember?.id || 0
    member.value = (membersRes.members || []).find((m) => m.id === memberId.value) || null
    if (!member.value) {
      error.value = 'Person not found'
      ledger.value = []
      return
    }
    const bal = await starapp.getMemberBalance({ memberId: memberId.value })
    balance.value = bal.balance ?? 0
    ledger.value = (await starapp.listLedger({ memberId: memberId.value, limit: 100 })).entries || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function award() {
  awarding.value = true
  awardError.value = ''
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
    awardError.value = e instanceof Error ? e.message : String(e)
  } finally {
    awarding.value = false
  }
}

async function revokeEntry(entry: StarLedgerEntry) {
  if (entry.entryType !== 'award' || entry.amount <= 0) return
  const amount = Math.min(entry.amount, balance.value)
  if (amount <= 0) return
  const note = entry.note ? `Revoke: ${entry.note}` : `Revoke award #${entry.id}`
  if (!confirm(`Revoke ${amount} star${amount === 1 ? '' : 's'} from this award?`)) return
  revokingId.value = entry.id
  ledgerError.value = ''
  try {
    const res = await starapp.revokeStars({
      childMemberId: memberId.value,
      amount,
      note,
    })
    balance.value = res.newBalance ?? balance.value
    await load()
  } catch (e) {
    ledgerError.value = e instanceof Error ? e.message : String(e)
  } finally {
    revokingId.value = null
  }
}

onMounted(load)
</script>

<template>
  <Section :title="member?.displayName || 'Person'" :icon="UserIcon">
    <template #toolbar>
      <RouterLink :to="{ name: 'familyPeople' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>People</span>
      </RouterLink>
      <RouterLink
        :to="{ name: 'familyPersonEdit', params: { id: memberId } }"
        class="button inline-icon good"
      >
        <HugeiconsIcon
          :icon="PencilEdit01Icon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>Edit</span>
      </RouterLink>
    </template>

    <p v-if="error" class="inline-notification error">{{ error }}</p>
    <p v-if="loading && !member" class="muted">Loading…</p>

    <template v-if="member">
      <div class="profile-row">
        <MemberAvatar :member="member" size="xl" />
        <div class="balance" :style="starStyle">★ {{ balance }}</div>
      </div>
    </template>
  </Section>

  <Section
    v-if="member && canAwardBonus"
    subtitle="Give extra stars outside of chores."
    :padding="true"
  >
    <template #title>
      <span class="section-title-with-icon">
        <HugeiconsIcon :icon="StarIcon" width="22" height="22" aria-hidden="true" />
        Bonus stars
      </span>
    </template>

    <FormLayout class="award-form" @submit.prevent="award">
      <p v-if="awardError" class="inline-notification error">{{ awardError }}</p>
      <FormField label="Amount" for="award-amount">
        <input id="award-amount" v-model.number="awardForm.amount" type="number" min="1" max="100" required />
      </FormField>
      <FormField label="Note" for="award-note">
        <input id="award-note" v-model="awardForm.note" type="text" placeholder="Optional" />
      </FormField>
      <template #actions>
        <button type="submit" class="good" :disabled="awarding">{{ awarding ? 'Awarding…' : 'Award' }}</button>
      </template>
    </FormLayout>
  </Section>

  <Section
    v-if="member"
    subtitle="Awards, revokes, and redemptions for this person."
    classes="ledger-list-section"
    :padding="false"
  >
    <template #title>
      <span class="section-title-with-icon">
        <HugeiconsIcon :icon="Clock01Icon" width="22" height="22" aria-hidden="true" />
        Star history
      </span>
    </template>

    <template #toolbar>
      <button
        type="button"
        class="inline-icon neutral"
        aria-label="Refresh"
        title="Refresh"
        :disabled="loading"
        @click="load"
      >
        <HugeiconsIcon
          :icon="Refresh01Icon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
      </button>
    </template>

    <p v-if="ledgerError" class="inline-notification error list-banner-pad">{{ ledgerError }}</p>
    <div v-if="loading && !ledger.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!ledger.length" class="inline-notification note list-banner-pad">No ledger entries yet.</p>

      <Table
        v-else
        class="list-table-wrap"
        :data="ledgerRows"
        :headers="ledgerHeaders"
        :show-pagination="ledger.length > 10"
      >
        <template #cell-amount="{ row }">
          <span :class="{ positive: row.amount > 0, negative: row.amount < 0 }">{{ row.amountLabel }}</span>
        </template>
        <template #cell-actions="{ row }">
          <div class="actions-cell">
            <button
              v-if="canRevoke && row.raw.entryType === 'award' && row.amount > 0 && balance > 0"
              type="button"
              class="bad small"
              :disabled="revokingId === row.id"
              @click="revokeEntry(row.raw)"
            >
              {{ revokingId === row.id ? 'Revoking…' : 'Revoke' }}
            </button>
          </div>
        </template>
      </Table>
    </template>
  </Section>
</template>

<style scoped>
.section-title-with-icon {
  display: inline-flex;
  align-items: center;
  gap: 0.45em;
  vertical-align: middle;
}

.list-banner-pad {
  padding-left: 1em;
  padding-right: 1em;
}

.list-table-wrap {
  margin-top: 0.5rem;
  margin-bottom: 1.5rem;
}

.actions-cell {
  text-align: right;
}

.profile-row {
  display: flex;
  gap: 1rem;
  align-items: center;
  margin-bottom: 1.5rem;
}
.balance {
  font-size: 1.75rem;
  margin-bottom: 0.5rem;
  font-weight: 700;
}
.award-form {
  max-width: 24rem;
  margin-bottom: 1.5rem;
}
.positive {
  color: var(--pico-ins-color);
  font-weight: 600;
}
.negative {
  color: var(--pico-del-color);
  font-weight: 600;
}
</style>
