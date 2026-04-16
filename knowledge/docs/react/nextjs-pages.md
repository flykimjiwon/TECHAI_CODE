# Next.js Pages Router (Legacy Reference)

For maintaining Next.js 12~13 Pages Router projects. Covers routing, data fetching, API routes, and migration path.

---

## 디렉토리 구조

```
pages/
  _app.tsx          # 글로벌 레이아웃, 프로바이더
  _document.tsx     # HTML 커스텀 (lang, font, script)
  _error.tsx        # 커스텀 에러 페이지
  404.tsx           # 404 페이지
  500.tsx           # 500 페이지
  index.tsx         # /
  about.tsx         # /about
  blog/
    index.tsx       # /blog
    [slug].tsx      # /blog/:slug (동적)
    [...slug].tsx   # /blog/a/b/c (catch-all)
    [[...slug]].tsx # /blog + /blog/a/b (optional catch-all)
  api/
    hello.ts        # /api/hello
    users/
      [id].ts       # /api/users/:id
public/             # 정적 파일 (/ 경로로 서빙)
styles/
  globals.css
```

---

## Routing

```tsx
import Link from 'next/link';
import { useRouter } from 'next/router';

// next/link — prefetch 기본값 true
function Nav() {
  return (
    <nav>
      <Link href="/">Home</Link>
      <Link href="/blog/my-post">Post</Link>
      {/* 동적 라우트 */}
      <Link href={{ pathname: '/blog/[slug]', query: { slug: 'my-post' } }}>
        Post
      </Link>
    </nav>
  );
}

// useRouter
function BlogPost() {
  const router = useRouter();
  const { slug, page = '1' } = router.query; // query: string | string[] | undefined

  const handleEdit = () => {
    router.push('/editor'); // 히스토리 추가
    // router.replace('/login'); // 히스토리 교체
    // router.back(); // 뒤로가기
  };

  // router.isReady: SSR 시 false, CSR hydration 후 true
  if (!router.isReady) return null;

  return <div>{slug}</div>;
}
```

---

## getServerSideProps (SSR)

```tsx
import { GetServerSideProps, InferGetServerSidePropsType } from 'next';

interface Props { user: { name: string; email: string } | null; }

export const getServerSideProps: GetServerSideProps<Props> = async (context) => {
  const { params, query, req, res, resolvedUrl, locale } = context;

  // 인증 쿠키 확인
  const token = req.cookies['auth-token'];
  if (!token) {
    return {
      redirect: { destination: '/login', permanent: false },
    };
  }

  try {
    const user = await fetchUser(token);
    return {
      props: { user },
    };
  } catch {
    return {
      notFound: true, // 404 반환
    };
  }
};

export default function ProfilePage({
  user,
}: InferGetServerSidePropsType<typeof getServerSideProps>) {
  if (!user) return null;
  return <div>{user.name}</div>;
}
```

---

## getStaticProps + ISR (SSG)

```tsx
import { GetStaticProps, InferGetStaticPropsType } from 'next';

interface Props { posts: Post[]; generatedAt: string; }

export const getStaticProps: GetStaticProps<Props> = async (context) => {
  const { locale, params } = context;

  const posts = await fetchPosts();

  return {
    props: {
      posts,
      generatedAt: new Date().toISOString(),
    },
    revalidate: 60, // ISR: 60초마다 백그라운드 재생성
    // revalidate 없으면 빌드 타임에만 생성
  };
};

export default function BlogPage({
  posts,
  generatedAt,
}: InferGetStaticPropsType<typeof getStaticProps>) {
  return (
    <div>
      <p>Generated: {generatedAt}</p>
      {posts.map(post => <PostCard key={post.id} post={post} />)}
    </div>
  );
}
```

---

## getStaticPaths (동적 SSG)

```tsx
import { GetStaticPaths, GetStaticProps } from 'next';

export const getStaticPaths: GetStaticPaths = async () => {
  const posts = await fetchAllPosts();

  return {
    paths: posts.map(post => ({
      params: { slug: post.slug },
      // locale: 'ko', // i18n 사용 시
    })),

    // fallback: false     → paths 외 경로는 404
    // fallback: true      → 첫 요청 시 폴백 UI 보여주고 백그라운드 생성
    // fallback: 'blocking' → 첫 요청 시 SSR처럼 대기 후 서빙 (SEO 유리)
    fallback: 'blocking',
  };
};

export const getStaticProps: GetStaticProps = async ({ params }) => {
  const slug = params?.slug as string;
  const post = await fetchPost(slug);

  if (!post) return { notFound: true };

  return {
    props: { post },
    revalidate: 300,
  };
};

// fallback: true 사용 시 router.isFallback 처리 필요
export default function PostPage({ post }: { post: Post }) {
  const router = useRouter();
  if (router.isFallback) return <Loading />;
  return <Article post={post} />;
}
```

---

## getInitialProps (레거시)

```tsx
// _app.tsx에서만 사용 권장 — 모든 페이지에서 자동 정적 최적화 비활성화됨
// 페이지 컴포넌트에서는 getServerSideProps / getStaticProps 사용

MyApp.getInitialProps = async (appContext: AppContext) => {
  const appProps = await App.getInitialProps(appContext);
  // 글로벌 데이터 (사용자 세션, 테마 등)
  return { ...appProps, globalData: {} };
};
```

---

## API Routes

```tsx
// pages/api/users/[id].ts
import type { NextApiRequest, NextApiResponse } from 'next';

interface UserResponse { id: string; name: string; }
interface ErrorResponse { message: string; }

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse<UserResponse | ErrorResponse>
) {
  const { id } = req.query;

  switch (req.method) {
    case 'GET': {
      const user = await db.user.findUnique({ where: { id: id as string } });
      if (!user) return res.status(404).json({ message: 'Not found' });
      return res.status(200).json(user);
    }

    case 'PUT': {
      const body = req.body as Partial<UserResponse>;
      const updated = await db.user.update({ where: { id: id as string }, data: body });
      return res.status(200).json(updated);
    }

    case 'DELETE': {
      await db.user.delete({ where: { id: id as string } });
      return res.status(204).end();
    }

    default:
      res.setHeader('Allow', ['GET', 'PUT', 'DELETE']);
      return res.status(405).json({ message: `Method ${req.method} not allowed` });
  }
}

// 설정: 기본 body parser 비활성화 (파일 업로드 등)
export const config = { api: { bodyParser: false } };
```

---

## _app.tsx

```tsx
import type { AppProps } from 'next/app';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import '../styles/globals.css';

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 60_000, retry: 1 } },
});

export default function MyApp({ Component, pageProps }: AppProps) {
  // Component: 현재 페이지 컴포넌트
  // pageProps: getServerSideProps/getStaticProps의 반환값

  // 페이지별 레이아웃 패턴
  const getLayout = (Component as any).getLayout ?? ((page: React.ReactNode) => page);

  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        {getLayout(<Component {...pageProps} />)}
      </AuthProvider>
    </QueryClientProvider>
  );
}

// 페이지에서 레이아웃 지정
// pages/dashboard.tsx
DashboardPage.getLayout = (page: React.ReactNode) => (
  <DashboardLayout>{page}</DashboardLayout>
);
```

---

## _document.tsx

```tsx
import { Html, Head, Main, NextScript } from 'next/document';

export default function Document() {
  return (
    <Html lang="ko">
      <Head>
        {/* 폰트, 파비콘 — 모든 페이지 공통 */}
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link
          href="https://fonts.googleapis.com/css2?family=Noto+Sans+KR&display=swap"
          rel="stylesheet"
        />
        <link rel="icon" href="/favicon.ico" />
      </Head>
      <body>
        <Main />
        <NextScript />
        {/* 모달 Portal 루트 */}
        <div id="modal-root" />
      </body>
    </Html>
  );
}
// 주의: _document는 서버에서만 렌더링 — React lifecycle, browser API 사용 불가
```

---

## next/head (SEO)

```tsx
import Head from 'next/head';

function ProductPage({ product }: { product: Product }) {
  return (
    <>
      <Head>
        <title>{product.name} | My Store</title>
        <meta name="description" content={product.description} />
        <meta property="og:title" content={product.name} />
        <meta property="og:image" content={product.imageUrl} />
        <meta name="robots" content="index, follow" />
        <link rel="canonical" href={`https://example.com/products/${product.slug}`} />
      </Head>
      <main>{/* ... */}</main>
    </>
  );
}
// 같은 키의 태그가 여러 번 선언되면 마지막 선언이 우선 (부모 < 자식)
```

---

## next/image (Pages Router)

```tsx
import Image from 'next/image';

function HeroSection() {
  return (
    <div style={{ position: 'relative', width: '100%', height: 400 }}>
      {/* fill: 부모 기준으로 채움 (부모에 position:relative 필요) */}
      <Image
        src="/hero.jpg"
        alt="Hero"
        fill
        style={{ objectFit: 'cover' }}
        priority // LCP 이미지에 적용 — preload
        sizes="100vw"
      />
    </div>
  );
}

function ProductCard({ product }: { product: Product }) {
  return (
    // width/height 명시 시 자동으로 공간 확보 (CLS 방지)
    <Image
      src={product.imageUrl}
      alt={product.name}
      width={300}
      height={200}
      placeholder="blur"         // 로딩 중 블러 처리
      blurDataURL={product.blurHash} // base64 blur placeholder
    />
  );
}

// 외부 이미지는 next.config.js에 도메인 등록 필요
// images: { domains: ['cdn.example.com'] }
```

---

## Middleware (Edge Runtime)

```tsx
// middleware.ts (루트 또는 src/ 아래)
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const token = request.cookies.get('auth-token')?.value;
  const { pathname } = request.nextUrl;

  // 인증 보호
  if (pathname.startsWith('/dashboard') && !token) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  // 헤더 추가
  const response = NextResponse.next();
  response.headers.set('x-pathname', pathname);
  return response;
}

export const config = {
  // middleware가 실행될 경로 패턴
  matcher: ['/dashboard/:path*', '/api/protected/:path*'],
};
// 주의: Edge Runtime — Node.js API(fs, crypto 등) 사용 불가
```

---

## 환경변수

```bash
# .env.local (git 제외)
DATABASE_URL=postgresql://...    # 서버 전용
NEXT_PUBLIC_API_URL=https://...  # 클라이언트+서버 모두 노출
NEXT_PUBLIC_GA_ID=G-XXXXX

# 접근
process.env.DATABASE_URL         # 서버 코드에서만
process.env.NEXT_PUBLIC_API_URL  # 어디서나

# .env.production, .env.development (공개 가능한 기본값)
```

---

## Pages Router → App Router 마이그레이션

| Pages Router | App Router |
|---|---|
| `pages/index.tsx` | `app/page.tsx` |
| `pages/blog/[slug].tsx` | `app/blog/[slug]/page.tsx` |
| `pages/_app.tsx` | `app/layout.tsx` |
| `pages/_document.tsx` | `app/layout.tsx` (html, body) |
| `pages/api/route.ts` | `app/api/route/route.ts` |
| `getServerSideProps` | async Server Component + fetch |
| `getStaticProps` + `revalidate` | fetch + `{ next: { revalidate: 60 } }` |
| `getStaticPaths` | `generateStaticParams()` |
| `useRouter` (next/router) | `useRouter` (next/navigation) |
| `router.query` | `useSearchParams()` + `useParams()` |
| `next/head` | `export const metadata` |
| `next/image` | `next/image` (동일, API 일부 변경) |

```tsx
// Before: getServerSideProps
export const getServerSideProps: GetServerSideProps = async ({ params }) => {
  const data = await fetch(`/api/posts/${params?.id}`).then(r => r.json());
  return { props: { data } };
};

// After: App Router Server Component
// app/posts/[id]/page.tsx
async function PostPage({ params }: { params: { id: string } }) {
  const data = await fetch(`/api/posts/${params.id}`, { cache: 'no-store' }).then(r => r.json());
  return <div>{data.title}</div>;
}

// Before: getStaticProps + revalidate (ISR)
export const getStaticProps = async () => ({
  props: { posts: await fetchPosts() },
  revalidate: 60,
});

// After: fetch with revalidate
async function BlogPage() {
  const posts = await fetch('/api/posts', { next: { revalidate: 60 } }).then(r => r.json());
  return <PostList posts={posts} />;
}

// Before: pages/api/hello.ts
export default function handler(req: NextApiRequest, res: NextApiResponse) {
  res.json({ message: 'hello' });
}

// After: app/api/hello/route.ts
export async function GET() {
  return Response.json({ message: 'hello' });
}
export async function POST(request: Request) {
  const body = await request.json();
  return Response.json({ received: body });
}

// useRouter 차이점
// Pages: import { useRouter } from 'next/router' → router.query.id
// App:   import { useRouter } from 'next/navigation' → router.push('/') 만
//        params는 useParams(), search는 useSearchParams() 별도 훅
```
