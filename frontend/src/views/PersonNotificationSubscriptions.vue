<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import { HugeiconsIcon } from '@hugeicons/vue'
import { ArrowLeft01Icon, Notification03Icon } from '@hugeicons/core-free-icons'
import ChoreNotificationSubscriptionsEditor from '../components/ChoreNotificationSubscriptionsEditor.vue'
import { starapp, type FamilyMember } from '../api/client'

const iconStrokeWidth = 2.5

const route = useRoute()
const memberId = computed(() => Number(route.params.id))
const member = ref<FamilyMember | null>(null)
const loading = ref(true)
const error = ref('')

const sectionTitle = computed(() =>
  member.value ? `${member.value.displayName} — chore notifications` : 'Chore notifications',
)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await starapp.listMembers()
    member.value = (res.members || []).find((m) => m.id === memberId.value) || null
    if (!member.value) {
      error.value = 'Person not found'
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

<template>
  <Section :title="sectionTitle" :icon="Notification03Icon" :padding="true">
    <template #toolbar>
      <RouterLink
        :to="{ name: 'familyPersonDetail', params: { id: memberId } }"
        class="button inline-icon neutral"
      >
        <HugeiconsIcon
          :icon="ArrowLeft01Icon"
          width="1em"
          height="1em"
          :strokeWidth="iconStrokeWidth"
          aria-hidden="true"
        />
        <span>Back</span>
      </RouterLink>
    </template>

    <p v-if="loading" class="subtle">Loading…</p>
    <p v-if="error" class="inline-notification error">{{ error }}</p>

    <ChoreNotificationSubscriptionsEditor
      v-if="member"
      :subscriber-member-id="member.id"
      :subscriber-display-name="member.displayName"
    />
  </Section>
</template>
