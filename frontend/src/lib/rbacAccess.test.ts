import assert from 'node:assert/strict'
import { test } from 'node:test'
import {
  canAccessControlPanelFromStatus,
  canAccessIamFromStatus,
  canAccessSettingsFromStatus,
  canAccessWebhooksFromStatus,
} from './rbacAccess.ts'

test('control panel is hidden when logged out', () => {
  assert.equal(canAccessControlPanelFromStatus({ isLoggedIn: false, rbacIsSuperuser: true }), false)
})

test('control panel is hidden without privilege', () => {
  assert.equal(canAccessControlPanelFromStatus({
    isLoggedIn: true,
    rbacIsSuperuser: false,
    rbacPermissions: [],
  }), false)
})

test('superuser can open the control panel', () => {
  assert.equal(canAccessControlPanelFromStatus({
    isLoggedIn: true,
    rbacIsSuperuser: true,
    rbacPermissions: [],
  }), true)
})

test('IAM view permission unlocks the control panel', () => {
  assert.equal(canAccessControlPanelFromStatus({
    isLoggedIn: true,
    rbacIsSuperuser: false,
    rbacPermissions: ['users.view'],
  }), true)
})

test('settings permission unlocks the control panel', () => {
  assert.equal(canAccessControlPanelFromStatus({
    isLoggedIn: true,
    rbacIsSuperuser: false,
    rbacPermissions: ['system.settings'],
  }), true)
})

test('unimplemented child permissions do not unlock the control panel', () => {
  assert.equal(canAccessControlPanelFromStatus({
    isLoggedIn: true,
    rbacIsSuperuser: false,
    rbacPermissions: ['system.logs', 'system.impersonate'],
  }), false)
})

test('IAM tile follows IAM privileges', () => {
  assert.equal(canAccessIamFromStatus({
    isLoggedIn: true,
    rbacIsSuperuser: false,
    rbacPermissions: ['usergroups.view'],
  }), true)
  assert.equal(canAccessIamFromStatus({
    isLoggedIn: true,
    rbacIsSuperuser: false,
    rbacPermissions: ['system.settings'],
  }), false)
})

test('settings tile follows system.settings', () => {
  assert.equal(canAccessSettingsFromStatus({
    isLoggedIn: true,
    rbacIsSuperuser: false,
    rbacPermissions: ['system.settings'],
  }), true)
  assert.equal(canAccessSettingsFromStatus({
    isLoggedIn: true,
    rbacIsSuperuser: false,
    rbacPermissions: ['users.view'],
  }), false)
})

test('webhooks tile follows system.settings', () => {
  assert.equal(canAccessWebhooksFromStatus({
    isLoggedIn: true,
    rbacIsSuperuser: false,
    rbacPermissions: ['system.settings'],
  }), true)
  assert.equal(canAccessWebhooksFromStatus({
    isLoggedIn: true,
    rbacIsSuperuser: false,
    rbacPermissions: ['users.view'],
  }), false)
})
