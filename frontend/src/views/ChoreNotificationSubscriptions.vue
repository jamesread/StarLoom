<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import Section from 'picocrank/vue/components/Section.vue'
import { Notification03Icon } from '@hugeicons/core-free-icons'
import ChoreNotificationSubscriptionsEditor from '../components/ChoreNotificationSubscriptionsEditor.vue'
import { starapp } from '../api/client'

const router = useRouter()

const loading = ref(true)
const error = ref('')
const subscriberMemberId = ref(0)
const subscriberDisplayName = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const fam = await starapp.getMyFamily()
    subscriberMemberId.value = fam.callerMember?.id || 0
    subscriberDisplayName.value = fam.callerMember?.displayName?.trim() || ''
    if (!subscriberMemberId.value) {
      error.value = 'You are not linked to a family person'
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
  <Section title="Chore notifications" :icon="Notification03Icon" :padding="true">
    <template #toolbar>
      <button type="button" class="inline-icon neutral" @click="router.push({ name: 'userControlPanel' })">
        Back
      </button>
    </template>

    <p v-if="loading" class="subtle">Loading…</p>
    <p v-if="error" class="inline-notification error">{{ error }}</p>

    <ChoreNotificationSubscriptionsEditor
      v-if="subscriberMemberId"
      :subscriber-member-id="subscriberMemberId"
      :subscriber-display-name="subscriberDisplayName"
    />
  </Section>
</template>
