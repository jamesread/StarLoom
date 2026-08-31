<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { starapp } from '../api/client'
import { fetchAppStatus, useStatus } from '../composables/useStatus'
import { canViewFamilyHomeFromStatus } from '../lib/rbacAccess'

const router = useRouter()
const statusState = useStatus()

onMounted(async () => {
  try {
    await fetchAppStatus()
    const res = await starapp.listStarCharts()
    const charts = (res.starCharts || []).filter((c) => c.active !== false)
    const first = charts[0] || res.starCharts?.[0]
    if (first) {
      await router.replace({ name: 'familyStarChartView', params: { id: first.id } })
      return
    }
    if (canViewFamilyHomeFromStatus(statusState.status)) {
      await router.replace({ name: 'familyStarCharts' })
      return
    }
    await router.replace({ name: 'home' })
  } catch {
    await router.replace({ name: 'home' })
  }
})
</script>

<template>
  <p class="subtle">Loading star chart…</p>
</template>
