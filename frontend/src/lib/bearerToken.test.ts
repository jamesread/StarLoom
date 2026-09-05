import assert from 'node:assert/strict'
import test from 'node:test'
import { hrefWithToken, tokenFromHref } from './bearerToken.ts'

test('tokenFromHref reads the token query parameter', () => {
  assert.equal(tokenFromHref('https://stars.example/?token=sa_abc'), 'sa_abc')
  assert.equal(tokenFromHref('https://stars.example/'), '')
  assert.equal(tokenFromHref('not-a-url'), '')
})

test('hrefWithToken appends the token to relative urls', () => {
  assert.equal(hrefWithToken('/avatars/3', 'sa_abc'), '/avatars/3?token=sa_abc')
  assert.equal(hrefWithToken('/avatars/3', ''), '/avatars/3')
})
