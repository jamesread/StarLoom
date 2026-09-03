import assert from 'node:assert/strict'
import { test } from 'node:test'
import { headerDisplayName } from './statusDisplayName.ts'

test('header is empty when logged out', () => {
  assert.equal(headerDisplayName({ isLoggedIn: false, username: 'admin', displayName: 'Alex' }), '')
})

test('header prefers the person display name', () => {
  assert.equal(headerDisplayName({ isLoggedIn: true, username: 'admin', displayName: 'Alex' }), 'Alex')
})

test('header falls back to username when no display name', () => {
  assert.equal(headerDisplayName({ isLoggedIn: true, username: 'admin' }), 'admin')
})
