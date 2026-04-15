# Testing Library (React) Reference

## Setup

```ts
import { render, screen, within, waitFor, waitForElementToBeRemoved } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import '@testing-library/jest-dom'
```

## render + screen

```tsx
const { container, rerender, unmount, baseElement } = render(<MyComponent prop="value" />)

// Re-render with new props
rerender(<MyComponent prop="updated" />)

// screen — global queries bound to document.body
screen.getByRole('button', { name: /submit/i })
screen.getByText('Hello World')
screen.debug()                  // prints DOM to console
screen.debug(screen.getByRole('form'))
```

## Query Priority (best → last resort)

```
1. getByRole          — semantic, accessibility-first
2. getByLabelText     — form fields
3. getByPlaceholderText
4. getByText          — non-interactive text
5. getByDisplayValue  — current value of input/select/textarea
6. getByAltText       — img alt text
7. getByTitle         — title attribute
8. getByTestId        — data-testid (last resort)
```

## Query Variants

```ts
// getBy* — throws if 0 or >1 match
screen.getByRole('button')

// queryBy* — returns null if not found (no throw), throws if >1
screen.queryByText('Loading...')

// findBy* — async, returns Promise, throws if not found after timeout
await screen.findByText('Loaded data')

// getAllBy* / queryAllBy* / findAllBy* — return arrays
screen.getAllByRole('listitem')
```

## getByRole Options

```ts
// Common ARIA roles: button, link, textbox, checkbox, radio,
// combobox, listbox, option, heading, img, list, listitem,
// dialog, alertdialog, alert, navigation, main, form, table

screen.getByRole('button', { name: /submit/i })       // accessible name (regex)
screen.getByRole('heading', { level: 2 })              // h2
screen.getByRole('checkbox', { checked: true })
screen.getByRole('textbox', { name: 'Email' })
screen.getByRole('combobox', { expanded: true })
```

## userEvent (v14 — always use setup())

```ts
const user = userEvent.setup()

// Click
await user.click(screen.getByRole('button'))
await user.dblClick(element)

// Typing
await user.type(screen.getByRole('textbox'), 'hello world')
await user.clear(screen.getByRole('textbox'))

// Combined type with clear
await user.clear(input)
await user.type(input, 'new value')

// Keyboard
await user.keyboard('{Enter}')
await user.keyboard('{Tab}')
await user.keyboard('{Escape}')
await user.keyboard('{ArrowDown}')
await user.keyboard('{Shift>}a{/Shift}')  // shift+a

// Tab navigation
await user.tab()
await user.tab({ shift: true })

// Select (native <select>)
await user.selectOptions(screen.getByRole('combobox'), ['option1', 'option2'])
await user.deselectOptions(select, 'option1')

// Hover / pointer
await user.hover(element)
await user.unhover(element)

// Upload file
await user.upload(fileInput, new File(['content'], 'test.png', { type: 'image/png' }))
```

## Async Utilities

```ts
// waitFor — polls until assertion passes or timeout
await waitFor(() => {
  expect(screen.getByText('Loaded')).toBeInTheDocument()
}, { timeout: 3000, interval: 50 })

// waitForElementToBeRemoved — waits until element disappears
await waitForElementToBeRemoved(() => screen.queryByText('Loading...'))
// or with element reference
const loader = screen.getByText('Loading...')
await waitForElementToBeRemoved(loader)

// findBy* is shorthand for waitFor + getBy
const el = await screen.findByText('Async content')
```

## within — Scoped Queries

```ts
const list = screen.getByRole('list')
const item = within(list).getByRole('listitem', { name: 'Item 1' })

// Useful for tables
const row = within(screen.getByRole('row', { name: /alice/i }))
expect(row.getByRole('cell', { name: '28' })).toBeInTheDocument()
```

## jest-dom Matchers

```ts
expect(el).toBeInTheDocument()
expect(el).toBeVisible()
expect(el).toBeDisabled()
expect(el).toBeEnabled()
expect(el).toBeChecked()
expect(el).toBeRequired()
expect(el).toHaveValue('text')
expect(el).toHaveDisplayValue('Label')
expect(el).toHaveTextContent(/hello/i)
expect(el).toHaveAttribute('href', '/home')
expect(el).toHaveClass('active')
expect(el).toHaveFocus()
expect(el).toContainElement(child)
expect(el).toHaveStyle({ display: 'none' })
expect(el).toBeEmptyDOMElement()
```

## Custom render with Providers

```tsx
// test-utils.tsx
import { render, RenderOptions } from '@testing-library/react'
import { ReactElement } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

function AllProviders({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>
}

function customRender(ui: ReactElement, options?: Omit<RenderOptions, 'wrapper'>) {
  return render(ui, { wrapper: AllProviders, ...options })
}

export * from '@testing-library/react'
export { customRender as render }
```

## Full Example

```tsx
import { render, screen, waitFor } from './test-utils'
import userEvent from '@testing-library/user-event'
import LoginForm from './LoginForm'

test('submits login form', async () => {
  const user = userEvent.setup()
  const onSubmit = vi.fn()

  render(<LoginForm onSubmit={onSubmit} />)

  await user.type(screen.getByLabelText(/email/i), 'test@example.com')
  await user.type(screen.getByLabelText(/password/i), 'secret')
  await user.click(screen.getByRole('button', { name: /sign in/i }))

  await waitFor(() => {
    expect(onSubmit).toHaveBeenCalledWith({
      email: 'test@example.com',
      password: 'secret',
    })
  })
})
```
