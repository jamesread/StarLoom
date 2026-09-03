<script setup lang="ts">
import { computed } from 'vue'
import { memberStarStyle } from '../lib/memberStarColor'
import type { TodaysChore } from '../api/client'

const props = withDefaults(
  defineProps<{
    chores: TodaysChore[]
    busyKey?: string
    interactive?: boolean
  }>(),
  { interactive: true },
)

const emit = defineEmits<{
  toggle: [chore: TodaysChore]
}>()

const dayLabel = computed(() => {
  const date = props.chores.find((chore) => chore.date)?.date
  if (!date) return 'Today'
  return new Date(date + 'T12:00:00').toLocaleDateString(undefined, { weekday: 'short' })
})

function rowKey(chore: TodaysChore) {
  return `${chore.choreId}-${chore.childMemberId}`
}

function cellClass(chore: TodaysChore) {
  if (chore.paused) return 'paused'
  if (chore.completed) return 'completed'
  return 'scheduled'
}

function canToggle(chore: TodaysChore) {
  return props.interactive && !chore.paused && props.busyKey !== rowKey(chore)
}
</script>

<template>
  <div class="chart-wrap">
    <table class="star-chart">
      <thead>
        <tr>
          <th class="chore-col">Chore</th>
          <th class="day-col">{{ dayLabel }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="chore in chores" :key="rowKey(chore)">
          <td class="chore-col">
            <span class="chore-title">{{ chore.title }}</span>
            <span class="reward">★{{ chore.starReward }}</span>
          </td>
          <td class="day-col" :class="cellClass(chore)">
            <button
              v-if="canToggle(chore)"
              type="button"
              class="cell-btn"
              :aria-pressed="Boolean(chore.completed)"
              :aria-label="chore.completed ? `Uncomplete ${chore.title}` : `Complete ${chore.title}`"
              @click="emit('toggle', chore)"
            >
              <span class="cell-mark">
                <span v-if="chore.completed" class="star" :style="memberStarStyle(chore.child)">★{{ chore.starReward }}</span>
                <span v-else class="dot">○</span>
              </span>
            </button>
            <span v-else-if="chore.completed" class="cell-mark star" :style="memberStarStyle(chore.child)">★{{ chore.starReward }}</span>
            <span v-else-if="chore.paused" class="cell-mark muted">—</span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.chart-wrap {
  overflow-x: auto;
}
.star-chart {
  width: 100%;
  table-layout: fixed;
}
.star-chart th,
.star-chart td {
  text-align: center;
  vertical-align: middle;
}
.star-chart th.chore-col,
.star-chart td.chore-col {
  width: 15em;
  max-width: 15em;
  text-align: left !important;
  overflow-wrap: anywhere;
}
.star-chart th.day-col,
.star-chart td.day-col {
  text-align: left !important;
}
.chore-title {
  display: block;
  font-weight: 600;
}
.reward {
  font-size: 0.85rem;
  color: var(--pico-muted-color);
}
.star-chart th.day-col,
.star-chart td.day-col {
  text-align: left !important;
  vertical-align: middle;
}
.cell-btn,
.cell-mark {
  box-sizing: border-box;
}
.cell-btn {
  display: inline-flex;
  align-items: center;
  justify-content: flex-start;
  background: none;
  border: none;
  padding: 0 0.35rem;
  cursor: pointer;
  height: 2.75rem;
  min-width: 2.75rem;
}
.cell-mark {
  display: inline-flex;
  align-items: center;
  justify-content: flex-start;
  height: 2rem;
  min-width: 2rem;
  font-size: 2rem;
  line-height: 1;
  white-space: nowrap;
}
.star {
  font-weight: 800;
  line-height: 1;
}
.dot {
  color: var(--pico-muted-color);
  line-height: 1;
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
</style>
