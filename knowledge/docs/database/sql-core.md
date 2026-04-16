# SQL Core Reference

## SELECT 기본

```sql
SELECT DISTINCT col1, col2, col3
FROM table_name
WHERE condition
GROUP BY col1, col2
HAVING aggregate_condition
ORDER BY col1 ASC, col2 DESC
LIMIT 10 OFFSET 20;  -- MySQL/PostgreSQL
-- Oracle: FETCH FIRST 10 ROWS ONLY (12c+), ROWNUM (레거시)
```

```sql
-- LIMIT/OFFSET 패턴
SELECT * FROM orders ORDER BY created_at DESC LIMIT 20 OFFSET 40;

-- DISTINCT
SELECT DISTINCT department_id FROM employees;

-- 별칭
SELECT e.first_name || ' ' || e.last_name AS full_name, d.name AS dept
FROM employees e
JOIN departments d ON e.department_id = d.id;
```

---

## JOIN

```sql
-- INNER JOIN: 양쪽 모두 존재하는 행만
SELECT o.id, c.name, o.total
FROM orders o
INNER JOIN customers c ON o.customer_id = c.id;

-- LEFT JOIN: 왼쪽 기준 (오른쪽 없으면 NULL)
SELECT c.name, COUNT(o.id) AS order_count
FROM customers c
LEFT JOIN orders o ON c.id = o.customer_id
GROUP BY c.id, c.name;

-- RIGHT JOIN: 오른쪽 기준
SELECT e.name, d.name AS department
FROM employees e
RIGHT JOIN departments d ON e.department_id = d.id;

-- FULL OUTER JOIN: 양쪽 모두 포함
SELECT a.id AS a_id, b.id AS b_id
FROM table_a a
FULL OUTER JOIN table_b b ON a.key = b.key;

-- CROSS JOIN: 카테시안 곱
SELECT a.size, a.color FROM attributes a
CROSS JOIN attributes b WHERE a.type = 'size' AND b.type = 'color';

-- Self-Join: 같은 테이블 두 번 사용
SELECT e.name AS employee, m.name AS manager
FROM employees e
LEFT JOIN employees m ON e.manager_id = m.id;
```

---

## 서브쿼리

```sql
-- WHERE IN
SELECT * FROM products
WHERE category_id IN (
    SELECT id FROM categories WHERE active = true
);

-- WHERE EXISTS (IN보다 큰 결과셋에서 빠름)
SELECT * FROM customers c
WHERE EXISTS (
    SELECT 1 FROM orders o
    WHERE o.customer_id = c.id AND o.status = 'completed'
);

-- 스칼라 서브쿼리 (단일 값 반환)
SELECT name,
       (SELECT COUNT(*) FROM orders WHERE customer_id = c.id) AS order_count
FROM customers c;

-- FROM절 인라인 뷰
SELECT dept_name, avg_salary
FROM (
    SELECT d.name AS dept_name, AVG(e.salary) AS avg_salary
    FROM employees e JOIN departments d ON e.dept_id = d.id
    GROUP BY d.name
) dept_stats
WHERE avg_salary > 50000;

-- 상관 서브쿼리 (외부 쿼리 참조)
SELECT * FROM employees e
WHERE salary > (
    SELECT AVG(salary) FROM employees
    WHERE department_id = e.department_id
);
```

---

## CTE (Common Table Expressions)

```sql
-- 기본 CTE
WITH active_customers AS (
    SELECT id, name, email
    FROM customers
    WHERE status = 'active'
),
recent_orders AS (
    SELECT customer_id, MAX(created_at) AS last_order
    FROM orders
    WHERE created_at > CURRENT_DATE - INTERVAL '90 days'
    GROUP BY customer_id
)
SELECT ac.name, ac.email, ro.last_order
FROM active_customers ac
JOIN recent_orders ro ON ac.id = ro.customer_id;

-- 재귀 CTE: 조직도 계층 구조
WITH RECURSIVE org_tree AS (
    -- Anchor: 최상위 노드
    SELECT id, name, manager_id, 1 AS level, name AS path
    FROM employees
    WHERE manager_id IS NULL

    UNION ALL

    -- Recursive: 자식 노드
    SELECT e.id, e.name, e.manager_id, ot.level + 1,
           ot.path || ' > ' || e.name
    FROM employees e
    JOIN org_tree ot ON e.manager_id = ot.id
)
SELECT level, path FROM org_tree ORDER BY path;

-- 재귀 CTE: 날짜 시퀀스 생성 (PostgreSQL)
WITH RECURSIVE date_series AS (
    SELECT CURRENT_DATE AS dt
    UNION ALL
    SELECT dt + 1 FROM date_series WHERE dt < CURRENT_DATE + 30
)
SELECT dt FROM date_series;
```

---

## 윈도우 함수

```sql
-- ROW_NUMBER: 중복 없는 순번
SELECT name, salary,
       ROW_NUMBER() OVER (ORDER BY salary DESC) AS rn
FROM employees;

-- RANK: 동점 건너뜀 (1,1,3)
-- DENSE_RANK: 동점 건너뛰지 않음 (1,1,2)
SELECT name, salary, department_id,
       RANK() OVER (PARTITION BY department_id ORDER BY salary DESC) AS rank,
       DENSE_RANK() OVER (PARTITION BY department_id ORDER BY salary DESC) AS dense_rank
FROM employees;

-- LEAD / LAG: 다음/이전 행 값
SELECT order_date, total,
       LAG(total) OVER (ORDER BY order_date) AS prev_total,
       LEAD(total) OVER (ORDER BY order_date) AS next_total,
       total - LAG(total) OVER (ORDER BY order_date) AS diff
FROM daily_sales;

-- 누적합 / 이동평균
SELECT order_date, amount,
       SUM(amount) OVER (ORDER BY order_date) AS running_total,
       AVG(amount) OVER (ORDER BY order_date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW) AS moving_avg_7d,
       SUM(amount) OVER (PARTITION BY DATE_TRUNC('month', order_date)) AS monthly_total
FROM orders;

-- FIRST_VALUE / LAST_VALUE
SELECT name, salary,
       FIRST_VALUE(salary) OVER (PARTITION BY dept_id ORDER BY salary DESC) AS highest_in_dept
FROM employees;

-- NTILE: 분위수
SELECT name, score,
       NTILE(4) OVER (ORDER BY score DESC) AS quartile
FROM test_results;
```

---

## 집계 함수

```sql
SELECT
    COUNT(*) AS total_rows,
    COUNT(email) AS non_null_emails,
    COUNT(DISTINCT customer_id) AS unique_customers,
    SUM(amount) AS total_amount,
    AVG(amount) AS avg_amount,
    MIN(amount) AS min_amount,
    MAX(amount) AS max_amount
FROM orders
WHERE created_at >= '2024-01-01';

-- GROUP_CONCAT (MySQL)
SELECT department_id,
       GROUP_CONCAT(name ORDER BY name SEPARATOR ', ') AS member_names
FROM employees
GROUP BY department_id;

-- STRING_AGG (PostgreSQL)
SELECT department_id,
       STRING_AGG(name, ', ' ORDER BY name) AS member_names
FROM employees
GROUP BY department_id;

-- HAVING: 집계 후 필터
SELECT customer_id, COUNT(*) AS order_count, SUM(total) AS revenue
FROM orders
GROUP BY customer_id
HAVING COUNT(*) >= 5 AND SUM(total) > 10000
ORDER BY revenue DESC;
```

---

## CASE WHEN

```sql
-- 단순 CASE
SELECT name,
       CASE status
           WHEN 'A' THEN 'Active'
           WHEN 'I' THEN 'Inactive'
           WHEN 'P' THEN 'Pending'
           ELSE 'Unknown'
       END AS status_label
FROM users;

-- 검색 CASE
SELECT name, salary,
       CASE
           WHEN salary < 30000 THEN 'Junior'
           WHEN salary < 60000 THEN 'Mid'
           WHEN salary < 100000 THEN 'Senior'
           ELSE 'Executive'
       END AS grade
FROM employees;

-- CASE in 집계 (조건부 집계 = 피벗 기초)
SELECT
    department_id,
    COUNT(CASE WHEN gender = 'M' THEN 1 END) AS male_count,
    COUNT(CASE WHEN gender = 'F' THEN 1 END) AS female_count,
    SUM(CASE WHEN status = 'active' THEN salary ELSE 0 END) AS active_payroll
FROM employees
GROUP BY department_id;
```

---

## 집합 연산

```sql
-- UNION: 중복 제거
SELECT name, email FROM customers
UNION
SELECT name, email FROM prospects;

-- UNION ALL: 중복 포함 (더 빠름)
SELECT 'Q1' AS quarter, SUM(amount) FROM orders WHERE month BETWEEN 1 AND 3
UNION ALL
SELECT 'Q2', SUM(amount) FROM orders WHERE month BETWEEN 4 AND 6;

-- INTERSECT: 교집합 (양쪽 모두 있는 행)
SELECT customer_id FROM orders WHERE year = 2023
INTERSECT
SELECT customer_id FROM orders WHERE year = 2024;

-- EXCEPT / MINUS: 차집합 (첫 번째에만 있는 행)
-- PostgreSQL/SQL Server: EXCEPT, Oracle: MINUS
SELECT id FROM all_users
EXCEPT
SELECT user_id FROM active_sessions;
```

---

## DML

```sql
-- INSERT 단건
INSERT INTO users (name, email, created_at)
VALUES ('홍길동', 'hong@example.com', NOW());

-- INSERT 다건
INSERT INTO products (name, price, category_id) VALUES
    ('상품A', 10000, 1),
    ('상품B', 20000, 2),
    ('상품C', 30000, 1);

-- INSERT ... SELECT
INSERT INTO order_archive
SELECT * FROM orders WHERE created_at < '2023-01-01';

-- UPDATE
UPDATE users
SET last_login = NOW(), login_count = login_count + 1
WHERE id = 123;

-- UPDATE with JOIN (MySQL)
UPDATE orders o
JOIN customers c ON o.customer_id = c.id
SET o.customer_name = c.name
WHERE o.customer_name IS NULL;

-- UPDATE with subquery
UPDATE products
SET price = price * 1.1
WHERE category_id IN (SELECT id FROM categories WHERE name = 'Premium');

-- DELETE
DELETE FROM sessions WHERE expires_at < NOW();

-- TRUNCATE (로그 없이 전체 삭제, 빠름, ROLLBACK 불가 주의)
TRUNCATE TABLE temp_staging;

-- MERGE / UPSERT (표준 SQL)
MERGE INTO target_table t
USING source_table s ON (t.id = s.id)
WHEN MATCHED THEN
    UPDATE SET t.value = s.value, t.updated_at = NOW()
WHEN NOT MATCHED THEN
    INSERT (id, value, created_at) VALUES (s.id, s.value, NOW());
```

---

## NULL 처리

```sql
-- COALESCE: 첫 번째 non-null 반환
SELECT COALESCE(nickname, username, 'Anonymous') AS display_name FROM users;
SELECT COALESCE(discount, 0) + price AS final_price FROM products;

-- NULLIF: 두 값이 같으면 NULL 반환 (0 나누기 방지)
SELECT total / NULLIF(quantity, 0) AS unit_price FROM order_items;

-- IS NULL / IS NOT NULL
SELECT * FROM employees WHERE manager_id IS NULL;  -- 최상위 직원
SELECT * FROM orders WHERE shipped_at IS NOT NULL;

-- NULL과 비교 주의: NULL = NULL은 항상 FALSE
-- 올바른 방법: IS NULL, IS NOT NULL
-- WHERE value = NULL -- 작동 안 함
-- WHERE value IS NULL -- 올바름
```

---

## 문자열 함수

```sql
-- 연결
SELECT CONCAT(first_name, ' ', last_name) AS full_name FROM employees;  -- MySQL
SELECT first_name || ' ' || last_name AS full_name FROM employees;       -- PostgreSQL/Oracle

-- 부분 문자열
SELECT SUBSTRING(phone, 1, 3) AS area_code FROM users;        -- MySQL/PG
SELECT SUBSTR(phone, 1, 3) AS area_code FROM users;            -- Oracle

-- 공백 제거
SELECT TRIM(name), LTRIM(name), RTRIM(name) FROM users;

-- 대소문자
SELECT UPPER(email), LOWER(username) FROM users;

-- 치환
SELECT REPLACE(phone, '-', '') AS clean_phone FROM users;

-- 길이
SELECT LENGTH(description) AS desc_len FROM articles;  -- PostgreSQL
SELECT LEN(description) AS desc_len FROM articles;     -- SQL Server

-- 위치 찾기
SELECT INSTR(email, '@') AS at_pos FROM users;   -- MySQL/Oracle
SELECT STRPOS(email, '@') AS at_pos FROM users;  -- PostgreSQL

-- 패딩
SELECT LPAD(employee_id::TEXT, 6, '0') AS emp_code FROM employees;
```

---

## 날짜 함수

```sql
-- 현재 날짜/시간
SELECT CURRENT_DATE, CURRENT_TIMESTAMP, NOW();

-- 날짜 더하기
-- MySQL
SELECT DATE_ADD(created_at, INTERVAL 30 DAY) AS expires_at FROM subscriptions;
-- PostgreSQL
SELECT created_at + INTERVAL '30 days' AS expires_at FROM subscriptions;

-- 날짜 차이
-- MySQL
SELECT DATEDIFF(NOW(), created_at) AS days_since FROM orders;
-- PostgreSQL
SELECT (NOW() - created_at)::INT AS days_since FROM orders;

-- 날짜 포맷
-- MySQL
SELECT DATE_FORMAT(created_at, '%Y-%m-%d %H:%i') AS formatted FROM orders;
-- PostgreSQL
SELECT TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI') AS formatted FROM orders;

-- 날짜 추출
SELECT EXTRACT(YEAR FROM created_at) AS year,
       EXTRACT(MONTH FROM created_at) AS month,
       EXTRACT(DOW FROM created_at) AS day_of_week
FROM orders;

-- 날짜 트런케이션 (PostgreSQL)
SELECT DATE_TRUNC('month', created_at) AS month_start FROM orders;
```

---

## 타입 변환

```sql
-- CAST (표준)
SELECT CAST(price AS DECIMAL(10,2)) FROM products;
SELECT CAST('2024-01-15' AS DATE) AS target_date;
SELECT CAST(user_id AS VARCHAR(20)) FROM orders;

-- CONVERT (MySQL)
SELECT CONVERT(price, DECIMAL(10,2)) FROM products;
SELECT CONVERT('2024-01-15', DATE) AS target_date;

-- PostgreSQL :: 연산자
SELECT '2024-01-15'::DATE, '42'::INTEGER, 3.14::TEXT;

-- 암시적 변환 주의: 인덱스 무효화 가능
-- WHERE CAST(user_id AS VARCHAR) = '123'  -- 인덱스 미사용
-- WHERE user_id = 123                     -- 타입 맞추는 게 최선
```

---

## 실전 패턴

### 누적합 (Running Total)
```sql
SELECT
    order_date,
    daily_amount,
    SUM(daily_amount) OVER (ORDER BY order_date) AS cumulative_total
FROM daily_sales;
```

### 이동평균 (7일)
```sql
SELECT
    dt,
    value,
    AVG(value) OVER (ORDER BY dt ROWS BETWEEN 6 PRECEDING AND CURRENT ROW) AS ma7
FROM metrics;
```

### 부서 내 순위 TOP 3
```sql
SELECT * FROM (
    SELECT name, salary, department_id,
           RANK() OVER (PARTITION BY department_id ORDER BY salary DESC) AS rnk
    FROM employees
) ranked
WHERE rnk <= 3;
```

### 피벗 (동적 컬럼 → 정적 예시)
```sql
SELECT
    user_id,
    MAX(CASE WHEN month = 1 THEN amount END) AS jan,
    MAX(CASE WHEN month = 2 THEN amount END) AS feb,
    MAX(CASE WHEN month = 3 THEN amount END) AS mar
FROM monthly_sales
GROUP BY user_id;
```

### 중복 제거 (최신 1건만 유지)
```sql
-- 중복 중 id가 가장 큰 것만 남기기
DELETE FROM logs
WHERE id NOT IN (
    SELECT MAX(id) FROM logs GROUP BY session_id
);

-- PostgreSQL CTE 방식
WITH ranked AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY id DESC) AS rn
    FROM logs
)
DELETE FROM logs WHERE id IN (SELECT id FROM ranked WHERE rn > 1);
```

### 페이징 최적화 (커서 기반)
```sql
-- OFFSET 방식: 대량 데이터 시 느려짐
SELECT * FROM posts ORDER BY id DESC LIMIT 20 OFFSET 10000;

-- 커서 기반: 이전 페이지의 마지막 id 활용
SELECT * FROM posts
WHERE id < :last_seen_id
ORDER BY id DESC
LIMIT 20;
```
