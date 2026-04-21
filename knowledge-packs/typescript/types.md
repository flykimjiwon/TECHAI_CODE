# TypeScript Type System Reference

## Basic Types
```typescript
let str: string = "hello"
let num: number = 42
let bool: boolean = true
let arr: number[] = [1, 2, 3]
let tuple: [string, number] = ["age", 30]
let obj: { name: string; age: number } = { name: "John", age: 30 }
let func: (x: number) => string = (x) => x.toString()
let any: any = "anything" // 피할 것
let unknown: unknown = "safe any" // 타입 체크 후 사용
let never: never // 절대 발생하지 않는 타입
let nullable: string | null = null
let optional: string | undefined
```

## Union & Intersection
```typescript
// Union: OR
type ID = string | number
let id: ID = "abc123" // or 123

// Intersection: AND
type User = { name: string }
type Admin = { role: string }
type AdminUser = User & Admin
let admin: AdminUser = { name: "John", role: "admin" }
```

## Generics
```typescript
function identity<T>(arg: T): T {
  return arg
}

const result = identity<string>("hello")

interface Box<T> {
  value: T
}

const box: Box<number> = { value: 42 }
```

## Type Guards
```typescript
function isString(value: unknown): value is string {
  return typeof value === "string"
}

if (isString(value)) {
  value.toUpperCase() // OK: value는 string
}

// typeof guard
if (typeof value === "string") { }

// instanceof guard
if (value instanceof Date) { }
```

## Utility Types

### Partial<T>
```typescript
type User = { name: string; age: number }
type PartialUser = Partial<User> // { name?: string; age?: number }
```

### Pick<T, K>
```typescript
type UserPreview = Pick<User, "name"> // { name: string }
```

### Omit<T, K>
```typescript
type UserWithoutAge = Omit<User, "age"> // { name: string }
```

### Record<K, V>
```typescript
type Roles = Record<string, boolean>
const roles: Roles = { admin: true, user: false }
```

### Required<T>
```typescript
type RequiredUser = Required<PartialUser> // { name: string; age: number }
```

### Readonly<T>
```typescript
type ReadonlyUser = Readonly<User>
const user: ReadonlyUser = { name: "John", age: 30 }
// user.name = "Jane" // Error
```

### ReturnType<T>
```typescript
function getUser() { return { name: "John", age: 30 } }
type User = ReturnType<typeof getUser> // { name: string; age: number }
```

### Parameters<T>
```typescript
function createUser(name: string, age: number) { }
type Params = Parameters<typeof createUser> // [string, number]
```

### Awaited<T>
```typescript
type Response = Promise<{ data: string }>
type Data = Awaited<Response> // { data: string }
```

## 자주 쓰는 패턴

### Type Assertion
```typescript
const input = document.getElementById("input") as HTMLInputElement
const value = someValue as string
```

### Interface vs Type
```typescript
// Interface (확장 가능)
interface User {
  name: string
}
interface User {
  age: number
}

// Type (union/intersection 가능)
type ID = string | number
type Point = { x: number } & { y: number }
```

### Conditional Types
```typescript
type IsString<T> = T extends string ? true : false
type A = IsString<string> // true
type B = IsString<number> // false
```

### Mapped Types
```typescript
type Optional<T> = {
  [K in keyof T]?: T[K]
}
```
