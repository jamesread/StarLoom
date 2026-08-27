import { reactive, readonly } from 'vue'
import { starapp, type InitResponse } from '../api/client'

const state = reactive({
  loaded: false,
  loading: false,
  error: '' as string | null,
  init: null as InitResponse | null,
})

export function useInit() {
  return readonly(state)
}

export async function loadInit() {
  if (state.loading) {
    return
  }
  state.loading = true
  state.error = null
  try {
    state.init = await starapp.init()
    state.loaded = true
  } catch (err) {
    state.error = err instanceof Error ? err.message : 'Init failed'
  } finally {
    state.loading = false
  }
}
