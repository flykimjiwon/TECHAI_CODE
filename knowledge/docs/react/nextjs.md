# Next.js 15 App Router Reference

## Directory Structure
```
app/
├── layout.tsx          # Root layout (필수)
├── page.tsx            # Home page
├── loading.tsx         # Loading UI (Suspense)
├── error.tsx           # Error boundary
├── not-found.tsx       # 404 page
├── api/
│   └── users/
│       └── route.ts    # API route
└── [slug]/
    └── page.tsx        # Dynamic route
```

## Root Layout
```tsx
// app/layout.tsx
export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ko">
      <body>{children}</body>
    </html>
  )
}
```

## Server vs Client Components
```tsx
// Server Component (기본값)
export default async function Page() {
  const data = await fetch('https://api.example.com/data')
  return <div>{data.title}</div>
}

// Client Component (상호작용 필요 시)
'use client'

import { useState } from 'react'

export default function Counter() {
  const [count, setCount] = useState(0)
  return <button onClick={() => setCount(count + 1)}>{count}</button>
}
```

## API Routes
```tsx
// app/api/users/route.ts
import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  const users = await db.users.findMany()
  return NextResponse.json(users)
}

export async function POST(request: NextRequest) {
  const body = await request.json()
  const user = await db.users.create({ data: body })
  return NextResponse.json(user, { status: 201 })
}
```

## Dynamic Routes
```tsx
// app/posts/[slug]/page.tsx
export default async function Post({ params }: { params: { slug: string } }) {
  const post = await getPost(params.slug)
  return <article>{post.content}</article>
}

// Generate static params
export async function generateStaticParams() {
  const posts = await getPosts()
  return posts.map(post => ({ slug: post.slug }))
}
```

## Metadata
```tsx
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Page Title',
  description: 'Description',
}

// Dynamic metadata
export async function generateMetadata({ params }: { params: { id: string } }): Promise<Metadata> {
  const post = await getPost(params.id)
  return {
    title: post.title,
    description: post.excerpt,
  }
}
```

## Middleware
```tsx
// middleware.ts (root)
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const token = request.cookies.get('token')
  if (!token) {
    return NextResponse.redirect(new URL('/login', request.url))
  }
  return NextResponse.next()
}

export const config = {
  matcher: '/dashboard/:path*',
}
```

## Loading & Error
```tsx
// app/loading.tsx
export default function Loading() {
  return <div>Loading...</div>
}

// app/error.tsx
'use client'

export default function Error({ error, reset }: { error: Error; reset: () => void }) {
  return (
    <div>
      <h2>Something went wrong!</h2>
      <button onClick={() => reset()}>Try again</button>
    </div>
  )
}
```

## Data Fetching
```tsx
// Server Component에서 직접 fetch
const data = await fetch('https://api.example.com/data', {
  next: { revalidate: 3600 } // ISR: 1시간마다 재생성
})

// No cache
const data = await fetch('https://api.example.com/data', {
  cache: 'no-store'
})
```

## Route Groups
```tsx
app/
├── (marketing)/
│   ├── layout.tsx
│   └── page.tsx
└── (dashboard)/
    ├── layout.tsx
    └── page.tsx
// (marketing), (dashboard)는 URL에 포함되지 않음
```
