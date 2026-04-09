# React 19 Hooks Reference

## useState
```tsx
const [count, setCount] = useState(0)
const [user, setUser] = useState<User | null>(null)

setCount(prev => prev + 1)
setUser({ name: 'John', age: 30 })
```

## useEffect
```tsx
useEffect(() => {
  // 실행할 side effect
  const subscription = subscribe()

  return () => {
    // cleanup
    subscription.unsubscribe()
  }
}, [dependency]) // 의존성 배열
```

## useMemo
```tsx
const expensiveValue = useMemo(() => {
  return computeExpensiveValue(a, b)
}, [a, b])
```

## useCallback
```tsx
const handleClick = useCallback(() => {
  doSomething(a, b)
}, [a, b])
```

## useRef
```tsx
const inputRef = useRef<HTMLInputElement>(null)
const countRef = useRef(0) // 리렌더링 없이 값 유지

useEffect(() => {
  inputRef.current?.focus()
}, [])

countRef.current += 1
```

## useTransition
```tsx
const [isPending, startTransition] = useTransition()

startTransition(() => {
  // 우선순위 낮은 상태 업데이트
  setSearchQuery(input)
})

{isPending && <Spinner />}
```

## use() (React 19 새 기능)
```tsx
// Promise를 직접 읽기
function UserProfile({ userPromise }: { userPromise: Promise<User> }) {
  const user = use(userPromise)
  return <div>{user.name}</div>
}

// Context 읽기
function Component() {
  const theme = use(ThemeContext)
  return <div>{theme}</div>
}
```

## useContext
```tsx
const ThemeContext = createContext<'light' | 'dark'>('light')

function Component() {
  const theme = useContext(ThemeContext)
  return <div className={theme}>Content</div>
}
```

## useReducer
```tsx
const [state, dispatch] = useReducer(reducer, initialState)

function reducer(state: State, action: Action) {
  switch (action.type) {
    case 'increment':
      return { count: state.count + 1 }
    default:
      return state
  }
}

dispatch({ type: 'increment' })
```

## useId
```tsx
const id = useId()
return (
  <>
    <label htmlFor={id}>Name</label>
    <input id={id} />
  </>
)
```

## 자주 쓰는 패턴

### Fetch with useEffect
```tsx
useEffect(() => {
  let cancelled = false

  async function fetchData() {
    const data = await fetch('/api/users')
    if (!cancelled) setUsers(data)
  }

  fetchData()
  return () => { cancelled = true }
}, [])
```

### Form handling
```tsx
const [form, setForm] = useState({ email: '', password: '' })

const handleChange = useCallback((e: ChangeEvent<HTMLInputElement>) => {
  setForm(prev => ({ ...prev, [e.target.name]: e.target.value }))
}, [])
```
