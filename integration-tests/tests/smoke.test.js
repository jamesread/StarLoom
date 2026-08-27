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
    })

    await waitForReady()

    const options = new chrome.Options().addArguments(
      '--headless=new',
      '--disable-gpu',
      '--no-sandbox',
    )
    driver = await new Builder().forBrowser('chrome').setChromeOptions(options).build()
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
    const loginRes = await fetch(`${baseURL}/starapp.api.v1.StarAppService/LoginWithUsernameAndPassword`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ username: 'admin', password: 'admin' }),
    })
    assert.equal(loginRes.status, 200)
    const loginBody = await loginRes.json()
    assert.equal(loginBody.standardResponse?.success, true)

    const res = await fetch(`${baseURL}/starapp.api.v1.StarAppService/Init`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: '{}',
    })
    assert.equal(res.status, 200)
    const body = await res.json()
    assert.equal(body.siteTitle, 'StarApp')
    assert.equal(body.showFooter, true)
  })

  it('renders the home page in the browser', async function () {
    await driver.get(baseURL)
    const userField = await driver.wait(until.elementLocated(By.css('input[autocomplete="username"], #username, input[placeholder="Username"]')), 10000)
    await userField.sendKeys('admin')
    const passField = await driver.findElement(By.css('input[type="password"]'))
    await passField.sendKeys('admin')
    const submit = await driver.findElement(By.css('button[type="submit"]'))
    await submit.click()
    await driver.wait(until.elementLocated(By.css('main')), 10000)
    const text = await driver.findElement(By.css('main')).getText()
    assert.match(text, /StarApp/)
  })
})
