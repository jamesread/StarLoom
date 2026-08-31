import assert from 'node:assert/strict'
import { test } from 'node:test'
import { setupSidebarNavigation, type SidebarNavigation } from './sidebarNavigation.ts'

function fakeNav(): SidebarNavigation & { links: unknown[] } {
  const links: unknown[] = []
  return {
    links,
    clearNavigationLinks() {
      links.length = 0
    },
    addSection(title, options) {
      links.push({ type: 'section', title, name: options?.name })
    },
    addRouterLink(link, altTitle, options) {
      links.push({ type: 'route', name: link, altTitle, options })
    },
  }
}

test('sidebar hides control panel without privilege', () => {
  const nav = fakeNav()
  setupSidebarNavigation(nav, { showControlPanel: false })
  assert.deepEqual(nav.links, [{ type: 'route', name: 'home', altTitle: undefined, options: undefined }])
})

test('sidebar adds family section for parents', () => {
  const nav = fakeNav()
  setupSidebarNavigation(nav, { showFamilyNav: true })
  assert.ok(nav.links.some((l) => l && typeof l === 'object' && 'name' in l && l.name === 'familyRewards'))
})

test('sidebar adds control panel as its own root section', () => {
  const nav = fakeNav()
  setupSidebarNavigation(nav, { showControlPanel: true })
  assert.deepEqual(nav.links, [
    { type: 'route', name: 'home', altTitle: undefined, options: undefined },
    { type: 'section', title: 'Control Panel', name: 'nav-control-panel' },
    {
      type: 'route',
      name: 'controlPanel',
      altTitle: null,
      options: { description: 'System administration' },
    },
  ])
})

test('top bar can exclude chores and control panel', () => {
  const nav = fakeNav()
  setupSidebarNavigation(nav, {
    showControlPanel: true,
    showFamilyNav: true,
    excludeRoutes: ['controlPanel'],
    flat: true,
  })
  assert.deepEqual(
    nav.links.map((l) => l && typeof l === 'object' && 'name' in l ? l.name : null),
    [
      'home',
      'familyStarCharts',
      'familyRewards',
    ],
  )
})
