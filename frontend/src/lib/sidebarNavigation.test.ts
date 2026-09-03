import assert from 'node:assert/strict'
import { test } from 'node:test'
import { setupSidebarNavigation, appendTopBarStarChartLinks, type SidebarNavigation } from './sidebarNavigation.ts'

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
  const names = nav.links
    .map((l) => (l && typeof l === 'object' && 'name' in l ? l.name : null))
  assert.ok(names.includes('familyStarCharts'))
  assert.ok(names.includes('familyRewards'))
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

test('top bar can exclude family admin links and control panel', () => {
  const nav = fakeNav()
  setupSidebarNavigation(nav, {
    showControlPanel: true,
    showFamilyNav: true,
    excludeRoutes: ['controlPanel', 'familyStarCharts', 'familyRewards'],
    flat: true,
  })
  assert.deepEqual(
    nav.links.map((l) => l && typeof l === 'object' && 'name' in l ? l.name : null),
    ['home'],
  )
})

test('top bar adds member star chart links', () => {
  const nav = fakeNav() as SidebarNavigation & {
    addCallback: (title: string, callback: () => void, options?: Record<string, unknown>) => void
  }
  nav.addCallback = (title, callback, options) => {
    nav.links.push({ type: 'callback', title, name: options?.name, callback })
  }
  setupSidebarNavigation(nav, { flat: true })
  appendTopBarStarChartLinks(nav, [{ id: 3, name: 'Bedroom chart', choreCount: 2 }], () => {}, null)
  assert.deepEqual(
    nav.links.map((l) => l && typeof l === 'object' && 'name' in l ? l.name : null),
    ['home', 'topbar-star-chart-3'],
  )
})
