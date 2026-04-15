# Supabase JS Client Reference

## Installation & Setup
```bash
npm install @supabase/supabase-js
```
```ts
// lib/supabase.ts
import { createClient } from "@supabase/supabase-js"
import type { Database } from "./database.types"   // generated types

export const supabase = createClient<Database>(
  process.env.NEXT_PUBLIC_SUPABASE_URL!,
  process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
)

// Server-side with service role (bypasses RLS)
export const supabaseAdmin = createClient<Database>(
  process.env.SUPABASE_URL!,
  process.env.SUPABASE_SERVICE_ROLE_KEY!
)

// Next.js App Router (server component / route handler)
import { createServerClient } from "@supabase/ssr"
import { cookies } from "next/headers"

export async function createServerSupabase() {
  const cookieStore = await cookies()
  return createServerClient<Database>(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    { cookies: { getAll: () => cookieStore.getAll(), setAll: (c) => c.forEach(({ name, value, options }) => cookieStore.set(name, value, options)) } }
  )
}
```

## Auth

### Sign Up / Sign In
```ts
// Email + Password
const { data, error } = await supabase.auth.signUp({
  email: "user@example.com",
  password: "securepassword",
  options: { data: { full_name: "Alice" } },   // custom metadata
})

const { data, error } = await supabase.auth.signInWithPassword({
  email: "user@example.com",
  password: "securepassword",
})

// OAuth
const { data, error } = await supabase.auth.signInWithOAuth({
  provider: "github",   // "google" | "kakao" | "apple" | etc.
  options: { redirectTo: `${location.origin}/auth/callback` },
})

// OTP (magic link)
await supabase.auth.signInWithOtp({ email: "user@example.com" })
```

### Session & User
```ts
const { data: { session } } = await supabase.auth.getSession()
const { data: { user } } = await supabase.auth.getUser()   // verifies JWT server-side

// Listen to auth state changes
const { data: { subscription } } = supabase.auth.onAuthStateChange((event, session) => {
  // event: "SIGNED_IN" | "SIGNED_OUT" | "TOKEN_REFRESHED" | "USER_UPDATED"
  if (event === "SIGNED_IN") console.log(session?.user.id)
})

// Cleanup
subscription.unsubscribe()

// Sign out
await supabase.auth.signOut()

// Update user
await supabase.auth.updateUser({ password: "newpassword", data: { name: "Bob" } })
```

## Database

### Select
```ts
// Basic
const { data, error } = await supabase.from("posts").select("*")

// Specific columns + relations
const { data } = await supabase
  .from("posts")
  .select(`
    id,
    title,
    created_at,
    author:users ( id, name, avatar_url ),
    categories ( name )
  `)

// Filters
const { data } = await supabase
  .from("posts")
  .select("*")
  .eq("published", true)          // =
  .neq("status", "deleted")       // !=
  .gt("views", 100)               // >
  .gte("rating", 4)               // >=
  .lt("price", 50)                // <
  .lte("age", 65)                 // <=
  .like("title", "%Next%")        // LIKE (case-sensitive)
  .ilike("title", "%next%")       // ILIKE (case-insensitive)
  .in("category", ["tech", "ai"]) // IN
  .is("deleted_at", null)         // IS NULL
  .not("deleted_at", "is", null)  // IS NOT NULL
  .contains("tags", ["react"])    // array contains
  .order("created_at", { ascending: false })
  .range(0, 9)                    // pagination (offset)
  .limit(10)
  .single()                       // expect exactly one row — throws if 0 or 2+
  .maybeSingle()                  // expect 0 or 1 row — returns null if 0
```

### Insert
```ts
const { data, error } = await supabase
  .from("posts")
  .insert({ title: "Hello", content: "World", author_id: user.id })
  .select()
  .single()

// Multiple rows
await supabase.from("tags").insert([
  { name: "react" },
  { name: "typescript" },
])
```

### Update
```ts
const { data, error } = await supabase
  .from("posts")
  .update({ title: "Updated", updated_at: new Date().toISOString() })
  .eq("id", postId)
  .select()
  .single()
```

### Upsert
```ts
await supabase
  .from("profiles")
  .upsert({ id: user.id, username: "alice", updated_at: new Date().toISOString() })
  .eq("id", user.id)
```

### Delete
```ts
await supabase.from("posts").delete().eq("id", postId)
```

## Realtime
```ts
// Listen to table changes
const channel = supabase
  .channel("posts-changes")
  .on(
    "postgres_changes",
    {
      event: "*",             // "INSERT" | "UPDATE" | "DELETE" | "*"
      schema: "public",
      table: "posts",
      filter: "author_id=eq.user-123",   // optional row-level filter
    },
    (payload) => {
      console.log(payload.eventType, payload.new, payload.old)
    }
  )
  .subscribe((status) => {
    console.log("Realtime status:", status)
  })

// Presence (online users)
const presenceChannel = supabase.channel("room-1")
presenceChannel
  .on("presence", { event: "sync" }, () => {
    const state = presenceChannel.presenceState()
    console.log("Online users:", state)
  })
  .subscribe(async (status) => {
    if (status === "SUBSCRIBED") {
      await presenceChannel.track({ user_id: user.id, online_at: new Date() })
    }
  })

// Cleanup
await supabase.removeChannel(channel)
```

## Storage
```ts
// Upload
const { data, error } = await supabase.storage
  .from("avatars")
  .upload(`${user.id}/avatar.png`, file, {
    cacheControl: "3600",
    upsert: true,
    contentType: "image/png",
  })

// Download
const { data } = await supabase.storage.from("avatars").download("path/to/file.png")

// Public URL
const { data: { publicUrl } } = supabase.storage
  .from("avatars")
  .getPublicUrl(`${user.id}/avatar.png`)

// Signed URL (private bucket)
const { data: { signedUrl } } = await supabase.storage
  .from("private-docs")
  .createSignedUrl("report.pdf", 3600)   // expires in 1 hour

// Delete
await supabase.storage.from("avatars").remove([`${user.id}/avatar.png`])

// List files
const { data: files } = await supabase.storage.from("avatars").list(user.id)
```

## Type Generation
```bash
# Generate TypeScript types from your database schema
npx supabase gen types typescript --project-id your-project-id > lib/database.types.ts
```
