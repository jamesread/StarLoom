<script setup lang="ts">
import type { StarChartDayProgress } from '../api/client'

defineProps<{
  progress: StarChartDayProgress[]
}>()

function progressLabel(entry: StarChartDayProgress) {
  const completed = entry.completed ?? 0
  const scheduled = entry.scheduled ?? 0
  if (entry.paused) {
    return `${completed}/${scheduled} · Paused`
  }
  return `${completed}/${scheduled} today`
}
</script>

<template>
  <div v-if="progress.length" class="today-progress" aria-label="Today's chore progress">
    <p class="today-progress-heading">Today's chores</p>
    <div v-for="entry in progress" :key="entry.starChartId" class="today-progress-row">
      <div class="today-progress-header">
        <span class="today-progress-name">{{ entry.starChartName }}</span>
        <span class="today-progress-count">{{ progressLabel(entry) }}</span>
      </div>
      <progress
        :value="entry.completed ?? 0"
        :max="Math.max(entry.scheduled ?? 0, 1)"
        :title="progressLabel(entry)"
      />
    </div>
  </div>
</template>

<style scoped>
.today-progress {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  width: 100%;
  margin-top: 0.35rem;
  padding-top: 0.35rem;
  border-top: 1px solid var(--pico-muted-border-color);
}
.today-progress-heading {
  margin: 0;
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  text-transform: uppercase;
  color: var(--muted-text-color, var(--pico-muted-color));
  text-align: center;
}
.today-progress-row {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.today-progress-header {
  display: flex;
  justify-content: space-between;
  gap: 0.35rem;
  font-size: 0.75rem;
  line-height: 1.2;
}
.today-progress-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.today-progress-count {
  flex-shrink: 0;
  color: var(--muted-text-color, var(--pico-muted-color));
}
progress {
  display: block;
  width: 100%;
  height: 0.65rem;
}
</style>
