# Zod Validation Reference

## Installation
```bash
npm install zod
```

## Core Types

### string
```ts
import { z } from "zod"

z.string()
z.string().min(2).max(100)
z.string().length(10)
z.string().email()
z.string().url()
z.string().uuid()
z.string().cuid()
z.string().regex(/^[a-z]+$/)
z.string().startsWith("prefix")
z.string().endsWith(".com")
z.string().includes("@")
z.string().trim()
z.string().toLowerCase()
z.string().toUpperCase()
z.string().datetime()       // ISO 8601 datetime string
z.string().ip()             // IPv4 or IPv6
z.string().emoji()
```

### number
```ts
z.number()
z.number().int()
z.number().positive()       // > 0
z.number().nonnegative()    // >= 0
z.number().negative()       // < 0
z.number().nonpositive()    // <= 0
z.number().min(0).max(100)
z.number().gt(0).lt(100)    // exclusive
z.number().gte(0).lte(100)  // inclusive (alias: min/max)
z.number().multipleOf(5)
z.number().finite()
z.number().safe()           // Number.MIN_SAFE_INTEGER to MAX
```

### boolean & date
```ts
z.boolean()
z.date()
z.date().min(new Date("2020-01-01"))
z.date().max(new Date())
z.literal(true)
z.literal("active")
z.literal(42)
```

## Objects & Arrays
```ts
// Object
const UserSchema = z.object({
  id:    z.string().uuid(),
  email: z.string().email(),
  name:  z.string().min(1),
  age:   z.number().int().min(0).max(150).optional(),
})

// Partial / Required
UserSchema.partial()                   // all fields optional
UserSchema.partial({ age: true })      // only age optional
UserSchema.required()                  // all fields required
UserSchema.pick({ id: true, email: true })
UserSchema.omit({ age: true })
UserSchema.extend({ phone: z.string() })
UserSchema.merge(AnotherSchema)

// Strip / Passthrough / Strict
z.object({ name: z.string() }).strip()       // default: remove unknown keys
z.object({ name: z.string() }).passthrough() // keep unknown keys
z.object({ name: z.string() }).strict()      // error on unknown keys

// Array
z.array(z.string())
z.array(z.number()).min(1).max(10)
z.array(z.string()).nonempty()
z.array(UserSchema)

// Tuple
z.tuple([z.string(), z.number(), z.boolean()])
z.tuple([z.string()]).rest(z.number())  // [...rest] items are numbers

// Record
z.record(z.string())                   // Record<string, string>
z.record(z.string(), z.number())       // Record<string, number>
```

## Optional / Nullable / Nullish / Default
```ts
z.string().optional()          // string | undefined
z.string().nullable()          // string | null
z.string().nullish()           // string | null | undefined
z.string().default("fallback") // undefined → "fallback"
z.string().optional().default("") // chaining

// In objects
z.object({
  name:     z.string(),
  nickname: z.string().optional(),    // name?: string
  bio:      z.string().nullable(),    // bio: string | null
  avatar:   z.string().nullish(),     // avatar?: string | null
  role:     z.string().default("user"),
})
```

## z.enum
```ts
const StatusSchema = z.enum(["active", "inactive", "pending"])
type Status = z.infer<typeof StatusSchema>  // "active" | "inactive" | "pending"

StatusSchema.options   // ["active", "inactive", "pending"]
StatusSchema.enum.active  // "active"

// From const array
const ROLES = ["user", "admin", "mod"] as const
const RoleSchema = z.enum(ROLES)

// Native enum
enum Direction { Up = "UP", Down = "DOWN" }
z.nativeEnum(Direction)
```

## z.union & z.discriminatedUnion
```ts
// Union
const StringOrNumber = z.union([z.string(), z.number()])
// shorthand:
z.string().or(z.number())

// Discriminated union (faster, better errors)
const ShapeSchema = z.discriminatedUnion("type", [
  z.object({ type: z.literal("circle"),    radius: z.number() }),
  z.object({ type: z.literal("square"),    side: z.number() }),
  z.object({ type: z.literal("rectangle"), width: z.number(), height: z.number() }),
])
```

## Type Inference
```ts
const PostSchema = z.object({
  id:      z.string().uuid(),
  title:   z.string().min(1).max(200),
  content: z.string().optional(),
  tags:    z.array(z.string()),
})

type Post = z.infer<typeof PostSchema>
// { id: string; title: string; content?: string; tags: string[] }

// Input vs Output types (when using .transform or .default)
type PostInput  = z.input<typeof PostSchema>
type PostOutput = z.output<typeof PostSchema>
```

## parse vs safeParse
```ts
// parse — throws ZodError on failure
const user = UserSchema.parse(rawData)

// safeParse — returns result object, never throws
const result = UserSchema.safeParse(rawData)
if (result.success) {
  console.log(result.data)      // typed
} else {
  console.log(result.error.issues)   // ZodIssue[]
  console.log(result.error.format()) // nested error object
  console.log(result.error.flatten()) // { formErrors, fieldErrors }
}

// parseAsync / safeParseAsync (for async refinements)
const result = await UserSchema.safeParseAsync(rawData)
```

## Refinement (.refine)
```ts
const PasswordSchema = z.string()
  .min(8)
  .refine((val) => /[A-Z]/.test(val), { message: "Must contain uppercase" })
  .refine((val) => /[0-9]/.test(val), { message: "Must contain number" })

// Cross-field refinement
const SignupSchema = z.object({
  password:        z.string().min(8),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords do not match",
  path: ["confirmPassword"],  // which field the error is attached to
})

// superRefine (multiple issues)
z.string().superRefine((val, ctx) => {
  if (val.length < 8) {
    ctx.addIssue({ code: z.ZodIssueCode.too_small, minimum: 8, type: "string", inclusive: true, message: "Too short" })
  }
  if (!/[A-Z]/.test(val)) {
    ctx.addIssue({ code: z.ZodIssueCode.custom, message: "No uppercase" })
  }
})
```

## Transform (.transform)
```ts
// Transform after validation
const TrimmedString = z.string().trim().transform((val) => val.toLowerCase())
type Trimmed = z.infer<typeof TrimmedString>   // string (output)

// String to number
const NumberFromString = z.string().transform((val) => parseInt(val, 10))

// Date transform
const DateSchema = z.string().datetime().transform((val) => new Date(val))

// Chaining pipe
const LimitedString = z.string().transform((s) => s.slice(0, 100)).pipe(z.string().min(1))
```

## z.preprocess
```ts
// Runs BEFORE validation — coerces unknown input
const NumberSchema = z.preprocess(
  (val) => (typeof val === "string" ? parseInt(val, 10) : val),
  z.number()
)

NumberSchema.parse("42")    // 42
NumberSchema.parse(42)      // 42

// Coerce shorthand (v3.20+)
z.coerce.number()           // coerces string/boolean/Date to number
z.coerce.string()           // coerces to string
z.coerce.boolean()          // coerces truthy/falsy
z.coerce.date()             // coerces string/number to Date
```

## 자주 쓰는 패턴

### API Request Validation
```ts
// app/api/posts/route.ts
const CreatePostSchema = z.object({
  title:   z.string().min(1).max(200),
  content: z.string().optional(),
  tags:    z.array(z.string()).max(10).default([]),
})

export async function POST(req: Request) {
  const body = await req.json()
  const result = CreatePostSchema.safeParse(body)
  if (!result.success) {
    return Response.json({ errors: result.error.flatten().fieldErrors }, { status: 400 })
  }
  const { title, content, tags } = result.data
  // ...
}
```

### React Hook Form Integration
```ts
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"

const schema = z.object({ email: z.string().email(), password: z.string().min(8) })
type FormData = z.infer<typeof schema>

const form = useForm<FormData>({ resolver: zodResolver(schema) })
```
