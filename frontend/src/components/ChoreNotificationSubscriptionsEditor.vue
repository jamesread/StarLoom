<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import FormLayout from 'picocrank/vue/components/FormLayout.vue'
import {
  starapp,
  type Chore,
  type ChoreNotificationSubscription,
  type FamilyMember,
} from '../api/client'

const props = defineProps<{
  subscriberMemberId: number
  subscriberDisplayName?: string
}>()

const loading = ref(true)
const saving = ref(false)
const error = ref('')
const success = ref('')
const members = ref<FamilyMember[]>([])
const chores = ref<Chore[]>([])
const selected = ref<Set<string>>(new Set())
const saved = ref<Set<string>>(new Set())

function personAllKey(memberId: number) {
  return `${memberId}:0`
}

function personChoreKey(memberId: number, choreId: number) {
  return `${memberId}:${choreId}`
}

function choreAnyKey(choreId: number) {
  return `0:${choreId}`
}

function subToKey(sub: ChoreNotificationSubscription) {
  const child = sub.childMemberId ?? 0
  const chore = sub.choreId ?? 0
  return `${child}:${chore}`
}

function keyToSub(key: string): ChoreNotificationSubscription {
  const [childRaw, choreRaw] = key.split(':')
  const child = Number(childRaw)
  const chore = Number(choreRaw)
  const out: ChoreNotificationSubscription = {}
  if (child > 0) out.childMemberId = child
  if (chore > 0) out.choreId = chore
  return out
}

function isChecked(key: string) {
  return selected.value.has(key)
}

function setChecked(key: string, on: boolean) {
  const next = new Set(selected.value)
  if (on) next.add(key)
  else next.delete(key)
  selected.value = next
}

function onToggle(key: string, event: Event) {
  setChecked(key, (event.target as HTMLInputElement).checked)
}

const activeChores = computed(() => chores.value.filter((chore) => chore.active !== false))

const membersWithChores = computed(() =>
  members.value.map((member) => ({
    member,
    assignedChores: activeChores.value.filter((chore) =>
      (chore.childMemberIds || []).includes(member.id),
    ),
  })),
)

const dirty = computed(() => {
  if (selected.value.size !== saved.value.size) return true
  for (const key of selected.value) {
    if (!saved.value.has(key)) return true
  }
  return false
})

const introText = computed(() => {
  const name = props.subscriberDisplayName?.trim()
  if (name) {
    return `Choose when Apprise should notify ${name} about chore completions in your family.`
  }
  return 'Choose when Apprise should notify this person about chore completions in your family.'
})

async function load() {
  if (!props.subscriberMemberId) return
  loading.value = true
  error.value = ''
  try {
    const [subsRes, choresRes] = await Promise.all([
      starapp.getMemberChoreNotificationSubscriptions({ memberId: props.subscriberMemberId }),
      starapp.listChores(),
    ])
    chores.value = choresRes.chores || []
    const membersRes = await starapp.listMembers()
    members.value = (membersRes.members || []).sort((a, b) =>
      (a.displayName || '').localeCompare(b.displayName || ''),
    )
    const keys = new Set((subsRes.subscriptions || []).map(subToKey))
    selected.value = new Set(keys)
    saved.value = new Set(keys)
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    const subscriptions = [...selected.value].map(keyToSub)
    const res = await starapp.saveMemberChoreNotificationSubscriptions({
      memberId: props.subscriberMemberId,
      subscriptions,
    })
    if (!res.standardResponse?.success) {
      error.value = res.standardResponse?.message || 'Save failed'
      return
    }
    saved.value = new Set(selected.value)
    success.value = 'Notification subscriptions saved'
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void load()
})

watch(
  () => props.subscriberMemberId,
  () => {
    void load()
  },
)
</script>

<template>
  <div class="subscriptions-editor">
    <p class="subtle">{{ introText }}</p>

    <p v-if="loading" class="subtle">Loading…</p>
    <p v-if="error" class="inline-notification error">{{ error }}</p>
    <p v-if="success" class="inline-notification note">{{ success }}</p>

    <FormLayout v-if="!loading" @submit.prevent="save">
      <div v-if="membersWithChores.length" class="notify-block">
        <h3>By person</h3>
        <p class="subtle">Notify when a specific person completes chores.</p>
        <div v-for="entry in membersWithChores" :key="entry.member.id" class="person-block">
          <h4>{{ entry.member.displayName }}</h4>
          <label class="notify-option">
            <input
              type="checkbox"
              :checked="isChecked(personAllKey(entry.member.id))"
              @change="onToggle(personAllKey(entry.member.id), $event)"
            />
            <span>All chores for {{ entry.member.displayName }}</span>
          </label>
          <label
            v-for="chore in entry.assignedChores"
            :key="`${entry.member.id}-${chore.id}`"
            class="notify-option indent"
          >
            <input
              type="checkbox"
              :checked="isChecked(personChoreKey(entry.member.id, chore.id))"
              @change="onToggle(personChoreKey(entry.member.id, chore.id), $event)"
            />
            <span>{{ chore.title }}</span>
          </label>
        </div>
      </div>

      <div v-if="activeChores.length" class="notify-block">
        <h3>By chore</h3>
        <p class="subtle">Notify when anyone completes a specific chore.</p>
        <label v-for="chore in activeChores" :key="`any-${chore.id}`" class="notify-option">
          <input
            type="checkbox"
            :checked="isChecked(choreAnyKey(chore.id))"
            @change="onToggle(choreAnyKey(chore.id), $event)"
          />
          <span>{{ chore.title }} — any person</span>
        </label>
      </div>

      <p v-if="!membersWithChores.length && !activeChores.length" class="subtle">
        No chores are configured yet. Add chores from Control Panel → Chores first.
      </p>

      <template #actions>
        <button type="submit" class="good" :disabled="saving || !dirty">
          {{ saving ? 'Saving…' : 'Save subscriptions' }}
        </button>
      </template>
    </FormLayout>
  </div>
</template>

<style scoped>
.subscriptions-editor :deep(form) {
  margin-top: 0.75rem;
}
.notify-block + .notify-block {
  margin-top: 1.5rem;
}
.notify-block h3 {
  margin: 0 0 0.25rem;
}
.person-block {
  margin-top: 1rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--pico-muted-border-color);
}
.person-block h4 {
  margin: 0 0 0.5rem;
}
.notify-option {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  margin: 0.35rem 0;
  cursor: pointer;
}
.notify-option.indent {
  margin-left: 1.25rem;
}
</style>
