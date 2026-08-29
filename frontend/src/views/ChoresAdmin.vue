<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import CheckGroup from 'picocrank/vue/components/CheckGroup.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { PauseIcon, PlusSignIcon, Refresh01Icon, Task01Icon } from '@hugeicons/core-free-icons'
import { starapp, type Chore, type ChorePause, type FamilyMember, type StarChart } from '../api/client'
import MemberAvatar from '../components/MemberAvatar.vue'

const iconStrokeWidth = 2.5

const route = useRoute()
const router = useRouter()

const chores = ref<Chore[]>([])
const starCharts = ref<StarChart[]>([])
const pauses = ref<ChorePause[]>([])
const people = ref<FamilyMember[]>([])
const error = ref('')
const createError = ref('')
const pauseError = ref('')
const loading = ref(true)
const creating = ref(false)
const pausing = ref(false)

const createDialog = ref<HTMLDialogElement | null>(null)
const pauseDialog = ref<HTMLDialogElement | null>(null)
const createTitleInput = ref<HTMLInputElement | null>(null)

const weekdayOptions = [
  { label: 'Mon', value: 1 },
  { label: 'Tue', value: 2 },
  { label: 'Wed', value: 3 },
  { label: 'Thu', value: 4 },
  { label: 'Fri', value: 5 },
  { label: 'Sat', value: 6 },
  { label: 'Sun', value: 7 },
]

const form = reactive({
  title: '',
  starReward: 1,
  weekdays: [1, 2, 3, 4, 5, 6, 7] as number[],
  childMemberIds: [] as number[],
  starChartId: 0,
})

const pauseForm = reactive({
  startDate: '',
  endDate: '',
  reason: '',
})

const personOptions = computed(() =>
  people.value.map((m) => ({ label: m.displayName, value: m.id })),
)

const starChartOptions = computed(() =>
  starCharts.value
    .filter((c) => c.active !== false)
    .map((c) => ({ label: c.name, value: c.id })),
)

const starChartNames = computed(() => {
  const map = new Map<number, string>()
  for (const chart of starCharts.value) {
    map.set(chart.id, chart.name)
  }
  return map
})

const tableHeaders = [
  { key: 'title', label: 'Title', sortable: true },
  { key: 'starChart', label: 'Star chart', sortable: true },
  { key: 'starReward', label: 'Stars', sortable: true, width: '6rem' },
  { key: 'days', label: 'Days', sortable: false },
  { key: 'people', label: 'People', sortable: false },
  { key: 'status', label: 'Status', sortable: true, width: '6rem' },
  { key: 'actions', label: 'Actions', sortable: false, width: '8rem' },
]

function formatWeekdays(days?: number[]) {
  if (!days?.length) return 'Every day'
  const labels = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
  return days.map((d) => labels[d - 1]).join(', ')
}

function membersForChore(ids?: number[]) {
  if (!ids?.length) return []
  return ids
    .map((id) => people.value.find((m) => m.id === id))
    .filter((member): member is FamilyMember => !!member)
}

const listRows = computed(() =>
  chores.value.map((c) => ({
    id: c.id,
    title: c.title,
    starChart: starChartNames.value.get(c.starChartId || 0) || '—',
    starReward: c.starReward,
    days: formatWeekdays(c.weekdays),
    people: membersForChore(c.childMemberIds),
    status: c.active ? 'Active' : 'Inactive',
    active: c.active,
    actions: '',
  })),
)

function resetCreateForm() {
  form.title = ''
  form.starReward = 1
  form.weekdays = [1, 2, 3, 4, 5, 6, 7]
  form.childMemberIds = []
  form.starChartId = starChartOptions.value[0]?.value || 0
}

function resetPauseForm() {
  pauseForm.startDate = ''
  pauseForm.endDate = ''
  pauseForm.reason = ''
}

async function load() {
  loading.value = true
  try {
    const [choreRes, pauseRes, memberRes, chartRes] = await Promise.all([
      starapp.listChores({ includeInactive: true }),
      starapp.listChorePauses(),
      starapp.listMembers(),
      starapp.listStarCharts({ includeInactive: true }),
    ])
    chores.value = choreRes.chores || []
    pauses.value = pauseRes.pauses || []
    people.value = memberRes.members || []
    starCharts.value = chartRes.starCharts || []
    if (!form.starChartId && starChartOptions.value.length) {
      form.starChartId = starChartOptions.value[0].value
    }
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

function openCreateDialog(preferredChartId?: number) {
  createError.value = ''
  resetCreateForm()
  if (preferredChartId && starCharts.value.some((c) => c.id === preferredChartId)) {
    form.starChartId = preferredChartId
  }
  createDialog.value?.showModal()
  nextTick(() => createTitleInput.value?.focus())
}

function closeCreateDialog() {
  createDialog.value?.close()
}

function openPauseDialog() {
  pauseError.value = ''
  resetPauseForm()
  pauseDialog.value?.showModal()
}

function closePauseDialog() {
  pauseDialog.value?.close()
}

async function createChore() {
  creating.value = true
  createError.value = ''
  try {
    await starapp.createChore({
      title: form.title.trim(),
      starReward: form.starReward,
      weekdays: form.weekdays,
      childMemberIds: form.childMemberIds,
      starChartId: form.starChartId || undefined,
    })
    closeCreateDialog()
    await load()
  } catch (e) {
    createError.value = e instanceof Error ? e.message : String(e)
  } finally {
    creating.value = false
  }
}

async function deactivate(id: number) {
  if (!confirm('Deactivate this chore?')) return
  try {
    await starapp.deleteChore({ id })
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function createPause() {
  pausing.value = true
  pauseError.value = ''
  try {
    await starapp.createChorePause({
      startDate: pauseForm.startDate,
      endDate: pauseForm.endDate,
      reason: pauseForm.reason.trim(),
    })
    closePauseDialog()
    await load()
  } catch (e) {
    pauseError.value = e instanceof Error ? e.message : String(e)
  } finally {
    pausing.value = false
  }
}

async function removePause(id: number) {
  if (!confirm('Remove this pause period?')) return
  try {
    await starapp.deleteChorePause({ id })
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(async () => {
  await load()
  if (route.query.create === '1') {
    const chartId = Number(route.query.starChartId)
    openCreateDialog(Number.isFinite(chartId) && chartId > 0 ? chartId : undefined)
    await router.replace({ name: 'familyChores' })
  }
})
</script>

<template>
  <dialog ref="createDialog" class="dialog" @close="resetCreateForm">
    <h2>Add chore</h2>
    <p>Define a recurring task, its star reward, schedule, and which people can complete it.</p>
    <form class="dialog-form" @submit.prevent="createChore">
      <FormField label="Title" for="chore-title">
        <input
          id="chore-title"
          ref="createTitleInput"
          v-model="form.title"
          type="text"
          required
          placeholder="Do your homework"
          :disabled="creating"
        />
      </FormField>
      <FormField label="Star reward" for="chore-reward">
        <input
          id="chore-reward"
          v-model.number="form.starReward"
          type="number"
          min="1"
          required
          :disabled="creating"
        />
      </FormField>
      <FormField v-if="starChartOptions.length > 1" label="Star chart" fake>
        <RadioGroup v-model="form.starChartId" :options="starChartOptions" name="chore-star-chart" />
      </FormField>
      <FormField label="Days of week" fake>
        <CheckGroup v-model="form.weekdays" :options="weekdayOptions" name="chore-weekdays" />
      </FormField>
      <FormField label="Assigned people" fake>
        <CheckGroup v-model="form.childMemberIds" :options="personOptions" name="chore-people" />
      </FormField>
      <p v-if="createError" class="inline-notification error">{{ createError }}</p>
      <div class="dialog-actions">
        <button type="button" class="neutral" :disabled="creating" @click="closeCreateDialog">Cancel</button>
        <button type="submit" class="good" :disabled="creating || !form.title.trim()">
          {{ creating ? 'Creating…' : 'Create' }}
        </button>
      </div>
    </form>
  </dialog>

  <dialog ref="pauseDialog" class="dialog" @close="resetPauseForm">
    <h2>Pause chores</h2>
    <p>Chores will not appear on the Star Chart during this date range.</p>
    <form class="dialog-form" @submit.prevent="createPause">
      <FormField label="Start date" for="pause-start">
        <input id="pause-start" v-model="pauseForm.startDate" type="date" required :disabled="pausing" />
      </FormField>
      <FormField label="End date" for="pause-end">
        <input id="pause-end" v-model="pauseForm.endDate" type="date" required :disabled="pausing" />
      </FormField>
      <FormField label="Reason" for="pause-reason">
        <input
          id="pause-reason"
          v-model="pauseForm.reason"
          type="text"
          placeholder="School holiday"
          :disabled="pausing"
        />
      </FormField>
      <p v-if="pauseError" class="inline-notification error">{{ pauseError }}</p>
      <div class="dialog-actions">
        <button type="button" class="neutral" :disabled="pausing" @click="closePauseDialog">Cancel</button>
        <button type="submit" class="good" :disabled="pausing">
          {{ pausing ? 'Saving…' : 'Pause' }}
        </button>
      </div>
    </form>
  </dialog>

  <Section
    subtitle="Recurring tasks and star rewards. Mark completions on the Star Chart."
    classes="chores-list"
    :padding="false"
  >
    <template #title>
      <span class="section-title-with-icon">
        <HugeiconsIcon :icon="Task01Icon" width="22" height="22" aria-hidden="true" />
        Chores
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
      <button
        type="button"
        class="inline-icon neutral"
        aria-label="Pause chores"
        title="Pause chores"
        :disabled="loading"
        @click="openPauseDialog"
      >
        <HugeiconsIcon
          :icon="PauseIcon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
      </button>
      <button
        type="button"
        class="inline-icon good"
        aria-label="Add chore"
        title="Add chore"
        :disabled="loading"
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
    <div v-if="loading && !chores.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!chores.length" class="inline-notification note list-banner-pad">
        No chores yet. Use <strong>+</strong> to add one, then mark completions on the
        <RouterLink :to="{ name: 'familyStarChart' }">Star Chart</RouterLink>.
      </p>

      <Table
        v-else
        class="list-table-wrap"
        :data="listRows"
        :headers="tableHeaders"
        :show-pagination="chores.length > 10"
      >
        <template #cell-title="{ value, row }">
          <RouterLink :to="{ name: 'familyChoreEdit', params: { id: row.id } }" class="title-link">
            <strong>{{ value }}</strong>
          </RouterLink>
        </template>
        <template #cell-people="{ row }">
          <span v-if="!row.people.length" class="muted">—</span>
          <span v-else class="child-avatars">
            <MemberAvatar
              v-for="person in row.people"
              :key="person.id"
              :member="person"
              size="md"
              :to="{ name: 'familyPersonDetail', params: { id: person.id } }"
            />
          </span>
        </template>
        <template #cell-actions="{ row }">
          <div class="actions-cell">
            <RouterLink :to="{ name: 'familyChoreEdit', params: { id: row.id } }" class="button neutral small">
              Edit
            </RouterLink>
            <button v-if="row.active" type="button" class="bad small" @click="deactivate(row.id)">
              Deactivate
            </button>
          </div>
        </template>
      </Table>
    </template>
  </Section>

  <Section title="Pause periods" subtitle="Date ranges when chores are hidden on the Star Chart." :padding="true">
    <table v-if="pauses.length">
      <thead>
        <tr>
          <th>From</th>
          <th>To</th>
          <th>Reason</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="p in pauses" :key="p.id">
          <td>{{ p.startDate }}</td>
          <td>{{ p.endDate }}</td>
          <td>{{ p.reason || '—' }}</td>
          <td>
            <button type="button" class="bad small" @click="removePause(p.id)">Remove</button>
          </td>
        </tr>
      </tbody>
    </table>
    <p v-else class="subtle">No pause periods.</p>
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

.title-link {
  font: inherit;
  color: var(--pico-primary);
  text-decoration: none;
}

.title-link:hover {
  text-decoration: underline;
}

.child-avatars {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
}

.dialog-form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>
