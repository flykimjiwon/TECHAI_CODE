# SQL DDL & Schema Design Reference

## CREATE TABLE

```sql
-- 기본 테이블 생성
CREATE TABLE users (
    id          BIGINT          NOT NULL AUTO_INCREMENT,  -- MySQL
    -- id       BIGSERIAL       NOT NULL,                 -- PostgreSQL
    -- id       NUMBER          GENERATED ALWAYS AS IDENTITY, -- Oracle 12c+
    username    VARCHAR(50)     NOT NULL,
    email       VARCHAR(255)    NOT NULL,
    password    VARCHAR(255)    NOT NULL,
    age         INT             CHECK (age >= 0 AND age <= 150),
    status      CHAR(1)         NOT NULL DEFAULT 'A',
    balance     DECIMAL(15, 2)  NOT NULL DEFAULT 0.00,
    is_admin    BOOLEAN         NOT NULL DEFAULT FALSE,
    bio         TEXT,
    created_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);
```

### 주요 데이터 타입 선택 가이드

| 용도 | MySQL | PostgreSQL | Oracle |
|------|-------|-----------|--------|
| 정수 PK | BIGINT AUTO_INCREMENT | BIGSERIAL / BIGINT | NUMBER / GENERATED AS IDENTITY |
| 짧은 문자열 | VARCHAR(n) | VARCHAR(n) | VARCHAR2(n) |
| 긴 텍스트 | TEXT | TEXT | CLOB |
| 금액 | DECIMAL(15,2) | NUMERIC(15,2) | NUMBER(15,2) |
| 날짜+시간 | DATETIME / TIMESTAMP | TIMESTAMP WITH TIME ZONE | TIMESTAMP |
| UUID | VARCHAR(36) / CHAR(36) | UUID | VARCHAR2(36) / RAW(16) |
| 불리언 | TINYINT(1) / BOOLEAN | BOOLEAN | NUMBER(1) / CHAR(1) |
| JSON | JSON | JSONB | CLOB + JSON constraint |
| 이진 데이터 | BLOB | BYTEA | BLOB / RAW |

---

## 제약조건

```sql
-- PRIMARY KEY
CREATE TABLE orders (
    id BIGINT PRIMARY KEY,
    -- 복합 PK
    -- PRIMARY KEY (order_id, item_id)
);

-- FOREIGN KEY: 참조 무결성
CREATE TABLE order_items (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_id    BIGINT NOT NULL,
    product_id  BIGINT NOT NULL,
    quantity    INT NOT NULL CHECK (quantity > 0),
    unit_price  DECIMAL(10, 2) NOT NULL,
    FOREIGN KEY (order_id)   REFERENCES orders(id)   ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT
    -- ON DELETE: CASCADE (같이 삭제), RESTRICT (삭제 막음), SET NULL (NULL로), NO ACTION (기본)
);

-- UNIQUE: 단일/복합
CREATE TABLE user_emails (
    id      BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    email   VARCHAR(255) NOT NULL,
    type    VARCHAR(20) NOT NULL,  -- 'primary', 'work', 'recovery'
    UNIQUE (email),
    UNIQUE (user_id, type)  -- 복합 유니크: 사용자당 타입 1개
);

-- CHECK
CREATE TABLE products (
    id       BIGINT PRIMARY KEY,
    name     VARCHAR(255) NOT NULL,
    price    DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    stock    INT NOT NULL DEFAULT 0 CHECK (stock >= 0),
    rating   DECIMAL(2,1) CHECK (rating BETWEEN 0.0 AND 5.0),
    status   VARCHAR(20) NOT NULL CHECK (status IN ('draft','active','discontinued'))
);

-- DEFAULT
CREATE TABLE sessions (
    token      CHAR(64) PRIMARY KEY,
    user_id    BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP + INTERVAL '24 HOUR'),
    is_active  BOOLEAN NOT NULL DEFAULT TRUE
);
```

---

## ALTER TABLE

```sql
-- 컬럼 추가
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;

-- 컬럼 수정 (MySQL)
ALTER TABLE users MODIFY COLUMN phone VARCHAR(30) NOT NULL DEFAULT '';
-- PostgreSQL
ALTER TABLE users ALTER COLUMN phone TYPE VARCHAR(30);
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;
ALTER TABLE users ALTER COLUMN phone SET DEFAULT '';

-- 컬럼 이름 변경
ALTER TABLE users RENAME COLUMN phone TO phone_number;  -- PostgreSQL/MySQL 8+
ALTER TABLE users CHANGE phone phone_number VARCHAR(30);  -- MySQL

-- 컬럼 삭제
ALTER TABLE users DROP COLUMN bio;

-- 제약조건 추가
ALTER TABLE orders ADD CONSTRAINT fk_customer
    FOREIGN KEY (customer_id) REFERENCES customers(id);
ALTER TABLE users ADD CONSTRAINT uq_email UNIQUE (email);
ALTER TABLE products ADD CONSTRAINT chk_price CHECK (price > 0);

-- 제약조건 삭제
ALTER TABLE orders DROP FOREIGN KEY fk_customer;  -- MySQL
ALTER TABLE orders DROP CONSTRAINT fk_customer;   -- PostgreSQL/Oracle

-- 인덱스 추가/삭제
ALTER TABLE users ADD INDEX idx_email (email);  -- MySQL
CREATE INDEX idx_email ON users (email);        -- 표준

-- 테이블 이름 변경
ALTER TABLE user_profiles RENAME TO profiles;  -- PostgreSQL
RENAME TABLE user_profiles TO profiles;        -- MySQL
```

---

## DROP / TRUNCATE

```sql
-- DROP: 구조 포함 완전 삭제
DROP TABLE IF EXISTS temp_data;
DROP TABLE IF EXISTS child_table, parent_table;  -- 의존성 주의

-- CASCADE: 참조하는 객체까지 삭제 (PostgreSQL)
DROP TABLE users CASCADE;

-- TRUNCATE: 데이터만 빠르게 삭제 (구조 유지, DDL이므로 autocommit)
TRUNCATE TABLE staging_data;
TRUNCATE TABLE logs RESTART IDENTITY;  -- PostgreSQL: 시퀀스도 리셋
```

---

## 인덱스

```sql
-- 단일 컬럼 인덱스
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_orders_created ON orders (created_at);

-- 유니크 인덱스
CREATE UNIQUE INDEX uq_users_email ON users (email);

-- 복합 인덱스: 순서가 중요 (leftmost prefix rule)
CREATE INDEX idx_orders_status_date ON orders (status, created_at);
-- 위 인덱스는 (status), (status, created_at) 쿼리에 사용됨
-- (created_at) 단독 쿼리에는 사용 안 됨

-- 커버링 인덱스: 쿼리에 필요한 컬럼을 인덱스에 모두 포함
-- SELECT name, email FROM users WHERE department_id = 5
CREATE INDEX idx_users_dept_cover ON users (department_id, name, email);
-- 인덱스만으로 쿼리 처리 → 테이블 접근 불필요 (Index Only Scan)

-- 내림차순 인덱스 (MySQL 8+, PostgreSQL)
CREATE INDEX idx_posts_created_desc ON posts (created_at DESC);

-- 부분 인덱스 (PostgreSQL): 조건에 맞는 행만 인덱싱
CREATE INDEX idx_orders_pending ON orders (created_at)
WHERE status = 'pending';
-- pending 주문 조회가 빈번하고, 전체의 소수일 때 효과적

-- 함수 기반 인덱스
CREATE INDEX idx_users_lower_email ON users (LOWER(email));
-- WHERE LOWER(email) = 'test@example.com' 쿼리에 사용
```

### 인덱스 설계 전략

```
1. 선택도(Selectivity) 높은 컬럼 우선 (카디널리티 높을수록 효과적)
   - 좋음: email, user_id, created_at
   - 나쁨: boolean, status (값이 2~3개)

2. WHERE, JOIN ON, ORDER BY, GROUP BY 순으로 우선순위
3. 복합 인덱스: 조건 많이 쓰는 컬럼 → 범위 조건 컬럼 순
4. 인덱스는 INSERT/UPDATE/DELETE 비용 증가 → 과도하게 만들지 않기
5. 테이블당 5~8개 이하 권장
6. 커버링 인덱스로 Index Only Scan 유도 (I/O 최소화)

-- 인덱스 사용 여부 확인
EXPLAIN SELECT * FROM users WHERE email = 'test@example.com';
-- type: ref 또는 eq_ref → 인덱스 사용
-- type: ALL → Full Scan → 인덱스 필요
```

---

## 정규화

```sql
-- 1NF: 원자값, 반복 그룹 제거
-- 위반 예: tags VARCHAR(255) = "java,python,go"  → 별도 테이블
CREATE TABLE user_tags (
    user_id BIGINT NOT NULL,
    tag     VARCHAR(50) NOT NULL,
    PRIMARY KEY (user_id, tag)
);

-- 2NF: 부분 함수 종속 제거 (복합 PK에서 일부 컬럼에만 종속된 속성 분리)
-- 위반 예: order_items(order_id, product_id, product_name)
--          product_name은 product_id에만 종속
-- 해결: product_name을 products 테이블로 분리

-- 3NF: 이행 함수 종속 제거
-- 위반 예: employees(id, dept_id, dept_name)
--          dept_name은 dept_id에 종속 (id → dept_id → dept_name)
-- 해결: departments(id, name) 별도 테이블
```

### 반정규화 전략 (성능 vs 무결성 트레이드오프)

```sql
-- 집계값 캐싱: 매번 COUNT(*) 대신 컬럼에 저장
ALTER TABLE posts ADD COLUMN comment_count INT NOT NULL DEFAULT 0;
-- INSERT/DELETE 시 트리거 또는 애플리케이션에서 업데이트

-- 자주 JOIN하는 컬럼 복제
ALTER TABLE orders ADD COLUMN customer_name VARCHAR(100);
-- customer_name이 변경되면 동기화 필요

-- 이력 테이블 분리 (파티셔닝 대안)
CREATE TABLE orders_archive LIKE orders;
-- 오래된 데이터는 archive로 이동하여 메인 테이블 슬림하게 유지
```

---

## ERD 설계 패턴

```sql
-- 1:1 관계 (같은 PK 공유 또는 FK + UNIQUE)
CREATE TABLE users      (id BIGINT PRIMARY KEY, ...);
CREATE TABLE user_profiles (
    user_id BIGINT PRIMARY KEY,
    avatar_url VARCHAR(500),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 1:N 관계 (FK를 N 쪽에)
CREATE TABLE departments (id BIGINT PRIMARY KEY, name VARCHAR(100));
CREATE TABLE employees (
    id BIGINT PRIMARY KEY,
    dept_id BIGINT,
    FOREIGN KEY (dept_id) REFERENCES departments(id)
);

-- M:N 관계 (중간 테이블)
CREATE TABLE students (id BIGINT PRIMARY KEY, name VARCHAR(100));
CREATE TABLE courses  (id BIGINT PRIMARY KEY, title VARCHAR(200));
CREATE TABLE enrollments (
    student_id  BIGINT NOT NULL,
    course_id   BIGINT NOT NULL,
    enrolled_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    grade       CHAR(2),
    PRIMARY KEY (student_id, course_id),
    FOREIGN KEY (student_id) REFERENCES students(id),
    FOREIGN KEY (course_id)  REFERENCES courses(id)
);
```

---

## 파티셔닝 개념

```sql
-- RANGE 파티셔닝 (날짜 기반이 일반적)
CREATE TABLE orders (
    id         BIGINT NOT NULL,
    created_at DATE NOT NULL,
    amount     DECIMAL(10,2)
)
PARTITION BY RANGE (YEAR(created_at)) (
    PARTITION p2022 VALUES LESS THAN (2023),
    PARTITION p2023 VALUES LESS THAN (2024),
    PARTITION p2024 VALUES LESS THAN (2025),
    PARTITION pmax  VALUES LESS THAN MAXVALUE
);

-- LIST 파티셔닝 (국가, 지역 등 열거형)
PARTITION BY LIST (country_code) (
    PARTITION p_kr VALUES IN ('KR'),
    PARTITION p_us VALUES IN ('US', 'CA'),
    PARTITION p_etc VALUES IN (DEFAULT)
);

-- HASH 파티셔닝 (균등 분산)
PARTITION BY HASH (user_id) PARTITIONS 8;
```

---

## 시퀀스 / AUTO_INCREMENT

```sql
-- MySQL AUTO_INCREMENT
CREATE TABLE items (id INT AUTO_INCREMENT PRIMARY KEY, ...);
ALTER TABLE items AUTO_INCREMENT = 1000;  -- 시작값 변경

-- PostgreSQL SERIAL / SEQUENCE
CREATE TABLE items (id SERIAL PRIMARY KEY, ...);
CREATE TABLE items (id BIGSERIAL PRIMARY KEY, ...);

-- 직접 시퀀스 생성 (PostgreSQL)
CREATE SEQUENCE order_seq START 1000 INCREMENT 1 CACHE 10;
SELECT NEXTVAL('order_seq');
SELECT CURRVAL('order_seq');

-- Oracle SEQUENCE
CREATE SEQUENCE emp_seq START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE;
INSERT INTO employees (id, name) VALUES (emp_seq.NEXTVAL, '홍길동');
```

---

## 뷰 (View)

```sql
-- 일반 뷰
CREATE OR REPLACE VIEW active_customers AS
SELECT id, name, email, created_at
FROM customers
WHERE status = 'active' AND deleted_at IS NULL;

-- 사용
SELECT * FROM active_customers WHERE name LIKE '김%';

-- Materialized View 개념 (PostgreSQL)
-- 실제 데이터를 저장 → 빠른 읽기, 수동 갱신 필요
CREATE MATERIALIZED VIEW monthly_stats AS
SELECT DATE_TRUNC('month', created_at) AS month, COUNT(*) AS orders, SUM(total) AS revenue
FROM orders GROUP BY 1;

REFRESH MATERIALIZED VIEW monthly_stats;  -- 수동 갱신
REFRESH MATERIALIZED VIEW CONCURRENTLY monthly_stats;  -- 락 없이 갱신
```

---

## 트랜잭션

```sql
-- 기본 트랜잭션
BEGIN;  -- 또는 START TRANSACTION

UPDATE accounts SET balance = balance - 10000 WHERE id = 1;
UPDATE accounts SET balance = balance + 10000 WHERE id = 2;

COMMIT;   -- 성공 시 확정
ROLLBACK; -- 오류 시 취소

-- SAVEPOINT: 부분 롤백
BEGIN;
INSERT INTO orders (customer_id, total) VALUES (1, 50000);
SAVEPOINT after_insert;

UPDATE inventory SET stock = stock - 1 WHERE product_id = 100;
-- 오류 발생 시
ROLLBACK TO SAVEPOINT after_insert;  -- INSERT는 유지, UPDATE만 취소
COMMIT;
```

### 격리 수준

| 수준 | Dirty Read | Non-Repeatable Read | Phantom Read |
|------|-----------|---------------------|--------------|
| READ UNCOMMITTED | 가능 | 가능 | 가능 |
| READ COMMITTED | 방지 | 가능 | 가능 |
| REPEATABLE READ | 방지 | 방지 | 가능 (MySQL은 방지) |
| SERIALIZABLE | 방지 | 방지 | 방지 |

```sql
-- 격리 수준 설정
SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;  -- MySQL
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;        -- 트랜잭션 단위
```

### 락 & 데드락 방지

```sql
-- 명시적 행 락 (SELECT FOR UPDATE)
BEGIN;
SELECT * FROM accounts WHERE id = 1 FOR UPDATE;  -- 행 락 획득
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
COMMIT;

-- 데드락 방지: 항상 같은 순서로 락 획득
-- 잘못된 예: 트랜잭션 A가 id=1→2, B가 id=2→1 순으로 락
-- 올바른 예: 항상 id 오름차순으로 락
BEGIN;
SELECT * FROM accounts WHERE id IN (1, 2) ORDER BY id FOR UPDATE;

-- NOWAIT: 락 대기 없이 즉시 오류
SELECT * FROM accounts WHERE id = 1 FOR UPDATE NOWAIT;

-- SKIP LOCKED: 락 걸린 행 건너뜀 (작업 큐 패턴)
SELECT * FROM job_queue WHERE status = 'pending'
ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED;
```
