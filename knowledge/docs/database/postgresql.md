# PostgreSQL Reference

## 데이터 타입

```sql
-- 정수 / 시퀀스
SMALLINT, INTEGER, BIGINT
SMALLSERIAL, SERIAL, BIGSERIAL  -- 자동 증가 (내부적으로 시퀀스)

-- 숫자
NUMERIC(p, s) / DECIMAL(p, s)   -- 정밀도 보장 (금액)
REAL, DOUBLE PRECISION           -- 부동소수 (근사값)

-- 문자열
VARCHAR(n), CHAR(n), TEXT        -- TEXT: 길이 제한 없음, VARCHAR와 성능 차이 없음

-- UUID
UUID  -- 예: 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
-- gen_random_uuid() (pgcrypto 또는 PG 13+)

-- 날짜/시간
DATE, TIME, TIMESTAMP
TIMESTAMP WITH TIME ZONE  -- 권장: 타임존 저장 (UTC 기준)
INTERVAL                  -- 예: INTERVAL '1 year 2 months 3 days'

-- JSON
JSON   -- 텍스트 저장, 매번 파싱
JSONB  -- 바이너리 저장, GIN 인덱스 지원 → 쿼리 성능 우수 (권장)

-- 배열
INTEGER[], TEXT[], JSONB[]

-- 네트워크
INET (IP 주소), CIDR (네트워크), MACADDR

-- 불리언
BOOLEAN  -- TRUE, FALSE, NULL

-- 기타
BYTEA  -- 이진 데이터
```

---

## JSONB 연산

```sql
-- 데이터 예시
CREATE TABLE events (
    id      BIGSERIAL PRIMARY KEY,
    payload JSONB NOT NULL
);
INSERT INTO events (payload) VALUES
('{"user_id": 1, "action": "login", "meta": {"ip": "1.2.3.4", "ua": "Chrome"}}');

-- -> : JSON 객체/배열 반환 (JSON 타입)
SELECT payload -> 'user_id' FROM events;           -- 결과: 1 (JSON)

-- ->> : 텍스트 반환
SELECT payload ->> 'action' FROM events;           -- 결과: login (TEXT)

-- 중첩
SELECT payload -> 'meta' ->> 'ip' FROM events;    -- 결과: 1.2.3.4

-- #> / #>> : 경로 배열로 접근
SELECT payload #>> '{meta,ip}' FROM events;       -- 결과: 1.2.3.4

-- @> : 포함 여부 (containment)
SELECT * FROM events WHERE payload @> '{"action": "login"}';

-- ? : 키 존재 여부
SELECT * FROM events WHERE payload ? 'user_id';

-- jsonb_build_object: 동적 객체 생성
SELECT jsonb_build_object(
    'id', id,
    'action', payload ->> 'action'
) FROM events;

-- jsonb_agg: 집계
SELECT jsonb_agg(payload) FROM events WHERE payload ->> 'action' = 'login';

-- jsonb_array_elements: 배열 펼치기
SELECT jsonb_array_elements(payload -> 'tags') AS tag
FROM articles WHERE payload ? 'tags';

-- JSONB 업데이트
UPDATE events
SET payload = payload || '{"processed": true}'  -- 키 추가/덮어쓰기
WHERE id = 1;

UPDATE events
SET payload = payload - 'action'  -- 키 삭제
WHERE id = 1;

-- GIN 인덱스 (JSONB 쿼리 최적화)
CREATE INDEX idx_events_payload ON events USING GIN (payload);
CREATE INDEX idx_events_action ON events USING GIN ((payload -> 'action'));
```

---

## 배열

```sql
-- 배열 컬럼
CREATE TABLE posts (
    id   BIGSERIAL PRIMARY KEY,
    tags TEXT[] NOT NULL DEFAULT '{}'
);

INSERT INTO posts (tags) VALUES (ARRAY['go', 'postgresql', 'backend']);
INSERT INTO posts (tags) VALUES ('{"go", "api"}');

-- 배열 접근 (1-based index)
SELECT tags[1] FROM posts;

-- ANY: 배열 원소 포함 여부
SELECT * FROM posts WHERE 'go' = ANY(tags);

-- 배열 연결
UPDATE posts SET tags = tags || ARRAY['featured'] WHERE id = 1;

-- 배열 길이
SELECT array_length(tags, 1) FROM posts;

-- unnest: 배열 → 행으로 펼치기
SELECT id, unnest(tags) AS tag FROM posts;

-- array_agg: 집계 → 배열로
SELECT user_id, array_agg(product_id ORDER BY product_id) AS purchased
FROM order_items
GROUP BY user_id;

-- 배열 포함 연산자
SELECT * FROM posts WHERE tags @> ARRAY['go'];      -- 'go' 포함
SELECT * FROM posts WHERE tags && ARRAY['go','api']; -- 교집합 존재

-- GIN 인덱스
CREATE INDEX idx_posts_tags ON posts USING GIN (tags);
```

---

## 문자열

```sql
-- 연결
SELECT 'Hello' || ' ' || 'World';

-- FORMAT (printf 스타일)
SELECT FORMAT('이름: %s, 나이: %s', name, age) FROM users;

-- 정규식
SELECT regexp_matches('2024-01-15', '(\d{4})-(\d{2})-(\d{2})');
SELECT regexp_replace(phone, '[^0-9]', '', 'g') AS clean_phone FROM users;
SELECT * FROM users WHERE email ~ '^[a-z]+@';   -- 대소문자 구분
SELECT * FROM users WHERE email ~* '^admin@';   -- 대소문자 무시

-- 문자열 → 배열
SELECT string_to_array('a,b,c', ',');  -- {a,b,c}
SELECT array_to_string(ARRAY['a','b'], ',');  -- a,b

-- ILIKE: 대소문자 무시 LIKE
SELECT * FROM products WHERE name ILIKE '%apple%';
```

---

## ON CONFLICT (UPSERT)

```sql
-- DO NOTHING: 충돌 시 무시
INSERT INTO user_settings (user_id, key, value)
VALUES (1, 'theme', 'dark')
ON CONFLICT (user_id, key) DO NOTHING;

-- DO UPDATE SET: 충돌 시 업데이트
INSERT INTO user_settings (user_id, key, value)
VALUES (1, 'theme', 'dark')
ON CONFLICT (user_id, key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW();
-- EXCLUDED: 삽입하려 했던 행의 값 참조

-- 여러 컬럼 업데이트
INSERT INTO products (sku, name, price, stock)
VALUES ('ABC123', '상품A', 10000, 100)
ON CONFLICT (sku) DO UPDATE
SET name  = EXCLUDED.name,
    price = EXCLUDED.price,
    stock = products.stock + EXCLUDED.stock;  -- 기존 + 신규

-- 제약조건 이름으로 지정
ON CONFLICT ON CONSTRAINT uq_user_settings DO UPDATE ...
```

---

## RETURNING

```sql
-- INSERT 후 생성된 값 반환
INSERT INTO users (name, email)
VALUES ('김철수', 'kim@example.com')
RETURNING id, created_at;

-- UPDATE 후 변경된 행 반환
UPDATE orders SET status = 'shipped', shipped_at = NOW()
WHERE id = 42
RETURNING id, status, shipped_at;

-- DELETE 후 삭제된 행 반환
DELETE FROM sessions WHERE expires_at < NOW()
RETURNING id, user_id;

-- CTE와 조합
WITH inserted AS (
    INSERT INTO orders (customer_id, total)
    VALUES (1, 50000)
    RETURNING id
)
INSERT INTO order_items (order_id, product_id, quantity)
SELECT id, 101, 2 FROM inserted;
```

---

## LATERAL JOIN

```sql
-- LATERAL: 서브쿼리에서 외부 쿼리 컬럼 참조 가능
SELECT c.id, c.name, recent.*
FROM customers c
CROSS JOIN LATERAL (
    SELECT id AS order_id, total, created_at
    FROM orders
    WHERE customer_id = c.id  -- 외부 c 참조
    ORDER BY created_at DESC
    LIMIT 3
) recent;

-- LEFT JOIN LATERAL: 결과 없는 고객도 포함
SELECT c.name, o.order_id, o.total
FROM customers c
LEFT JOIN LATERAL (
    SELECT id AS order_id, total
    FROM orders WHERE customer_id = c.id
    ORDER BY created_at DESC LIMIT 1
) o ON TRUE;
```

---

## GENERATE_SERIES

```sql
-- 정수 시퀀스
SELECT generate_series(1, 10);
SELECT generate_series(0, 100, 10);  -- 0,10,20,...,100

-- 날짜 시퀀스 (일별)
SELECT generate_series(
    '2024-01-01'::DATE,
    '2024-12-31'::DATE,
    '1 day'::INTERVAL
)::DATE AS dt;

-- 날짜별 통계 (데이터 없는 날도 포함)
SELECT d.dt, COALESCE(COUNT(o.id), 0) AS order_count
FROM generate_series(
    '2024-01-01'::DATE,
    '2024-01-31'::DATE,
    '1 day'
) AS d(dt)
LEFT JOIN orders o ON o.created_at::DATE = d.dt
GROUP BY d.dt
ORDER BY d.dt;
```

---

## EXPLAIN ANALYZE

```sql
EXPLAIN ANALYZE
SELECT u.name, COUNT(o.id)
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.status = 'active'
GROUP BY u.id, u.name;
```

### 실행 계획 읽는 법

```
Seq Scan       — 전체 테이블 순차 스캔 (인덱스 없음, 대량 데이터 시 느림)
Index Scan     — 인덱스로 행 찾은 후 테이블 접근
Index Only Scan— 테이블 접근 없이 인덱스만으로 처리 (커버링 인덱스)
Bitmap Scan    — 인덱스에서 비트맵 생성 후 일괄 테이블 접근 (IN/OR에 효율)
Hash Join      — 작은 테이블을 해시 테이블로 만들어 조인 (equi-join)
Nested Loop    — 외부 행마다 내부 테이블 반복 (소량 데이터, 인덱스 있을 때 빠름)
Merge Join     — 정렬된 양쪽 테이블을 병합 (대량 정렬 조인)

중요 수치:
  cost=0.00..123.45      — 예상 비용 (시작..종료)
  rows=1000              — 예상 행 수
  actual time=0.1..5.2   — 실제 시간 (ms)
  actual rows=950        — 실제 행 수
  loops=1                — 반복 횟수

rows 예상 vs 실제 차이가 크면 → ANALYZE 실행하여 통계 갱신 필요
```

---

## 인덱스 종류

```sql
-- B-tree (기본): =, <, >, BETWEEN, LIKE 'prefix%'
CREATE INDEX idx_users_email ON users (email);

-- GIN: JSONB, 배열, 전문 검색
CREATE INDEX idx_events_payload ON events USING GIN (payload);
CREATE INDEX idx_posts_tags ON posts USING GIN (tags);

-- GiST: 지리 데이터(PostGIS), 범위 타입, 트리 구조
CREATE INDEX idx_locations_geo ON locations USING GIST (geom);

-- BRIN: 물리 순서와 상관관계가 높은 대용량 테이블 (시계열, 로그)
-- 매우 작은 인덱스 크기, 범위 조회에 적합
CREATE INDEX idx_logs_created ON logs USING BRIN (created_at);

-- 부분 인덱스 (Partial Index)
CREATE INDEX idx_orders_pending ON orders (customer_id, created_at)
WHERE status = 'pending';
```

---

## 슬로우 쿼리 & 통계

```sql
-- 현재 실행 중인 쿼리
SELECT pid, now() - pg_stat_activity.query_start AS duration,
       query, state
FROM pg_stat_activity
WHERE state != 'idle'
  AND query_start < NOW() - INTERVAL '5 seconds'
ORDER BY duration DESC;

-- 쿼리 강제 종료
SELECT pg_terminate_backend(pid) FROM pg_stat_activity
WHERE pid <> pg_backend_pid() AND query_start < NOW() - INTERVAL '1 hour';

-- 테이블 통계 (순차 스캔 vs 인덱스 스캔)
SELECT relname, seq_scan, idx_scan, n_live_tup, n_dead_tup
FROM pg_stat_user_tables
ORDER BY seq_scan DESC;
-- seq_scan이 높고 n_live_tup이 많으면 인덱스 추가 검토

-- 인덱스 사용률
SELECT indexrelname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC;
-- idx_scan = 0 인 인덱스는 제거 검토

-- pg_stat_statements (확장 활성화 필요)
SELECT query, calls, total_exec_time/calls AS avg_ms, rows
FROM pg_stat_statements
ORDER BY avg_ms DESC LIMIT 20;
```

---

## VACUUM / ANALYZE

```sql
-- VACUUM: 데드 튜플 정리 (UPDATE/DELETE 후 공간 회수)
VACUUM users;
VACUUM VERBOSE users;  -- 상세 출력

-- VACUUM FULL: 테이블 재작성 (락 발생, 공간 OS 반환)
VACUUM FULL users;

-- ANALYZE: 통계 갱신 (쿼리 플래너용)
ANALYZE users;
ANALYZE;  -- 전체 DB

-- 둘 다
VACUUM ANALYZE users;

-- autovacuum: 자동 실행 (기본 활성화)
-- 설정 확인
SHOW autovacuum;
SELECT * FROM pg_settings WHERE name LIKE 'autovacuum%';
```

---

## 파티셔닝 (PG 10+)

```sql
-- RANGE 파티셔닝
CREATE TABLE orders (
    id         BIGSERIAL,
    created_at TIMESTAMPTZ NOT NULL,
    total      NUMERIC(10,2)
) PARTITION BY RANGE (created_at);

CREATE TABLE orders_2024_q1 PARTITION OF orders
    FOR VALUES FROM ('2024-01-01') TO ('2024-04-01');
CREATE TABLE orders_2024_q2 PARTITION OF orders
    FOR VALUES FROM ('2024-04-01') TO ('2024-07-01');
CREATE TABLE orders_default PARTITION OF orders DEFAULT;

-- LIST 파티셔닝
CREATE TABLE events (country_code CHAR(2), ...) PARTITION BY LIST (country_code);
CREATE TABLE events_kr PARTITION OF events FOR VALUES IN ('KR');
CREATE TABLE events_us PARTITION OF events FOR VALUES IN ('US', 'CA');

-- HASH 파티셔닝
CREATE TABLE logs (...) PARTITION BY HASH (user_id);
CREATE TABLE logs_0 PARTITION OF logs FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE logs_1 PARTITION OF logs FOR VALUES WITH (MODULUS 4, REMAINDER 1);
```

---

## 슬로우 쿼리 개선 플로우

```sql
-- 1. 슬로우 쿼리 찾기
SELECT query, avg_exec_time FROM pg_stat_statements ORDER BY avg_exec_time DESC LIMIT 5;

-- 2. 실행 계획 확인
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM orders WHERE customer_id = 123 AND status = 'pending';

-- 3. Seq Scan 발견 → 인덱스 추가
CREATE INDEX CONCURRENTLY idx_orders_customer_status
ON orders (customer_id, status)
WHERE status = 'pending';  -- 부분 인덱스 (pending이 소수일 때)

-- 4. 인덱스 추가 후 재확인
EXPLAIN ANALYZE SELECT * FROM orders WHERE customer_id = 123 AND status = 'pending';
-- Index Scan 또는 Index Only Scan 확인

-- 5. 통계가 오래됐으면
ANALYZE orders;
-- 이후 다시 EXPLAIN으로 예상 rows 확인

-- CONCURRENTLY: 운영 중 인덱스 추가 (락 없음, 단 시간 더 걸림)
CREATE INDEX CONCURRENTLY idx_new ON table_name (column);
```
