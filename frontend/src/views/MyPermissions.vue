<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import Table from 'picocrank/vue/components/Table.vue'
import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, Refresh01Icon, UserShield01Icon } from '@hugeicons/core-free-icons'
import { starapp } from '../api/client'

const iconStrokeWidth = 2.5

const audit = ref<Awaited<ReturnType<typeof starapp.getMyPermissionsAudit>> | null>(null)
const error = ref('')
const loading = ref(true)

const tableHeaders = [
  { key: 'permission', label: 'Permission', sortable: true },
  { key: 'granted', label: 'Granted', sortable: true, width: '7rem' },
  { key: 'groups', label: 'Granting groups', sortable: true },
]

const listRows = computed(() => {
  if (!audit.value) return []
  const isSuperuser = audit.value.isSuperuser === true
  return (audit.value.permissions || []).map((p) => ({
    permission: p.permission,
    granted: p.granted ? 'Yes' : 'No',
    grantedValue: p.granted === true,
    groups:
      (p.grantingGroups || []).join(', ') || (isSuperuser ? 'Superuser' : '—'),
  }))
})

const groupSummary = computed(() => (audit.value?.groupNames || []).join(', ') || '—')
const roleSummary = computed(() => (audit.value?.roleNames || []).join(', ') || '—')

async function load() {
  loading.value = true
  try {
    audit.value = await starapp.getMyPermissionsAudit()
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Load failed'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <Section
    subtitle="Effective permissions from your groups and roles."
    classes="my-permissions-list"
    :padding="false"
  >
    <template #title>
      <span class="section-title-with-icon">
        <HugeiconsIcon :icon="UserShield01Icon" width="22" height="22" aria-hidden="true" />
        My Permissions
      </span>
    </template>

    <template #toolbar>
      <RouterLink :to="{ name: 'userControlPanel' }" class="button inline-icon neutral">
        <HugeiconsIcon
          :icon="ArrowLeft01Icon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>Back</span>
      </RouterLink>
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
    </template>

    <p v-if="error" class="inline-notification error list-banner-pad">{{ error }}</p>
    <div v-if="loading && !audit" class="list-banner-pad muted">Loading…</div>

    <template v-else-if="audit">
      <div v-if="audit.isSuperuser" class="list-banner-pad">
        <NotificationBlock type="note" message="Superuser — all permissions granted." />
      </div>

      <dl class="membership-summary list-banner-pad">
        <dt>Groups</dt>
        <dd>{{ groupSummary }}</dd>
        <dt>Roles</dt>
        <dd>{{ roleSummary }}</dd>
      </dl>

      <p v-if="!listRows.length" class="inline-notification note list-banner-pad">
        No permissions to show.
      </p>

      <Table v-else class="list-table-wrap" :data="listRows" :headers="tableHeaders">
        <template #cell-permission="{ value }">
          <strong>{{ value }}</strong>
        </template>
        <template #cell-granted="{ value, row }">
          <span :class="row.grantedValue ? 'granted-yes' : 'granted-no'">{{ value }}</span>
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

.membership-summary {
  display: grid;
  grid-template-columns: 5.5rem 1fr;
  column-gap: 0.75em;
  margin: 0 0 1rem;
}

.membership-summary dt {
  font-weight: 600;
}

.membership-summary dd {
  margin: 0;
}

.list-table-wrap {
  margin-top: 0.5rem;
  margin-bottom: 1.5rem;
}

.granted-yes {
  font-weight: 600;
  color: var(--pico-ins-color, var(--pico-primary));
}

.granted-no {
  color: var(--pico-muted-color);
}
</style>
