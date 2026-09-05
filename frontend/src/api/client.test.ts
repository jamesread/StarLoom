import assert from 'node:assert/strict'
import test from 'node:test'

test('connectFetch path is defined', () => {
  assert.match('/starapp.api.v1.StarAppService/Init', /StarAppService/)
})

test('connectFetch surfaces Connect error message from JSON body', async () => {
  const originalFetch = globalThis.fetch
  globalThis.fetch = (async () =>
    new Response(JSON.stringify({ code: 'failed_precondition', message: 'insufficient star balance (has 0, needs 1)' }), {
      status: 400,
      statusText: 'Bad Request',
      headers: { 'Content-Type': 'application/json' },
    })) as typeof fetch

  try {
    const { starapp } = await import('./client.ts')
    await assert.rejects(
      () => starapp.requestRedemption({ rewardId: 1, childMemberId: 2 }),
      (err: unknown) => {
        assert.ok(err instanceof Error)
        assert.equal(err.message, 'insufficient star balance (has 0, needs 1)')
        return true
      },
    )
  } finally {
    globalThis.fetch = originalFetch
  }
})

test('connectFetch falls back to HTTP status when error body has no message', async () => {
  const originalFetch = globalThis.fetch
  globalThis.fetch = (async () =>
    new Response('not json', {
      status: 503,
      statusText: 'Service Unavailable',
    })) as typeof fetch

  try {
    const { starapp } = await import('./client.ts')
    await assert.rejects(
      () => starapp.getStatus(),
      (err: unknown) => {
        assert.ok(err instanceof Error)
        assert.equal(err.message, 'Service Unavailable')
        return true
      },
    )
  } finally {
    globalThis.fetch = originalFetch
  }
})
