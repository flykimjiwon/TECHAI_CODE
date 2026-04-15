# Vitest Reference

## Basic Structure

```ts
import { describe, it, test, expect, beforeEach, afterEach, beforeAll, afterAll } from 'vitest'

describe('MyModule', () => {
  beforeAll(() => { /* once before all tests */ })
  afterAll(() => { /* once after all tests */ })
  beforeEach(() => { /* before each test */ })
  afterEach(() => { /* after each test */ })

  it('does something', () => {
    expect(1 + 1).toBe(2)
  })

  test('alias for it', () => {
    expect('hello').toContain('ell')
  })
})
```

## Expect Matchers

```ts
// Equality
expect(val).toBe(2)              // strict ===
expect(obj).toEqual({ a: 1 })   // deep equal
expect(obj).toStrictEqual({})   // deep equal + same type (no undefined props)

// Strings / Arrays
expect('hello world').toContain('world')
expect([1, 2, 3]).toContain(2)
expect('foo@bar.com').toMatch(/^[\w]+@/)
expect('hello').toMatch('ell')

// Booleans / Nulls
expect(null).toBeNull()
expect(undefined).toBeUndefined()
expect(0).toBeFalsy()
expect(1).toBeTruthy()
expect(val).toBeDefined()

// Numbers
expect(0.1 + 0.2).toBeCloseTo(0.3, 5)
expect(10).toBeGreaterThan(5)
expect(10).toBeLessThanOrEqual(10)

// Errors
expect(() => JSON.parse('{')).toThrow()
expect(() => fn()).toThrow('expected message')
expect(() => fn()).toThrow(TypeError)

// Arrays / Objects
expect([1, 2, 3]).toHaveLength(3)
expect({ a: 1, b: 2 }).toHaveProperty('a', 1)
expect([1, 2, 3]).toEqual(expect.arrayContaining([1, 3]))
expect({ a: 1 }).toEqual(expect.objectContaining({ a: 1 }))

// Negation
expect(val).not.toBe(0)
```

## Mocking with vi

```ts
import { vi, expect } from 'vitest'

// Basic mock function
const fn = vi.fn()
fn('arg')
expect(fn).toHaveBeenCalled()
expect(fn).toHaveBeenCalledWith('arg')
expect(fn).toHaveBeenCalledTimes(1)

// Return values
const fn2 = vi.fn().mockReturnValue(42)
const fn3 = vi.fn().mockReturnValueOnce('first').mockReturnValue('rest')

// Async mocks
const asyncFn = vi.fn().mockResolvedValue({ data: 'ok' })
const rejectFn = vi.fn().mockRejectedValue(new Error('fail'))

// Custom implementation
const impl = vi.fn().mockImplementation((x: number) => x * 2)
const implOnce = vi.fn().mockImplementationOnce(() => 99)

// Spy on existing method
import * as utils from './utils'
const spy = vi.spyOn(utils, 'fetchData').mockReturnValue('mocked')
// Restore after test
spy.mockRestore()

// Mock entire module
vi.mock('./api', () => ({
  fetchUser: vi.fn().mockResolvedValue({ id: 1, name: 'Alice' }),
}))

// Partial mock
vi.mock('./config', async (importOriginal) => {
  const actual = await importOriginal<typeof import('./config')>()
  return { ...actual, DEBUG: true }
})

// Reset / restore
vi.clearAllMocks()    // clears calls/instances but keeps implementation
vi.resetAllMocks()    // also resets implementations
vi.restoreAllMocks()  // restores spies to originals
```

## Module Mocking Patterns

```ts
// Auto-mock (replaces all exports with vi.fn())
vi.mock('./heavy-module')

// Mock with factory (hoisted above imports)
vi.mock('@/lib/db', () => ({
  default: { query: vi.fn() },
}))

// Dynamic import in tests
const { fetchData } = await import('./api')
vi.mocked(fetchData).mockResolvedValue([])
```

## Snapshot Testing

```ts
it('matches snapshot', () => {
  const result = render(<Component />)
  expect(result).toMatchSnapshot()
  // Updates on first run, then asserts equality
})

// Inline snapshot
expect(obj).toMatchInlineSnapshot(`
  {
    "a": 1,
    "b": 2,
  }
`)

// Update snapshots: vitest --update-snapshots
```

## Setup & Teardown

```ts
// vitest.config.ts
import { defineConfig } from 'vitest/config'
export default defineConfig({
  test: {
    globals: true,           // no need to import describe/it/expect
    environment: 'jsdom',    // 'node' | 'jsdom' | 'happy-dom'
    setupFiles: ['./src/test/setup.ts'],
    globalSetup: ['./src/test/globalSetup.ts'],
  },
})

// setup.ts — runs before each test file
import '@testing-library/jest-dom'
import { afterEach } from 'vitest'
import { cleanup } from '@testing-library/react'
afterEach(() => cleanup())
```

## Test Control

```ts
test.skip('skipped test', () => { /* won't run */ })
test.only('only this runs', () => { /* isolates */ })
test.todo('not yet implemented')

// Concurrent tests (run in parallel within describe)
test.concurrent('parallel test 1', async () => { /* ... */ })
test.concurrent('parallel test 2', async () => { /* ... */ })

// Parameterized tests
test.each([
  [1, 1, 2],
  [2, 3, 5],
])('adds %i + %i = %i', (a, b, expected) => {
  expect(a + b).toBe(expected)
})

// Object form
test.each([
  { a: 1, b: 2, sum: 3 },
])('$a + $b = $sum', ({ a, b, sum }) => {
  expect(a + b).toBe(sum)
})
```

## Coverage Config

```ts
// vitest.config.ts
export default defineConfig({
  test: {
    coverage: {
      provider: 'v8',           // or 'istanbul'
      reporter: ['text', 'json', 'html'],
      include: ['src/**/*.ts'],
      exclude: ['src/**/*.d.ts', 'src/test/**'],
      thresholds: {
        lines: 80,
        functions: 80,
        branches: 70,
        statements: 80,
      },
    },
  },
})
// Run: vitest run --coverage
```

## Timer Mocks

```ts
vi.useFakeTimers()
vi.setSystemTime(new Date('2024-01-01'))

setTimeout(() => result = 'done', 1000)
vi.advanceTimersByTime(1000)
expect(result).toBe('done')

vi.runAllTimers()      // flush all pending
vi.runAllTimersAsync() // flush including promises
vi.useRealTimers()     // restore
```
