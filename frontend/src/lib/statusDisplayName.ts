import type { GetStatusResponse } from '../api/client'

export function headerDisplayName(status?: GetStatusResponse | null) {
  if (!status?.isLoggedIn) {
    return ''
  }
  const displayName = status.displayName?.trim()
  return displayName || status.username || ''
}
