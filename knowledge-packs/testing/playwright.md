# Playwright E2E Testing Reference

## Basic Structure

```ts
import { test, expect } from '@playwright/test'

test.describe('Feature', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
  })

  test.afterEach(async ({ page }) => {
    // cleanup if needed
  })

  test('user can log in', async ({ page }) => {
    await page.getByLabel('Email').fill('user@example.com')
    await page.getByLabel('Password').fill('secret')
    await page.getByRole('button', { name: 'Sign in' }).click()
    await expect(page).toHaveURL('/dashboard')
  })
})
```

## Locators

```ts
// Preferred — semantic / accessible
page.getByRole('button', { name: /submit/i })
page.getByRole('textbox', { name: 'Email' })
page.getByRole('heading', { level: 2 })
page.getByRole('checkbox', { checked: true })

page.getByLabel('Password')
page.getByPlaceholder('Search...')
page.getByText('Welcome back')
page.getByText(/partial match/i)
page.getByAltText('Profile photo')
page.getByTitle('Close dialog')
page.getByTestId('user-card')      // data-testid attribute

// CSS / XPath
page.locator('.my-class')
page.locator('#main-content')
page.locator('input[type="email"]')
page.locator('xpath=//button[@data-action="delete"]')

// Chaining locators
page.locator('.card').getByRole('button', { name: 'Edit' })
page.getByRole('list').locator('li').first()

// Filtering
page.getByRole('listitem').filter({ hasText: 'Alice' })
page.getByRole('listitem').filter({ has: page.getByRole('button') })
page.locator('.item').nth(2)       // 0-indexed
page.locator('.item').first()
page.locator('.item').last()
```

## Assertions (expect)

```ts
// Visibility
await expect(locator).toBeVisible()
await expect(locator).toBeHidden()
await expect(locator).toBeAttached()

// Text
await expect(locator).toHaveText('Exact text')
await expect(locator).toHaveText(/regex match/)
await expect(locator).toContainText('partial')
await expect(locator).toContainText(['item1', 'item2'])  // all items in list

// Attributes / State
await expect(locator).toHaveAttribute('href', '/home')
await expect(locator).toHaveClass('active')
await expect(locator).toHaveClass(/btn-/)
await expect(locator).toBeDisabled()
await expect(locator).toBeEnabled()
await expect(locator).toBeChecked()
await expect(locator).toBeFocused()

// Values
await expect(locator).toHaveValue('current value')
await expect(locator).toHaveValues(['opt1', 'opt2'])  // multi-select

// Count
await expect(locator).toHaveCount(3)

// Page-level
await expect(page).toHaveTitle('My App')
await expect(page).toHaveTitle(/My App/)
await expect(page).toHaveURL('/dashboard')
await expect(page).toHaveURL(/\/dashboard/)

// Soft assertions — don't abort test on failure
await expect.soft(locator).toBeVisible()
await expect.soft(locator).toHaveText('foo')
// All soft failures collected and reported at end

// Negation
await expect(locator).not.toBeVisible()
```

## Actions

```ts
// Navigation
await page.goto('https://example.com')
await page.goto('/relative-path')
await page.goBack()
await page.goForward()
await page.reload()

// Input
await page.getByLabel('Name').fill('Alice')
await page.getByLabel('Name').clear()
await page.getByRole('textbox').pressSequentially('slow typing', { delay: 50 })

// Click / pointer
await locator.click()
await locator.click({ button: 'right' })
await locator.dblclick()
await locator.hover()
await locator.focus()
await locator.blur()

// Keyboard
await page.keyboard.press('Enter')
await page.keyboard.press('Tab')
await page.keyboard.press('Control+a')
await locator.press('Escape')

// Select
await page.getByRole('combobox').selectOption('option-value')
await page.getByRole('combobox').selectOption({ label: 'Option Label' })

// File upload
await page.getByLabel('Upload').setInputFiles('path/to/file.png')
await page.getByLabel('Upload').setInputFiles(['file1.png', 'file2.png'])

// Wait (prefer auto-waiting locators over manual waits)
await page.waitForURL('/dashboard')
await page.waitForLoadState('networkidle')
await locator.waitFor({ state: 'visible' })
await page.waitForSelector('.spinner', { state: 'detached' })
await page.waitForResponse('**/api/users')

// Evaluate JS in browser
const title = await page.evaluate(() => document.title)
await page.evaluate((val) => localStorage.setItem('key', val), 'value')
```

## Custom Fixtures (test.extend)

```ts
// fixtures.ts
import { test as base, expect } from '@playwright/test'

type Fixtures = {
  loggedInPage: Page
}

type WorkerFixtures = {
  workerToken: string
}

export const test = base.extend<Fixtures, WorkerFixtures>({
  // Test-scoped fixture (runs per test)
  loggedInPage: async ({ page }, use) => {
    // Setup
    await page.goto('/login')
    await page.getByLabel('Email').fill('admin@example.com')
    await page.getByLabel('Password').fill('password')
    await page.getByRole('button', { name: 'Sign in' }).click()
    await page.waitForURL('/dashboard')

    // Hand off to test
    await use(page)

    // Teardown (after test)
    await page.goto('/logout')
  },

  // Worker-scoped fixture (shared across tests in same worker)
  workerToken: [async ({}, use) => {
    const token = await fetchAuthToken()
    await use(token)
  }, { scope: 'worker' }],
})

export { expect }
```

```ts
// mytest.spec.ts — using custom fixtures
import { test, expect } from './fixtures'

test('admin sees dashboard', async ({ loggedInPage }) => {
  await expect(loggedInPage.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
})
```

## playwright.config.ts

```ts
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'Mobile Safari', use: { ...devices['iPhone 14'] } },
  ],
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },
})
```

## Tips

```ts
// Page Object Model pattern
class LoginPage {
  constructor(private page: Page) {}
  async goto() { await this.page.goto('/login') }
  async login(email: string, password: string) {
    await this.page.getByLabel('Email').fill(email)
    await this.page.getByLabel('Password').fill(password)
    await this.page.getByRole('button', { name: 'Sign in' }).click()
  }
}

// Intercept network requests
await page.route('**/api/users', route =>
  route.fulfill({ json: [{ id: 1, name: 'Alice' }] })
)

// Mock API errors
await page.route('**/api/data', route =>
  route.fulfill({ status: 500, body: 'Server error' })
)

// test.skip / test.only
test.skip('feature behind flag', async ({ page }) => {})
test.only('focus on this', async ({ page }) => {})
```
