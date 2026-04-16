# Tibero Reference

## 개요

Tibero는 TmaxSoft(티맥스소프트)가 개발한 Oracle 호환 RDBMS. 공공기관·금융·기업 전산 시스템에 Oracle 대체재로 도입. Oracle SQL/PL/SQL 문법을 대부분 지원하므로 Oracle 지식이 그대로 적용됨.

---

## Oracle 호환 기능

다음 Oracle 문법은 Tibero에서 동일하게 동작한다.

```sql
-- ROWNUM
SELECT * FROM (
    SELECT * FROM employees ORDER BY salary DESC
) WHERE ROWNUM <= 10;

-- FETCH FIRST (Tibero 6+)
SELECT * FROM employees ORDER BY salary DESC FETCH FIRST 10 ROWS ONLY;

-- DUAL
SELECT SYSDATE FROM DUAL;
SELECT 1 + 1 FROM DUAL;

-- NVL / NVL2
SELECT NVL(commission, 0) FROM employees;
SELECT NVL2(commission, '있음', '없음') FROM employees;

-- DECODE
SELECT DECODE(status, 'A', 'Active', 'I', 'Inactive', 'Unknown') FROM users;

-- CONNECT BY 계층형 쿼리
SELECT LEVEL, LPAD(' ', (LEVEL-1)*2) || name AS tree,
       SYS_CONNECT_BY_PATH(name, ' > ') AS path
FROM employees
START WITH manager_id IS NULL
CONNECT BY PRIOR id = manager_id
ORDER SIBLINGS BY name;

-- 시퀀스
CREATE SEQUENCE emp_seq START WITH 1 INCREMENT BY 1 CACHE 20;
SELECT emp_seq.NEXTVAL FROM DUAL;
INSERT INTO employees (id, name) VALUES (emp_seq.NEXTVAL, '홍길동');

-- 힌트 (대부분 호환)
SELECT /*+ INDEX(e idx_emp_dept) */ * FROM employees e WHERE dept_id = 10;
SELECT /*+ FULL(e) */ * FROM employees e;
SELECT /*+ USE_NL(e d) */ e.name, d.name FROM employees e JOIN departments d ON e.dept_id = d.id;

-- MERGE INTO
MERGE INTO target t
USING source s ON (t.id = s.id)
WHEN MATCHED THEN UPDATE SET t.value = s.value
WHEN NOT MATCHED THEN INSERT (id, value) VALUES (s.id, s.value);
```

### PL/SQL 호환

```sql
-- 프로시저 / 함수 / 패키지: Oracle 문법 그대로 사용
CREATE OR REPLACE PROCEDURE update_status(p_id NUMBER, p_status VARCHAR2) AS
BEGIN
    UPDATE users SET status = p_status WHERE id = p_id;
    COMMIT;
EXCEPTION
    WHEN OTHERS THEN
        ROLLBACK;
        RAISE;
END;
/

-- BULK COLLECT / FORALL 지원
DECLARE
    TYPE id_tbl IS TABLE OF employees.id%TYPE;
    v_ids id_tbl;
BEGIN
    SELECT id BULK COLLECT INTO v_ids FROM employees WHERE dept_id = 10;
    FORALL i IN 1..v_ids.COUNT
        UPDATE employees SET processed = 'Y' WHERE id = v_ids(i);
    COMMIT;
END;
/

-- EXECUTE IMMEDIATE
DECLARE
    v_sql VARCHAR2(500);
BEGIN
    v_sql := 'UPDATE employees SET salary = :1 WHERE id = :2';
    EXECUTE IMMEDIATE v_sql USING 60000, 100;
    COMMIT;
END;
/

-- DBMS_OUTPUT.PUT_LINE
BEGIN
    DBMS_OUTPUT.PUT_LINE('Tibero PL/SQL 실행');
END;
/
```

---

## Tibero 고유 관리 도구 및 환경

### 환경 변수

```bash
# 필수 환경 변수
export TB_HOME=/opt/tmaxdb/tibero6   # Tibero 설치 경로
export TB_SID=tibero                  # 인스턴스 이름
export PATH=$TB_HOME/bin:$PATH
export LD_LIBRARY_PATH=$TB_HOME/lib:$LD_LIBRARY_PATH
```

### 시작 / 종료

```bash
# Tibero 인스턴스 시작
tbboot

# 정상 종료 (접속 대기)
tbdown

# 즉시 종료
tbdown immediate

# 리스너 시작/종료
tblistener start
tblistener stop

# 상태 확인
tbboot status
```

### 설정 파일

```
$TB_HOME/config/
  tibero.tip          -- 인스턴스 파라미터 (Oracle의 init.ora / spfile 대응)
  tbdsn.tbr           -- 접속 정보 (Oracle의 tnsnames.ora 대응)
```

```
# tbdsn.tbr 예시
TIBERO =
(
    (INSTANCE
        (HOST=localhost)
        (PORT=8629)
        (DB_NAME=tibero)
    )
)
```

```
# tibero.tip 주요 파라미터 예시
DB_NAME=tibero
LISTENER_PORT=8629
MAX_SESSION_COUNT=100
MEMORY_TARGET=2G
LOG_DIR=$TB_HOME/log
```

---

## tbSQL (Tibero SQL 도구)

Oracle의 SQL*Plus에 대응하는 CLI 도구.

```bash
# 접속
tbsql sys/tibero@tibero          # SYS로 접속
tbsql user1/password@TIBERO      # 일반 사용자 (tbdsn.tbr의 alias 사용)

# 로컬 접속 (리스너 없이)
tbsql sys/tibero
```

```sql
-- tbSQL 내부 명령
DESC employees;                    -- 테이블 구조 확인
SET SERVEROUTPUT ON;               -- DBMS_OUTPUT 출력 활성화
SET LINESIZE 200;                  -- 출력 라인 너비
SET PAGESIZE 50;                   -- 페이지 행 수

-- 쿼리 실행
SELECT * FROM employees WHERE dept_id = 10;

-- 프로시저 실행
EXEC update_status(1, 'A');

-- PL/SQL 블록 실행 (/ 로 종료)
BEGIN
    DBMS_OUTPUT.PUT_LINE('Hello Tibero');
END;
/

-- 스크립트 파일 실행
@/opt/scripts/deploy.sql
START /opt/scripts/deploy.sql

-- 결과를 파일로 저장
SPOOL /tmp/output.txt
SELECT * FROM employees;
SPOOL OFF;

-- 접속 종료
EXIT;
QUIT;
```

---

## JDBC 연결

```xml
<!-- pom.xml 의존성 (내부 nexus 또는 로컬 설치) -->
<dependency>
    <groupId>com.tmax.tibero</groupId>
    <artifactId>tibero-jdbc</artifactId>
    <version>6.0</version>
    <scope>system</scope>
    <systemPath>${TB_HOME}/client/lib/jar/tibero6-jdbc.jar</systemPath>
</dependency>
```

```java
// JDBC 드라이버 클래스
Class.forName("com.tmax.tibero.jdbc.TbDriver");

// 연결 문자열
// jdbc:tibero:thin:@호스트:포트:DB명
String url = "jdbc:tibero:thin:@localhost:8629:tibero";
Connection conn = DriverManager.getConnection(url, "user1", "password");
```

```yaml
# Spring Boot application.yml
spring:
  datasource:
    driver-class-name: com.tmax.tibero.jdbc.TbDriver
    url: jdbc:tibero:thin:@localhost:8629:tibero
    username: user1
    password: password

# JPA 방언 (Hibernate)
  jpa:
    database-platform: org.hibernate.dialect.Oracle12cDialect
    # Tibero 전용 dialect가 없으면 Oracle dialect 사용
```

---

## Oracle과의 차이점 및 주의사항

### 지원 제한 기능

```
- DBMS_SCHEDULER: 제한적 지원 (기본 기능은 가능, 고급 기능 일부 미지원)
- Oracle Advanced Queuing (AQ): 미지원
- Oracle Spatial (SDO_*): 미지원 또는 제한
- Oracle Text (CONTAINS, CTXSYS): 미지원
- Oracle Label Security: 미지원
- Flashback Query / Flashback Table: 버전에 따라 제한
- CONNECT BY ... CYCLE 감지: 버전에 따라 동작 차이
- 일부 내장 패키지(DBMS_*)가 Oracle과 동작 차이 있을 수 있음
```

### 힌트 동작 차이

```sql
-- 힌트는 대부분 호환되나 내부 옵티마이저 동작이 다를 수 있음
-- Oracle에서 잘 작동하던 힌트가 Tibero에서 다르게 작동하는 경우:
-- 1. EXPLAIN PLAN으로 실행 계획 확인
EXPLAIN PLAN FOR SELECT * FROM employees WHERE dept_id = 10;
SELECT * FROM TABLE(DBMS_XPLAN.DISPLAY);

-- 2. 힌트 없이 먼저 테스트 후 필요 시 추가
-- 3. Tibero 전용 힌트 문서 참고 ($TB_HOME/doc)
```

### 문자셋 주의

```
- 설치 시 DB 문자셋 결정 (변경 불가)
- 일반적으로 UTF-8 또는 KO16KSC5601(EUC-KR) 사용
- JDBC URL에 characterEncoding 명시 권장:
  jdbc:tibero:thin:@host:8629:tibero?characterEncoding=UTF-8
```

---

## Oracle → Tibero 마이그레이션 체크리스트

```
기본 SQL
  [ ] SELECT / INSERT / UPDATE / DELETE: 호환
  [ ] ROWNUM, DUAL, NVL, DECODE: 호환
  [ ] CONNECT BY 계층형: 호환
  [ ] MERGE INTO: 호환
  [ ] 시퀀스: 호환

PL/SQL
  [ ] 프로시저 / 함수 / 패키지: 대부분 호환
  [ ] BULK COLLECT / FORALL: 호환
  [ ] EXECUTE IMMEDIATE: 호환
  [ ] EXCEPTION 처리: 호환
  [ ] 커서: 호환

패키지 확인 필요
  [ ] DBMS_OUTPUT: 호환
  [ ] DBMS_UTILITY: 일부 호환
  [ ] DBMS_SCHEDULER: 제한적
  [ ] UTL_FILE: 호환
  [ ] DBMS_LOB: 호환

연결 / 드라이버
  [ ] JDBC 드라이버: tibero6-jdbc.jar로 교체
  [ ] 연결 문자열: jdbc:tibero:thin:@ 형식으로 변경
  [ ] Hibernate dialect: Oracle dialect 유지 또는 Tibero dialect 확인

운영 도구
  [ ] SQL*Plus → tbSQL
  [ ] Oracle Enterprise Manager → tbAdmin (웹 관리 도구)
  [ ] tnsnames.ora → tbdsn.tbr
  [ ] export/import (expdp/impdp) → Tibero 제공 마이그레이션 툴 사용
```

---

## 실전 패턴

### tbSQL로 자주 쓰는 조회

```sql
-- 현재 세션 정보
SELECT SYS_CONTEXT('USERENV', 'SESSION_USER') AS usr,
       SYS_CONTEXT('USERENV', 'DB_NAME') AS db
FROM DUAL;

-- 테이블 목록
SELECT table_name FROM user_tables ORDER BY table_name;

-- 인덱스 목록
SELECT index_name, column_name, column_position
FROM user_ind_columns
WHERE table_name = 'EMPLOYEES'
ORDER BY index_name, column_position;

-- 현재 세션 락 확인
SELECT sid, type, lmode, request, block FROM v$lock WHERE block = 1;

-- 슬로우 쿼리 (v$sql)
SELECT sql_text, elapsed_time/1000 AS elapsed_ms, executions,
       elapsed_time/NULLIF(executions,0)/1000 AS avg_ms
FROM v$sql
ORDER BY elapsed_time DESC
FETCH FIRST 20 ROWS ONLY;

-- 프로시저 실행 및 결과 출력
SET SERVEROUTPUT ON SIZE 1000000;
BEGIN
    update_salary(100, 10, :v_sal);
    DBMS_OUTPUT.PUT_LINE('결과: ' || :v_sal);
END;
/
```
