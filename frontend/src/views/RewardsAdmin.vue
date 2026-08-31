<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { GiftIcon, PlusSignIcon, Refresh01Icon, TaskDone01Icon } from '@hugeicons/core-free-icons'
import { starapp, type FamilyMember, type Redemption, type Reward } from '../api/client'

const iconStrokeWidth = 2.5

const rewards = ref<Reward[]>([])
const members = ref<FamilyMember[]>([])
const pendingRedemptions = ref<Redemption[]>([])
const error = ref('')
const redemptionsError = ref('')
const loading = ref(true)
const redemptionsLoading = ref(true)

const awardDialog = ref<HTMLDialogElement | null>(null)
const awardReward = ref<Reward | null>(null)
const awardChildId = ref(0)
const awardError = ref('')
const awardMessage = ref('')
const awarding = ref(false)

const childOptions = computed(() =>
  members.value
    .filter((m) => m.role === 'child')
    .sort((a, b) => a.displayName.localeCompare(b.displayName, undefined, { sensitivity: 'base' }))
    .map((m) => ({ label: m.displayName, value: m.id })),
)

const tableHeaders = [
  { key: 'title', label: 'Title', sortable: true },
  { key: 'costStars', label: 'Cost', sortable: true, width: '6rem' },
  { key: 'availability', label: 'Availability', sortable: false, width: '10rem' },
  { key: 'status', label: 'Status', sortable: true, width: '6rem' },
  { key: 'approval', label: 'Approval', sortable: true, width: '8rem' },
  { key: 'actions', label: 'Actions', sortable: false, width: '10rem' },
]

function availabilityLabel(expression?: string) {
  const trimmed = expression?.trim()
  return trimmed ? trimmed : 'Always'
}

const listRows = computed(() =>
  rewards.value.map((r) => ({
    id: r.id,
    title: r.title,
    costStars: r.costStars,
    availability: availabilityLabel(r.availabilityExpression),
    status: r.active !== false ? 'Active' : 'Inactive',
    approval: r.approvalRequired !== false ? 'Required' : 'Auto',
    active: r.active !== false,
    actions: '',
  })),
)

const redemptionHeaders = [
  { key: 'childDisplayName', label: 'Person', sortable: true },
  { key: 'rewardTitle', label: 'Reward', sortable: true },
  { key: 'starsSpent', label: 'Stars', sortable: true, width: '6rem' },
  { key: 'createdAt', label: 'Requested', sortable: true, width: '11rem' },
  { key: 'actions', label: 'Actions', sortable: false, width: '10rem' },
]

const redemptionRows = computed(() =>
  pendingRedemptions.value.map((r) => ({
    id: r.id,
    childDisplayName: r.childDisplayName || '—',
    rewardTitle: r.rewardTitle || '—',
    starsSpent: r.starsSpent,
    createdAt: r.createdAt || '—',
    actions: '',
  })),
)

async function loadRewards() {
  loading.value = true
  try {
    const [rewardRes, memberRes] = await Promise.all([
      starapp.listRewards({ includeInactive: true }),
      starapp.listMembers(),
    ])
    rewards.value = rewardRes.rewards || []
    members.value = memberRes.members || []
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function loadRedemptions() {
  redemptionsLoading.value = true
  try {
    const res = await starapp.listRedemptions({ status: 'pending' })
    pendingRedemptions.value = res.redemptions || []
    redemptionsError.value = ''
  } catch (e) {
    redemptionsError.value = e instanceof Error ? e.message : String(e)
  } finally {
    redemptionsLoading.value = false
  }
}

async function load() {
  await Promise.all([loadRewards(), loadRedemptions()])
}

function openAwardDialog(rewardId: number) {
  const reward = rewards.value.find((r) => r.id === rewardId)
  if (!reward) return
  awardReward.value = reward
  awardChildId.value = childOptions.value[0]?.value || 0
  awardError.value = ''
  awardMessage.value = ''
  awardDialog.value?.showModal()
}

function closeAwardDialog() {
  awardDialog.value?.close()
}

function resetAwardForm() {
  awardReward.value = null
  awardChildId.value = 0
  awardError.value = ''
  awarding.value = false
}

async function awardRewardToChild() {
  if (!awardReward.value?.id || !awardChildId.value) {
    awardError.value = 'Select a person.'
    return
  }
  awarding.value = true
  awardError.value = ''
  awardMessage.value = ''
  try {
    const res = await starapp.requestRedemption({
      rewardId: awardReward.value.id,
      childMemberId: awardChildId.value,
    })
    awardMessage.value = res.standardResponse?.message || 'Reward awarded'
    await load()
    await nextTick()
    closeAwardDialog()
  } catch (e) {
    awardError.value = e instanceof Error ? e.message : String(e)
  } finally {
    awarding.value = false
  }
}

async function approveRedemption(id: number) {
  try {
    await starapp.approveRedemption({ redemptionId: id })
    await loadRedemptions()
  } catch (e) {
    redemptionsError.value = e instanceof Error ? e.message : String(e)
  }
}

async function rejectRedemption(id: number) {
  try {
    await starapp.rejectRedemption({ redemptionId: id })
    await loadRedemptions()
  } catch (e) {
    redemptionsError.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(load)
</script>

<template>
  <dialog ref="awardDialog" class="dialog" @close="resetAwardForm">
    <h2>Award reward</h2>
    <p v-if="awardReward">
      Give <strong>{{ awardReward.title }}</strong> ({{ awardReward.costStars }} stars) to a family member.
    </p>
    <FormLayout v-if="childOptions.length" @submit.prevent="awardRewardToChild">
      <FormField label="Person" component-has-label>
        <RadioGroup
          v-model="awardChildId"
          variant="list"
          :options="childOptions"
          name="award-reward-person"
        />
      </FormField>
      <p v-if="awardError" class="inline-notification error">{{ awardError }}</p>
      <p v-else-if="awardMessage" class="inline-notification note">{{ awardMessage }}</p>
      <template #actions>
        <button type="button" class="neutral" :disabled="awarding" @click="closeAwardDialog">Cancel</button>
        <button type="submit" class="good" :disabled="awarding || !awardChildId">
          {{ awarding ? 'Awarding…' : 'Award' }}
        </button>
      </template>
    </FormLayout>
    <template v-else>
      <p class="inline-notification note">No family members to award rewards to yet.</p>
      <div class="dialog-actions">
        <button type="button" class="neutral" @click="closeAwardDialog">Close</button>
      </div>
    </template>
  </dialog>

  <Section
    subtitle="Privileges children can redeem with stars."
    classes="rewards-list"
    :padding="false"
  >
    <template #title>
      <span class="section-title-with-icon">
        <HugeiconsIcon :icon="GiftIcon" width="22" height="22" aria-hidden="true" />
        Rewards
      </span>
    </template>

    <template #toolbar>
      <button
        type="button"
        class="inline-icon neutral"
        aria-label="Refresh"
        title="Refresh"
        :disabled="loading || redemptionsLoading"
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
      <RouterLink
        :to="{ name: 'familyRewardCreate' }"
        class="button inline-icon good"
        aria-label="Add reward"
        title="Add reward"
      >
        <HugeiconsIcon
          :icon="PlusSignIcon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
      </RouterLink>
    </template>

    <p v-if="error" class="inline-notification error list-banner-pad">{{ error }}</p>
    <div v-if="loading && !rewards.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!rewards.length" class="inline-notification note list-banner-pad">
        No rewards yet.
        <RouterLink :to="{ name: 'familyRewardCreate' }">Add a reward</RouterLink>.
      </p>

      <Table
        v-else
        class="list-table-wrap"
        :data="listRows"
        :headers="tableHeaders"
        :show-pagination="rewards.length > 10"
      >
        <template #cell-title="{ value, row }">
          <RouterLink :to="{ name: 'familyRewardEdit', params: { id: row.id } }" class="title-link">
            <strong>{{ value }}</strong>
          </RouterLink>
        </template>
        <template #cell-availability="{ value }">
          <span class="availability-cell" :title="value">{{ value }}</span>
        </template>
        <template #cell-actions="{ row }">
          <div class="actions-cell">
            <RouterLink :to="{ name: 'familyRewardEdit', params: { id: row.id } }" class="button neutral small">
              Edit
            </RouterLink>
            <button
              v-if="row.active"
              type="button"
              class="good small"
              @click="openAwardDialog(row.id)"
            >
              Award
            </button>
          </div>
        </template>
      </Table>
    </template>
  </Section>

  <Section
    subtitle="Pending redemption requests awaiting parent approval."
    :padding="false"
  >
    <template #title>
      <span class="section-title-with-icon">
        <HugeiconsIcon :icon="TaskDone01Icon" width="22" height="22" aria-hidden="true" />
        Redemption requests
      </span>
    </template>

    <template #toolbar>
      <button
        type="button"
        class="inline-icon neutral"
        aria-label="Refresh redemption requests"
        title="Refresh redemption requests"
        :disabled="redemptionsLoading"
        @click="loadRedemptions"
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

    <p v-if="redemptionsError" class="inline-notification error list-banner-pad">{{ redemptionsError }}</p>
    <div v-if="redemptionsLoading && !pendingRedemptions.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!pendingRedemptions.length" class="inline-notification note list-banner-pad">
        No pending redemption requests.
      </p>

      <Table
        v-else
        class="list-table-wrap"
        :data="redemptionRows"
        :headers="redemptionHeaders"
        :show-pagination="pendingRedemptions.length > 10"
      >
        <template #cell-actions="{ row }">
          <div class="actions-cell">
            <button type="button" class="good small" @click="approveRedemption(row.id)">Approve</button>
            <button type="button" class="neutral small" @click="rejectRedemption(row.id)">Reject</button>
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
  display: flex;
  justify-content: flex-end;
  gap: 0.35rem;
}

.title-link {
  font: inherit;
  color: var(--pico-primary);
  text-decoration: none;
}

.title-link:hover {
  text-decoration: underline;
}

.availability-cell {
  display: inline-block;
  max-width: 10rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: ui-monospace, monospace;
  font-size: 0.85em;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
