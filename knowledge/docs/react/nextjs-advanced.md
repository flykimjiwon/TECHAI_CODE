# Next.js 15 Advanced Reference

## Server Actions
```tsx
// app/actions.ts
'use server'

import { revalidatePath, revalidateTag } from 'next/cache'
import { redirect } from 'next/navigation'

export async function createPost(formData: FormData) {
  const title = formData.get('title') as string
  await db.post.create({ data: { title } })

  revalidatePath('/posts')           // 특정 경로 캐시 무효화
  revalidateTag('posts')             // 태그 기반 무효화
  redirect('/posts')                 // 리다이렉트 (throw 기반)
}

// 컴포넌트에서 사용
export default function NewPostForm() {
  return (
    <form action={createPost}>
      <input name="title" required />
      <button type="submit">작성</button>
    </form>
  )
}
```

## Caching
```tsx
// fetch 캐시 옵션
const data = await fetch('https://api.example.com/posts', {
  cache: 'force-cache',           // 기본값: 영구 캐시
  next: { revalidate: 3600 },     // ISR: 1시간마다 재검증
  cache: 'no-store',              // 캐시 없음 (동적)
  next: { tags: ['posts'] },      // 태그 기반 무효화 가능
})

// unstable_cache: DB 쿼리 등 fetch 외 캐싱
import { unstable_cache } from 'next/cache'

const getCachedPosts = unstable_cache(
  async (userId: number) => {
    return db.post.findMany({ where: { userId } })
  },
  ['user-posts'],                  // 캐시 키
  {
    revalidate: 60,                // 60초마다 재검증
    tags: ['posts'],               // revalidateTag('posts') 로 무효화
  }
)

// 사용
const posts = await getCachedPosts(userId)
```

## next/image
```tsx
import Image from 'next/image'

// 고정 크기
<Image
  src="/hero.jpg"
  alt="히어로 이미지"
  width={800}
  height={400}
  priority           // LCP 이미지에 사용 (preload)
  quality={85}       // 기본 75
/>

// fill (부모 컨테이너 채우기)
<div className="relative h-64 w-full">
  <Image
    src="/banner.jpg"
    alt="배너"
    fill
    sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
    className="object-cover"
  />
</div>

// placeholder blur (로컬 이미지 자동, 외부는 blurDataURL 필요)
<Image
  src="/photo.jpg"
  alt="사진"
  width={400}
  height={300}
  placeholder="blur"
/>

// 외부 이미지 — next.config.ts에 도메인 허용 필요
// next.config.ts
images: {
  remotePatterns: [{ hostname: 'images.unsplash.com' }],
}
```

## next/font
```tsx
// Google Fonts
import { Inter, Noto_Sans_KR } from 'next/font/google'

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
  display: 'swap',
})

const notoSansKR = Noto_Sans_KR({
  subsets: ['latin'],
  weight: ['400', '700'],
  variable: '--font-noto',
})

// app/layout.tsx
export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ko" className={`${inter.variable} ${notoSansKR.variable}`}>
      <body className="font-sans">{children}</body>
    </html>
  )
}

// tailwind.config.ts — CSS 변수 연결
theme: {
  extend: {
    fontFamily: {
      sans: ['var(--font-noto)', 'var(--font-inter)'],
    },
  },
}

// 로컬 폰트
import localFont from 'next/font/local'

const pretendard = localFont({
  src: './fonts/PretendardVariable.woff2',
  variable: '--font-pretendard',
  display: 'swap',
})
```

## Metadata API
```tsx
// app/layout.tsx — 정적 메타데이터
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: { template: '%s | 내사주', default: '내사주' },
  description: '사주 + MBTI 통합 플랫폼',
  openGraph: {
    title: '내사주',
    description: '사주 + MBTI 통합 플랫폼',
    url: 'https://naesaju.kr',
    siteName: '내사주',
    images: [{ url: '/og-image.png', width: 1200, height: 630 }],
    locale: 'ko_KR',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: '내사주',
    images: ['/og-image.png'],
  },
}

// app/posts/[id]/page.tsx — 동적 메타데이터
export async function generateMetadata(
  { params }: { params: { id: string } }
): Promise<Metadata> {
  const post = await getPost(params.id)
  return {
    title: post.title,
    openGraph: { images: [{ url: post.thumbnail }] },
  }
}

// opengraph-image.tsx — 동적 OG 이미지 생성
// app/posts/[id]/opengraph-image.tsx
import { ImageResponse } from 'next/og'

export const runtime = 'edge'
export const size = { width: 1200, height: 630 }

export default async function OgImage({ params }: { params: { id: string } }) {
  const post = await getPost(params.id)
  return new ImageResponse(
    <div style={{ fontSize: 48, background: 'white', width: '100%', height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
      {post.title}
    </div>,
    size
  )
}
```

## Parallel Routes
```tsx
// app/layout.tsx — 여러 슬롯을 동시에 렌더링
export default function Layout({
  children,
  team,
  analytics,
}: {
  children: React.ReactNode
  team: React.ReactNode       // @team 슬롯
  analytics: React.ReactNode  // @analytics 슬롯
}) {
  return (
    <div>
      {children}
      <div className="grid grid-cols-2">
        {team}
        {analytics}
      </div>
    </div>
  )
}

// 디렉토리 구조
// app/
// ├── layout.tsx
// ├── page.tsx
// ├── @team/
// │   └── page.tsx
// └── @analytics/
//     ├── page.tsx
//     └── default.tsx  ← 매칭 안 될 때 폴백
```

## Intercepting Routes
```tsx
// 현재 경로를 유지하면서 다른 경로를 오버레이로 렌더링
// 예: 피드에서 사진 클릭 시 모달, URL은 /photos/1로 변경

// 디렉토리 구조
// app/
// ├── feed/
// │   └── page.tsx
// ├── photos/
// │   └── [id]/
// │       └── page.tsx        ← 직접 접근 시 전체 페이지
// └── @modal/
//     ├── default.tsx          ← null 반환
//     └── (.)photos/           ← (.) = 같은 레벨 가로채기
//         └── [id]/
//             └── page.tsx     ← 모달로 렌더링

// (.) 같은 레벨 | (..) 한 단계 위 | (...) root

// app/@modal/(.)photos/[id]/page.tsx
export default function PhotoModal({ params }: { params: { id: string } }) {
  return (
    <dialog open>
      <img src={`/photos/${params.id}.jpg`} alt="사진" />
    </dialog>
  )
}
```

## 자주 쓰는 패턴

### Server Action + redirect + revalidate
```tsx
'use server'
import { revalidatePath } from 'next/cache'
import { redirect } from 'next/navigation'

export async function deletePost(id: number) {
  await db.post.delete({ where: { id } })
  revalidatePath('/posts')
  redirect('/posts')
}
```

### 조건부 캐시 (사용자별 데이터)
```tsx
import { cookies } from 'next/headers'
import { unstable_cache } from 'next/cache'

async function getPersonalizedData(userId: string) {
  // 공개 데이터: 캐싱
  const publicData = await unstable_cache(
    () => fetchPublicData(),
    ['public-data'],
    { revalidate: 3600 }
  )()

  // 개인 데이터: no-store
  const personalData = await fetch(`/api/user/${userId}`, {
    cache: 'no-store',
  })

  return { publicData, personalData: await personalData.json() }
}
```

### Parallel Routes로 모달 패턴
```tsx
// app/@modal/default.tsx
export default function Default() {
  return null // 매칭 없을 때 아무것도 렌더링 안 함
}

// app/layout.tsx
export default function RootLayout({ children, modal }: { children: React.ReactNode; modal: React.ReactNode }) {
  return (
    <html>
      <body>
        {children}
        {modal}
      </body>
    </html>
  )
}
```
