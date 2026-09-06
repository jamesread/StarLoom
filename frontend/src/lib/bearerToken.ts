const storageKey = 'starapp-bearer-token'

export function tokenFromHref(href: string): string {
  try {
    return new URL(href).searchParams.get('token')?.trim() || ''
  } catch {
    return ''
  }
}

export function hrefWithToken(href: string, token: string): string {
  if (!token) return href
  try {
    const url = new URL(href, 'http://localhost')
    url.searchParams.set('token', token)
    return href.startsWith('http') ? url.toString() : url.pathname + url.search
  } catch {
    return href
  }
}

let cached: string | null = null

function readStored(): string {
  try {
    return window.sessionStorage.getItem(storageKey) || ''
  } catch {
    return ''
  }
}

function writeStored(token: string) {
  try {
    window.sessionStorage.setItem(storageKey, token)
  } catch {
    // Storage unavailable.
  }
}

export function bearerToken(): string {
  if (cached !== null) return cached
  if (typeof window === 'undefined') return ''
  const fromHref = tokenFromHref(window.location.href)
  if (fromHref) {
    writeStored(fromHref)
  }
  cached = fromHref || readStored()
  return cached
}

export function clearBearerToken() {
  cached = ''
  try {
    window.sessionStorage.removeItem(storageKey)
  } catch {
    // Storage unavailable.
  }
}
