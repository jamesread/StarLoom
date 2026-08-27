import assert from 'node:assert/strict'
import test from 'node:test'

test('connectFetch path is defined', () => {
  assert.match('/starapp.api.v1.StarAppService/Init', /StarAppService/)
})
