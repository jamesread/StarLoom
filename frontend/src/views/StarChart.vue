<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, ArrowRight01Icon, PlusSignIcon, Refresh01Icon, StarIcon } from '@hugeicons/core-free-icons'
import {
  starapp,
  type StarChart as StarChartMeta,
  type WeeklyStarChart,
  type WeeklyStarChartBonusChild,
  type WeeklyStarChartChild,
  type WeeklyStarChartDay,
} from '../api/client'
import MemberAvatar from '../components/MemberAvatar.vue'
import { fetchAppStatus, useStatus } from '../composables/useStatus'
import { canCompleteChoresFromStatus, canViewFamilyHomeFromStatus } from '../lib/rbacAccess'
import { memberStarStyle } from '../lib/memberStarColor'

const statusState = useStatus()
const route = useRoute()
const router = useRouter()

const starChartId = computed(() => Number(route.params.id))
const chart = ref<WeeklyStarChart | null>(null)
const starCharts = ref<StarChartMeta[]>([])
const error = ref('')
const loading = ref(false)
const weekStart = ref('')

const iconStrokeWidth = 2.5

const canComplete = computed(() => canCompleteChoresFromStatus(statusState.status))
const canAddChores = computed(() => canViewFamilyHomeFromStatus(statusState.status))
const hasChores = computed(() => (chart.value?.rows?.length ?? 0) > 0)
const sectionTitle = computed(() => chart.value?.starChartName || 'Star Chart')
const activeStarCharts = computed(() => starCharts.value.filter((c) => c.active !== false))
const createChoreQuery = computed(() => ({
  create: '1',
  starChartId: String(starChartId.value),
}))

const dayLabels = computed(() => {
  if (!chart.value?.weekStart) return ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
  const start = new Date(chart.value.weekStart + 'T12:00:00')
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(start)
    d.setDate(d.getDate() + i)
    return d.toLocaleDateString(undefined, { weekday: 'short' })
  })
})

const weekLabel = computed(() => {
  if (!chart.value?.weekStart || !chart.value?.weekEnd) return ''
  return `${chart.value.weekStart} — ${chart.value.weekEnd}`
})

function mondayOf(date: Date): string {
  const d = new Date(date)
  const wd = d.getDay()
  const diff = wd === 0 ? -6 : 1 - wd
  d.setDate(d.getDate() + diff)
  return d.toISOString().slice(0, 10)
}

function shiftWeek(delta: number) {
  const base = weekStart.value || chart.value?.weekStart || mondayOf(new Date())
  const d = new Date(base + 'T12:00:00')
  d.setDate(d.getDate() + delta * 7)
  weekStart.value = d.toISOString().slice(0, 10)
  load()
}

async function load() {
  loading.value = true
  try {
    await fetchAppStatus()
    const [chartRes, chartsRes] = await Promise.all([
      starapp.getWeeklyStarChart({
        weekStart: weekStart.value || undefined,
        starChartId: starChartId.value,
      }),
      starapp.listStarCharts(),
    ])
    chart.value = chartRes
    starCharts.value = chartsRes.starCharts || []
    if (!weekStart.value && chart.value?.weekStart) {
      weekStart.value = chart.value.weekStart
    }
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

function switchChart(id: number) {
  if (id === starChartId.value) return
  router.push({ name: 'familyStarChartView', params: { id } })
}

function dayForChild(child: WeeklyStarChartChild, index: number): WeeklyStarChartDay | undefined {
  return child.days?.[index]
}

function bonusStarsForChild(child: WeeklyStarChartBonusChild, dayIndex: number): number {
  return child.days?.[dayIndex]?.starsEarned || 0
}

function bonusChildHasStars(child: WeeklyStarChartBonusChild): boolean {
  return (child.days || []).some((day) => (day.starsEarned || 0) > 0)
}

const visibleBonusChildren = computed(() =>
  (chart.value?.bonusChildren || []).filter(bonusChildHasStars),
)

function starsForRow(children: WeeklyStarChartChild[] | undefined, index: number): number {
  let total = 0
  for (const child of children || []) {
    total += dayForChild(child, index)?.starsEarned || 0
  }
  return total
}

function cellClass(day: WeeklyStarChartDay | undefined): string {
  if (!day) return ''
  if (day.paused) return 'paused'
  if (day.completed) return 'completed'
  if (day.scheduled) return 'scheduled'
  return 'off-day'
}

async function toggleCell(
  choreId: number,
  child: WeeklyStarChartChild,
  dayIndex: number,
) {
  if (!canComplete.value) return
  const day = dayForChild(child, dayIndex)
  if (!day?.scheduled || day.paused) return
  const childId = child.child?.id
  if (!childId || !day.date) return
  try {
    if (day.completed) {
      await starapp.uncompleteChore({ choreId, childMemberId: childId, date: day.date })
    } else {
      await starapp.completeChore({ choreId, childMemberId: childId, date: day.date })
    }
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(load)
watch(starChartId, () => {
  weekStart.value = ''
  load()
})
</script>

<template>
  <Section :title="sectionTitle" :icon="StarIcon" :padding="false">
    <template #toolbar>
      <label v-if="activeStarCharts.length > 1" class="chart-select">
        <span class="visually-hidden">Star chart</span>
        <select :value="starChartId" @change="switchChart(Number(($event.target as HTMLSelectElement).value))">
          <option v-for="sc in activeStarCharts" :key="sc.id" :value="sc.id">{{ sc.name }}</option>
        </select>
      </label>
      <button type="button" class="inline-icon neutral" :disabled="loading" @click="shiftWeek(-1)">
        <HugeiconsIcon
          :icon="ArrowLeft01Icon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>Previous</span>
      </button>
      <span v-if="weekLabel" class="week-label">{{ weekLabel }}</span>
      <button type="button" class="inline-icon neutral" :disabled="loading" @click="shiftWeek(1)">
        <HugeiconsIcon
          :icon="ArrowRight01Icon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>Next</span>
      </button>
      <button type="button" class="inline-icon neutral" :disabled="loading" aria-label="Refresh" @click="load">
        <HugeiconsIcon
          :icon="Refresh01Icon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
      </button>
    </template>

    <p v-if="error" class="inline-notification error list-banner-pad">{{ error }}</p>

    <p v-if="loading" class="list-banner-pad muted">Loading…</p>

    <div v-else-if="chart && !hasChores" class="list-banner-pad empty-chores">
      <p class="inline-notification note">
        <template v-if="canAddChores">
          No chores are defined yet. Add chores to start tracking stars on the chart.
        </template>
        <template v-else>No chores are defined yet. Ask a parent to add people and assign chores.</template>
      </p>
      <RouterLink
        v-if="canAddChores"
        :to="{ name: 'familyChores', query: createChoreQuery }"
        class="button inline-icon good"
      >
        <HugeiconsIcon
          :icon="PlusSignIcon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>Add chore</span>
      </RouterLink>
    </div>

    <div v-else-if="chart" class="chart-wrap">
      <table class="star-chart">
        <thead>
          <tr>
            <th class="chore-col">Chore</th>
            <th v-for="(label, i) in dayLabels" :key="i">{{ label }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in chart.rows" :key="row.choreId">
            <td class="chore-col">
              <span class="chore-title">{{ row.title }}</span>
              <span class="reward">★{{ row.starReward }}</span>
              <span class="child-icons">
                <template v-for="child in row.children" :key="child.assignmentId">
                  <MemberAvatar
                    v-if="child.child?.id"
                    :member="child.child"
                    size="md"
                    :to="{ name: 'familyPersonDetail', params: { id: child.child.id } }"
                  />
                </template>
              </span>
            </td>
            <td
              v-for="(_, dayIndex) in dayLabels"
              :key="dayIndex"
              :class="cellClass(dayForChild(row.children?.[0], dayIndex))"
            >
              <template v-if="row.children?.length === 1">
                <button
                  v-if="canComplete && dayForChild(row.children[0], dayIndex)?.scheduled && !dayForChild(row.children[0], dayIndex)?.paused"
                  type="button"
                  class="cell-btn"
                  @click="toggleCell(row.choreId!, row.children[0], dayIndex)"
                >
                  <span v-if="dayForChild(row.children[0], dayIndex)?.completed" class="star star-lg" :style="memberStarStyle(row.children[0].child)">★{{ row.starReward }}</span>
                  <span v-else class="dot dot-lg">○</span>
                </button>
                <span v-else-if="dayForChild(row.children[0], dayIndex)?.completed" class="star star-lg" :style="memberStarStyle(row.children[0].child)">★{{ row.starReward }}</span>
                <span v-else-if="dayForChild(row.children[0], dayIndex)?.paused" class="muted">—</span>
                <span v-else-if="!dayForChild(row.children[0], dayIndex)?.scheduled" class="muted">·</span>
              </template>
              <template v-else>
                <div class="multi-child">
                  <button
                    v-for="child in row.children"
                    :key="child.assignmentId"
                    type="button"
                    class="cell-btn mini"
                    :disabled="!canComplete || !dayForChild(child, dayIndex)?.scheduled || dayForChild(child, dayIndex)?.paused"
                    @click="toggleCell(row.choreId!, child, dayIndex)"
                  >
                    <span v-if="dayForChild(child, dayIndex)?.completed" class="star star-md" :style="memberStarStyle(child.child)">
                      ★{{ row.starReward }}
                    </span>
                    <span v-else-if="dayForChild(child, dayIndex)?.scheduled && !dayForChild(child, dayIndex)?.paused" class="dot dot-md">○</span>
                  </button>
                </div>
                <span v-if="starsForRow(row.children, dayIndex) > 0 && !canComplete" class="star star-md bonus-stars">
                  ★{{ starsForRow(row.children, dayIndex) }}
                </span>
              </template>
            </td>
          </tr>
          <tr v-if="visibleBonusChildren.length" class="bonus-row">
            <td class="chore-col">
              <strong>Bonus stars</strong>
              <span class="child-icons">
                <template v-for="child in visibleBonusChildren" :key="child.child?.id">
                  <MemberAvatar
                    v-if="child.child?.id"
                    :member="child.child"
                    size="md"
                    :to="{ name: 'familyPersonDetail', params: { id: child.child.id } }"
                  />
                </template>
              </span>
            </td>
            <td v-for="(_, dayIndex) in dayLabels" :key="dayIndex">
              <div class="multi-child">
                <span
                  v-for="child in visibleBonusChildren"
                  :key="`${child.child?.id}-${dayIndex}`"
                  v-show="bonusStarsForChild(child, dayIndex) > 0"
                  class="star star-md bonus-stars"
                  :style="memberStarStyle(child.child)"
                >
                  ★{{ bonusStarsForChild(child, dayIndex) }}
                </span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <p v-else class="list-banner-pad muted">No chart data.</p>
  </Section>
</template>

<style scoped>
.week-label {
  font-weight: 600;
  white-space: nowrap;
}
.chart-select select {
  min-width: 10rem;
}
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
.list-banner-pad {
  padding-left: 1em;
  padding-right: 1em;
}

.empty-chores {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.chart-wrap {
  overflow-x: auto;
}
.star-chart {
  width: 100%;
  min-width: 36rem;
}
.star-chart th,
.star-chart td {
  text-align: center;
  vertical-align: middle;
}
.chore-col {
  text-align: left !important;
  min-width: 10rem;
}
.chore-title {
  display: block;
  font-weight: 600;
}
.reward {
  font-size: 0.85rem;
  color: var(--pico-muted-color);
}
.child-icons {
  display: flex;
  gap: 0.25rem;
  margin-top: 0.25rem;
}
.cell-btn {
  background: none;
  border: none;
  padding: 0.35rem;
  cursor: pointer;
  min-width: 2.75rem;
}
.cell-btn.mini {
  font-size: 0.85rem;
  min-width: 2.5rem;
}
.cell-btn:disabled {
  cursor: default;
  opacity: 0.5;
}
.star {
  font-weight: 800;
  line-height: 1;
}
.star-lg {
  font-size: 2rem;
}
.star-md {
  font-size: 1.5rem;
}
.bonus-stars {
  color: var(--pico-primary);
}
.dot {
  color: var(--pico-muted-color);
}
.dot-lg {
  font-size: 1.75rem;
}
.dot-md {
  font-size: 1.35rem;
}
.muted {
  color: var(--pico-muted-color);
}
td.scheduled {
  background: color-mix(in srgb, var(--pico-primary) 8%, transparent);
}
td.completed {
  background: color-mix(in srgb, var(--pico-primary) 18%, transparent);
}
td.paused {
  background: color-mix(in srgb, var(--pico-muted-color) 12%, transparent);
}
.bonus-row td {
  border-top: 2px solid var(--pico-muted-border-color);
}
.multi-child {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}
</style>
