<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Section from 'picocrank/vue/components/Section.vue'
import { TaskDone01Icon } from '@hugeicons/core-free-icons'
import { starapp, type Redemption } from '../api/client'

const pending = ref<Redemption[]>([])
const error = ref('')

async function load() {
  try {
    const res = await starapp.listRedemptions({ status: 'pending' })
    pending.value = res.redemptions || []
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function approve(id: number) {
  try {
    await starapp.approveRedemption({ redemptionId: id })
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

async function reject(id: number) {
  try {
    await starapp.rejectRedemption({ redemptionId: id })
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(load)
</script>

<template>
  <Section title="Redemptions" :icon="TaskDone01Icon">
    <p class="subtle">Pending redemption requests awaiting parent approval.</p>
    <p v-if="error" class="error">{{ error }}</p>

    <table v-if="pending.length">
      <thead>
        <tr>
          <th>Person</th>
          <th>Reward</th>
          <th>Stars</th>
          <th>Requested</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in pending" :key="r.id">
          <td>{{ r.childDisplayName }}</td>
          <td>{{ r.rewardTitle }}</td>
          <td>{{ r.starsSpent }}</td>
          <td>{{ r.createdAt }}</td>
          <td>
            <button type="button" @click="approve(r.id)">Approve</button>
            <button type="button" class="secondary outline" @click="reject(r.id)">Reject</button>
          </td>
        </tr>
      </tbody>
    </table>
    <p v-else class="subtle">No pending redemptions.</p>
  </Section>
</template>

<style scoped>
.error {
  color: var(--pico-del-color);
}
</style>
