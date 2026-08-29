<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Navigation from 'picocrank/vue/components/Navigation.vue'
import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import { StarIcon } from '@hugeicons/core-free-icons'
import { starapp, type ParentHomeSummary, type ChildHomeSummary, type StarChart } from '../api/client'
import MemberAvatar from '../components/MemberAvatar.vue'
import { fetchAppStatus, useStatus } from '../composables/useStatus'
import {
  canViewChildHomeFromStatus,
  canViewFamilyHomeFromStatus,
  hasPermission,
} from '../lib/rbacAccess'
import { memberStarStyle } from '../lib/memberStarColor'

const statusState = useStatus()
const router = useRouter()
const error = ref('')
const loading = ref(true)
const familyName = ref('')
const creating = ref(false)

const parentSummary = ref<ParentHomeSummary | null>(null)
const childSummary = ref<ChildHomeSummary | null>(null)
const starCharts = ref<StarChart[]>([])
const starChartNavRef = ref<InstanceType<typeof Navigation> | null>(null)

const isParentHome = computed(
  () =>
    canViewFamilyHomeFromStatus(statusState.status) ||
    hasPermission(statusState.status, 'family.manage'),
)
const needsFamily = computed(
  () => isParentHome.value && !parentSummary.value?.family?.id,
)
const canCreateFamily = computed(() => hasPermission(statusState.status, 'family.manage'))

const sectionTitle = computed(() => {
  if (loading.value) return 'Home'
  if (isParentHome.value) {
    if (needsFamily.value) return 'Home'
    return parentSummary.value?.family?.name || 'Home'
  }
  return childSummary.value?.member?.displayName || 'My stars'
})

async function loadParent() {
  error.value = ''
  try {
    parentSummary.value = await starapp.getParentHomeSummary()
    if (parentSummary.value?.family?.id) {
      await loadStarCharts()
    } else {
      starCharts.value = []
    }
    return
  } catch {
    parentSummary.value = null
    starCharts.value = []
  }
  try {
    const fam = await starapp.getMyFamily()
    if (fam.family?.id) {
      parentSummary.value = { family: fam.family, children: [] }
      await loadStarCharts()
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function loadStarCharts() {
  try {
    const res = await starapp.listStarCharts()
    starCharts.value = (res.starCharts || []).filter((chart) => chart.active !== false)
  } catch {
    starCharts.value = []
  }
  await nextTick()
  setupStarChartNav()
}

function setupStarChartNav() {
  const nav = starChartNavRef.value
  if (!nav) return
  nav.clearNavigationLinks()
  for (const chart of starCharts.value) {
    const count = chart.choreCount ?? 0
    nav.addCallback(
      chart.name,
      () => {
        void router.push({ name: 'familyStarChartView', params: { id: chart.id } })
      },
      {
        name: `star-chart-${chart.id}`,
        icon: StarIcon,
        description: count === 1 ? '1 chore' : `${count} chores`,
      },
    )
  }
}

async function loadChild() {
  try {
    childSummary.value = await starapp.getChildHomeSummary()
    error.value = ''
  } catch (e) {
    error.value =
      'Your account is not linked to a family profile. A parent must add you via Family → People (not IAM → Users).'
    if (e instanceof Error && !e.message.includes('failed_precondition')) {
      error.value = e.message
    }
  }
}

async function load() {
  loading.value = true
  error.value = ''
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
        'Your account is not linked to a family. Sign in as a parent, create the family on Home, then add people from Family → People.'
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

async function redeem(rewardId: number) {
  try {
    await starapp.requestRedemption({ rewardId })
    await loadChild()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(load)
watch(() => statusState.status?.isLoggedIn, () => {
  void load()
})
watch(starChartNavRef, () => {
  setupStarChartNav()
})
</script>

<template>
  <Section :title="sectionTitle" :icon="StarIcon">
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
          Pending redemptions:
          <RouterLink :to="{ name: 'familyRedemptions' }">{{ parentSummary.pendingRedemptions }}</RouterLink>
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
        <div class="balance large" :style="memberStarStyle(childSummary?.member)">
          ★ {{ childSummary?.balance ?? 0 }} stars
        </div>
      </div>

      <h3>Recent awards</h3>
      <ul v-if="childSummary?.recentAwards?.length" class="award-list">
        <li v-for="entry in childSummary.recentAwards" :key="entry.id">
          +{{ entry.amount }}
          <span v-if="entry.note"> {{ entry.note }}</span>
          <span class="subtle"> — {{ entry.createdAt }}</span>
        </li>
      </ul>
      <p v-else class="subtle">No awards yet.</p>

      <h3>Rewards you can get</h3>
      <ul v-if="childSummary?.rewards?.length" class="reward-list">
        <li v-for="reward in childSummary.rewards" :key="reward.id">
          {{ reward.title }} — {{ reward.costStars }} stars
          <button
            type="button"
            class="outline"
            :disabled="(childSummary.balance ?? 0) < reward.costStars"
            @click="redeem(reward.id)"
          >
            Redeem
          </button>
        </li>
      </ul>
      <p v-else class="subtle">No rewards available.</p>
    </template>
  </Section>

  <Section
    v-if="isParentHome && !needsFamily && !loading"
    title="Star charts"
    subtitle="Open a weekly chart to mark chore completions."
    :padding="true"
  >
    <Navigation v-if="starCharts.length" ref="starChartNavRef">
      <NavigationGrid />
    </Navigation>
    <p v-else class="subtle">
      No star charts yet.
      <RouterLink :to="{ name: 'familyStarCharts' }">Manage star charts</RouterLink>
    </p>
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
  margin-bottom: 1.5rem;
}
.award-list, .reward-list {
  padding-left: 1.25rem;
}
.last-award {
  font-size: 0.85rem;
  text-align: center;
}
.error {
  color: var(--pico-del-color);
}
</style>
