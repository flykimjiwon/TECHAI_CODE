# MySQL Reference

## 스토리지 엔진: InnoDB vs MyISAM

| 항목 | InnoDB | MyISAM |
|------|--------|--------|
| 트랜잭션 | 지원 (ACID) | 미지원 |
| 외래키 | 지원 | 미지원 |
| 행 수준 락 | 지원 | 테이블 락만 |
| 크래시 복구 | 자동 (redo log) | 수동 |
| FULLTEXT 인덱스 | 5.6+부터 지원 | 지원 |
| 용도 | **운영 DB 표준** | 읽기 전용 통계 테이블 (레거시) |

InnoDB가 MySQL 5.5+의 기본 엔진. 특별한 이유 없으면 InnoDB 사용.

---

## 데이터 타입

```sql
-- 정수
TINYINT     -- 1바이트, -128~127 (UNSIGNED: 0~255)
SMALLINT    -- 2바이트
MEDIUMINT   -- 3바이트
INT         -- 4바이트, -21억~21억
BIGINT      -- 8바이트, -922경~922경 (PK에 권장)

-- 실수
DECIMAL(p, s)  -- 정밀도 보장 (금액에 사용, NUMERIC과 동일)
FLOAT, DOUBLE  -- 부동소수 (근사값, 금액에 사용 금지)

-- 문자열
CHAR(n)        -- 고정 길이, 패딩 있음 (상태코드, 코드값)
VARCHAR(n)     -- 가변 길이, n은 문자 수 기준 (utf8mb4: 최대 ~16383)
TINYTEXT       -- 최대 255바이트
TEXT           -- 최대 65KB
MEDIUMTEXT     -- 최대 16MB
LONGTEXT       -- 최대 4GB

-- VARCHAR vs TEXT 선택 기준:
-- VARCHAR: 인덱스 가능, 255자 이하 권장
-- TEXT: 인덱스 불가 (prefix 인덱스는 가능), 긴 본문 내용

-- 날짜/시간
DATE           -- 날짜만 (YYYY-MM-DD)
TIME           -- 시간만
DATETIME       -- 날짜+시간, 타임존 미저장, 범위 1000~9999년
TIMESTAMP      -- 날짜+시간, UTC 저장 후 세션 타임존 변환, 범위 1970~2038년
               -- DATETIME 권장 (2038 문제 없음)
YEAR           -- 연도만

-- ENUM: 허용값 고정 (내부적으로 정수 저장, 빠름, 변경 시 ALTER 필요)
status ENUM('pending', 'active', 'inactive', 'deleted') NOT NULL DEFAULT 'pending'

-- JSON (MySQL 5.7.8+)
payload JSON

-- 이진
BINARY(n), VARBINARY(n)
BLOB, MEDIUMBLOB, LONGBLOB
```

---

## AUTO_INCREMENT / UUID

```sql
-- AUTO_INCREMENT
CREATE TABLE orders (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    ...
);

-- 현재 AUTO_INCREMENT 값 확인
SELECT AUTO_INCREMENT FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'mydb' AND TABLE_NAME = 'orders';

-- 시작값 변경
ALTER TABLE orders AUTO_INCREMENT = 10000;

-- 마지막 삽입 ID
SELECT LAST_INSERT_ID();

-- UUID (텍스트 저장)
INSERT INTO tokens (id, user_id) VALUES (UUID(), 1);

-- UUID 정렬 최적화 (MySQL 8.0+)
-- UUID_TO_BIN(uuid, 1): 시간순 정렬 가능한 바이너리로 변환
CREATE TABLE sessions (
    id BINARY(16) NOT NULL DEFAULT (UUID_TO_BIN(UUID(), 1)) PRIMARY KEY
);
SELECT BIN_TO_UUID(id, 1) AS uuid FROM sessions;
```

---

## JSON 함수 (MySQL 5.7+)

```sql
CREATE TABLE events (
    id      BIGINT AUTO_INCREMENT PRIMARY KEY,
    payload JSON NOT NULL
);

INSERT INTO events (payload) VALUES
('{"user_id": 1, "action": "login", "tags": ["web", "mobile"]}');

-- JSON_EXTRACT: 값 추출
SELECT JSON_EXTRACT(payload, '$.user_id') FROM events;    -- 1
SELECT JSON_EXTRACT(payload, '$.tags[0]') FROM events;    -- "web"

-- -> 단축 연산자 (JSON 타입 반환)
SELECT payload -> '$.user_id' FROM events;

-- ->> 단축 연산자 (텍스트 반환, 따옴표 제거)
SELECT payload ->> '$.action' FROM events;                -- login

-- WHERE에서 사용
SELECT * FROM events WHERE payload ->> '$.action' = 'login';

-- JSON_SET: 값 추가/수정
UPDATE events
SET payload = JSON_SET(payload, '$.processed', true, '$.ts', NOW())
WHERE id = 1;

-- JSON_REMOVE: 키 삭제
UPDATE events SET payload = JSON_REMOVE(payload, '$.tags') WHERE id = 1;

-- JSON_ARRAYAGG: 집계 (MySQL 5.7.22+)
SELECT user_id, JSON_ARRAYAGG(product_id) AS products
FROM order_items GROUP BY user_id;

-- JSON_OBJECTAGG: 키-값 집계
SELECT JSON_OBJECTAGG(setting_key, setting_value)
FROM user_settings WHERE user_id = 1;

-- 배열 원소 포함 여부
SELECT * FROM events
WHERE JSON_CONTAINS(payload -> '$.tags', '"mobile"');

-- JSON 인덱스: 가상 컬럼 활용
ALTER TABLE events
    ADD COLUMN action VARCHAR(50) GENERATED ALWAYS AS (payload ->> '$.action') VIRTUAL;
CREATE INDEX idx_events_action ON events (action);
-- 이후 WHERE action = 'login' 쿼리는 인덱스 사용
```

---

## 인덱스

```sql
-- PRIMARY KEY (클러스터드 인덱스: 데이터 자체가 PK 순서로 저장)
-- UNIQUE INDEX
CREATE UNIQUE INDEX uq_users_email ON users (email);

-- 일반 INDEX
CREATE INDEX idx_orders_created ON orders (created_at);

-- 복합 인덱스: leftmost prefix rule
CREATE INDEX idx_orders_status_customer ON orders (status, customer_id, created_at);
-- 사용 가능한 조합:
--   (status)
--   (status, customer_id)
--   (status, customer_id, created_at)
-- 사용 불가:
--   (customer_id)
--   (created_at)
--   (customer_id, created_at)

-- FULLTEXT 인덱스 (자연어 검색)
CREATE FULLTEXT INDEX ft_products_name ON products (name, description);
SELECT * FROM products WHERE MATCH(name, description) AGAINST ('스마트폰' IN BOOLEAN MODE);

-- 인덱스 삭제
DROP INDEX idx_orders_created ON orders;
ALTER TABLE orders DROP INDEX idx_orders_created;
```

---

## EXPLAIN 읽는 법

```sql
EXPLAIN SELECT u.name, COUNT(o.id)
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.status = 'active'
GROUP BY u.id;
```

| 컬럼 | 의미 |
|------|------|
| type | **접근 방식** (아래 참고) |
| key | 실제 사용된 인덱스 |
| rows | 예상 처리 행 수 |
| Extra | 추가 정보 (Using index, Using filesort 등) |

**type 값 (좋음 → 나쁨 순):**
```
system   — 테이블에 행이 1개
const    — PK/UNIQUE로 정확히 1행 (WHERE id = 1)
eq_ref   — JOIN 시 PK/UNIQUE 1:1 매칭
ref      — 인덱스로 여러 행 (WHERE status = 'active')
range    — 인덱스 범위 스캔 (BETWEEN, >, <, IN)
index    — 인덱스 전체 스캔 (테이블 스캔보다 낫지만 주의)
ALL      — 풀 테이블 스캔 ← 인덱스 추가 검토
```

**Extra 주요 값:**
```
Using index         — 커버링 인덱스 (테이블 접근 없음, 좋음)
Using where         — WHERE 조건 후처리
Using filesort      — ORDER BY에 인덱스 미사용 (정렬 비용 발생)
Using temporary     — 임시 테이블 사용 (GROUP BY, DISTINCT 등)
Using index condition — ICP (Index Condition Pushdown), 스토리지 엔진 레벨 필터
```

---

## 슬로우 쿼리 로그

```sql
-- 설정 (my.cnf 또는 동적)
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 1;           -- 1초 이상 쿼리 기록
SET GLOBAL slow_query_log_file = '/var/log/mysql/slow.log';
SET GLOBAL log_queries_not_using_indexes = 'ON';  -- 인덱스 미사용 쿼리도

-- 현재 설정 확인
SHOW VARIABLES LIKE 'slow%';
SHOW VARIABLES LIKE 'long_query_time';

-- mysqldumpslow으로 분석 (CLI)
-- mysqldumpslow -s t -t 10 /var/log/mysql/slow.log
-- -s t: 평균 실행시간 정렬, -t 10: 상위 10개

-- performance_schema 활용
SELECT digest_text, count_star, avg_timer_wait/1e9 AS avg_ms
FROM performance_schema.events_statements_summary_by_digest
ORDER BY avg_timer_wait DESC LIMIT 20;
```

---

## 문자셋: utf8mb4

```sql
-- utf8 (MySQL의 utf8)은 3바이트만 지원 → 이모지 저장 불가
-- utf8mb4: 완전한 UTF-8 (4바이트), 이모지 포함 모든 유니코드

-- DB/테이블 생성 시 명시
CREATE DATABASE myapp CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE posts (
    content TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci
);

-- 기존 테이블 변환
ALTER TABLE users CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- COLLATION 선택:
-- utf8mb4_unicode_ci  — 유니코드 표준 정렬, 일반적으로 권장
-- utf8mb4_general_ci  — 빠르지만 일부 문자 처리 다름
-- utf8mb4_bin         — 바이너리 비교, 대소문자 구분
-- utf8mb4_0900_ai_ci  — MySQL 8.0+ 기본, 최신 유니코드, 권장

-- 연결 설정 (JDBC URL)
-- jdbc:mysql://host/db?useUnicode=true&characterEncoding=utf8mb4
```

---

## INSERT ... ON DUPLICATE KEY UPDATE

```sql
-- PK 또는 UNIQUE 충돌 시 UPDATE
INSERT INTO page_views (page_id, date, view_count)
VALUES (1, '2024-01-15', 1)
ON DUPLICATE KEY UPDATE view_count = view_count + 1;

-- VALUES() 함수로 삽입하려던 값 참조
INSERT INTO products (sku, name, price, stock)
VALUES ('ABC', '상품A', 10000, 100)
ON DUPLICATE KEY UPDATE
    price = VALUES(price),
    stock = stock + VALUES(stock);

-- 여러 행 한 번에
INSERT INTO scores (user_id, game_id, score)
VALUES (1, 10, 500), (2, 10, 700), (3, 10, 300)
ON DUPLICATE KEY UPDATE score = GREATEST(score, VALUES(score));
```

---

## GROUP_CONCAT / FIND_IN_SET

```sql
-- GROUP_CONCAT: 그룹 내 값을 문자열로 합침
SELECT order_id,
       GROUP_CONCAT(product_name ORDER BY product_name SEPARATOR ', ') AS items,
       GROUP_CONCAT(DISTINCT category_id) AS categories
FROM order_items oi
JOIN products p ON oi.product_id = p.id
GROUP BY order_id;

-- 길이 제한 조정 (기본 1024)
SET SESSION group_concat_max_len = 65536;

-- FIND_IN_SET: 콤마 구분 문자열에서 값 찾기 (레거시 데이터 처리용)
SELECT * FROM users WHERE FIND_IN_SET('admin', roles) > 0;
-- roles = 'editor,admin,viewer' 인 행 찾기
-- 주의: 인덱스 사용 불가 → 정규화된 설계가 권장
```

---

## 실전 패턴

### 페이징 최적화

```sql
-- 나쁜 예: OFFSET이 크면 앞에서부터 다 스캔
SELECT * FROM posts ORDER BY id DESC LIMIT 20 OFFSET 100000;

-- 좋은 예: 커서 기반 (마지막 id 기억)
SELECT * FROM posts
WHERE id < :last_seen_id
ORDER BY id DESC
LIMIT 20;

-- Deferred Join 패턴 (OFFSET 불가피할 때)
SELECT p.*
FROM posts p
JOIN (
    SELECT id FROM posts ORDER BY id DESC LIMIT 20 OFFSET 10000
) ids ON p.id = ids.id;
-- 서브쿼리에서 인덱스만으로 id 목록 추출 → 본 테이블 접근 최소화
```

### 대량 INSERT (Bulk)

```sql
-- 단건 반복보다 다건 한 번에 (10~100배 빠름)
INSERT INTO logs (user_id, action, created_at) VALUES
    (1, 'login', NOW()),
    (2, 'view', NOW()),
    (3, 'purchase', NOW());

-- LOAD DATA INFILE (가장 빠름, CSV 파일 직접 로드)
LOAD DATA LOCAL INFILE '/tmp/data.csv'
INTO TABLE orders
FIELDS TERMINATED BY ',' ENCLOSED BY '"'
LINES TERMINATED BY '\n'
IGNORE 1 ROWS  -- 헤더 제외
(customer_id, total, created_at);

-- INSERT IGNORE: 중복 행 오류 무시
INSERT IGNORE INTO user_tags (user_id, tag) VALUES (1, 'vip'), (1, 'new');
```

---

## MySQL 5.7 vs 8.0 주요 차이

| 기능 | 5.7 | 8.0 |
|------|-----|-----|
| CTE (WITH) | 미지원 | 지원 |
| 윈도우 함수 | 미지원 | 지원 |
| ROLE | 미지원 | 지원 |
| JSON 개선 | 기본 지원 | 대폭 개선 (JSON_TABLE 등) |
| 기본 문자셋 | utf8 | utf8mb4 |
| 기본 인증 | mysql_native_password | caching_sha2_password |
| DESCENDING 인덱스 | 미지원 | 지원 |
| Invisible 인덱스 | 미지원 | 지원 |
| EXPLAIN ANALYZE | 미지원 | 지원 |

```sql
-- 8.0 전용 기능 예시

-- CTE
WITH ranked AS (
    SELECT *, RANK() OVER (PARTITION BY dept_id ORDER BY salary DESC) AS rnk
    FROM employees
)
SELECT * FROM ranked WHERE rnk = 1;

-- JSON_TABLE: JSON 배열 → 관계형 테이블
SELECT jt.*
FROM orders o,
JSON_TABLE(o.items, '$[*]' COLUMNS (
    product_id INT PATH '$.product_id',
    qty        INT PATH '$.quantity'
)) jt;

-- Invisible Index: 인덱스 비활성화 테스트 (삭제 전 영향 확인)
ALTER TABLE users ALTER INDEX idx_email INVISIBLE;
-- 쿼리 영향 확인 후
ALTER TABLE users ALTER INDEX idx_email VISIBLE;
-- 또는 DROP INDEX
```

---

## 락: InnoDB

```sql
-- 행 락 (Row Lock): UPDATE/DELETE/SELECT FOR UPDATE 시 자동
-- Gap Lock: 범위 쿼리 시 삽입 방지 (REPEATABLE READ 수준에서)
-- Next-Key Lock: 행 락 + Gap Lock 조합

-- 데드락 감지: InnoDB가 자동 감지 후 하나 롤백
-- 마지막 데드락 확인
SHOW ENGINE INNODB STATUS;  -- LATEST DETECTED DEADLOCK 섹션

-- 데드락 방지 패턴
-- 1. 항상 같은 순서로 테이블/행 접근
-- 2. 트랜잭션 짧게 유지
-- 3. 필요한 인덱스 추가 (넓은 범위 락 방지)
-- 4. SELECT FOR UPDATE → 필요할 때만 사용

-- 현재 락 대기 확인 (MySQL 8.0)
SELECT * FROM performance_schema.data_lock_waits;
SELECT * FROM performance_schema.data_locks;

-- innodb_lock_wait_timeout 설정 (기본 50초)
SET SESSION innodb_lock_wait_timeout = 5;
```

---

## 백업

```sql
-- mysqldump: 논리적 백업 (SQL 파일)
-- 단일 DB
-- mysqldump -u root -p mydb > mydb_backup.sql

-- 특정 테이블
-- mysqldump -u root -p mydb orders users > partial_backup.sql

-- 전체 DB
-- mysqldump -u root -p --all-databases > full_backup.sql

-- 옵션:
-- --single-transaction: InnoDB 일관된 스냅샷 (락 없음)
-- --routines: 프로시저/함수 포함
-- --triggers: 트리거 포함
-- --no-data: 스키마만

-- 복원
-- mysql -u root -p mydb < mydb_backup.sql

-- mysqlpump (5.7+): 병렬 백업, 더 빠름
-- mysqlpump -u root -p --default-parallelism=4 mydb > backup.sql

-- 물리적 백업: Percona XtraBackup (운영 환경 권장, 무중단)
```
