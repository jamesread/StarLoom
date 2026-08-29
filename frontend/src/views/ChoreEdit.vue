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
import { starapp, type Chore, type FamilyMember } from '../api/client'

const route = useRoute()
const router = useRouter()
const choreId = computed(() => Number(route.params.id))

const chore = ref<Chore | null>(null)
const children = ref<FamilyMember[]>([])
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
})

const booleanOptions = [
  { label: 'Active', value: true },
  { label: 'Inactive', value: false },
]

const childOptions = computed(() =>
  children.value.map((c) => ({ label: c.displayName, value: c.id })),
)

const sectionTitle = computed(() => chore.value?.title || 'Edit chore')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [choreRes, memberRes] = await Promise.all([
      starapp.listChores({ includeInactive: true }),
      starapp.listMembers(),
    ])
    children.value = (memberRes.members || []).filter((m) => m.role === 'child')
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
    })
    router.push({ name: 'familyChores' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

async function deactivate() {
  if (!chore.value || !confirm('Deactivate this chore?')) return
  try {
    await starapp.deleteChore({ id: chore.value.id })
    router.push({ name: 'familyChores' })
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
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
      <FormField label="Days of week" fake>
        <CheckGroup v-model="form.weekdays" :options="weekdayOptions" name="chore-edit-weekdays" />
      </FormField>
      <FormField label="Assigned children" fake>
        <CheckGroup v-model="form.childMemberIds" :options="childOptions" name="chore-edit-children" />
      </FormField>
      <FormField label="Status" fake>
        <RadioGroup v-model="form.active" name="chore-edit-active" :options="booleanOptions" />
      </FormField>
      <template #actions>
        <button type="submit" class="good" :disabled="saving">{{ saving ? 'Saving…' : 'Save' }}</button>
        <RouterLink :to="{ name: 'familyChores' }" class="button neutral">Cancel</RouterLink>
        <button v-if="form.active" type="button" class="bad" @click="deactivate">Deactivate</button>
      </template>
    </FormLayout>
  </Section>
</template>
