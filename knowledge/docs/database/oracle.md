# Oracle Database Reference

## 데이터 타입

```sql
-- 숫자
NUMBER(p, s)    -- 범용 숫자 (p: 전체 자리, s: 소수점 이하)
NUMBER(10)      -- 정수 10자리
NUMBER(15, 2)   -- 금액 (소수 2자리)
NUMBER          -- 가변 정밀도

-- 문자열
VARCHAR2(n)     -- 가변 길이 (최대 4000바이트, Extended: 32767)
CHAR(n)         -- 고정 길이
NVARCHAR2(n)    -- 유니코드 가변
CLOB            -- 최대 4GB 텍스트
NCLOB           -- 유니코드 CLOB

-- 날짜/시간
DATE            -- 날짜+시간 (초까지), Oracle의 DATE는 시간 포함 주의
TIMESTAMP(n)    -- 날짜+시간 (소수점 n자리 초, 기본 6)
TIMESTAMP WITH TIME ZONE
TIMESTAMP WITH LOCAL TIME ZONE
INTERVAL YEAR TO MONTH
INTERVAL DAY TO SECOND

-- 이진
RAW(n)          -- 최대 2000바이트 이진
BLOB            -- 최대 4GB 이진
BFILE           -- 외부 파일 참조

-- 기타
ROWID           -- 행의 물리적 주소
XMLTYPE         -- XML 데이터
```

---

## Oracle 고유 문법

```sql
-- ROWNUM: 결과 행 번호 (WHERE 처리 후 부여)
-- 주의: ORDER BY 전에 ROWNUM 부여 → 정렬 후 상위 N개는 서브쿼리 필요
SELECT * FROM (
    SELECT * FROM employees ORDER BY salary DESC
) WHERE ROWNUM <= 10;

-- FETCH FIRST (Oracle 12c+, 표준 SQL)
SELECT * FROM employees
ORDER BY salary DESC
FETCH FIRST 10 ROWS ONLY;

-- OFFSET + FETCH
SELECT * FROM employees
ORDER BY salary DESC
OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY;

-- DUAL: 가상 테이블 (함수 테스트, 상수 조회)
SELECT SYSDATE FROM DUAL;
SELECT 1 + 1 FROM DUAL;
SELECT SYS_GUID() FROM DUAL;  -- UUID 생성
SELECT seq_name.NEXTVAL FROM DUAL;

-- NVL: NULL이면 대체값 반환 (COALESCE와 유사)
SELECT NVL(commission, 0) FROM employees;

-- NVL2: NULL 여부에 따라 다른 값
SELECT NVL2(commission, '커미션 있음', '커미션 없음') FROM employees;

-- DECODE: if-else 단축 (CASE WHEN의 Oracle 전통 방식)
SELECT DECODE(status,
    'A', 'Active',
    'I', 'Inactive',
    'P', 'Pending',
    'Unknown'  -- ELSE
) AS status_label FROM users;

-- NULLIF
SELECT NULLIF(a, b) FROM t;  -- a = b면 NULL, 아니면 a
```

---

## 시퀀스

```sql
-- 시퀀스 생성
CREATE SEQUENCE emp_seq
    START WITH 1000
    INCREMENT BY 1
    MAXVALUE 9999999999
    NOCYCLE         -- 최대값 도달 시 오류 (CYCLE: 처음부터 다시)
    NOCACHE;        -- CACHE n: 미리 n개 캐시 (성능 향상, RAC 환경 주의)

-- 시퀀스 생성 (캐시 사용)
CREATE SEQUENCE order_seq START WITH 1 INCREMENT BY 1 CACHE 20;

-- 사용
SELECT emp_seq.NEXTVAL FROM DUAL;   -- 다음 값 (부여)
SELECT emp_seq.CURRVAL FROM DUAL;   -- 현재 값 (같은 세션)

INSERT INTO employees (id, name) VALUES (emp_seq.NEXTVAL, '홍길동');

-- 시퀀스 수정
ALTER SEQUENCE emp_seq INCREMENT BY 10;
ALTER SEQUENCE emp_seq CACHE 50;

-- 시퀀스 삭제
DROP SEQUENCE emp_seq;

-- 현재 값 확인
SELECT sequence_name, last_number FROM user_sequences;
```

---

## CONNECT BY: 계층형 쿼리

```sql
-- 조직도 예시
CREATE TABLE employees (
    id         NUMBER PRIMARY KEY,
    name       VARCHAR2(100),
    manager_id NUMBER REFERENCES employees(id)
);

-- 계층 조회
SELECT
    LEVEL,
    LPAD(' ', (LEVEL-1)*2) || name AS org_chart,
    SYS_CONNECT_BY_PATH(name, ' > ') AS full_path,
    CONNECT_BY_ISLEAF AS is_leaf
FROM employees
START WITH manager_id IS NULL        -- 루트 노드
CONNECT BY PRIOR id = manager_id    -- 부모 → 자식 방향
ORDER SIBLINGS BY name;             -- 같은 레벨 내 정렬

-- 특정 노드에서 시작
SELECT LEVEL, name
FROM employees
START WITH id = 5
CONNECT BY PRIOR id = manager_id;

-- 자식 → 부모 방향 (역방향)
SELECT LEVEL, name
FROM employees
START WITH id = 100   -- 말단 직원
CONNECT BY id = PRIOR manager_id;

-- 순환 감지 방지
CONNECT BY NOCYCLE PRIOR id = manager_id

-- BOM(Bill of Materials) 패턴 - 부품 구성
SELECT
    LEVEL AS depth,
    component_id,
    SYS_CONNECT_BY_PATH(component_id, '/') AS path
FROM bom
START WITH parent_id IS NULL
CONNECT BY PRIOR component_id = parent_id;
```

---

## PL/SQL 기본

### DECLARE / BEGIN / EXCEPTION / END

```sql
DECLARE
    v_name      employees.name%TYPE;       -- 컬럼 타입 참조
    v_salary    NUMBER(10, 2) := 0;        -- 초기화
    v_emp_rec   employees%ROWTYPE;         -- 행 전체 타입
    v_count     PLS_INTEGER;
BEGIN
    -- 단건 조회
    SELECT name, salary INTO v_name, v_salary
    FROM employees WHERE id = 100;

    DBMS_OUTPUT.PUT_LINE('이름: ' || v_name || ', 급여: ' || v_salary);

    -- 행 전체 조회
    SELECT * INTO v_emp_rec FROM employees WHERE id = 100;
    DBMS_OUTPUT.PUT_LINE(v_emp_rec.name);

EXCEPTION
    WHEN NO_DATA_FOUND THEN
        DBMS_OUTPUT.PUT_LINE('데이터 없음');
    WHEN TOO_MANY_ROWS THEN
        DBMS_OUTPUT.PUT_LINE('여러 행 반환됨');
    WHEN OTHERS THEN
        DBMS_OUTPUT.PUT_LINE('오류: ' || SQLERRM);
        RAISE;  -- 상위로 재전파
END;
/
```

### 커서

```sql
DECLARE
    CURSOR emp_cur IS
        SELECT id, name, salary FROM employees WHERE department_id = 10;
    v_rec emp_cur%ROWTYPE;
BEGIN
    OPEN emp_cur;
    LOOP
        FETCH emp_cur INTO v_rec;
        EXIT WHEN emp_cur%NOTFOUND;
        DBMS_OUTPUT.PUT_LINE(v_rec.name || ': ' || v_rec.salary);
    END LOOP;
    CLOSE emp_cur;
END;
/

-- FOR 루프 커서 (자동 OPEN/FETCH/CLOSE)
BEGIN
    FOR emp IN (SELECT id, name FROM employees WHERE dept_id = 10) LOOP
        DBMS_OUTPUT.PUT_LINE(emp.name);
    END LOOP;
END;
/

-- 커서 파라미터
DECLARE
    CURSOR dept_emp(p_dept_id NUMBER) IS
        SELECT name FROM employees WHERE department_id = p_dept_id;
BEGIN
    FOR emp IN dept_emp(20) LOOP
        DBMS_OUTPUT.PUT_LINE(emp.name);
    END LOOP;
END;
/
```

### 예외 처리 / RAISE_APPLICATION_ERROR

```sql
DECLARE
    v_balance NUMBER;
    insufficient_funds EXCEPTION;
    PRAGMA EXCEPTION_INIT(insufficient_funds, -20001);
BEGIN
    SELECT balance INTO v_balance FROM accounts WHERE id = 1;

    IF v_balance < 10000 THEN
        RAISE_APPLICATION_ERROR(-20001, '잔액이 부족합니다: ' || v_balance);
    END IF;

    UPDATE accounts SET balance = balance - 10000 WHERE id = 1;
    COMMIT;
EXCEPTION
    WHEN insufficient_funds THEN
        ROLLBACK;
        DBMS_OUTPUT.PUT_LINE('처리 실패: 잔액 부족');
    WHEN OTHERS THEN
        ROLLBACK;
        RAISE;
END;
/
-- RAISE_APPLICATION_ERROR: -20000 ~ -20999 범위 사용자 정의 오류
```

### 프로시저 / 함수

```sql
-- 프로시저
CREATE OR REPLACE PROCEDURE update_salary(
    p_emp_id    IN  NUMBER,
    p_pct       IN  NUMBER,
    p_new_sal   OUT NUMBER
) AS
BEGIN
    UPDATE employees
    SET salary = salary * (1 + p_pct / 100)
    WHERE id = p_emp_id
    RETURNING salary INTO p_new_sal;

    IF SQL%ROWCOUNT = 0 THEN
        RAISE_APPLICATION_ERROR(-20002, '직원을 찾을 수 없습니다: ' || p_emp_id);
    END IF;

    COMMIT;
EXCEPTION
    WHEN OTHERS THEN
        ROLLBACK;
        RAISE;
END update_salary;
/

-- 실행
DECLARE
    v_new_salary NUMBER;
BEGIN
    update_salary(100, 10, v_new_salary);
    DBMS_OUTPUT.PUT_LINE('새 급여: ' || v_new_salary);
END;
/

-- 함수 (RETURN 값 있음)
CREATE OR REPLACE FUNCTION get_full_name(p_id NUMBER)
RETURN VARCHAR2 AS
    v_name VARCHAR2(200);
BEGIN
    SELECT first_name || ' ' || last_name INTO v_name
    FROM employees WHERE id = p_id;
    RETURN v_name;
EXCEPTION
    WHEN NO_DATA_FOUND THEN RETURN NULL;
END;
/

-- SQL에서 사용
SELECT get_full_name(100) FROM DUAL;
```

### 패키지

```sql
-- 패키지 스펙 (인터페이스)
CREATE OR REPLACE PACKAGE emp_pkg AS
    g_min_salary CONSTANT NUMBER := 30000;

    PROCEDURE hire(p_name VARCHAR2, p_dept_id NUMBER, p_salary NUMBER);
    PROCEDURE fire(p_id NUMBER);
    FUNCTION get_headcount(p_dept_id NUMBER) RETURN NUMBER;
END emp_pkg;
/

-- 패키지 바디 (구현)
CREATE OR REPLACE PACKAGE BODY emp_pkg AS
    PROCEDURE hire(p_name VARCHAR2, p_dept_id NUMBER, p_salary NUMBER) AS
    BEGIN
        INSERT INTO employees (id, name, department_id, salary)
        VALUES (emp_seq.NEXTVAL, p_name, p_dept_id, p_salary);
        COMMIT;
    END hire;

    PROCEDURE fire(p_id NUMBER) AS
    BEGIN
        DELETE FROM employees WHERE id = p_id;
        COMMIT;
    END fire;

    FUNCTION get_headcount(p_dept_id NUMBER) RETURN NUMBER AS
        v_cnt NUMBER;
    BEGIN
        SELECT COUNT(*) INTO v_cnt FROM employees WHERE department_id = p_dept_id;
        RETURN v_cnt;
    END get_headcount;
END emp_pkg;
/

-- 사용
EXEC emp_pkg.hire('김철수', 10, 50000);
SELECT emp_pkg.get_headcount(10) FROM DUAL;
```

---

## 힌트 (Optimizer Hint)

```sql
-- 힌트는 쿼리 옵티마이저 지시 (잘못 쓰면 오히려 느려짐)

-- FULL: 인덱스 무시하고 풀 스캔 강제
SELECT /*+ FULL(e) */ * FROM employees e WHERE status = 'A';

-- INDEX: 특정 인덱스 사용
SELECT /*+ INDEX(e idx_emp_dept) */ * FROM employees e WHERE dept_id = 10;

-- LEADING: 조인 순서 지정 (첫 번째 드라이빙 테이블)
SELECT /*+ LEADING(d e) */ d.name, e.name
FROM departments d JOIN employees e ON d.id = e.dept_id;

-- USE_NL: Nested Loop Join
SELECT /*+ USE_NL(e d) */ e.name, d.name
FROM employees e JOIN departments d ON e.dept_id = d.id;

-- USE_HASH: Hash Join
SELECT /*+ USE_HASH(e d) */ e.name, d.name
FROM employees e JOIN departments d ON e.dept_id = d.id;

-- PARALLEL: 병렬 실행
SELECT /*+ PARALLEL(o, 4) */ COUNT(*) FROM orders o;

-- NO_MERGE: 인라인 뷰 병합 방지
SELECT /*+ NO_MERGE(v) */ * FROM (
    SELECT dept_id, AVG(salary) avg_sal FROM employees GROUP BY dept_id
) v WHERE avg_sal > 50000;
```

---

## MERGE INTO (UPSERT)

```sql
MERGE INTO target t
USING (
    SELECT 1 AS id, '홍길동' AS name, 50000 AS salary FROM DUAL
) s
ON (t.id = s.id)
WHEN MATCHED THEN
    UPDATE SET t.name = s.name, t.salary = s.salary, t.updated_at = SYSDATE
    -- DELETE WHERE t.status = 'inactive'  -- 조건부 삭제도 가능
WHEN NOT MATCHED THEN
    INSERT (id, name, salary, created_at)
    VALUES (s.id, s.name, s.salary, SYSDATE);

-- 실무: 스테이징 테이블 → 운영 테이블 동기화
MERGE INTO products t
USING staging_products s ON (t.product_code = s.product_code)
WHEN MATCHED THEN
    UPDATE SET t.price = s.price, t.stock = s.stock
WHEN NOT MATCHED THEN
    INSERT (product_code, name, price, stock)
    VALUES (s.product_code, s.name, s.price, s.stock);
```

---

## EXECUTE IMMEDIATE (동적 SQL)

```sql
DECLARE
    v_sql    VARCHAR2(1000);
    v_count  NUMBER;
    v_table  VARCHAR2(30) := 'EMPLOYEES';
BEGIN
    -- DDL 실행
    EXECUTE IMMEDIATE 'CREATE TABLE temp_log (id NUMBER, msg VARCHAR2(200))';

    -- 동적 SELECT INTO
    v_sql := 'SELECT COUNT(*) FROM ' || v_table;
    EXECUTE IMMEDIATE v_sql INTO v_count;
    DBMS_OUTPUT.PUT_LINE('행 수: ' || v_count);

    -- 바인드 변수 사용 (SQL Injection 방지, 성능 향상)
    v_sql := 'UPDATE employees SET salary = :1 WHERE id = :2';
    EXECUTE IMMEDIATE v_sql USING 60000, 100;

    COMMIT;
END;
/
```

---

## 대량 데이터 처리: BULK COLLECT / FORALL

```sql
DECLARE
    TYPE emp_id_tbl IS TABLE OF employees.id%TYPE;
    TYPE emp_sal_tbl IS TABLE OF employees.salary%TYPE;

    v_ids  emp_id_tbl;
    v_sals emp_sal_tbl;
BEGIN
    -- BULK COLLECT: 여러 행을 한 번에 컬렉션으로 가져옴
    SELECT id, salary
    BULK COLLECT INTO v_ids, v_sals
    FROM employees
    WHERE department_id = 10;

    -- FORALL: 컬렉션 기반 DML을 한 번에 실행 (row-by-row보다 10~100배 빠름)
    FORALL i IN 1..v_ids.COUNT
        UPDATE employees
        SET salary = v_sals(i) * 1.1
        WHERE id = v_ids(i);

    COMMIT;
    DBMS_OUTPUT.PUT_LINE('처리 건수: ' || v_ids.COUNT);
END;
/

-- LIMIT 옵션: 대용량 테이블을 나눠서 처리
DECLARE
    CURSOR emp_cur IS SELECT id FROM employees;
    TYPE id_tbl IS TABLE OF employees.id%TYPE;
    v_ids id_tbl;
BEGIN
    OPEN emp_cur;
    LOOP
        FETCH emp_cur BULK COLLECT INTO v_ids LIMIT 1000;
        EXIT WHEN v_ids.COUNT = 0;

        FORALL i IN 1..v_ids.COUNT
            UPDATE employees SET processed = 'Y' WHERE id = v_ids(i);
        COMMIT;
    END LOOP;
    CLOSE emp_cur;
END;
/
```
