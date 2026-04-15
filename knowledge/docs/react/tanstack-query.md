# TanStack Query v5 Reference

## Setup
```tsx
// app/providers.tsx
'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { useState } from 'react'

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 60 * 1000,  // 1분
        gcTime: 5 * 60 * 1000, // 5분 (구 cacheTime)
        retry: 1,
        refetchOnWindowFocus: false,
      },
    },
  })
}

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => makeQueryClient())
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  )
}
```

## useQuery
```tsx
import { useQuery } from '@tanstack/react-query'

function UserProfile({ userId }: { userId: number }) {
  const {
    data,
    isLoading,
    isError,
    error,
    isFetching,
    refetch,
  } = useQuery({
    queryKey: ['user', userId],        // 배열 — userId 바뀌면 자동 재요청
    queryFn: () => fetchUser(userId),
    staleTime: 5 * 60 * 1000,         // 5분간 fresh
    gcTime: 10 * 60 * 1000,           // 10분 후 GC
    enabled: userId > 0,              // false면 요청 안 함
    select: (data) => data.profile,   // 데이터 변환 (원본 캐시 유지)
    placeholderData: previousData => previousData, // 페이지네이션 깜빡임 방지
  })

  if (isLoading) return <p>로딩 중...</p>
  if (isError) return <p>에러: {error.message}</p>

  return <div>{data.name}</div>
}
```

## useMutation
```tsx
import { useMutation, useQueryClient } from '@tanstack/react-query'

function AddTodo() {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: (newTodo: { title: string }) => createTodo(newTodo),

    // Optimistic update
    onMutate: async (newTodo) => {
      await queryClient.cancelQueries({ queryKey: ['todos'] })

      const previousTodos = queryClient.getQueryData<Todo[]>(['todos'])

      queryClient.setQueryData<Todo[]>(['todos'], (old) => [
        ...(old ?? []),
        { id: Date.now(), ...newTodo, completed: false },
      ])

      return { previousTodos } // context로 전달
    },

    // 에러 시 롤백
    onError: (_err, _variables, context) => {
      if (context?.previousTodos) {
        queryClient.setQueryData(['todos'], context.previousTodos)
      }
    },

    // 성공/실패 무관하게 캐시 무효화
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] })
    },
  })

  return (
    <button
      onClick={() => mutation.mutate({ title: '새 할 일' })}
      disabled={mutation.isPending}
    >
      {mutation.isPending ? '추가 중...' : '추가'}
    </button>
  )
}
```

## Cache Invalidation
```tsx
const queryClient = useQueryClient()

// 특정 키 prefix 전체 무효화
queryClient.invalidateQueries({ queryKey: ['todos'] })
// ['todos'], ['todos', 1], ['todos', 'list'] 모두 무효화

// 정확히 일치하는 키만
queryClient.invalidateQueries({ queryKey: ['todos', 1], exact: true })

// 조건부 무효화
queryClient.invalidateQueries({
  predicate: (query) =>
    query.queryKey[0] === 'todos' && query.state.data?.userId === 1,
})
```

## QueryClient 유틸 메서드
```tsx
// 수동으로 캐시 데이터 설정
queryClient.setQueryData<User>(['user', 1], (old) => ({
  ...old!,
  name: '새 이름',
}))

// 캐시 데이터 읽기
const user = queryClient.getQueryData<User>(['user', 1])

// 서버 요청 전에 미리 패치
await queryClient.prefetchQuery({
  queryKey: ['user', userId],
  queryFn: () => fetchUser(userId),
})

// 진행 중인 요청 취소
await queryClient.cancelQueries({ queryKey: ['todos'] })

// 캐시에서 완전히 제거
queryClient.removeQueries({ queryKey: ['todos'] })
```

## 자주 쓰는 패턴

### 의존 쿼리 (dependent queries)
```tsx
const { data: user } = useQuery({
  queryKey: ['user', userId],
  queryFn: () => fetchUser(userId),
})

const { data: posts } = useQuery({
  queryKey: ['posts', user?.id],
  queryFn: () => fetchPosts(user!.id),
  enabled: !!user?.id, // user 로드 후에만 실행
})
```

### 무한 스크롤
```tsx
import { useInfiniteQuery } from '@tanstack/react-query'

const { data, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
  queryKey: ['posts'],
  queryFn: ({ pageParam }) => fetchPosts({ page: pageParam }),
  initialPageParam: 1,
  getNextPageParam: (lastPage, pages) =>
    lastPage.hasMore ? pages.length + 1 : undefined,
})

const allPosts = data?.pages.flatMap((page) => page.items) ?? []
```

### 쿼리 키 팩토리 (타입 안전)
```tsx
const todoKeys = {
  all: ['todos'] as const,
  lists: () => [...todoKeys.all, 'list'] as const,
  list: (filters: string) => [...todoKeys.lists(), filters] as const,
  details: () => [...todoKeys.all, 'detail'] as const,
  detail: (id: number) => [...todoKeys.details(), id] as const,
}

useQuery({ queryKey: todoKeys.detail(1), queryFn: () => fetchTodo(1) })
queryClient.invalidateQueries({ queryKey: todoKeys.lists() })
```
