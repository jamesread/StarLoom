import { reactive, readonly } from 'vue'
import { starapp, type GetStatusResponse } from '../api/client'

const state = reactive({
  loaded: false,
  loading: false,
  error: '' as string | null,
  status: null as GetStatusResponse | null,
})

let inflight: Promise<GetStatusResponse> | null = null

export function useStatus() {
  return readonly(state)
}

export function applyStatusToWindow(st: GetStatusResponse | null) {
  if (typeof window === 'undefined') return
  window.userRbacPermissions = st?.rbacPermissions || []
  window.userRbacIsSuperuser = Boolean(st?.rbacIsSuperuser)
}

export function invalidateAppStatus() {
  state.loaded = false
  state.status = null
  inflight = null
}

export async function fetchAppStatus(opts?: { force?: boolean }) {
  if (!opts?.force && state.loaded && state.status) {
    return state.status
  }
  if (!opts?.force && inflight) {
    return inflight
  }
  state.loading = true
  state.error = null
  inflight = starapp
    .getStatus()
    .then((st) => {
      state.status = st
      state.loaded = true
      applyStatusToWindow(st)
      return st
    })
    .catch((err: unknown) => {
      state.error = err instanceof Error ? err.message : 'Status failed'
      throw err
    })
    .finally(() => {
      state.loading = false
      inflight = null
    })
  return inflight
}

declare global {
  interface Window {
    userRbacPermissions?: string[]
    userRbacIsSuperuser?: boolean
  }
}
