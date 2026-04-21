# Drizzle ORM Reference

## Installation
```bash
npm install drizzle-orm
npm install drizzle-kit -D
# Driver: pg | postgres-js | @neondatabase/serverless | better-sqlite3
npm install postgres
```

## Schema (pgTable)
```ts
// db/schema.ts
import { pgTable, text, varchar, integer, boolean, timestamp, pgEnum, serial, uuid, jsonb, index, uniqueIndex } from "drizzle-orm/pg-core"
import { relations } from "drizzle-orm"

export const roleEnum = pgEnum("role", ["user", "admin", "moderator"])

export const users = pgTable("users", {
  id:        uuid("id").primaryKey().defaultRandom(),
  email:     varchar("email", { length: 255 }).notNull().unique(),
  name:      text("name"),
  role:      roleEnum("role").default("user").notNull(),
  age:       integer("age"),
  active:    boolean("active").default(true).notNull(),
  metadata:  jsonb("metadata").$type<{ plan: string; features: string[] }>(),
  createdAt: timestamp("created_at").defaultNow().notNull(),
  updatedAt: timestamp("updated_at").defaultNow().$onUpdate(() => new Date()),
}, (table) => ({
  emailIdx:   index("email_idx").on(table.email),
  activeIdx:  uniqueIndex("active_email_idx").on(table.email).where(sql`${table.active} = true`),
}))

export const posts = pgTable("posts", {
  id:        serial("id").primaryKey(),
  title:     text("title").notNull(),
  content:   text("content"),
  published: boolean("published").default(false),
  authorId:  uuid("author_id").notNull().references(() => users.id, { onDelete: "cascade" }),
  createdAt: timestamp("created_at").defaultNow().notNull(),
})

// Relations (for query builder joins)
export const usersRelations = relations(users, ({ many }) => ({
  posts: many(posts),
}))

export const postsRelations = relations(posts, ({ one }) => ({
  author: one(users, { fields: [posts.authorId], references: [users.id] }),
}))
```

## Client Setup
```ts
// db/index.ts
import { drizzle } from "drizzle-orm/postgres-js"
import postgres from "postgres"
import * as schema from "./schema"

const client = postgres(process.env.DATABASE_URL!)
export const db = drizzle(client, { schema })

// Neon serverless
import { neon } from "@neondatabase/serverless"
import { drizzle } from "drizzle-orm/neon-http"
const sql = neon(process.env.DATABASE_URL!)
export const db = drizzle(sql, { schema })
```

## Queries

### Select
```ts
import { eq, and, or, ne, gt, gte, lt, lte, like, ilike, inArray, isNull, isNotNull, desc, asc, count, sql } from "drizzle-orm"

// Basic
const allUsers = await db.select().from(users)

// With columns
const userNames = await db.select({ id: users.id, name: users.name }).from(users)

// With filters
const activeAdmins = await db
  .select()
  .from(users)
  .where(and(eq(users.active, true), eq(users.role, "admin")))
  .orderBy(desc(users.createdAt), asc(users.name))
  .limit(10)
  .offset(20)

// OR / complex conditions
const result = await db
  .select()
  .from(users)
  .where(or(eq(users.role, "admin"), gt(users.age, 30)))

// Pattern matching
const found = await db.select().from(users).where(ilike(users.email, "%@gmail%"))

// inArray
const selected = await db.select().from(users).where(inArray(users.id, ["id1", "id2"]))

// Aggregates
const [{ total }] = await db.select({ total: count() }).from(users)
```

### Joins
```ts
// Inner join
const postsWithAuthor = await db
  .select({
    postId:     posts.id,
    title:      posts.title,
    authorName: users.name,
  })
  .from(posts)
  .innerJoin(users, eq(posts.authorId, users.id))
  .where(eq(posts.published, true))

// Left join
const allUsersWithPosts = await db
  .select()
  .from(users)
  .leftJoin(posts, eq(users.id, posts.authorId))

// Relations query (relational API — requires relations defined)
const usersWithPosts = await db.query.users.findMany({
  with: {
    posts: {
      where: eq(posts.published, true),
      orderBy: desc(posts.createdAt),
      limit: 5,
    },
  },
  where: eq(users.active, true),
})
```

### Insert
```ts
// Single
const [newUser] = await db
  .insert(users)
  .values({ email: "user@example.com", name: "Alice" })
  .returning()

// Multiple
await db.insert(posts).values([
  { title: "Post 1", authorId: "user-id" },
  { title: "Post 2", authorId: "user-id" },
])

// Upsert
await db
  .insert(users)
  .values({ email: "user@example.com", name: "Alice" })
  .onConflictDoUpdate({
    target: users.email,
    set: { name: "Alice Updated", updatedAt: new Date() },
  })
```

### Update
```ts
const [updated] = await db
  .update(users)
  .set({ name: "Bob", updatedAt: new Date() })
  .where(eq(users.id, "user-id"))
  .returning()

// Increment
await db
  .update(users)
  .set({ age: sql`${users.age} + 1` })
  .where(eq(users.id, "user-id"))
```

### Delete
```ts
await db.delete(users).where(eq(users.id, "user-id"))

const deleted = await db
  .delete(posts)
  .where(and(eq(posts.authorId, "user-id"), eq(posts.published, false)))
  .returning()
```

### Transactions
```ts
const result = await db.transaction(async (tx) => {
  const [user] = await tx.insert(users).values({ email: "a@b.com" }).returning()
  await tx.insert(posts).values({ title: "First", authorId: user.id })
  return user
})
```

## drizzle.config.ts
```ts
// drizzle.config.ts
import { defineConfig } from "drizzle-kit"

export default defineConfig({
  schema:    "./db/schema.ts",
  out:       "./drizzle",
  dialect:   "postgresql",
  dbCredentials: {
    url: process.env.DATABASE_URL!,
  },
  verbose:   true,
  strict:    true,
})
```

## Migration Commands
```bash
# Generate SQL migration files from schema changes
npx drizzle-kit generate

# Apply migrations to the database
npx drizzle-kit migrate

# Push schema directly (no migration files, for prototyping)
npx drizzle-kit push

# Pull schema from existing database
npx drizzle-kit pull

# Open Drizzle Studio (visual browser)
npx drizzle-kit studio

# Check for schema drift
npx drizzle-kit check
```
