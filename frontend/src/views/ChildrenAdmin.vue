<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { PlusSignIcon, Refresh01Icon, UserMultipleIcon } from '@hugeicons/core-free-icons'
import { starapp, type FamilyMember } from '../api/client'
import MemberAvatar from '../components/MemberAvatar.vue'
import { memberStarColor } from '../lib/memberStarColor'

const iconStrokeWidth = 2.5

const members = ref<FamilyMember[]>([])
const error = ref('')
const loading = ref(true)

const tableHeaders = [
  { key: 'displayName', label: 'Name', sortable: true },
  { key: 'role', label: 'Role', sortable: true, width: '7rem' },
  { key: 'username', label: 'Username', sortable: true },
  { key: 'starColor', label: 'Star color', sortable: false, width: '7rem' },
  { key: 'actions', label: 'Actions', sortable: false, width: '8rem' },
]

function roleLabel(role?: string) {
  if (role === 'parent') return 'Parent'
  return 'Member'
}

const listRows = computed(() =>
  members.value.map((m) => ({
    id: m.id,
    displayName: m.displayName,
    role: roleLabel(m.role),
    username: m.username || '—',
    starColor: memberStarColor(m),
    member: m,
    actions: '',
  })),
)

async function load() {
  loading.value = true
  try {
    const res = await starapp.listMembers()
    members.value = res.members || []
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <Section
    subtitle="Family members who can earn stars and be assigned to chores."
    classes="people-list"
    :padding="false"
  >
    <template #title>
      <span class="section-title-with-icon">
        <HugeiconsIcon :icon="UserMultipleIcon" width="22" height="22" aria-hidden="true" />
        People
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
      <RouterLink
        :to="{ name: 'familyPersonCreate' }"
        class="button inline-icon good"
        aria-label="Add person"
        title="Add person"
      >
        <HugeiconsIcon
          :icon="PlusSignIcon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
      </RouterLink>
    </template>

    <p v-if="error" class="inline-notification error list-banner-pad">{{ error }}</p>
    <div v-if="loading && !members.length" class="list-banner-pad muted">Loading…</div>

    <template v-else>
      <p v-if="!members.length" class="inline-notification note list-banner-pad">
        No people yet.
        <RouterLink :to="{ name: 'familyPersonCreate' }">Add a person</RouterLink>.
      </p>

      <Table
        v-else
        class="list-table-wrap"
        :data="listRows"
        :headers="tableHeaders"
        :show-pagination="members.length > 10"
      >
        <template #cell-displayName="{ value, row }">
          <RouterLink :to="{ name: 'familyPersonDetail', params: { id: row.id } }" class="title-link name-cell">
            <MemberAvatar :member="row.member" size="sm" />
            <strong>{{ value }}</strong>
          </RouterLink>
        </template>
        <template #cell-starColor="{ value }">
          <span class="color-swatch" :style="{ backgroundColor: value }" :title="value" />
        </template>
        <template #cell-actions="{ row }">
          <div class="actions-cell">
            <RouterLink :to="{ name: 'familyPersonDetail', params: { id: row.id } }" class="button neutral small">
              Manage
            </RouterLink>
          </div>
        </template>
      </Table>
    </template>
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

.name-cell {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.color-swatch {
  display: inline-block;
  width: 1.25rem;
  height: 1.25rem;
  border-radius: 50%;
  vertical-align: middle;
  border: 1px solid color-mix(in srgb, var(--pico-color) 20%, transparent);
}
</style>
