<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { GiftIcon, PlusSignIcon, Refresh01Icon, TaskDone01Icon, ArrowLeft01Icon } from '@hugeicons/core-free-icons'
import { starapp, type FamilyMember, type Redemption, type Reward } from '../api/client'

const iconStrokeWidth = 2.5

const rewards = ref<Reward[]>([])
const members = ref<FamilyMember[]>([])
const myMemberId = ref(0)
const pendingRedemptions = ref<Redemption[]>([])
const error = ref('')
const redemptionsError = ref('')
const loading = ref(true)
const redemptionsLoading = ref(true)

const awardDialog = ref<HTMLDialogElement | null>(null)
const createDialog = ref<HTMLDialogElement | null>(null)
const createTitleInput = ref<HTMLInputElement | null>(null)
const giveReward = ref<Reward | null>(null)
const giveChildId = ref(0)
const memberBalances = ref<Record<number, number>>({})
const giveBalancesLoading = ref(false)
const giveError = ref('')
const giveMessage = ref('')
const giving = ref(false)
const creating = ref(false)
const createError = ref('')

const createForm = reactive({
  title: '',
  description: '',
  costStars: 5,
  approvalRequired: true,
  active: true,
  availabilityExpression: '',
})

const booleanOptions = [
  { label: 'Yes', value: true },
  { label: 'No', value: false },
]

const statusOptions = [
  { label: 'Active', value: true },
  { label: 'Inactive', value: false },
]

const personOptions = computed(() =>
  members.value
    .filter((m) => m.id !== myMemberId.value)
    .sort((a, b) => a.displayName.localeCompare(b.displayName, undefined, { sensitivity: 'base' }))
    .map((m) => {
      const balance = memberBalances.value[m.id]
      const starsSuffix = balance === undefined ? '' : ` (${balance} stars)`
      return { label: `${m.displayName}${starsSuffix}`, value: m.id }
    }),
)

const selectedGivePerson = computed(() =>
  members.value.find((m) => m.id === giveChildId.value) ?? null,
)

const selectedGiveBalance = computed(() => {
  if (!giveChildId.value) return null
  const balance = memberBalances.value[giveChildId.value]
  return balance === undefined ? null : balance
})

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
    approval: r.approvalRequired === true ? 'Required' : 'Auto',
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
    childMemberId: r.childMemberId,
    childDisplayName: r.childDisplayName || '—',
    rewardTitle: r.rewardTitle || '—',
    starsSpent: r.starsSpent,
    createdAt: r.createdAt || '—',
    isOwn: Boolean(myMemberId.value) && r.childMemberId === myMemberId.value,
    actions: '',
  })),
)

async function loadRewards() {
  loading.value = true
  try {
    const [rewardRes, memberRes, familyRes] = await Promise.all([
      starapp.listRewards({ includeInactive: true }),
      starapp.listMembers(),
      starapp.getMyFamily().catch(() => null),
    ])
    rewards.value = rewardRes.rewards || []
    members.value = memberRes.members || []
    myMemberId.value = familyRes?.callerMember?.id || 0
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

function resetCreateForm() {
  createForm.title = ''
  createForm.description = ''
  createForm.costStars = 5
  createForm.approvalRequired = true
  createForm.active = true
  createForm.availabilityExpression = ''
  createError.value = ''
  creating.value = false
}

function openCreateDialog() {
  resetCreateForm()
  createDialog.value?.showModal()
  nextTick(() => createTitleInput.value?.focus())
}

function closeCreateDialog() {
  createDialog.value?.close()
}

async function submitCreateReward() {
  if (!createForm.title.trim()) {
    createError.value = 'Title is required.'
    return
  }
  creating.value = true
  createError.value = ''
  try {
    const res = await starapp.createReward({
      title: createForm.title.trim(),
      description: createForm.description.trim(),
      costStars: createForm.costStars,
      approvalRequired: createForm.approvalRequired,
      availabilityExpression: createForm.availabilityExpression.trim(),
    })
    const created = res.reward
    if (created?.id && !createForm.active) {
      await starapp.updateReward({
        id: created.id,
        title: createForm.title.trim(),
        description: createForm.description.trim(),
        costStars: createForm.costStars,
        active: false,
        approvalRequired: createForm.approvalRequired,
        availabilityExpression: createForm.availabilityExpression.trim(),
      })
    }
    closeCreateDialog()
    await load()
  } catch (e) {
    createError.value = e instanceof Error ? e.message : String(e)
  } finally {
    creating.value = false
  }
}

async function loadGiveBalances(memberIds: number[]) {
  giveBalancesLoading.value = true
  memberBalances.value = {}
  try {
    const entries = await Promise.all(
      memberIds.map(async (id) => {
        const res = await starapp.getMemberBalance({ memberId: id })
        return [id, res.balance ?? 0] as const
      }),
    )
    memberBalances.value = Object.fromEntries(entries)
  } catch (e) {
    giveError.value = e instanceof Error ? e.message : String(e)
  } finally {
    giveBalancesLoading.value = false
  }
}

function openGiveDialog(rewardId: number) {
  const reward = rewards.value.find((r) => r.id === rewardId)
  if (!reward) return
  giveReward.value = reward
  giveChildId.value = personOptions.value[0]?.value || 0
  giveError.value = ''
  giveMessage.value = ''
  const memberIds = members.value
    .filter((m) => m.id !== myMemberId.value)
    .map((m) => m.id)
  void loadGiveBalances(memberIds)
  awardDialog.value?.showModal()
}

function closeGiveDialog() {
  awardDialog.value?.close()
}

function resetGiveForm() {
  giveReward.value = null
  giveChildId.value = 0
  memberBalances.value = {}
  giveBalancesLoading.value = false
  giveError.value = ''
  giving.value = false
}

async function giveRewardToChild() {
  if (!giveReward.value?.id || !giveChildId.value) {
    giveError.value = 'Select a person.'
    return
  }
  if (myMemberId.value && giveChildId.value === myMemberId.value) {
    giveError.value = 'You cannot give a reward to yourself.'
    return
  }
  giving.value = true
  giveError.value = ''
  giveMessage.value = ''
  try {
    const res = await starapp.requestRedemption({
      rewardId: giveReward.value.id,
      childMemberId: giveChildId.value,
    })
    giveMessage.value = res.standardResponse?.message || 'Reward given'
    await load()
    await nextTick()
    closeGiveDialog()
  } catch (e) {
    giveError.value = e instanceof Error ? e.message : String(e)
  } finally {
    giving.value = false
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
  <dialog ref="createDialog" class="dialog" @close="resetCreateForm">
    <h2>Add reward</h2>
    <p>Define a privilege children can redeem with stars.</p>
    <FormLayout @submit.prevent="submitCreateReward">
      <FormField label="Title" for="reward-create-title">
        <input
          id="reward-create-title"
          ref="createTitleInput"
          v-model="createForm.title"
          type="text"
          required
          placeholder="Extra screen time"
          :disabled="creating"
        />
      </FormField>
      <FormField label="Description" for="reward-create-description">
        <textarea
          id="reward-create-description"
          v-model="createForm.description"
          rows="2"
          :disabled="creating"
        />
      </FormField>
      <FormField label="Cost (stars)" for="reward-create-cost">
        <input
          id="reward-create-cost"
          v-model.number="createForm.costStars"
          type="number"
          min="1"
          required
          :disabled="creating"
        />
      </FormField>
      <FormField label="Requires approval" component-has-label>
        <RadioGroup
          v-model="createForm.approvalRequired"
          name="reward-create-approval"
          variant="boolean"
          :options="booleanOptions"
        />
      </FormField>
      <FormField label="Status" component-has-label>
        <RadioGroup
          v-model="createForm.active"
          name="reward-create-active"
          variant="boolean"
          :options="statusOptions"
        />
      </FormField>
      <FormField
        label="Availability expression"
        for="reward-create-availability"
        description="Optional. expr language; leave blank for always available. Variables include balance, costStars, countPerDay, countPerWeek, hour, dayName, etc."
      >
        <textarea
          id="reward-create-availability"
          v-model="createForm.availabilityExpression"
          rows="3"
          placeholder='(hour > 9 && hour < 18) && (dayName == "Sat" || dayName == "Sun")'
          :disabled="creating"
        />
      </FormField>
      <p v-if="createError" class="inline-notification error">{{ createError }}</p>
      <template #actions>
        <button type="button" class="neutral" :disabled="creating" @click="closeCreateDialog">Cancel</button>
        <button type="submit" class="good" :disabled="creating || !createForm.title.trim()">
          {{ creating ? 'Creating…' : 'Create' }}
        </button>
      </template>
    </FormLayout>
  </dialog>

  <dialog ref="awardDialog" class="dialog" @close="resetGiveForm">
    <h2>Give Reward</h2>
    <p v-if="giveReward">
      Give <strong>{{ giveReward.title }}</strong> ({{ giveReward.costStars }} stars) to a family member.
    </p>
    <FormLayout v-if="personOptions.length" @submit.prevent="giveRewardToChild">
      <FormField label="Person" component-has-label>
        <RadioGroup
          v-model="giveChildId"
          variant="list"
          :options="personOptions"
          name="give-reward-person"
        />
      </FormField>
      <p v-if="giveBalancesLoading" class="person-balance muted">Loading star balances…</p>
      <p v-else-if="selectedGivePerson && selectedGiveBalance !== null" class="person-balance">
        <strong>{{ selectedGivePerson.displayName }}</strong> has {{ selectedGiveBalance }} stars.
      </p>
      <NotificationBlock v-if="giveError" type="error" :message="giveError" />
      <p v-else-if="giveMessage" class="inline-notification note">{{ giveMessage }}</p>
      <template #actions>
        <button type="button" class="neutral" :disabled="giving" @click="closeGiveDialog">Cancel</button>
        <button type="submit" class="good" :disabled="giving || !giveChildId || giveBalancesLoading">
          {{ giving ? 'Giving…' : 'Give' }}
        </button>
      </template>
    </FormLayout>
    <template v-else>
      <p class="inline-notification note">No other family members to give rewards to.</p>
      <div class="dialog-actions">
        <button type="button" class="neutral" @click="closeGiveDialog">Close</button>
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
      <RouterLink :to="{ name: 'controlPanel' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Control Panel</span>
      </RouterLink>
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
      <button
        type="button"
        class="inline-icon good"
        aria-label="Add reward"
        title="Add reward"
        @click="openCreateDialog"
      >
        <HugeiconsIcon
          :icon="PlusSignIcon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
      </button>
    </template>

    <p v-if="error" class="inline-notification error list-banner-pad">{{ error }}</p>
    <div v-if="loading && !rewards.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!rewards.length" class="inline-notification note list-banner-pad">
        No rewards yet.
        <button type="button" class="linkish" @click="openCreateDialog">Add a reward</button>.
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
              @click="openGiveDialog(row.id)"
            >
              Give
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
            <template v-if="row.isOwn">
              <span class="subtle">Awaiting another parent</span>
            </template>
            <template v-else>
              <button type="button" class="good small" @click="approveRedemption(row.id)">Approve</button>
              <button type="button" class="neutral small" @click="rejectRedemption(row.id)">Reject</button>
            </template>
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

.person-balance {
  margin: 0;
}

.linkish {
  background: none;
  border: none;
  padding: 0;
  font: inherit;
  color: var(--pico-primary);
  text-decoration: underline;
  cursor: pointer;
}
</style>
