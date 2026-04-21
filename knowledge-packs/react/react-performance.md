# React Performance

Memoization, code splitting, virtualization, React 18 concurrent features, and anti-patterns.

---

## React.memo

```tsx
// 언제 쓸지: props가 자주 변경되지 않는 순수 컴포넌트, 렌더 비용이 높을 때
const ExpensiveList = React.memo(function ExpensiveList({
  items,
  onItemClick,
}: {
  items: string[];
  onItemClick: (item: string) => void;
}) {
  return (
    <ul>
      {items.map(item => (
        <li key={item} onClick={() => onItemClick(item)}>{item}</li>
      ))}
    </ul>
  );
});

// 커스텀 비교 함수 (얕은 비교로 부족할 때)
const MemoizedChart = React.memo(Chart, (prev, next) => {
  return prev.data.length === next.data.length &&
    prev.data.every((d, i) => d.value === next.data[i].value);
});

// 쓰지 말아야 할 때:
// - props가 매번 달라지는 컴포넌트 (비교 비용만 추가)
// - 렌더가 이미 빠른 단순 컴포넌트
// - children을 받는 컴포넌트 (JSX는 매번 새 객체)
```

---

## useMemo

```tsx
function ProductTable({ products, filter }: { products: Product[]; filter: string }) {
  // 비싼 계산 캐싱
  const filteredProducts = useMemo(
    () => products.filter(p => p.name.toLowerCase().includes(filter.toLowerCase())),
    [products, filter]
  );

  // 참조 안정성 — 자식에 객체/배열 전달 시 리렌더 방지
  const config = useMemo(() => ({ pageSize: 20, sortBy: 'name' }), []);

  return <Table data={filteredProducts} config={config} />;
}

// useMemo 없이 인라인 객체 → 매번 새 참조 → memo된 자식도 리렌더
// 나쁜 예:
function Bad({ items }: { items: string[] }) {
  return <MemoizedChild style={{ color: 'red' }} items={items} />;
  // { color: 'red' }가 매번 새 객체 → MemoizedChild 항상 리렌더
}

// 좋은 예:
const STYLE = { color: 'red' }; // 모듈 레벨 상수
function Good({ items }: { items: string[] }) {
  return <MemoizedChild style={STYLE} items={items} />;
}
```

---

## useCallback

```tsx
function Parent() {
  const [count, setCount] = useState(0);
  const [text, setText] = useState('');

  // text가 바뀌어도 handleSubmit 참조 유지 → MemoizedChild 리렌더 방지
  const handleSubmit = useCallback((value: string) => {
    console.log('submit', value, count);
  }, [count]); // count가 deps — count 바뀌면 함수 재생성

  return (
    <>
      <input value={text} onChange={e => setText(e.target.value)} />
      <MemoizedChild onSubmit={handleSubmit} />
      <button onClick={() => setCount(c => c + 1)}>{count}</button>
    </>
  );
}

// useCallback은 항상 React.memo와 함께 쓸 때 의미 있음
// memo 없는 자식에 useCallback 전달 → 아무 효과 없음
```

---

## memo + useMemo + useCallback 조합 전략

```tsx
// 목표: 무거운 자식 컴포넌트가 부모의 무관한 state 변경에 리렌더되지 않도록
interface DataGridProps {
  rows: Row[];
  columns: Column[];
  onRowClick: (row: Row) => void;
  onSort: (key: string) => void;
}

const DataGrid = React.memo(({ rows, columns, onRowClick, onSort }: DataGridProps) => {
  // 내부에서도 비싼 계산은 useMemo
  const sortedRows = useMemo(() => [...rows].sort((a, b) => a.id - b.id), [rows]);
  return (/* ... */);
});

function Dashboard() {
  const [filter, setFilter] = useState('');
  const [selectedId, setSelectedId] = useState<string | null>(null);

  // rows, columns가 filter에 의존하지 않으면 useMemo로 안정화
  const columns = useMemo<Column[]>(() => [
    { key: 'name', label: 'Name' },
    { key: 'status', label: 'Status' },
  ], []);

  const handleRowClick = useCallback((row: Row) => {
    setSelectedId(row.id);
  }, []); // 의존성 없음 → 항상 동일 참조

  const handleSort = useCallback((key: string) => {
    console.log('sort by', key);
  }, []);

  return (
    <>
      <input value={filter} onChange={e => setFilter(e.target.value)} />
      {/* filter 입력 시 DataGrid 리렌더 안 됨 */}
      <DataGrid rows={rows} columns={columns} onRowClick={handleRowClick} onSort={handleSort} />
    </>
  );
}
```

---

## React DevTools Profiler

```
1. DevTools → Profiler 탭 → Record 클릭
2. 사용자 인터랙션 수행
3. Record 중지 → 각 커밋의 렌더 시간 확인
4. 회색 컴포넌트 = 렌더 안 됨, 색상 = 렌더됨 (노란색→빨간색 = 더 오래 걸림)
5. 컴포넌트 클릭 → "Why did this render?" 확인:
   - "Props changed" → memo + useCallback/useMemo 적용
   - "State changed" → 상태 분리 고려
   - "Parent rendered" → 부모에 memo 적용 또는 children 분리
```

---

## why-did-you-render

```tsx
// src/wdyr.ts (개발 전용)
import React from 'react';
if (process.env.NODE_ENV === 'development') {
  const whyDidYouRender = require('@welldone-software/why-did-you-render');
  whyDidYouRender(React, {
    trackAllPureComponents: false, // true 시 모든 컴포넌트 추적 (노이즈 많음)
    trackExtraHooks: [[require('react-redux'), 'useSelector']],
  });
}

// 특정 컴포넌트만 추적
MyComponent.whyDidYouRender = true;

// 콘솔 출력 예시:
// [why-did-you-render] Re-rendered MyComponent
// different props: { items: [old] → [new] }
```

---

## 가상화 (대량 리스트)

```tsx
// react-window — 고정 높이 리스트
import { FixedSizeList } from 'react-window';

function VirtualList({ items }: { items: string[] }) {
  const Row = ({ index, style }: { index: number; style: React.CSSProperties }) => (
    <div style={style}>{items[index]}</div>
  );

  return (
    <FixedSizeList height={600} width="100%" itemCount={items.length} itemSize={48}>
      {Row}
    </FixedSizeList>
  );
}

// react-virtuoso — 가변 높이, 그룹, 무한 스크롤 지원
import { Virtuoso } from 'react-virtuoso';

function DynamicList({ items }: { items: { id: string; content: string }[] }) {
  return (
    <Virtuoso
      style={{ height: '600px' }}
      totalCount={items.length}
      itemContent={index => <ItemCard item={items[index]} />}
      endReached={() => loadMore()} // 무한 스크롤
    />
  );
}
```

---

## Code Splitting

```tsx
// React.lazy + Suspense — 컴포넌트 레벨
const HeavyChart = React.lazy(() => import('./HeavyChart'));
const AdminPanel = React.lazy(() => import('./AdminPanel'));

function App() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <HeavyChart />
    </Suspense>
  );
}

// Route-based splitting (Next.js Pages Router)
const DashboardPage = dynamic(() => import('../components/Dashboard'), {
  loading: () => <Skeleton />,
  ssr: false, // 클라이언트 전용 컴포넌트
});

// 조건부 로딩
function FeatureFlag({ enabled }: { enabled: boolean }) {
  const LazyFeature = React.lazy(() => import('./NewFeature'));
  if (!enabled) return null;
  return (
    <Suspense fallback={<Spinner />}>
      <LazyFeature />
    </Suspense>
  );
}
```

---

## Suspense 전략

```tsx
// 중첩 Suspense 경계 — 로딩 단계 분리
function Page() {
  return (
    <Suspense fallback={<PageSkeleton />}>     {/* 전체 페이지 */}
      <Header />
      <Suspense fallback={<ContentSkeleton />}> {/* 주요 컨텐츠 */}
        <MainContent />
        <Suspense fallback={<CommentSkeleton />}> {/* 부가 컨텐츠 */}
          <Comments />
        </Suspense>
      </Suspense>
    </Suspense>
  );
}

// React 18 — use() + Promise (Server Component 외 실험적)
// TanStack Query와 함께 사용 (안정적)
function ProductDetail({ id }: { id: string }) {
  const { data } = useSuspenseQuery({
    queryKey: ['product', id],
    queryFn: () => fetchProduct(id),
  });
  // data는 항상 존재 (로딩 중엔 Suspense가 처리)
  return <div>{data.name}</div>;
}
```

---

## React 18: useTransition, useDeferredValue

```tsx
// useTransition — 긴급하지 않은 state 업데이트를 낮은 우선순위로
function SearchResults() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<string[]>([]);
  const [isPending, startTransition] = useTransition();

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setQuery(e.target.value); // 즉시 (입력 반응성 유지)
    startTransition(() => {
      setResults(expensiveSearch(e.target.value)); // 낮은 우선순위
    });
  };

  return (
    <>
      <input value={query} onChange={handleSearch} />
      {isPending ? <Spinner /> : <ResultList items={results} />}
    </>
  );
}

// useDeferredValue — 외부 값에 적용 (props로 받을 때)
function FilteredList({ query }: { query: string }) {
  const deferredQuery = useDeferredValue(query);
  const isStale = query !== deferredQuery;

  const filtered = useMemo(
    () => expensiveFilter(deferredQuery),
    [deferredQuery]
  );

  return (
    <div style={{ opacity: isStale ? 0.6 : 1 }}>
      {filtered.map(item => <Item key={item.id} item={item} />)}
    </div>
  );
}
```

---

## 이미지 최적화 (Intersection Observer)

```tsx
function LazyImage({ src, alt }: { src: string; alt: string }) {
  const [isVisible, setIsVisible] = useState(false);
  const imgRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) { setIsVisible(true); observer.disconnect(); } },
      { rootMargin: '100px' } // 100px 전에 미리 로드
    );
    if (imgRef.current) observer.observe(imgRef.current);
    return () => observer.disconnect();
  }, []);

  return (
    <div ref={imgRef}>
      {isVisible
        ? <img src={src} alt={alt} loading="lazy" />
        : <div className="img-placeholder" />
      }
    </div>
  );
}
```

---

## 실전: 무한 스크롤 (IntersectionObserver + 가상화)

```tsx
function InfiniteList() {
  const [pages, setPages] = useState<Item[][]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const sentinelRef = useRef<HTMLDivElement>(null);
  const pageRef = useRef(0);

  const loadMore = useCallback(async () => {
    if (isLoading || !hasMore) return;
    setIsLoading(true);
    const data = await fetchPage(pageRef.current + 1);
    pageRef.current += 1;
    setPages(prev => [...prev, data.items]);
    setHasMore(data.hasNextPage);
    setIsLoading(false);
  }, [isLoading, hasMore]);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) loadMore(); },
      { threshold: 0.1 }
    );
    if (sentinelRef.current) observer.observe(sentinelRef.current);
    return () => observer.disconnect();
  }, [loadMore]);

  // 초기 로드
  useEffect(() => { loadMore(); }, []); // eslint-disable-line

  const allItems = useMemo(() => pages.flat(), [pages]);

  return (
    <Virtuoso
      style={{ height: '100vh' }}
      data={allItems}
      itemContent={(_, item) => <ItemCard item={item} />}
      components={{ Footer: () => (
        <div ref={sentinelRef} style={{ height: 40 }}>
          {isLoading && <Spinner />}
          {!hasMore && <p>모든 항목을 불러왔습니다</p>}
        </div>
      )}}
    />
  );
}
```

---

## 안티패턴

```tsx
// 1. 불필요한 state — 계산 가능한 값
// 나쁨
const [fullName, setFullName] = useState('');
useEffect(() => setFullName(`${firstName} ${lastName}`), [firstName, lastName]);

// 좋음
const fullName = `${firstName} ${lastName}`; // 렌더 시 직접 계산

// 2. 객체/배열 리터럴 in JSX — 매 렌더마다 새 참조
// 나쁨
<MemoizedChild options={{ limit: 10 }} tags={['a', 'b']} />

// 좋음 (모듈 레벨 상수 또는 useMemo)
const OPTIONS = { limit: 10 };
const TAGS = ['a', 'b'];
<MemoizedChild options={OPTIONS} tags={TAGS} />

// 3. 인라인 함수 남용
// 나쁨 — 매 렌더마다 새 함수 → memo된 자식 항상 리렌더
<MemoizedButton onClick={() => handleAction(id)} />

// 좋음
const handleButtonClick = useCallback(() => handleAction(id), [id]);
<MemoizedButton onClick={handleButtonClick} />

// 4. useEffect에서 state 기반 파생 계산
// 나쁨 — 렌더 2번 발생
const [doubled, setDoubled] = useState(0);
useEffect(() => setDoubled(count * 2), [count]);

// 좋음 — 렌더 1번
const doubled = count * 2;

// 5. Context로 모든 것 전달 — Context 값 변경 시 모든 소비자 리렌더
// 고주파 업데이트 (마우스 좌표, 스크롤)는 Context 대신 ref 또는 Zustand 사용
```

---

## 번들 분석

```bash
# Next.js
npm install @next/bundle-analyzer
# next.config.js에 withBundleAnalyzer 적용 후
ANALYZE=true npm run build

# CRA / Vite
npx source-map-explorer 'build/static/js/*.js'

# 확인 포인트:
# - 중복 패키지 (lodash vs lodash-es)
# - moment.js → date-fns로 교체 (tree-shaking 지원)
# - 아이콘 라이브러리 전체 import → named import로
# 나쁨: import * as Icons from 'react-icons/fa'
# 좋음: import { FaUser } from 'react-icons/fa'
```
