<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { starapp } from '../api/client'

const router = useRouter()

onMounted(async () => {
  try {
    const res = await starapp.listStarCharts()
    const charts = (res.starCharts || []).filter((c) => c.active !== false)
    const first = charts[0] || res.starCharts?.[0]
    if (first) {
      await router.replace({ name: 'familyStarChartView', params: { id: first.id } })
      return
    }
    await router.replace({ name: 'familyStarCharts' })
  } catch {
    await router.replace({ name: 'home' })
  }
})
</script>

<template>
  <p class="subtle">Loading star chart…</p>
</template>
