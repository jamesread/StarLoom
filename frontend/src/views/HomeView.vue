<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Navigation from 'picocrank/vue/components/Navigation.vue'
import RewardNavigationGrid from '../components/RewardNavigationGrid.vue'
import RecentAwardsList from '../components/RecentAwardsList.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { StarIcon, PlusSignIcon, GiftIcon } from '@hugeicons/core-free-icons'
import { HugeiconsIcon } from '@hugeicons/vue'
import { starapp, type ParentHomeSummary, type ChildHomeSummary, type Reward, type TodaysChore } from '../api/client'
import MemberAvatar from '../components/MemberAvatar.vue'
import TodaysStarChart from '../components/TodaysStarChart.vue'
import TodayChoreProgress from '../components/TodayChoreProgress.vue'
import { fetchAppStatus, useStatus } from '../composables/useStatus'
import {
  canViewChildHomeFromStatus,
  canViewFamilyHomeFromStatus,
  canApproveRedemptionsFromStatus,
  hasPermission,
} from '../lib/rbacAccess'
import { memberStarStyle } from '../lib/memberStarColor'

const iconStrokeWidth = 2.5

const statusState = useStatus()
const error = ref('')
const loading = ref(true)
const familyName = ref('')
const creating = ref(false)

const parentSummary = ref<ParentHomeSummary | null>(null)
const childSummary = ref<ChildHomeSummary | null>(null)
const personalSummary = ref<ChildHomeSummary | null>(null)
const rewardNavRef = ref<InstanceType<typeof Navigation> | null>(null)
const redeemingRewardId = ref<number | null>(null)
const redeemError = ref('')
const todaysChoreError = ref('')
const todaysChoreBusyKey = ref('')

const isParentHome = computed(
  () =>
    canViewFamilyHomeFromStatus(statusState.status) ||
    hasPermission(statusState.status, 'family.manage'),
)
const needsFamily = computed(
  () => isParentHome.value && !parentSummary.value?.family?.id,
)
const canCreateFamily = computed(() => hasPermission(statusState.status, 'family.manage'))
const showAddPerson = computed(
  () => isParentHome.value && !needsFamily.value && !loading.value,
)
const isChildHome = computed(
  () => !isParentHome.value && canViewChildHomeFromStatus(statusState.status),
)
const showChildRewardsSection = computed(() => isChildHome.value && !loading.value && Boolean(childSummary.value))
const showChildAwardsSection = computed(() => isChildHome.value && !loading.value && Boolean(childSummary.value))
const showPersonalRewardsSection = computed(
  () =>
    isParentHome.value &&
    !loading.value &&
    Boolean(personalSummary.value) &&
    hasPermission(statusState.status, 'redemptions.request'),
)
const showPersonalAwardsSection = computed(
  () =>
    isParentHome.value &&
    !loading.value &&
    Boolean(personalSummary.value) &&
    hasPermission(statusState.status, 'stars.view_own'),
)
const rewardsSummary = computed(() =>
  isChildHome.value ? childSummary.value : personalSummary.value,
)
const todaysChores = computed<TodaysChore[]>(() => {
  if (isParentHome.value) {
    return parentSummary.value?.todaysChores || []
  }
  return childSummary.value?.todaysChores || []
})
const todaysChoresByChart = computed(() => {
  const groups = new Map<number, { id: number; name: string; chores: TodaysChore[] }>()
  for (const chore of todaysChores.value) {
    const id = chore.starChartId ?? 0
    const name = chore.starChartName?.trim() || 'Star Chart'
    let group = groups.get(id)
    if (!group) {
      group = { id, name, chores: [] }
      groups.set(id, group)
    }
    group.chores.push(chore)
  }
  return [...groups.values()].sort((a, b) =>
    a.name.localeCompare(b.name, undefined, { sensitivity: 'base' }),
  )
})
const showTodaysChores = computed(
  () => !loading.value && !needsFamily.value && todaysChoresByChart.value.length > 0,
)

const pendingRewardIds = computed(
  () => new Set(rewardsSummary.value?.pendingRewardIds || []),
)

const unavailableRewardIds = computed(
  () => new Set(rewardsSummary.value?.unavailableRewardIds || []),
)

const sectionTitle = computed(() => {
  if (loading.value) return 'Home'
  if (isParentHome.value) {
    if (needsFamily.value) return 'Home'
    return parentSummary.value?.family?.name || 'Home'
  }
  return childSummary.value?.member?.displayName || 'My stars'
})

async function loadPersonalSummary() {
  personalSummary.value = null
  if (!hasPermission(statusState.status, 'redemptions.request')) {
    return
  }
  try {
    personalSummary.value = await starapp.getChildHomeSummary()
    redeemError.value = ''
    await nextTick()
    setupRewardNav()
  } catch {
    personalSummary.value = null
  }
}

async function loadParent() {
  error.value = ''
  try {
    parentSummary.value = await starapp.getParentHomeSummary()
    if (parentSummary.value?.family?.id) {
      await loadPersonalSummary()
    } else {
      personalSummary.value = null
    }
    return
  } catch {
    parentSummary.value = null
    personalSummary.value = null
  }
  try {
    const fam = await starapp.getMyFamily()
    if (fam.family?.id) {
      parentSummary.value = { family: fam.family, children: [] }
      await loadPersonalSummary()
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

function starsNeeded(reward: Pick<Reward, 'costStars'>) {
  const balance = rewardsSummary.value?.balance ?? 0
  const cost = reward.costStars ?? 0
  return Math.max(0, cost - balance)
}

function isPendingApproval(rewardId: number) {
  return pendingRewardIds.value.has(rewardId)
}

function isUnavailableNow(rewardId: number) {
  return unavailableRewardIds.value.has(rewardId)
}

/** Approvers cannot self-redeem rewards that need approval (another parent must grant them). */
function isSelfApprovalBlocked(reward: Pick<Reward, 'approvalRequired'>) {
  return Boolean(reward.approvalRequired) && canApproveRedemptionsFromStatus(statusState.status)
}

function rewardDescriptionParts(reward: Reward): { description: string; starCount?: number } {
  if (isPendingApproval(reward.id)) {
    return { description: 'Pending approval' }
  }
  if (isSelfApprovalBlocked(reward)) {
    return { description: 'Needs another parent to award' }
  }
  if (isUnavailableNow(reward.id)) {
    return { description: 'Not currently available' }
  }
  const needed = starsNeeded(reward)
  if (needed > 0) {
    return {
      starCount: needed,
      description: needed === 1 ? 'more star' : 'more stars',
    }
  }
  const cost = reward.costStars ?? 0
  return {
    starCount: cost,
    description: cost === 1 ? 'star · Tap to redeem' : 'stars · Tap to redeem',
  }
}

function setupRewardNav() {
  const nav = rewardNavRef.value
  if (!nav) return
  nav.clearNavigationLinks()
  for (const reward of rewardsSummary.value?.rewards || []) {
    const needed = starsNeeded(reward)
    const pending = isPendingApproval(reward.id)
    const unavailable = isUnavailableNow(reward.id)
    const selfBlocked = isSelfApprovalBlocked(reward)
    const blocked = pending || unavailable || selfBlocked || needed > 0
    const disabled = blocked || redeemingRewardId.value === reward.id
    const { description, starCount } = rewardDescriptionParts(reward)
    nav.addNavigationLink({
      name: `reward-${reward.id}`,
      type: 'callback',
      icon: GiftIcon,
      callback: () => {
        void redeem(reward.id)
      },
      title: reward.title,
      description,
      starCount,
      disabled,
      rewardBlocked: disabled,
    })
  }
}

async function loadChild() {
  try {
    childSummary.value = await starapp.getChildHomeSummary()
    error.value = ''
    redeemError.value = ''
    await nextTick()
    setupRewardNav()
  } catch (e) {
    error.value =
      'Your account is not linked to a family profile. A parent must add you via Control Panel → People (not IAM → Users).'
    if (e instanceof Error && !e.message.includes('failed_precondition')) {
      error.value = e.message
    }
  }
}

async function load() {
  loading.value = true
  error.value = ''
  todaysChoreError.value = ''
  try {
    await fetchAppStatus()
    const st = statusState.status
    if (canViewFamilyHomeFromStatus(st)) {
      await loadParent()
    } else if (canViewChildHomeFromStatus(st)) {
      await loadChild()
    } else if (hasPermission(st, 'family.manage')) {
      await loadParent()
    } else {
      error.value =
        'Your account is not linked to a family. Sign in as a parent, create the family on Home, then add people from Control Panel → People.'
    }
  } finally {
    loading.value = false
  }
}

async function createFamily() {
  if (!familyName.value.trim()) return
  creating.value = true
  try {
    await starapp.createFamily({ name: familyName.value.trim() })
    familyName.value = ''
    await loadParent()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    creating.value = false
  }
}

function formatAward(entry?: { amount?: number; note?: string; createdAt?: string }) {
  if (!entry) return 'No recent awards'
  const note = entry.note ? ` ${entry.note}` : ''
  return `+${entry.amount}${note}`
}

async function toggleTodaysChore(chore: TodaysChore) {
  if (chore.paused || !chore.choreId || !chore.childMemberId) return
  todaysChoreBusyKey.value = `${chore.choreId}-${chore.childMemberId}`
  todaysChoreError.value = ''
  try {
    const body = {
      choreId: chore.choreId,
      childMemberId: chore.childMemberId,
      date: chore.date || undefined,
    }
    if (chore.completed) {
      await starapp.uncompleteChore(body)
    } else {
      await starapp.completeChore(body)
    }
    if (isParentHome.value) {
      await loadParent()
    } else {
      await loadChild()
    }
  } catch (e) {
    todaysChoreError.value = e instanceof Error ? e.message : String(e)
  } finally {
    todaysChoreBusyKey.value = ''
  }
}

async function redeem(rewardId: number) {
  const reward = rewardsSummary.value?.rewards?.find((entry) => entry.id === rewardId)
  if (
    !reward ||
    isPendingApproval(rewardId) ||
    isUnavailableNow(rewardId) ||
    isSelfApprovalBlocked(reward) ||
    starsNeeded(reward) > 0
  ) {
    return
  }
  redeemingRewardId.value = rewardId
  redeemError.value = ''
  setupRewardNav()
  try {
    await starapp.requestRedemption({ rewardId })
    if (isParentHome.value) {
      await loadParent()
    } else {
      await loadChild()
    }
  } catch (e) {
    redeemError.value = e instanceof Error ? e.message : String(e)
  } finally {
    redeemingRewardId.value = null
    await nextTick()
    setupRewardNav()
  }
}

onMounted(load)
watch(() => statusState.status?.isLoggedIn, () => {
  void load()
})
watch(rewardNavRef, () => {
  setupRewardNav()
})
watch(
  () => [
    rewardsSummary.value?.balance,
    rewardsSummary.value?.rewards,
    rewardsSummary.value?.pendingRewardIds,
    rewardsSummary.value?.unavailableRewardIds,
  ],
  () => {
    void nextTick(setupRewardNav)
  },
  { deep: true },
)
</script>

<template>
  <Section :title="sectionTitle" :icon="StarIcon">
    <template v-if="showAddPerson" #toolbar>
      <RouterLink
        :to="{ name: 'familyPersonCreate' }"
        class="button inline-icon good"
        aria-label="Add person"
        title="Add person"
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

    <p v-if="loading" class="subtle">Loading…</p>

    <template v-else-if="isParentHome">
      <p v-if="error" class="error">{{ error }}</p>

      <div v-if="needsFamily && canCreateFamily">
        <p>Create your family to start awarding stars.</p>
        <FormLayout @submit.prevent="createFamily">
          <FormField label="Family name" for="family-name">
            <input id="family-name" v-model="familyName" type="text" placeholder="The Smith Family" required />
          </FormField>
          <template #actions>
            <button type="submit" class="good" :disabled="creating">{{ creating ? 'Creating…' : 'Create family' }}</button>
          </template>
        </FormLayout>
      </div>

      <div v-else-if="needsFamily" class="subtle">
        <p>No family has been set up yet. An account with family management permission must create one.</p>
      </div>

      <template v-else>
        <p v-if="parentSummary?.pendingRedemptions" class="subtle">
          Pending redemption requests:
          <RouterLink :to="{ name: 'familyRewards' }">{{ parentSummary.pendingRedemptions }}</RouterLink>
        </p>

        <div v-if="!parentSummary?.children?.length" class="subtle">
          <p>No people yet.</p>
          <RouterLink :to="{ name: 'familyPersonCreate' }">Add your first person</RouterLink>
        </div>

        <div v-else class="people-grid">
          <RouterLink
            v-for="person in parentSummary?.children"
            :key="person.member?.id"
            class="people-card"
            :to="{ name: 'familyPersonDetail', params: { id: person.member?.id } }"
          >
            <MemberAvatar v-if="person.member" :member="person.member" size="lg" />
            <strong>{{ person.member?.displayName }}</strong>
            <div class="balance" :style="memberStarStyle(person.member)">★ {{ person.balance ?? 0 }}</div>
            <div class="subtle last-award">{{ formatAward(person.lastAward) }}</div>
            <TodayChoreProgress :progress="person.todayStarChartProgress || []" />
          </RouterLink>
        </div>
      </template>
    </template>

    <template v-else>
      <p v-if="error" class="error">{{ error }}</p>
      <div class="child-home-header">
        <MemberAvatar
          v-if="childSummary?.member"
          :member="childSummary.member"
          size="xl"
        />
        <div class="child-balance">
          <p class="child-balance-label">Your stars</p>
          <div class="balance large" :style="memberStarStyle(childSummary?.member)">
            ★ {{ childSummary?.balance ?? 0 }}
          </div>
        </div>
      </div>
    </template>
  </Section>

  <template v-if="showTodaysChores">
    <p v-if="todaysChoreError" class="inline-notification error list-banner-pad todays-chore-error">
      {{ todaysChoreError }}
    </p>
    <Section
      v-for="group in todaysChoresByChart"
      :key="group.id"
      subtitle="Tap today's cell to mark a chore done."
      :padding="false"
    >
      <template #title>
        <span class="section-title-with-icon">
          <HugeiconsIcon :icon="StarIcon" width="22" height="22" aria-hidden="true" />
          <RouterLink
            :to="
              group.id > 0
                ? { name: 'familyStarChartView', params: { id: group.id } }
                : { name: 'familyStarChart' }
            "
            class="title-link"
          >
            {{ group.name }}
          </RouterLink>
          <span>: Today's chores</span>
        </span>
      </template>
      <TodaysStarChart
        :chores="group.chores"
        :busy-key="todaysChoreBusyKey"
        @toggle="toggleTodaysChore"
      />
    </Section>
  </template>

  <Section
    v-if="showPersonalRewardsSection"
    title="Reward Shop"
    subtitle="Save up stars and tap a reward to redeem it."
    :icon="GiftIcon"
    :padding="true"
  >
    <p v-if="redeemError" class="inline-notification error">{{ redeemError }}</p>
    <Navigation v-if="personalSummary?.rewards?.length" ref="rewardNavRef">
      <RewardNavigationGrid :member="personalSummary?.member" />
    </Navigation>
    <p v-else class="subtle">No rewards available right now.</p>
  </Section>

  <Section
    v-if="showChildRewardsSection"
    title="Reward Shop"
    subtitle="Save up stars and tap a reward to redeem it."
    :icon="GiftIcon"
    :padding="true"
  >
    <p v-if="redeemError" class="inline-notification error">{{ redeemError }}</p>
    <Navigation v-if="childSummary?.rewards?.length" ref="rewardNavRef">
      <RewardNavigationGrid :member="childSummary?.member" />
    </Navigation>
    <p v-else class="subtle">No rewards available right now.</p>
  </Section>

  <Section
    v-if="showPersonalAwardsSection"
    title="Your recent awards"
    subtitle="Stars you earned lately."
    :icon="StarIcon"
    :padding="true"
  >
    <RecentAwardsList
      :entries="personalSummary?.recentAwards || []"
      :member="personalSummary?.member"
    />
  </Section>

  <Section
    v-if="showChildAwardsSection"
    title="Recent awards"
    subtitle="Stars your family gave you lately."
    :icon="StarIcon"
    :padding="true"
  >
    <RecentAwardsList
      :entries="childSummary?.recentAwards || []"
      :member="childSummary?.member"
    />
  </Section>
</template>

<style scoped>
.people-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(10rem, 1fr));
  gap: 1rem;
}
.people-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.35rem;
  padding: 1rem;
  border: 1px solid var(--pico-muted-border-color);
  border-radius: var(--pico-border-radius);
  text-decoration: none;
  color: inherit;
}
.people-card:hover {
  border-color: var(--pico-primary);
}
.balance {
  font-size: 1.25rem;
  font-weight: 700;
}
.balance.large {
  font-size: 1.75rem;
}
.child-home-header {
  display: flex;
  align-items: center;
  gap: 1rem;
}
.child-balance-label {
  margin: 0 0 0.15rem;
  font-size: 0.95rem;
  color: var(--muted-text-color, var(--pico-muted-color));
}
.child-balance {
  display: flex;
  flex-direction: column;
}
.last-award {
  font-size: 0.85rem;
  text-align: center;
}
.error {
  color: var(--pico-del-color);
}
.list-banner-pad {
  padding-left: 1em;
  padding-right: 1em;
}
.todays-chore-error {
  margin-bottom: 0.5rem;
}
.section-title-with-icon {
  display: inline-flex;
  align-items: center;
  gap: 0.45em;
  vertical-align: middle;
}
.title-link {
  color: var(--pico-primary);
  text-decoration: none;
}
.title-link:hover {
  text-decoration: underline;
}
</style>
