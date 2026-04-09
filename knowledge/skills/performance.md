# Performance Optimization

## Web Performance - Core Web Vitals

**LCP (Largest Contentful Paint)**: <2.5s
- Optimize images (WebP, compression)
- Preload critical resources
- Server-side rendering (SSR)

**FID (First Input Delay)**: <100ms
- Code splitting (defer non-critical JS)
- Remove blocking scripts
- Web Workers for heavy computation

**CLS (Cumulative Layout Shift)**: <0.1
- Set image width/height
- Reserve space for dynamic content
- Avoid inserting content above existing

## Image Optimization

```html
<!-- Responsive images -->
<img
  src="image.webp"
  srcset="image-320.webp 320w, image-640.webp 640w, image-1280.webp 1280w"
  sizes="(max-width: 640px) 100vw, 640px"
  loading="lazy"
  width="640"
  height="480"
/>
```

**Format**: WebP > JPEG > PNG. Compression 80% quality.

## Bundle Splitting

```js
// Dynamic import (code splitting)
const Chart = lazy(() => import('./Chart'))

// Route-based splitting
const Dashboard = lazy(() => import('./pages/Dashboard'))
```

## Server Performance

### N+1 Query Problem

```python
# ❌ Bad (N+1)
users = User.query.all()
for user in users:
    print(user.orders)  # separate query per user

# ✅ Good (batch)
users = User.query.options(joinedload(User.orders)).all()
```

```sql
-- ❌ N+1
SELECT * FROM users;
SELECT * FROM orders WHERE user_id = 1;
SELECT * FROM orders WHERE user_id = 2;
-- ...

-- ✅ JOIN
SELECT * FROM users LEFT JOIN orders ON users.id = orders.user_id;
```

### DB Indexing

```sql
-- Add index on WHERE/JOIN columns
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_orders_user_id ON orders(user_id);

-- Composite index for multi-column WHERE
CREATE INDEX idx_orders_status_date ON orders(status, created_at);
```

### Connection Pooling

```python
# ❌ Bad (new connection per request)
conn = psycopg2.connect(...)

# ✅ Good (connection pool)
pool = psycopg2.pool.SimpleConnectionPool(1, 20, ...)
conn = pool.getconn()
```

## Caching

```python
# Redis cache
@cache.memoize(timeout=300)
def get_user_profile(user_id):
    return db.query(User).get(user_id)

# In-memory cache
from functools import lru_cache

@lru_cache(maxsize=128)
def expensive_calculation(n):
    return n ** 2
```

## Pagination

```sql
-- ✅ Offset pagination
SELECT * FROM posts ORDER BY created_at DESC LIMIT 20 OFFSET 0;

-- ✅ Cursor pagination (better for large datasets)
SELECT * FROM posts WHERE id > 1000 ORDER BY id LIMIT 20;
```

**원칙**: Measure first. Profile before optimizing. 80/20 rule (20% code → 80% performance impact).
