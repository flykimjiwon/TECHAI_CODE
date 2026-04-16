# React Class Components (Legacy Reference)

For maintaining React 15~17 codebases. Covers lifecycle, state, refs, context, HOC, and migration path.

---

## Basic Class Component

```tsx
import React, { Component } from 'react';

interface Props {
  title: string;
  onClose: () => void;
}

interface State {
  count: number;
  loading: boolean;
}

class MyComponent extends Component<Props, State> {
  state: State = {
    count: 0,
    loading: false,
  };

  render() {
    const { title, onClose } = this.props;
    const { count, loading } = this.state;

    return (
      <div>
        <h1>{title}</h1>
        <p>{count}</p>
        {loading && <span>Loading...</span>}
        <button onClick={onClose}>Close</button>
      </div>
    );
  }
}
```

---

## Lifecycle — Mounting

```tsx
class DataLoader extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    // constructor: 초기 state 설정, 이벤트 핸들러 바인딩
    // setState 호출 금지, side-effect 금지
    this.state = { data: null, error: null };
    this.handleClick = this.handleClick.bind(this);
  }

  componentDidMount() {
    // DOM 접근 가능, 네트워크 요청, 구독 설정
    fetch('/api/data')
      .then(res => res.json())
      .then(data => this.setState({ data }))
      .catch(error => this.setState({ error }));
  }

  handleClick() { /* ... */ }

  render() { return <div>{this.state.data}</div>; }
}
```

---

## Lifecycle — Updating

```tsx
class Counter extends Component<{ step: number }, { count: number }> {
  state = { count: 0 };

  static getDerivedStateFromProps(
    nextProps: { step: number },
    prevState: { count: number }
  ) {
    // props로부터 state 동기화. 순수 함수, side-effect 금지.
    // null 반환 시 state 변경 없음.
    if (nextProps.step > 10) {
      return { count: 0 }; // state 초기화
    }
    return null;
  }

  shouldComponentUpdate(nextProps: { step: number }, nextState: { count: number }) {
    // false 반환 시 리렌더링 스킵 (PureComponent가 자동 처리)
    return nextState.count !== this.state.count || nextProps.step !== this.props.step;
  }

  getSnapshotBeforeUpdate(prevProps: { step: number }, prevState: { count: number }) {
    // DOM 업데이트 직전 스냅샷 (예: 스크롤 위치)
    // 반환값이 componentDidUpdate의 3번째 인자로 전달됨
    return document.getElementById('list')?.scrollHeight ?? null;
  }

  componentDidUpdate(
    prevProps: { step: number },
    prevState: { count: number },
    snapshot: number | null
  ) {
    // 업데이트 후 side-effect (네트워크 요청 등)
    // prevProps/prevState와 비교해서 조건부 실행 필수 (무한루프 방지)
    if (prevProps.step !== this.props.step) {
      this.fetchData();
    }
    if (snapshot !== null) {
      const list = document.getElementById('list');
      if (list) list.scrollTop += list.scrollHeight - snapshot;
    }
  }

  fetchData() { /* ... */ }
  render() { return <div>{this.state.count}</div>; }
}
```

---

## Lifecycle — Unmounting & Deprecated

```tsx
class Subscription extends Component {
  private timer: ReturnType<typeof setInterval> | null = null;
  private eventSource: EventSource | null = null;

  componentDidMount() {
    this.timer = setInterval(this.tick, 1000);
    this.eventSource = new EventSource('/events');
    this.eventSource.onmessage = this.handleMessage;
  }

  componentWillUnmount() {
    // 구독 해제, 타이머 제거, 비동기 취소 — 메모리 누수 방지
    if (this.timer) clearInterval(this.timer);
    if (this.eventSource) this.eventSource.close();
  }

  tick = () => { /* ... */ };
  handleMessage = (e: MessageEvent) => { /* ... */ };
  render() { return <div />; }
}

// DEPRECATED (레거시 코드에서만 마주침, Strict Mode에서 경고)
class LegacyComponent extends Component {
  UNSAFE_componentWillMount() {
    // constructor로 이전 권장
  }
  UNSAFE_componentWillReceiveProps(nextProps: unknown) {
    // getDerivedStateFromProps + componentDidUpdate로 이전 권장
  }
  UNSAFE_componentWillUpdate(nextProps: unknown, nextState: unknown) {
    // getSnapshotBeforeUpdate로 이전 권장
  }
  render() { return null; }
}
```

---

## setState

```tsx
class Form extends Component<{}, { count: number; text: string }> {
  state = { count: 0, text: '' };

  // 객체 형태 — 비동기 배칭, 이전 state 참조 시 stale 위험
  handleClick = () => {
    this.setState({ count: this.state.count + 1 }); // stale 가능
  };

  // 함수 형태 — 최신 state 보장, 연속 호출 안전
  increment = () => {
    this.setState(prev => ({ count: prev.count + 1 }));
    this.setState(prev => ({ count: prev.count + 1 })); // 올바르게 +2
  };

  // 콜백: setState 완료 후 실행 보장
  afterUpdate = () => {
    this.setState({ text: 'done' }, () => {
      console.log('State updated:', this.state.text);
    });
  };

  render() { return <div>{this.state.count}</div>; }
}
```

---

## Refs

```tsx
class InputFocus extends Component {
  // createRef (React 16.3+)
  private inputRef = React.createRef<HTMLInputElement>();

  // Callback ref (모든 버전)
  private divRef: HTMLDivElement | null = null;
  private setDivRef = (el: HTMLDivElement | null) => { this.divRef = el; };

  componentDidMount() {
    this.inputRef.current?.focus();
  }

  render() {
    return (
      <div>
        <input ref={this.inputRef} />
        <div ref={this.setDivRef} />
      </div>
    );
  }
}

// forwardRef — 부모가 자식 DOM에 접근
const FancyInput = React.forwardRef<HTMLInputElement, { placeholder: string }>(
  (props, ref) => <input ref={ref} placeholder={props.placeholder} />
);

class Parent extends Component {
  private fancyRef = React.createRef<HTMLInputElement>();

  componentDidMount() {
    this.fancyRef.current?.focus();
  }

  render() {
    return <FancyInput ref={this.fancyRef} placeholder="Enter text" />;
  }
}
```

---

## Context (Class)

```tsx
const ThemeContext = React.createContext<'light' | 'dark'>('light');

// 방법 1: static contextType (단일 컨텍스트만 가능)
class ThemedButton extends Component {
  static contextType = ThemeContext;
  declare context: React.ContextType<typeof ThemeContext>;

  render() {
    return <button className={this.context}>{this.props.children}</button>;
  }
}

// 방법 2: Consumer (다중 컨텍스트, 구 패턴)
class MultiContextComponent extends Component {
  render() {
    return (
      <ThemeContext.Consumer>
        {theme => (
          <div className={theme}>
            {/* 중첩 Consumer로 다중 컨텍스트 */}
          </div>
        )}
      </ThemeContext.Consumer>
    );
  }
}
```

---

## HOC (Higher-Order Component)

```tsx
// withLogger HOC
function withLogger<P extends object>(WrappedComponent: React.ComponentType<P>) {
  return class WithLogger extends Component<P> {
    componentDidMount() {
      console.log(`${WrappedComponent.displayName || WrappedComponent.name} mounted`);
    }
    render() {
      return <WrappedComponent {...this.props} />;
    }
  };
}

// withAuth HOC
interface WithAuthProps {
  isAuthenticated: boolean;
  user: { name: string } | null;
}
function withAuth<P extends WithAuthProps>(WrappedComponent: React.ComponentType<P>) {
  return function WithAuth(props: Omit<P, keyof WithAuthProps>) {
    const isAuthenticated = Boolean(localStorage.getItem('token'));
    const user = isAuthenticated ? { name: 'User' } : null;
    return <WrappedComponent {...(props as P)} isAuthenticated={isAuthenticated} user={user} />;
  };
}

// connect 패턴 (Redux)
// const ConnectedComponent = connect(mapStateToProps, mapDispatchToProps)(MyComponent);
```

---

## Render Props

```tsx
interface MousePosition { x: number; y: number; }

class MouseTracker extends Component<
  { children: (pos: MousePosition) => React.ReactNode },
  MousePosition
> {
  state: MousePosition = { x: 0, y: 0 };

  handleMouseMove = (e: React.MouseEvent) => {
    this.setState({ x: e.clientX, y: e.clientY });
  };

  render() {
    return (
      <div onMouseMove={this.handleMouseMove}>
        {this.props.children(this.state)}
      </div>
    );
  }
}

// 사용
<MouseTracker>
  {({ x, y }) => <p>Mouse: {x}, {y}</p>}
</MouseTracker>
```

---

## Error Boundary

```tsx
interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundary extends Component<
  { fallback: React.ReactNode; children: React.ReactNode },
  ErrorBoundaryState
> {
  state: ErrorBoundaryState = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    // 렌더링 단계에서 호출 — state 업데이트만, side-effect 금지
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    // 커밋 단계 — 에러 로깅 서비스 호출 가능
    console.error('ErrorBoundary caught:', error, info.componentStack);
  }

  render() {
    if (this.state.hasError) return this.props.fallback;
    return this.props.children;
  }
}

// 함수 컴포넌트에서는 Error Boundary 불가 — 클래스 필수
```

---

## PureComponent vs Component

```tsx
// PureComponent: shouldComponentUpdate에서 props/state 얕은 비교 자동 수행
class PureList extends React.PureComponent<{ items: string[] }> {
  render() {
    return <ul>{this.props.items.map(i => <li key={i}>{i}</li>)}</ul>;
  }
}

// 주의: 중첩 객체/배열은 얕은 비교로 감지 불가
// items 배열의 내부 값이 바뀌어도 참조가 같으면 리렌더링 안 됨
// → 불변 업데이트 필수: setState({ items: [...items, newItem] })
```

---

## Class → Hook 마이그레이션 대응표

| Class Lifecycle | Hook 대응 |
|---|---|
| `constructor` (state init) | `useState` 초기값 |
| `constructor` (binding) | 불필요 (화살표 함수) |
| `componentDidMount` | `useEffect(() => {}, [])` |
| `componentDidUpdate` | `useEffect(() => {}, [deps])` |
| `componentWillUnmount` | `useEffect(() => { return () => cleanup(); }, [])` |
| `getDerivedStateFromProps` | `useEffect` + `useState` 또는 렌더 중 직접 계산 |
| `shouldComponentUpdate` | `React.memo` + `useMemo` |
| `getSnapshotBeforeUpdate` | `useRef` + `useLayoutEffect` |
| `componentDidCatch` | 여전히 클래스 필수 (Error Boundary) |
| `PureComponent` | `React.memo` |
| `this.state` | `useState` / `useReducer` |
| `this.setState(fn)` | `setState(prev => ...)` |
| `createRef` | `useRef` |
| `static contextType` | `useContext` |
| Render Props | Custom Hook |
| HOC | Custom Hook |

```tsx
// Before (class)
class Timer extends Component<{}, { tick: number }> {
  state = { tick: 0 };
  private id: ReturnType<typeof setInterval> | null = null;

  componentDidMount() { this.id = setInterval(() => this.setState(s => ({ tick: s.tick + 1 })), 1000); }
  componentWillUnmount() { if (this.id) clearInterval(this.id); }
  render() { return <div>{this.state.tick}</div>; }
}

// After (hook)
function Timer() {
  const [tick, setTick] = useState(0);
  useEffect(() => {
    const id = setInterval(() => setTick(t => t + 1), 1000);
    return () => clearInterval(id);
  }, []);
  return <div>{tick}</div>;
}
```
