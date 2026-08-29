<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import FormField from 'picocrank/vue/components/FormField.vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import CheckGroup from 'picocrank/vue/components/CheckGroup.vue'
import RadioGroup from 'picocrank/vue/components/RadioGroup.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, Task01Icon } from '@hugeicons/core-free-icons'
import { starapp, type Chore, type FamilyMember, type StarChart } from '../api/client'

const route = useRoute()
const router = useRouter()
const choreId = computed(() => Number(route.params.id))

const chore = ref<Chore | null>(null)
const people = ref<FamilyMember[]>([])
const starCharts = ref<StarChart[]>([])
const error = ref('')
const saving = ref(false)
const loading = ref(true)

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
  weekdays: [] as number[],
  childMemberIds: [] as number[],
  active: true,
  starChartId: 0,
})

const booleanOptions = [
  { label: 'Active', value: true },
  { label: 'Inactive', value: false },
]

const personOptions = computed(() =>
  people.value.map((m) => ({ label: m.displayName, value: m.id })),
)

const starChartOptions = computed(() =>
  starCharts.value
    .filter((c) => c.active !== false)
    .map((c) => ({ label: c.name, value: c.id })),
)

const sectionTitle = computed(() => chore.value?.title || 'Edit chore')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [choreRes, memberRes, chartRes] = await Promise.all([
      starapp.listChores({ includeInactive: true }),
      starapp.listMembers(),
      starapp.listStarCharts({ includeInactive: true }),
    ])
    people.value = memberRes.members || []
    starCharts.value = chartRes.starCharts || []
    chore.value = (choreRes.chores || []).find((c) => c.id === choreId.value) || null
    if (!chore.value) {
      error.value = 'Chore not found'
      return
    }
    form.title = chore.value.title
    form.starReward = chore.value.starReward
    form.weekdays = [...(chore.value.weekdays?.length ? chore.value.weekdays : [1, 2, 3, 4, 5, 6, 7])]
    form.childMemberIds = [...(chore.value.childMemberIds || [])]
    form.active = chore.value.active !== false
    form.starChartId = chore.value.starChartId || starChartOptions.value[0]?.value || 0
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!chore.value) return
  saving.value = true
  error.value = ''
  try {
    await starapp.updateChore({
      id: chore.value.id,
      title: form.title.trim(),
      starReward: form.starReward,
      weekdays: form.weekdays,
      childMemberIds: form.childMemberIds,
      active: form.active,
      starChartId: form.starChartId || undefined,
    })
    router.push({ name: 'familyChores' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <Section :title="sectionTitle" :icon="Task01Icon" :padding="true">
    <template #toolbar>
      <RouterLink :to="{ name: 'familyChores' }" class="button inline-icon neutral">
        <HugeiconsIcon :icon="ArrowLeft01Icon" width="1em" height="1em" aria-hidden="true" />
        <span>Chores</span>
      </RouterLink>
    </template>

    <p v-if="loading" class="subtle">Loading…</p>
    <p v-else-if="error && !chore" class="inline-notification error">{{ error }}</p>

    <FormLayout v-else @submit.prevent="save">
      <p v-if="error" class="inline-notification error">{{ error }}</p>
      <FormField label="Title" for="chore-edit-title">
        <input id="chore-edit-title" v-model="form.title" type="text" required />
      </FormField>
      <FormField label="Star reward" for="chore-edit-reward">
        <input id="chore-edit-reward" v-model.number="form.starReward" type="number" min="1" required />
      </FormField>
      <FormField v-if="starChartOptions.length > 1" label="Star chart" fake>
        <RadioGroup
          v-model="form.starChartId"
          variant="list"
          :options="starChartOptions"
          name="chore-edit-star-chart"
        />
      </FormField>
      <FormField label="Days of week" fake>
        <CheckGroup v-model="form.weekdays" :options="weekdayOptions" name="chore-edit-weekdays" />
      </FormField>
      <FormField label="Assigned people" fake>
        <CheckGroup v-model="form.childMemberIds" :options="personOptions" name="chore-edit-people" />
      </FormField>
      <FormField label="Status" fake>
        <RadioGroup
          v-model="form.active"
          name="chore-edit-active"
          variant="boolean"
          :options="booleanOptions"
        />
      </FormField>
      <template #actions>
        <button type="submit" class="good" :disabled="saving">{{ saving ? 'Saving…' : 'Save' }}</button>
        <RouterLink :to="{ name: 'familyChores' }" class="button neutral">Cancel</RouterLink>
      </template>
    </FormLayout>
  </Section>
</template>
