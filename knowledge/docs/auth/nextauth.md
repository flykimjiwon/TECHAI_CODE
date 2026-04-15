# Auth.js v5 (NextAuth) Reference

## Installation
```bash
npm install next-auth@beta
npx auth secret          # generates AUTH_SECRET
```

## auth.ts Config
```ts
// auth.ts (project root)
import NextAuth from "next-auth"
import GitHub from "next-auth/providers/github"
import Google from "next-auth/providers/google"
import Credentials from "next-auth/providers/credentials"
import { PrismaAdapter } from "@auth/prisma-adapter"
import { prisma } from "@/lib/prisma"
import { z } from "zod"

export const { handlers, auth, signIn, signOut } = NextAuth({
  adapter: PrismaAdapter(prisma),
  session: { strategy: "jwt" },   // "jwt" | "database"
  pages: {
    signIn:  "/login",
    signOut: "/logout",
    error:   "/auth/error",
  },
  providers: [
    GitHub({
      clientId:     process.env.GITHUB_ID!,
      clientSecret: process.env.GITHUB_SECRET!,
    }),
    Google({
      clientId:     process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
    }),
    Credentials({
      credentials: {
        email:    { label: "Email", type: "email" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        const parsed = z.object({
          email:    z.string().email(),
          password: z.string().min(8),
        }).safeParse(credentials)

        if (!parsed.success) return null

        const user = await prisma.user.findUnique({
          where: { email: parsed.data.email },
        })
        if (!user?.hashedPassword) return null

        const isValid = await bcrypt.compare(parsed.data.password, user.hashedPassword)
        if (!isValid) return null

        return { id: user.id, email: user.email, name: user.name }
      },
    }),
  ],
  callbacks: {
    async jwt({ token, user, account }) {
      if (user) {
        token.id   = user.id
        token.role = user.role    // custom field — add to User type
      }
      return token
    },
    async session({ session, token }) {
      if (token && session.user) {
        session.user.id   = token.id as string
        session.user.role = token.role as string
      }
      return session
    },
    async signIn({ user, account, profile }) {
      // Return false to block sign-in
      if (!user.email) return false
      return true
    },
    async redirect({ url, baseUrl }) {
      // Allow relative URLs and same-origin absolute URLs
      if (url.startsWith("/")) return `${baseUrl}${url}`
      if (new URL(url).origin === baseUrl) return url
      return baseUrl
    },
  },
})
```

## TypeScript Augmentation
```ts
// types/next-auth.d.ts
import "next-auth"

declare module "next-auth" {
  interface Session {
    user: {
      id:    string
      role:  string
      email: string
      name:  string | null
      image: string | null
    }
  }
  interface User {
    role: string
  }
}

declare module "next-auth/jwt" {
  interface JWT {
    id:   string
    role: string
  }
}
```

## Route Handler Setup
```ts
// app/api/auth/[...nextauth]/route.ts
import { handlers } from "@/auth"

export const { GET, POST } = handlers
```

## Server Component — auth()
```ts
// app/dashboard/page.tsx
import { auth } from "@/auth"
import { redirect } from "next/navigation"

export default async function DashboardPage() {
  const session = await auth()

  if (!session) redirect("/login")

  return (
    <div>
      <p>Welcome, {session.user.name}</p>
      <p>Role: {session.user.role}</p>
    </div>
  )
}
```

## Server Action — signIn / signOut
```ts
// app/login/actions.ts
"use server"
import { signIn, signOut } from "@/auth"
import { AuthError } from "next-auth"

export async function loginWithGitHub() {
  await signIn("github", { redirectTo: "/dashboard" })
}

export async function loginWithCredentials(formData: FormData) {
  try {
    await signIn("credentials", {
      email:       formData.get("email"),
      password:    formData.get("password"),
      redirectTo:  "/dashboard",
    })
  } catch (error) {
    if (error instanceof AuthError) {
      switch (error.type) {
        case "CredentialsSignin": return { error: "Invalid credentials" }
        default:                  return { error: "Something went wrong" }
      }
    }
    throw error   // re-throw redirect errors
  }
}

export async function logout() {
  await signOut({ redirectTo: "/" })
}
```

## Client Component — signIn / signOut
```tsx
"use client"
import { signIn, signOut, useSession } from "next-auth/react"

export function LoginButton() {
  const { data: session, status } = useSession()

  if (status === "loading") return <p>Loading...</p>

  if (session) {
    return (
      <div>
        <p>Signed in as {session.user.email}</p>
        <button onClick={() => signOut({ callbackUrl: "/" })}>Sign out</button>
      </div>
    )
  }

  return (
    <div>
      <button onClick={() => signIn("github")}>Sign in with GitHub</button>
      <button onClick={() => signIn("google", { callbackUrl: "/dashboard" })}>
        Sign in with Google
      </button>
    </div>
  )
}
```

## SessionProvider (App Router)
```tsx
// app/providers.tsx
"use client"
import { SessionProvider } from "next-auth/react"

export function Providers({ children }: { children: React.ReactNode }) {
  return <SessionProvider>{children}</SessionProvider>
}

// app/layout.tsx
import { Providers } from "./providers"

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html>
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  )
}
```

## Middleware Protection
```ts
// middleware.ts (project root)
import { auth } from "@/auth"
import { NextResponse } from "next/server"

export default auth((req) => {
  const { nextUrl, auth: session } = req

  const isLoggedIn    = !!session
  const isAuthPage    = nextUrl.pathname.startsWith("/login")
  const isProtected   = nextUrl.pathname.startsWith("/dashboard")
  const isApiProtected = nextUrl.pathname.startsWith("/api/protected")

  if (isAuthPage && isLoggedIn) {
    return NextResponse.redirect(new URL("/dashboard", req.url))
  }

  if ((isProtected || isApiProtected) && !isLoggedIn) {
    const loginUrl = new URL("/login", req.url)
    loginUrl.searchParams.set("callbackUrl", nextUrl.pathname)
    return NextResponse.redirect(loginUrl)
  }

  return NextResponse.next()
})

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
}
```

## Environment Variables
```env
AUTH_SECRET=your-secret-here          # npx auth secret
AUTH_URL=http://localhost:3000         # production: your domain

GITHUB_ID=your-github-client-id
GITHUB_SECRET=your-github-client-secret

GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
```
