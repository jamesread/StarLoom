<script setup lang="ts">
import { computed } from 'vue'
import type { FamilyMember, StarLedgerEntry } from '../api/client'
import { memberStarColor } from '../lib/memberStarColor'

const props = defineProps<{
  entries: StarLedgerEntry[]
  member?: FamilyMember | null
}>()

const starColor = computed(() => memberStarColor(props.member))

const awardRows = computed(() =>
  props.entries.map((entry) => ({
    entry,
    headline: awardHeadline(entry),
    countLabel: starsLabel(entry.amount),
    whenLabel: friendlyWhen(entry.createdAt),
  })),
)

function awardHeadline(entry: StarLedgerEntry) {
  const note = entry.note?.trim()
  if (note) return note
  return 'You earned stars!'
}

function starsLabel(amount?: number) {
  const n = amount ?? 0
  if (n === 1) return '1 star'
  return `${n} stars`
}

function friendlyWhen(createdAt?: string) {
  if (!createdAt) return ''
  const parsed = new Date(createdAt)
  if (Number.isNaN(parsed.getTime())) return createdAt

  const now = new Date()
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const startOfThatDay = new Date(parsed.getFullYear(), parsed.getMonth(), parsed.getDate())
  const dayDiff = Math.round((startOfToday.getTime() - startOfThatDay.getTime()) / 86_400_000)

  if (dayDiff === 0) return 'Today'
  if (dayDiff === 1) return 'Yesterday'
  if (dayDiff > 1 && dayDiff < 7) {
    return parsed.toLocaleDateString(undefined, { weekday: 'long' })
  }
  return parsed.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}
</script>

<template>
  <ul v-if="entries.length" class="recent-awards">
    <li v-for="row in awardRows" :key="row.entry.id" class="recent-award-card">
      <div class="recent-award-stars" :style="{ '--award-star-color': starColor }">
        <span class="recent-award-amount">+{{ row.entry.amount }}</span>
        <span class="recent-award-star" aria-hidden="true">★</span>
      </div>
      <div class="recent-award-body">
        <p class="recent-award-headline">{{ row.headline }}</p>
        <p class="recent-award-meta">
          <span class="recent-award-count">{{ row.countLabel }}</span>
          <span v-if="row.whenLabel" class="recent-award-sep" aria-hidden="true">·</span>
          <span v-if="row.whenLabel" class="recent-award-when">{{ row.whenLabel }}</span>
        </p>
      </div>
    </li>
  </ul>
  <p v-else class="recent-awards-empty">
    No stars yet. When you do something great, your family can award you stars here!
  </p>
</template>

<style scoped>
.recent-awards {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.recent-award-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.85rem 1rem;
  border: 1px solid var(--border-color, var(--pico-muted-border-color));
  border-radius: var(--pico-border-radius);
  background: var(--standout-bg-color, var(--section-bg-color));
}

.recent-award-stars {
  flex: 0 0 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 4.25rem;
  padding: 0.5rem 0.65rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--award-star-color) 18%, var(--standout-bg-color, #fff));
  border: 2px solid color-mix(in srgb, var(--award-star-color) 55%, transparent);
  color: var(--award-star-color);
  line-height: 1;
}

.recent-award-amount {
  font-size: 1.35rem;
  font-weight: 800;
}

.recent-award-star {
  font-size: 1.1rem;
  margin-top: 0.1rem;
}

.recent-award-body {
  min-width: 0;
}

.recent-award-headline {
  margin: 0 0 0.25rem;
  font-size: 1.05rem;
  font-weight: 700;
  line-height: 1.3;
}

.recent-award-meta {
  margin: 0;
  font-size: 0.95rem;
  color: var(--muted-text-color, var(--pico-muted-color));
}

.recent-award-count {
  font-weight: 600;
}

.recent-award-sep {
  margin: 0 0.35rem;
}

.recent-awards-empty {
  margin: 0;
  padding: 1rem;
  border-radius: var(--pico-border-radius);
  background: var(--standout-bg-color, var(--section-bg-color));
  color: var(--muted-text-color, var(--pico-muted-color));
  text-align: center;
  line-height: 1.5;
}
</style>
