# React Patterns

Production patterns: compound components, composition, state architecture, forms, portals, custom hooks.

---

## Compound Components

```tsx
// Context 기반 Compound Component (권장)
interface TabsContextValue {
  activeTab: string;
  setActiveTab: (id: string) => void;
}

const TabsContext = React.createContext<TabsContextValue | null>(null);

function useTabs() {
  const ctx = React.useContext(TabsContext);
  if (!ctx) throw new Error('useTabs must be used within <Tabs>');
  return ctx;
}

function Tabs({ defaultTab, children }: { defaultTab: string; children: React.ReactNode }) {
  const [activeTab, setActiveTab] = useState(defaultTab);
  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab }}>
      <div className="tabs">{children}</div>
    </TabsContext.Provider>
  );
}

function TabList({ children }: { children: React.ReactNode }) {
  return <div role="tablist">{children}</div>;
}

function Tab({ id, children }: { id: string; children: React.ReactNode }) {
  const { activeTab, setActiveTab } = useTabs();
  return (
    <button
      role="tab"
      aria-selected={activeTab === id}
      onClick={() => setActiveTab(id)}
    >
      {children}
    </button>
  );
}

function TabPanel({ id, children }: { id: string; children: React.ReactNode }) {
  const { activeTab } = useTabs();
  if (activeTab !== id) return null;
  return <div role="tabpanel">{children}</div>;
}

Tabs.List = TabList;
Tabs.Tab = Tab;
Tabs.Panel = TabPanel;

// 사용
<Tabs defaultTab="profile">
  <Tabs.List>
    <Tabs.Tab id="profile">Profile</Tabs.Tab>
    <Tabs.Tab id="settings">Settings</Tabs.Tab>
  </Tabs.List>
  <Tabs.Panel id="profile"><ProfileContent /></Tabs.Panel>
  <Tabs.Panel id="settings"><SettingsContent /></Tabs.Panel>
</Tabs>
```

---

## Render Props → Custom Hook 변환

```tsx
// Render Props (레거시)
function MouseTracker({ children }: { children: (pos: { x: number; y: number }) => React.ReactNode }) {
  const [pos, setPos] = useState({ x: 0, y: 0 });
  return (
    <div onMouseMove={e => setPos({ x: e.clientX, y: e.clientY })}>
      {children(pos)}
    </div>
  );
}
// 사용: <MouseTracker>{({ x, y }) => <p>{x}, {y}</p>}</MouseTracker>

// Custom Hook (현대)
function useMouse() {
  const [pos, setPos] = useState({ x: 0, y: 0 });
  useEffect(() => {
    const handler = (e: MouseEvent) => setPos({ x: e.clientX, y: e.clientY });
    window.addEventListener('mousemove', handler);
    return () => window.removeEventListener('mousemove', handler);
  }, []);
  return pos;
}
// 사용: const { x, y } = useMouse();
```

---

## Container / Presentational

```tsx
// Presentational — UI만 담당, 순수 props
function UserCard({ name, avatar, onFollow }: {
  name: string;
  avatar: string;
  onFollow: () => void;
}) {
  return (
    <div className="card">
      <img src={avatar} alt={name} />
      <h3>{name}</h3>
      <button onClick={onFollow}>Follow</button>
    </div>
  );
}

// Container — 데이터 패칭, 비즈니스 로직
function UserCardContainer({ userId }: { userId: string }) {
  const { data: user } = useQuery({ queryKey: ['user', userId], queryFn: () => fetchUser(userId) });
  const followMutation = useMutation({ mutationFn: () => followUser(userId) });

  if (!user) return <Skeleton />;
  return (
    <UserCard
      name={user.name}
      avatar={user.avatar}
      onFollow={() => followMutation.mutate()}
    />
  );
}

// 현대적 관점: hooks로 로직 분리 후 단일 컴포넌트도 충분
// Container/Presentational 분리는 테스트 용이성이 목적일 때만 적용
```

---

## Provider 패턴

```tsx
interface AuthUser { id: string; name: string; role: 'admin' | 'user'; }
interface AuthContextValue {
  user: AuthUser | null;
  login: (credentials: { email: string; password: string }) => Promise<void>;
  logout: () => void;
}

const AuthContext = React.createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);

  const login = async (credentials: { email: string; password: string }) => {
    const data = await authApi.login(credentials);
    setUser(data.user);
  };

  const logout = () => {
    authApi.logout();
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within <AuthProvider>');
  return ctx;
}

// 성능: 자주 변경되는 값은 분리된 Context로 분리
// AuthProvider에서 user와 actions를 별도 Context로 쪼개면 불필요한 리렌더 방지
```

---

## 상태 관리 아키텍처

```tsx
// 서버 상태 vs 클라이언트 상태 구분

// 서버 상태 → TanStack Query (캐싱, 동기화, 재시도 자동)
function ProductList() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['products', { category: 'tech' }],
    queryFn: () => fetchProducts({ category: 'tech' }),
    staleTime: 5 * 60 * 1000, // 5분
  });
  // ...
}

// 전역 클라이언트 상태 → Zustand (간단, 보일러플레이트 없음)
import { create } from 'zustand';

interface CartStore {
  items: CartItem[];
  addItem: (item: CartItem) => void;
  removeItem: (id: string) => void;
  total: number;
}

const useCartStore = create<CartStore>((set, get) => ({
  items: [],
  addItem: (item) => set(state => ({ items: [...state.items, item] })),
  removeItem: (id) => set(state => ({ items: state.items.filter(i => i.id !== id) })),
  get total() { return get().items.reduce((sum, i) => sum + i.price, 0); },
}));

// 로컬 UI 상태 → useState / useReducer
// 판단 기준:
// - 다른 컴포넌트에서도 필요? → 전역
// - 서버 데이터인가? → TanStack Query
// - 복잡한 상태 전환? → useReducer
// - 단순 토글/값? → useState
```

---

## Controlled vs Uncontrolled

```tsx
// Controlled — React가 값 소유, 검증/포맷/의존 필드에 필요
function ControlledInput() {
  const [value, setValue] = useState('');
  return (
    <input
      value={value}
      onChange={e => setValue(e.target.value.toUpperCase())} // 포맷 적용
    />
  );
}

// Uncontrolled — DOM이 값 소유, 성능 중요한 단순 폼
function UncontrolledForm() {
  const nameRef = useRef<HTMLInputElement>(null);
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    console.log(nameRef.current?.value);
  };
  return (
    <form onSubmit={handleSubmit}>
      <input ref={nameRef} defaultValue="initial" />
      <button type="submit">Submit</button>
    </form>
  );
}
```

---

## forwardRef + useImperativeHandle

```tsx
interface DialogHandle {
  open: () => void;
  close: () => void;
}

const Dialog = forwardRef<DialogHandle, { title: string; children: React.ReactNode }>(
  ({ title, children }, ref) => {
    const [isOpen, setIsOpen] = useState(false);

    // 부모에게 노출할 API만 선택적으로 공개 (DOM 직접 노출 대신)
    useImperativeHandle(ref, () => ({
      open: () => setIsOpen(true),
      close: () => setIsOpen(false),
    }), []);

    if (!isOpen) return null;
    return (
      <div role="dialog" aria-label={title}>
        <h2>{title}</h2>
        {children}
        <button onClick={() => setIsOpen(false)}>Close</button>
      </div>
    );
  }
);

function Parent() {
  const dialogRef = useRef<DialogHandle>(null);
  return (
    <>
      <button onClick={() => dialogRef.current?.open()}>Open Dialog</button>
      <Dialog ref={dialogRef} title="Confirm">Are you sure?</Dialog>
    </>
  );
}
```

---

## children 패턴

```tsx
import React, { Children, cloneElement, isValidElement } from 'react';

// ReactNode vs ReactElement
interface Props {
  children: React.ReactNode;      // 모든 것 (string, number, null, element)
  label: React.ReactElement;      // React 엘리먼트만
}

// cloneElement — children에 추가 props 주입 (Compound Component 대안)
function RadioGroup({ value, onChange, children }: {
  value: string;
  onChange: (v: string) => void;
  children: React.ReactNode;
}) {
  return (
    <div role="radiogroup">
      {Children.map(children, child => {
        if (!isValidElement(child)) return child;
        return cloneElement(child as React.ReactElement<{ checked: boolean; onChange: () => void }>, {
          checked: child.props.value === value,
          onChange: () => onChange(child.props.value),
        });
      })}
    </div>
  );
}
```

---

## Portal

```tsx
import { createPortal } from 'react-dom';

function Modal({ isOpen, onClose, children }: {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;
}) {
  if (!isOpen) return null;

  // #modal-root에 렌더링 — z-index, overflow 문제 해결
  return createPortal(
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        {children}
      </div>
    </div>,
    document.getElementById('modal-root')!
  );
}

// Tooltip도 동일 패턴 — 부모의 overflow:hidden 영향 벗어나기
function Tooltip({ text, children }: { text: string; children: React.ReactElement }) {
  const [visible, setVisible] = useState(false);
  const [coords, setCoords] = useState({ top: 0, left: 0 });
  const ref = useRef<HTMLSpanElement>(null);

  const show = () => {
    const rect = ref.current?.getBoundingClientRect();
    if (rect) setCoords({ top: rect.bottom + 8, left: rect.left });
    setVisible(true);
  };

  return (
    <>
      <span ref={ref} onMouseEnter={show} onMouseLeave={() => setVisible(false)}>
        {children}
      </span>
      {visible && createPortal(
        <div className="tooltip" style={{ position: 'fixed', ...coords }}>{text}</div>,
        document.body
      )}
    </>
  );
}
```

---

## 폴더 구조

```
# Feature-based (권장 — 관련 파일이 한 곳에)
src/
  features/
    auth/
      components/LoginForm.tsx
      hooks/useAuth.ts
      api/authApi.ts
      types.ts
      index.ts          # public API만 export
    products/
      components/
      hooks/
      ...
  shared/
    components/Button.tsx
    hooks/useDebounce.ts
    utils/

# Layer-based (소규모 앱에 적합)
src/
  components/
  hooks/
  pages/
  services/
  types/
```

---

## Custom Hook 패턴

```tsx
// useToggle
function useToggle(initial = false) {
  const [value, setValue] = useState(initial);
  const toggle = useCallback(() => setValue(v => !v), []);
  const setTrue = useCallback(() => setValue(true), []);
  const setFalse = useCallback(() => setValue(false), []);
  return [value, { toggle, setTrue, setFalse }] as const;
}

// useDebounce
function useDebounce<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);
  return debounced;
}

// usePrevious
function usePrevious<T>(value: T): T | undefined {
  const ref = useRef<T>();
  useEffect(() => { ref.current = value; });
  return ref.current;
}

// useEventListener
function useEventListener<K extends keyof WindowEventMap>(
  type: K,
  handler: (event: WindowEventMap[K]) => void,
  element: EventTarget = window
) {
  const savedHandler = useRef(handler);
  useEffect(() => { savedHandler.current = handler; }, [handler]);
  useEffect(() => {
    const listener = (e: Event) => savedHandler.current(e as WindowEventMap[K]);
    element.addEventListener(type, listener);
    return () => element.removeEventListener(type, listener);
  }, [type, element]);
}
```

---

## 실전: Controlled Form + Validation

```tsx
interface FormValues { email: string; password: string; }
type FormErrors = Partial<Record<keyof FormValues, string>>;

function validate(values: FormValues): FormErrors {
  const errors: FormErrors = {};
  if (!values.email.includes('@')) errors.email = '유효한 이메일을 입력하세요';
  if (values.password.length < 8) errors.password = '비밀번호는 8자 이상이어야 합니다';
  return errors;
}

function LoginForm({ onSubmit }: { onSubmit: (values: FormValues) => Promise<void> }) {
  const [values, setValues] = useState<FormValues>({ email: '', password: '' });
  const [errors, setErrors] = useState<FormErrors>({});
  const [touched, setTouched] = useState<Partial<Record<keyof FormValues, boolean>>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setValues(prev => ({ ...prev, [name]: value }));
    if (touched[name as keyof FormValues]) {
      setErrors(validate({ ...values, [name]: value }));
    }
  };

  const handleBlur = (e: React.FocusEvent<HTMLInputElement>) => {
    const { name } = e.target;
    setTouched(prev => ({ ...prev, [name]: true }));
    setErrors(validate(values));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const errs = validate(values);
    if (Object.keys(errs).length > 0) {
      setErrors(errs);
      setTouched({ email: true, password: true });
      return;
    }
    setIsSubmitting(true);
    try {
      await onSubmit(values);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} noValidate>
      <div>
        <input
          type="email" name="email" value={values.email}
          onChange={handleChange} onBlur={handleBlur}
          aria-invalid={!!errors.email}
        />
        {touched.email && errors.email && <span role="alert">{errors.email}</span>}
      </div>
      <div>
        <input
          type="password" name="password" value={values.password}
          onChange={handleChange} onBlur={handleBlur}
          aria-invalid={!!errors.password}
        />
        {touched.password && errors.password && <span role="alert">{errors.password}</span>}
      </div>
      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? 'Logging in...' : 'Login'}
      </button>
    </form>
  );
}
```
