import { reactive, readonly } from 'vue'
import { starapp, type InitResponse } from '../api/client'

const state = reactive({
  loaded: false,
  loading: false,
  error: '' as string | null,
  init: null as InitResponse | null,
})

let inflight: Promise<InitResponse> | null = null

export function useInit() {
  return readonly(state)
}

export function invalidateInit() {
  state.loaded = false
  state.init = null
  inflight = null
}

export async function loadInit(opts?: { force?: boolean }) {
  if (!opts?.force && state.loaded && state.init) {
    return state.init
  }
  if (!opts?.force && inflight) {
    return inflight
  }
  state.loading = true
  state.error = null
  inflight = starapp
    .init()
    .then((init) => {
      state.init = init
      state.loaded = true
      return init
    })
    .catch((err: unknown) => {
      state.error = err instanceof Error ? err.message : 'Init failed'
      throw err
    })
    .finally(() => {
      state.loading = false
      inflight = null
    })
  return inflight
}
