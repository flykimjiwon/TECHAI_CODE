# Prisma ORM Reference

## Installation
```bash
npm install prisma @prisma/client
npx prisma init
```

## schema.prisma
```prisma
datasource db {
  provider = "postgresql"   // "mysql" | "sqlite" | "mongodb"
  url      = env("DATABASE_URL")
}

generator client {
  provider = "prisma-client-js"
}

enum Role {
  USER
  ADMIN
  MODERATOR
}

model User {
  id        String    @id @default(cuid())
  email     String    @unique
  name      String?
  role      Role      @default(USER)
  createdAt DateTime  @default(now())
  updatedAt DateTime  @updatedAt

  posts     Post[]
  profile   Profile?

  @@index([email])
  @@map("users")          // custom table name
}

model Profile {
  id     String @id @default(cuid())
  bio    String?
  userId String @unique

  user   User   @relation(fields: [userId], references: [id], onDelete: Cascade)
}

model Post {
  id          String    @id @default(cuid())
  title       String
  content     String?
  published   Boolean   @default(false)
  authorId    String
  tags        String[]  // PostgreSQL array
  metadata    Json?

  author      User      @relation(fields: [authorId], references: [id])
  categories  PostCategory[]

  @@index([authorId, published])
}

model PostCategory {
  postId     String
  categoryId String

  post       Post     @relation(fields: [postId], references: [id])
  category   Category @relation(fields: [categoryId], references: [id])

  @@id([postId, categoryId])   // composite primary key
}
```

## Client Setup
```ts
// lib/prisma.ts
import { PrismaClient } from "@prisma/client"

const globalForPrisma = globalThis as unknown as { prisma: PrismaClient }

export const prisma =
  globalForPrisma.prisma ?? new PrismaClient({ log: ["query"] })

if (process.env.NODE_ENV !== "production") globalForPrisma.prisma = prisma
```

## CRUD Operations

### findMany
```ts
// Basic
const users = await prisma.user.findMany()

// With filters
const posts = await prisma.post.findMany({
  where: {
    published: true,
    author: { role: "ADMIN" },
    title: { contains: "Next.js", mode: "insensitive" },
    createdAt: { gte: new Date("2024-01-01") },
    OR: [{ title: { startsWith: "How" } }, { title: { startsWith: "Why" } }],
  },
  orderBy: [{ createdAt: "desc" }, { title: "asc" }],
  select: { id: true, title: true, author: { select: { name: true } } },
  skip: 0,
  take: 10,
})

// With relations
const usersWithPosts = await prisma.user.findMany({
  include: {
    posts: { where: { published: true }, orderBy: { createdAt: "desc" } },
    profile: true,
  },
})
```

### findUnique / findFirst
```ts
const user = await prisma.user.findUnique({
  where: { id: "user-id" },
  include: { posts: true },
})

const post = await prisma.post.findFirst({
  where: { title: { contains: "prisma" } },
  orderBy: { createdAt: "desc" },
})
```

### create
```ts
const user = await prisma.user.create({
  data: {
    email: "user@example.com",
    name: "Alice",
    posts: {
      create: [
        { title: "First post", published: true },
        { title: "Draft" },
      ],
    },
    profile: {
      create: { bio: "Developer" },
    },
  },
  include: { posts: true },
})

// createMany
await prisma.post.createMany({
  data: [{ title: "A", authorId: "id1" }, { title: "B", authorId: "id2" }],
  skipDuplicates: true,
})
```

### update
```ts
const updated = await prisma.user.update({
  where: { id: "user-id" },
  data: {
    name: "Bob",
    posts: {
      connect: { id: "post-id" },           // add existing relation
      disconnect: { id: "other-post-id" },  // remove relation
      update: { where: { id: "post-id" }, data: { published: true } },
      deleteMany: { published: false },
    },
  },
})

// updateMany
await prisma.post.updateMany({
  where: { authorId: "user-id" },
  data: { published: false },
})
```

### delete
```ts
await prisma.user.delete({ where: { id: "user-id" } })

await prisma.post.deleteMany({ where: { published: false } })
```

### upsert
```ts
const user = await prisma.user.upsert({
  where: { email: "user@example.com" },
  update: { name: "Updated Name" },
  create: { email: "user@example.com", name: "New User" },
})
```

### Transactions
```ts
// Sequential
const [post, user] = await prisma.$transaction([
  prisma.post.create({ data: { title: "Post", authorId: "id" } }),
  prisma.user.update({ where: { id: "id" }, data: { name: "Updated" } }),
])

// Interactive (with logic)
await prisma.$transaction(async (tx) => {
  const user = await tx.user.findUnique({ where: { id: "id" } })
  if (!user) throw new Error("User not found")
  return tx.post.create({ data: { title: "Post", authorId: user.id } })
})
```

## Migrations
```bash
# Development: create + apply migration
npx prisma migrate dev --name add_user_role

# Production: apply pending migrations
npx prisma migrate deploy

# Generate client after schema change
npx prisma generate

# Push schema without migration file (prototyping)
npx prisma db push

# Pull schema from existing database
npx prisma db pull

# Open Prisma Studio (visual DB browser)
npx prisma studio
```

## Filtering Reference
```ts
where: {
  name: { equals: "Alice" },
  name: { not: "Bob" },
  age:  { gt: 18, lt: 65 },
  age:  { gte: 18, lte: 64 },
  name: { in: ["Alice", "Bob"] },
  name: { notIn: ["Charlie"] },
  email: { contains: "@gmail" },
  email: { startsWith: "alice" },
  email: { endsWith: ".com" },
  name: { not: null },          // IS NOT NULL
  name: null,                   // IS NULL
  AND: [...],
  OR: [...],
  NOT: [...],
}
```
