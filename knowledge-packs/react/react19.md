# React 19 New Features Reference

## Actions (async transitions in forms)
```tsx
// form action으로 async 함수 직접 전달
async function updateName(formData: FormData) {
  'use server'
  const name = formData.get('name') as string
  await db.user.update({ name })
}

export default function Form() {
  return (
    <form action={updateName}>
      <input name="name" />
      <button type="submit">저장</button>
    </form>
  )
}
```

## useActionState
```tsx
import { useActionState } from 'react'

async function submitAction(prevState: State, formData: FormData) {
  const name = formData.get('name') as string
  if (!name) return { error: '이름을 입력하세요' }
  await updateUser(name)
  return { error: null }
}

function Form() {
  const [state, submitAction, isPending] = useActionState(submitAction, { error: null })

  return (
    <form action={submitAction}>
      {state.error && <p>{state.error}</p>}
      <input name="name" />
      <button type="submit" disabled={isPending}>
        {isPending ? '저장 중...' : '저장'}
      </button>
    </form>
  )
}
```

## useFormStatus
```tsx
// 반드시 <form> 안에 있는 컴포넌트에서 사용
import { useFormStatus } from 'react-dom'

function SubmitButton() {
  const { pending, data, method, action } = useFormStatus()

  return (
    <button type="submit" disabled={pending}>
      {pending ? '제출 중...' : '제출'}
    </button>
  )
}

function Form() {
  return (
    <form action={serverAction}>
      <input name="email" />
      <SubmitButton /> {/* 반드시 form 안에 위치 */}
    </form>
  )
}
```

## useOptimistic
```tsx
import { useOptimistic } from 'react'

function MessageList({ messages }: { messages: Message[] }) {
  const [optimisticMessages, addOptimisticMessage] = useOptimistic(
    messages,
    (state, newMessage: string) => [
      ...state,
      { id: Date.now(), text: newMessage, sending: true },
    ]
  )

  async function handleSubmit(formData: FormData) {
    const text = formData.get('text') as string
    addOptimisticMessage(text) // 즉시 UI 반영
    await sendMessage(text)    // 서버 요청 완료 후 실제 데이터로 교체
  }

  return (
    <form action={handleSubmit}>
      {optimisticMessages.map(msg => (
        <p key={msg.id} style={{ opacity: msg.sending ? 0.5 : 1 }}>
          {msg.text}
        </p>
      ))}
      <input name="text" />
      <button type="submit">전송</button>
    </form>
  )
}
```

## use() hook
```tsx
import { use, Suspense } from 'react'

// Promise 읽기 (conditionally 호출 가능)
function UserProfile({ userPromise }: { userPromise: Promise<User> }) {
  const user = use(userPromise) // Suspense와 함께 사용
  return <div>{user.name}</div>
}

function Page() {
  const userPromise = fetchUser(1)
  return (
    <Suspense fallback={<p>로딩 중...</p>}>
      <UserProfile userPromise={userPromise} />
    </Suspense>
  )
}

// Context 읽기 (조건부 호출 가능)
function Component({ show }: { show: boolean }) {
  if (!show) return null
  const theme = use(ThemeContext) // if 블록 안에서도 호출 가능
  return <div>{theme}</div>
}
```

## Server Components
```tsx
// Server Component (기본값, async 가능)
// app/page.tsx
export default async function Page() {
  const data = await fetch('https://api.example.com/data')
  const json = await data.json()
  return <div>{json.title}</div>
}

// Client Component
'use client'

import { useState } from 'react'

export default function Counter() {
  const [count, setCount] = useState(0)
  return <button onClick={() => setCount(c => c + 1)}>{count}</button>
}

// Server Action
'use server'

export async function createPost(formData: FormData) {
  const title = formData.get('title') as string
  await db.post.create({ data: { title } })
}
```

## ref as prop (forwardRef 불필요)
```tsx
// React 19: ref를 일반 prop으로 전달
function Input({ ref, ...props }: React.ComponentProps<'input'>) {
  return <input ref={ref} {...props} />
}

// 사용
function Form() {
  const inputRef = useRef<HTMLInputElement>(null)
  return <Input ref={inputRef} name="email" />
}
```

## Context as Provider
```tsx
const ThemeContext = createContext<'light' | 'dark'>('light')

// React 19: <Context.Provider> 대신 <Context> 직접 사용
function App() {
  return (
    <ThemeContext value="dark">
      <Page />
    </ThemeContext>
  )
}
```

## Document Metadata
```tsx
// 어디서든 선언 가능 — React가 <head>로 호이스팅
function BlogPost({ post }: { post: Post }) {
  return (
    <article>
      <title>{post.title}</title>
      <meta name="description" content={post.excerpt} />
      <link rel="canonical" href={`https://example.com/posts/${post.slug}`} />
      <h1>{post.title}</h1>
      <p>{post.content}</p>
    </article>
  )
}
```

## 자주 쓰는 패턴

### Server Action + useActionState 조합
```tsx
'use server'
async function loginAction(prevState: { error: string | null }, formData: FormData) {
  const email = formData.get('email') as string
  const password = formData.get('password') as string
  const result = await signIn(email, password)
  if (!result.ok) return { error: result.message }
  redirect('/dashboard')
}

// Client
'use client'
function LoginForm() {
  const [state, action, isPending] = useActionState(loginAction, { error: null })
  return (
    <form action={action}>
      {state.error && <p className="text-red-500">{state.error}</p>}
      <input name="email" type="email" required />
      <input name="password" type="password" required />
      <button disabled={isPending}>{isPending ? '로그인 중...' : '로그인'}</button>
    </form>
  )
}
```

### Optimistic delete
```tsx
const [optimisticItems, removeOptimistic] = useOptimistic(
  items,
  (state, id: number) => state.filter(item => item.id !== id)
)

async function handleDelete(id: number) {
  removeOptimistic(id)
  await deleteItem(id)
}
```
