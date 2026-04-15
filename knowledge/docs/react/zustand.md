# Zustand State Management Reference

## create store (기본)
```tsx
import { create } from 'zustand'

interface BearStore {
  bears: number
  increase: () => void
  reset: () => void
}

const useBearStore = create<BearStore>()((set) => ({
  bears: 0,
  increase: () => set((state) => ({ bears: state.bears + 1 })),
  reset: () => set({ bears: 0 }),
}))

// 사용
function Component() {
  const bears = useBearStore((state) => state.bears)
  const increase = useBearStore((state) => state.increase)
  return <button onClick={increase}>Bears: {bears}</button>
}
```

## TypeScript double invocation 패턴
```tsx
// TypeScript에서 미들웨어와 함께 쓸 때 권장
const useStore = create<StoreState>()(
  (set, get) => ({
    count: 0,
    double: () => get().count * 2,
    increment: () => set((s) => ({ count: s.count + 1 })),
  })
)
```

## Selectors (렌더링 최적화)
```tsx
// 단일 값 — 변경 시에만 리렌더
const bears = useBearStore((state) => state.bears)

// 여러 값 — useShallow로 shallow 비교
import { useShallow } from 'zustand/react/shallow'

const { bears, increase } = useBearStore(
  useShallow((state) => ({ bears: state.bears, increase: state.increase }))
)
```

## Async actions (미들웨어 불필요)
```tsx
interface UserStore {
  user: User | null
  loading: boolean
  fetchUser: (id: number) => Promise<void>
}

const useUserStore = create<UserStore>()((set) => ({
  user: null,
  loading: false,
  fetchUser: async (id) => {
    set({ loading: true })
    try {
      const user = await api.getUser(id)
      set({ user, loading: false })
    } catch {
      set({ loading: false })
    }
  },
}))
```

## persist middleware
```tsx
import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

interface SettingsStore {
  theme: 'light' | 'dark'
  language: string
  setTheme: (theme: 'light' | 'dark') => void
}

const useSettingsStore = create<SettingsStore>()(
  persist(
    (set) => ({
      theme: 'light',
      language: 'ko',
      setTheme: (theme) => set({ theme }),
    }),
    {
      name: 'settings-storage',             // localStorage 키
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ theme: state.theme }), // 일부만 저장
      version: 1,
      migrate: (persistedState, version) => {
        if (version === 0) {
          // 마이그레이션 로직
          return { ...persistedState as SettingsStore, language: 'ko' }
        }
        return persistedState as SettingsStore
      },
    }
  )
)
```

## devtools middleware
```tsx
import { devtools } from 'zustand/middleware'

const useStore = create<StoreState>()(
  devtools(
    (set) => ({
      count: 0,
      increment: () => set((s) => ({ count: s.count + 1 }), false, 'increment'),
      //                                                     ^^^^  ^^^^^^^^^^^^^
      //                                                   replace  action name
    }),
    { name: 'MyStore' } // Redux DevTools에 표시될 이름
  )
)
```

## Middleware 조합 순서
```tsx
import { create } from 'zustand'
import { persist, devtools, immer } from 'zustand/middleware'

// 권장 순서: devtools > persist > immer
const useStore = create<StoreState>()(
  devtools(
    persist(
      (set) => ({
        items: [],
        addItem: (item) => set((s) => { s.items.push(item) }),
      }),
      { name: 'my-store' }
    ),
    { name: 'MyStore' }
  )
)
```

## Vanilla store (비-React 환경)
```tsx
import { createStore } from 'zustand/vanilla'

const store = createStore<BearStore>()((set) => ({
  bears: 0,
  increase: () => set((s) => ({ bears: s.bears + 1 })),
}))

// 외부에서 직접 접근
store.getState().increase()
store.setState({ bears: 10 })

const unsubscribe = store.subscribe((state) => {
  console.log('changed:', state.bears)
})
```

## 자주 쓰는 패턴

### 슬라이스 패턴 (대형 스토어 분리)
```tsx
interface BearSlice {
  bears: number
  addBear: () => void
}

interface FishSlice {
  fishes: number
  addFish: () => void
}

const createBearSlice = (set: SetState<BearSlice & FishSlice>): BearSlice => ({
  bears: 0,
  addBear: () => set((s) => ({ bears: s.bears + 1 })),
})

const createFishSlice = (set: SetState<BearSlice & FishSlice>): FishSlice => ({
  fishes: 0,
  addFish: () => set((s) => ({ fishes: s.fishes + 1 })),
})

const useStore = create<BearSlice & FishSlice>()((...a) => ({
  ...createBearSlice(...a),
  ...createFishSlice(...a),
}))
```

### 스토어 초기화 (로그아웃 등)
```tsx
const initialState = { user: null, token: null }

const useAuthStore = create<AuthStore>()((set) => ({
  ...initialState,
  login: (user, token) => set({ user, token }),
  logout: () => set(initialState), // 전체 초기화
}))
```
