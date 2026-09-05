<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowDown01Icon, ArrowLeft01Icon, ArrowUp01Icon, StarIcon } from '@hugeicons/core-free-icons'
import { starapp, type Chore, type StarChart } from '../api/client'

const route = useRoute()
const router = useRouter()
const chartId = computed(() => Number(route.params.id))

const chart = ref<StarChart | null>(null)
const chores = ref<Chore[]>([])
const dragIndex = ref(-1)
const error = ref('')
const saving = ref(false)
const loading = ref(true)

const form = reactive({
  name: '',
  sortOrder: 0,
  active: true,
})

const booleanOptions = [
  { label: 'Active', value: true },
  { label: 'Inactive', value: false },
]

const sectionTitle = computed(() => chart.value?.name || 'Edit star chart')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await starapp.listStarCharts({ includeInactive: true })
    chart.value = (res.starCharts || []).find((c) => c.id === chartId.value) || null
    if (!chart.value) {
      error.value = 'Star chart not found'
      return
    }
    form.name = chart.value.name
    form.sortOrder = chart.value.sortOrder ?? 0
    form.active = chart.value.active !== false
    const choreRes = await starapp.listChores({ starChartId: chartId.value })
    chores.value = choreRes.chores || []
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!chart.value) return
  saving.value = true
  error.value = ''
  try {
    await starapp.updateStarChart({
      id: chart.value.id,
      name: form.name.trim(),
      sortOrder: form.sortOrder,
      active: form.active,
    })
    router.push({ name: 'familyStarCharts' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

async function move(from: number, to: number) {
  if (to < 0 || to >= chores.value.length || from === to) return
  const next = chores.value.slice()
  const [moved] = next.splice(from, 1)
  next.splice(to, 0, moved)
  chores.value = next
  error.value = ''
  try {
    await starapp.reorderChores({ choreIds: next.map((c) => c.id) })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

function onDrop(index: number) {
  const from = dragIndex.value
  dragIndex.value = -1
  if (from >= 0) move(from, index)
}

async function removeChart() {
  if (!chart.value) return
  if ((chart.value.choreCount ?? 0) > 0) {
    error.value = 'Remove or reassign all chores before deleting this chart.'
    return
  }
  if (!confirm(`Delete "${chart.value.name}"?`)) return
  try {
    await starapp.deleteStarChart({ id: chart.value.id })
    router.push({ name: 'familyStarCharts' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(load)
</script>

<template>
  <Section :title="sectionTitle" :icon="StarIcon" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'familyStarCharts' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Star Charts</span>
      </RouterLink>
    </template>

    <p v-if="loading" class="subtle">Loading…</p>
    <p v-else-if="error && !chart" class="inline-notification error">{{ error }}</p>

    <FormLayout v-else @submit.prevent="save">
      <p v-if="error" class="inline-notification error">{{ error }}</p>
      <FormField label="Name" for="star-chart-edit-name">
        <input id="star-chart-edit-name" v-model="form.name" type="text" required />
      </FormField>
      <FormField label="Sort order" for="star-chart-edit-sort">
        <input id="star-chart-edit-sort" v-model.number="form.sortOrder" type="number" min="0" />
      </FormField>
      <FormField label="Status" component-has-label>
        <RadioGroup
          v-model="form.active"
          name="star-chart-edit-active"
          variant="boolean"
          :options="booleanOptions"
        />
      </FormField>
      <p v-if="chart" class="subtle">
        {{ chart.choreCount ?? 0 }} chore(s) assigned to this chart.
        <RouterLink :to="{ name: 'familyStarChartView', params: { id: chart.id } }">View chart</RouterLink>
      </p>
      <template #actions>
        <button type="submit" class="good" :disabled="saving">{{ saving ? 'Saving…' : 'Save' }}</button>
        <RouterLink :to="{ name: 'familyStarCharts' }" class="button neutral">Cancel</RouterLink>
        <button
          v-if="chart && (chart.choreCount ?? 0) === 0"
          type="button"
          class="bad"
          @click="removeChart"
        >
          Delete
        </button>
      </template>
    </FormLayout>

    <template v-if="!loading && chart">
      <h3>Chore order</h3>
      <p class="subtle">Drag a row, or use the arrows, to set the order chores appear on the chart.</p>
      <table v-if="chores.length" class="chore-order">
        <tbody>
          <tr
            v-for="(chore, index) in chores"
            :key="chore.id"
            draggable="true"
            @dragstart="dragIndex = index"
            @dragover.prevent
            @drop.prevent="onDrop(index)"
          >
            <td>{{ chore.title }}</td>
            <td class="order-actions">
              <button
                type="button"
                class="inline-icon neutral"
                :aria-label="`Move ${chore.title} up`"
                :disabled="index === 0"
                @click="move(index, index - 1)"
              >
                <HugeiconsIcon
                  :icon="ArrowUp01Icon"
                  width="1em"
                  height="1em"
                  :strokeWidth="2.5"
                  aria-hidden="true"
                />
              </button>
              <button
                type="button"
                class="inline-icon neutral"
                :aria-label="`Move ${chore.title} down`"
                :disabled="index === chores.length - 1"
                @click="move(index, index + 1)"
              >
                <HugeiconsIcon
                  :icon="ArrowDown01Icon"
                  width="1em"
                  height="1em"
                  :strokeWidth="2.5"
                  aria-hidden="true"
                />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-else class="subtle">No chores on this chart yet.</p>
    </template>
  </Section>
</template>

<style scoped>
.chore-order {
  width: 100%;
}
.chore-order tr {
  cursor: grab;
}
.order-actions {
  text-align: right;
  white-space: nowrap;
}
.order-actions button:disabled {
  opacity: 0.35;
}
</style>
