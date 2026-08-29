import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { setTimeout as sleep } from 'node:timers/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { unlinkSync } from 'node:fs'
import { Builder, By, until } from 'selenium-webdriver'
import chrome from 'selenium-webdriver/chrome.js'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(__dirname, '../..')
const configDir = __dirname
const baseURL = 'http://127.0.0.1:18080'
const dbPath = path.join(configDir, 'starapp-test.db')

const session = { cookie: '' }

function storeCookies(response) {
  const setCookies = response.headers.getSetCookie?.() ?? []
  if (setCookies.length === 0) {
    const single = response.headers.get('set-cookie')
    if (single) {
      setCookies.push(single)
    }
  }
  if (setCookies.length > 0) {
    session.cookie = setCookies.map((entry) => entry.split(';')[0]).join('; ')
  }
}

async function apiFetch(url, options = {}) {
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  }
  if (session.cookie) {
    headers.Cookie = session.cookie
  }
  const res = await fetch(url, { ...options, headers })
  storeCookies(res)
  return res
}

async function waitForReady(timeoutMs = 20000) {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    try {
      const res = await fetch(`${baseURL}/metrics`)
      if (res.ok) {
        return
      }
    } catch {
      // retry
    }
    await sleep(250)
  }
  throw new Error('backend did not become ready')
}

async function loginSession() {
  const loginRes = await apiFetch(`${baseURL}/starapp.api.v1.StarAppService/LoginWithUsernameAndPassword`, {
    method: 'POST',
    body: JSON.stringify({ username: 'admin', password: 'admin' }),
  })
  assert.equal(loginRes.status, 200)
  const loginBody = await loginRes.json()
  assert.equal(loginBody.standardResponse?.success, true)
  assert.notEqual(session.cookie, '')
}

async function fetchInit() {
  const res = await apiFetch(`${baseURL}/starapp.api.v1.StarAppService/Init`, {
    method: 'POST',
    body: '{}',
  })
  assert.equal(res.status, 200)
  return res.json()
}

async function updateCvar(key, valueInt) {
  const res = await apiFetch(`${baseURL}/starapp.api.v1.StarAppService/UpdateCvar`, {
    method: 'POST',
    body: JSON.stringify({ key, valueInt }),
  })
  assert.equal(res.status, 200)
  return res.json()
}

async function resetSiteDisplayCvars() {
  await loginSession()
  await updateCvar('show_version_number', 1)
  await updateCvar('show_footer', 1)
}

function expectFalse(value, message) {
  assert.equal(value ?? false, false, message)
}

function expectEmpty(value, message) {
  assert.equal(value ?? '', '', message)
}

describe('starapp smoke', function () {
  this.timeout(120000)

  let backend
  let driver

  before(async function () {
    try {
      unlinkSync(dbPath)
    } catch {
      // ignore missing db
    }

    await new Promise((resolve, reject) => {
      const build = spawn('go', ['build', '-o', 'starapp', '.'], {
        cwd: path.join(root, 'service'),
        stdio: 'inherit',
      })
      build.on('exit', (code) => (code === 0 ? resolve() : reject(new Error('go build failed'))))
    })

    await new Promise((resolve, reject) => {
      const build = spawn('npm', ['run', 'build'], {
        cwd: path.join(root, 'frontend'),
        stdio: 'inherit',
      })
      build.on('exit', (code) => (code === 0 ? resolve() : reject(new Error('frontend build failed'))))
    })

    backend = spawn('./starapp', ['-configdir', configDir], {
      cwd: path.join(root, 'service'),
      stdio: 'inherit',
      env: {
        ...process.env,
        STARAPP_SECURE_COOKIES: 'false',
      },
    })

    await waitForReady()

    const options = new chrome.Options().addArguments(
      '--headless=new',
      '--disable-gpu',
      '--no-sandbox',
    )
    options.setPageLoadStrategy('none')
    driver = await new Builder().forBrowser('chrome').setChromeOptions(options).build()
    await driver.manage().setTimeouts({ implicit: 0, pageLoad: 30000, script: 30000 })
  })

  after(async function () {
    if (driver) {
      await driver.quit()
    }
    if (backend) {
      backend.kill('SIGTERM')
    }
    try {
      unlinkSync(dbPath)
    } catch {
      // ignore
    }
  })

  it('Init returns StarApp metadata', async function () {
    await loginSession()
    const body = await fetchInit()
    assert.equal(body.siteTitle, 'StarApp')
    assert.equal(body.showFooter, true)
    assert.equal(body.showVersionNumber, true)
    assert.equal(body.currentVersion, 'dev')
  })

  it('Init redacts version when show_version_number is off', async function () {
    await loginSession()
    try {
      await updateCvar('show_version_number', 0)
      const body = await fetchInit()
      expectFalse(body.showVersionNumber, 'showVersionNumber')
      expectEmpty(body.currentVersion, 'currentVersion')
      expectEmpty(body.availableVersion, 'availableVersion')
      expectFalse(body.showNewVersions, 'showNewVersions')
    } finally {
      await updateCvar('show_version_number', 1)
    }
  })

  it('Init hides footer when show_footer is off', async function () {
    await loginSession()
    try {
      await updateCvar('show_footer', 0)
      const body = await fetchInit()
      expectFalse(body.showFooter, 'showFooter')
    } finally {
      await updateCvar('show_footer', 1)
    }
  })

  it('renders the home page in the browser', async function () {
    await resetSiteDisplayCvars()
    await driver.manage().deleteAllCookies()
    await driver.get(baseURL)
    const userField = await driver.wait(
      until.elementLocated(By.css('#username, input[autocomplete="username"]')),
      15000,
    )
    await userField.sendKeys('admin')
    const passField = await driver.findElement(By.css('#password, input[type="password"]'))
    await passField.sendKeys('admin')
    const submit = await driver.findElement(By.css('button[type="submit"]'))
    await submit.click()
    await driver.wait(until.elementLocated(By.css('#layout')), 15000)
    await driver.navigate().refresh()
    await driver.wait(until.elementLocated(By.css('footer')), 15000)
    const headerTitle = await driver.findElement(By.css('header h1')).getText()
    assert.match(headerTitle, /StarApp/)
    const footerText = await driver.findElement(By.css('footer')).getText()
    assert.match(footerText, /StarApp/)
    assert.match(footerText, /dev/)
    assert.match(footerText, /Docs/)
  })
})
